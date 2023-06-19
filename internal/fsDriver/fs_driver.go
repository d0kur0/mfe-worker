package fsDriver

import (
	"errors"
	"fmt"
	"github.com/samber/lo"
	"github.com/xanzy/go-gitlab"
	"io"
	"log"
	"mfe-worker/internal/configMap"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const StorageSubDir = "images"

type FSDriver struct {
	configMap  *configMap.ConfigMap
	ImagesPath string
}

func NewFSDriver(configMap *configMap.ConfigMap) (*FSDriver, error) {
	if _, err := os.Stat(configMap.StoragePath); os.IsNotExist(err) {
		return nil, fmt.Errorf(`storage directory was not found (%s), create it first with correct access rights`, configMap.StoragePath)
	}

	imagesPath := path.Join(configMap.StoragePath, StorageSubDir)
	_, imagesPathHasExists := os.Stat(imagesPath)

	if imagesPathHasExists != nil {
		if err := os.Mkdir(imagesPath, 755); err != nil {
			return nil, errors.Join(fmt.Errorf(`failed on create dir: %s`, imagesPath), err)
		}
	}

	return &FSDriver{configMap: configMap, ImagesPath: imagesPath}, nil
}

func (d *FSDriver) IsDirExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (d *FSDriver) CreateDir(path string) error {
	if d.IsDirExists(path) {
		return errors.New("trying create dir what is already exists")
	}

	return os.Mkdir(path, 0755)
}

func (d *FSDriver) GetProjectPath(projectId string) string {
	return filepath.Join(d.ImagesPath, projectId)
}

func (d *FSDriver) HasProjectDir(projectId string) bool {
	return d.IsDirExists(d.GetProjectPath(projectId))
}

func (d *FSDriver) CreateProjectDir(projectId string) error {
	return d.CreateDir(d.GetProjectPath(projectId))
}

func (d *FSDriver) GetProjectBranchPath(projectId string, branch string) string {
	return filepath.Join(d.ImagesPath, projectId, branch)
}

func (d *FSDriver) HasProjectBranchDir(projectId string, branch string) bool {
	return d.IsDirExists(d.GetProjectBranchPath(projectId, branch))
}

func (d *FSDriver) CreateProjectBranchDir(projectId string, branch string) error {
	return d.CreateDir(d.GetProjectBranchPath(projectId, branch))
}

func (d *FSDriver) GetBranchRevisionPath(projectId string, branch string, revision string) string {
	return filepath.Join(d.ImagesPath, projectId, branch, revision)
}

func (d *FSDriver) HasBranchRevisionDir(projectId string, branch string, revision string) bool {
	return d.IsDirExists(d.GetBranchRevisionPath(projectId, branch, revision))
}

func (d *FSDriver) CreateBranchRevisionDir(projectId string, branch string, revision string) error {
	return d.CreateDir(d.GetBranchRevisionPath(projectId, branch, revision))
}

func (d *FSDriver) CopyFile(source string, dest string) (err error) {
	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}

	defer func(sourceFile *os.File) {
		err := sourceFile.Close()
		if err != nil {
			log.Println(err)
		}
	}(sourceFile)

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}

	defer func(destFile *os.File) {
		err := destFile.Close()
		if err != nil {
			log.Println(err)
		}
	}(destFile)

	_, err = io.Copy(destFile, sourceFile)
	if err == nil {
		sourceInfo, err := os.Stat(source)
		if err != nil {
			err = os.Chmod(dest, sourceInfo.Mode())
		}
	}

	return
}

func (d *FSDriver) CopyDir(source string, dest string) (err error) {
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dest, sourceInfo.Mode())
	if err != nil {
		return err
	}

	directory, _ := os.Open(source)
	objects, err := directory.Readdir(-1)

	for _, obj := range objects {
		sourceFilePointer := path.Join(source, obj.Name())
		destFilePointer := path.Join(dest, obj.Name())

		if obj.IsDir() {
			err = d.CopyDir(sourceFilePointer, destFilePointer)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			err = d.CopyFile(sourceFilePointer, destFilePointer)
			if err != nil {
				fmt.Println(err)
			}
		}

	}
	return
}

type PickedFile struct {
	Path    string
	WebPath string
}

func (d *FSDriver) PickFilesToWebStorage(project *configMap.Project, glBranch *gitlab.Branch, tmpPath string) (pickedFiles []PickedFile, err error) {
	branchPath, err := filepath.Abs(d.GetProjectBranchPath(project.ProjectID, glBranch.Name))
	if err != nil {
		return pickedFiles, err
	}

	revisionPath, err := filepath.Abs(d.GetBranchRevisionPath(project.ProjectID, glBranch.Name, glBranch.Commit.ShortID))
	if err != nil {
		return pickedFiles, err
	}

	filesWithGlob := lo.Flatten(lo.Map(project.DistFiles, func(filePath string, index int) []string {
		files, _ := filepath.Glob(path.Join(tmpPath, filePath))
		return files
	}))

	for _, filePath := range filesWithGlob {
		filePathWithoutTmpDir := strings.Trim(strings.ReplaceAll(filePath, tmpPath, ""), "/")
		destPath := path.Join(revisionPath, filePathWithoutTmpDir)

		objInfo, err := os.Stat(filePath)
		if err != nil {
			return pickedFiles, err
		}

		if objInfo.IsDir() {
			if err = d.CopyDir(filePath, destPath); err != nil {
				return pickedFiles, err
			}
		}

		if !objInfo.IsDir() {
			destDir, _ := filepath.Split(filePathWithoutTmpDir)
			destDirSegments := filepath.SplitList(destDir)

			for index, seg := range destDirSegments {
				prevSegPath := lo.Slice(destDirSegments, 0, index)
				currSegPath := []string{revisionPath}
				currSegPath = append(currSegPath, prevSegPath...)
				currSegPath = append(currSegPath, seg)
				segPath := path.Join(currSegPath...)
				if !d.IsDirExists(segPath) {
					if err := d.CreateDir(segPath); err != nil {
						return pickedFiles, err
					}
				}
			}

			if err = d.CopyFile(filePath, destPath); err != nil {
				return pickedFiles, err
			}
		}

		pickedFiles = append(
			pickedFiles,
			PickedFile{
				Path: filePathWithoutTmpDir,
				WebPath: fmt.Sprintf(
					"%s/static/%s/%s/%s/%s",
					d.configMap.HttpBaseUrl, project.ProjectID,
					glBranch.Name, glBranch.Commit.ShortID, filePathWithoutTmpDir,
				),
			},
		)
	}

	return pickedFiles, os.Symlink(revisionPath, path.Join(branchPath, "@latest"))
}

func (d *FSDriver) GetTmpPathForBuild(projectId string, branch string, revision string) string {
	return path.Join(
		d.configMap.StoragePath,
		fmt.Sprintf("%s-%s-%s", projectId, branch, revision),
	)
}

func (d *FSDriver) HasTmpDirForBuild(projectId string, branch string, revision string) bool {
	_, err := os.Stat(d.GetTmpPathForBuild(projectId, branch, revision))
	return err == nil
}

func (d *FSDriver) RemoveTmpDirForBuild(projectId string, branch string, revision string) error {
	return os.RemoveAll(d.GetTmpPathForBuild(projectId, branch, revision))
}
