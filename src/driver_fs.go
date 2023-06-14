package src

import (
	"errors"
	"fmt"
	"github.com/xanzy/go-gitlab"
	"os"
	"path"
	"path/filepath"
)

type FSDriver struct {
	configMap  *ConfigMap
	imagesPath string
}

func NewFSDriver(configMap *ConfigMap) (*FSDriver, error) {
	if _, err := os.Stat(configMap.StoragePath); os.IsNotExist(err) {
		return nil, fmt.Errorf(`storage directory was not found (%s), create it first with correct access rights`, configMap.StoragePath)
	}

	imagesPath := path.Join(configMap.StoragePath, "/images")
	_, imagesPathHasExists := os.Stat(imagesPath)

	if imagesPathHasExists != nil {
		if err := os.Mkdir(imagesPath, 755); err != nil {
			return nil, errors.Join(fmt.Errorf(`failed on create dir: %s`, imagesPath), err)
		}
	}

	return &FSDriver{configMap: configMap, imagesPath: imagesPath}, nil
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
	return filepath.Join(d.imagesPath, projectId)
}

func (d *FSDriver) HasProjectDir(projectId string) bool {
	return d.IsDirExists(d.GetProjectPath(projectId))
}

func (d *FSDriver) CreateProjectDir(projectId string) error {
	return d.CreateDir(d.GetProjectPath(projectId))
}

func (d *FSDriver) GetProjectBranchPath(projectId string, branch string) string {
	return filepath.Join(d.imagesPath, projectId, branch)
}

func (d *FSDriver) HasProjectBranchDir(projectId string, branch string) bool {
	return d.IsDirExists(d.GetProjectBranchPath(projectId, branch))
}

func (d *FSDriver) CreateProjectBranchDir(projectId string, branch string) error {
	return d.CreateDir(d.GetProjectBranchPath(projectId, branch))
}

func (d *FSDriver) GetBranchRevisionPath(projectId string, branch string, revision string) string {
	return filepath.Join(d.imagesPath, projectId, branch, revision)
}

func (d *FSDriver) HasBranchRevisionDir(projectId string, branch string, revision string) bool {
	return d.IsDirExists(d.GetBranchRevisionPath(projectId, branch, revision))
}

func (d *FSDriver) CreateBranchRevisionDir(projectId string, branch string, revision string) error {
	return d.CreateDir(d.GetBranchRevisionPath(projectId, branch, revision))
}

func (d *FSDriver) PickFilesToWebStorage(project *Project, glBranch *gitlab.Branch, tmpPath string) error {
	branchRevisionPath := d.GetBranchRevisionPath(project.ProjectID, glBranch.Name, glBranch.Commit.ShortID)

	for _, filePath := range project.DistFiles {
		input, err := os.ReadFile(path.Join(tmpPath, filePath))
		if err != nil {
			return err
		}

		destDir, _ := filepath.Split(filePath)
		destDirSegments := filepath.SplitList(destDir)

		for _, seg := range destDirSegments {
			segPath := path.Join(branchRevisionPath, seg)
			if !d.IsDirExists(segPath) {
				if err := d.CreateDir(segPath); err != nil {
					return err
				}
			}
		}

		err = os.WriteFile(path.Join(branchRevisionPath, filePath), input, 0644)
		if err != nil {
			return err
		}
	}

	return nil
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
