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

func (d *FSDriver) getProjectPath(projectId string) string {
	return filepath.Join(d.imagesPath, projectId)
}

func (d *FSDriver) hasProject(projectId string) bool {
	_, err := os.Stat(d.getProjectPath(projectId))
	return err == nil
}

func (d *FSDriver) createProject(projectId string) error {
	if alreadyExist := d.hasProject(projectId); alreadyExist {
		return errors.New("trying create project what is already exists")
	}

	projectPath := d.getProjectPath(projectId)

	if err := os.Mkdir(projectPath, 0755); err != nil {
		return errors.Join(fmt.Errorf("failed on create project dir: `%s`", projectPath), err)
	}

	return nil
}

func (d *FSDriver) getProjectBranchPath(projectId string, branch string) string {
	return filepath.Join(d.imagesPath, projectId, branch)
}

func (d *FSDriver) hasProjectBranch(projectId string, branch string) bool {
	_, err := os.Stat(d.getProjectBranchPath(projectId, branch))
	return err == nil
}

func (d *FSDriver) createProjectBranch(projectId string, branch string) error {
	if alreadyExist := d.hasProjectBranch(projectId, branch); alreadyExist {
		return errors.New("trying create project branch what is already exists")
	}

	projectBranchPath := d.getProjectBranchPath(projectId, branch)

	if err := os.Mkdir(projectBranchPath, 0755); err != nil {
		return errors.Join(fmt.Errorf("failed on create project branch dir: `%s`", projectBranchPath), err)
	}

	return nil
}

func (d *FSDriver) getBranchRevisionPath(projectId string, branch string, revision string) string {
	return filepath.Join(d.imagesPath, projectId, branch, revision)
}

func (d *FSDriver) hasBranchRevision(projectId string, branch string, revision string) bool {
	_, err := os.Stat(d.getBranchRevisionPath(projectId, branch, revision))
	return err == nil
}

func (d *FSDriver) createBranchRevision(projectId string, branch string, revision string) error {
	if alreadyExist := d.hasBranchRevision(projectId, branch, revision); alreadyExist {
		return errors.New("trying create branch revision what is already exists")
	}

	projectBranchPath := d.getBranchRevisionPath(projectId, branch, revision)

	if err := os.Mkdir(projectBranchPath, 0755); err != nil {
		return errors.Join(fmt.Errorf("failed on create branch revision dir: `%s`", projectBranchPath), err)
	}

	return nil
}

func (d *FSDriver) pickFilesToWebStorage(project *Project, glBranch *gitlab.Branch, tmpPath string) error {
	branchRevisionPath := d.getBranchRevisionPath(project.ProjectID, glBranch.Name, glBranch.Commit.ShortID)
	for _, file := range project.DistFiles {
		input, err := os.ReadFile(path.Join(tmpPath, file))
		if err != nil {
			return err
		}

		err = os.WriteFile(path.Join(branchRevisionPath, filepath.Base(file)), input, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *FSDriver) getTmpPathForBuild(projectId string, branch string, revision string) string {
	return path.Join(
		d.configMap.StoragePath,
		fmt.Sprintf("%s-%s-%s", projectId, branch, revision),
	)
}

func (d *FSDriver) hasTmpDirForBuild(projectId string, branch string, revision string) bool {
	_, err := os.Stat(d.getTmpPathForBuild(projectId, branch, revision))
	return err == nil
}

func (d *FSDriver) removeTmpDirForBuild(projectId string, branch string, revision string) error {
	return os.RemoveAll(d.getTmpPathForBuild(projectId, branch, revision))
}
