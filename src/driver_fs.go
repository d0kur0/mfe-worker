package src

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type FSDriver struct {
	configMap *ConfigMap
}

func NewFSDriver(configMap *ConfigMap) (*FSDriver, error) {
	if _, err := os.Stat(configMap.StoragePath); os.IsNotExist(err) {
		return nil, fmt.Errorf(`artifact directory was not found (%s), create it first with correct access rights`, configMap.StoragePath)
	}

	return &FSDriver{configMap}, nil
}

func (ctx *FSDriver) hasProject(projectId string) bool {
	_, err := os.Stat(filepath.Join(ctx.configMap.StoragePath, projectId))
	return err != nil
}

func (ctx *FSDriver) createProject(projectId string) error {
	if alreadyExist := ctx.hasProject(projectId); alreadyExist {
		return errors.New("trying create project what is already exists")
	}

	path := filepath.Join(ctx.configMap.StoragePath, projectId)

	if err := os.Mkdir(path, 0755); err != nil {
		return errors.Join(fmt.Errorf("failed on create project dir: `%s`", path), err)
	}

	return nil
}

func (ctx *FSDriver) hasProjectBranch(projectId string, branch string) bool {
	_, err := os.Stat(filepath.Join(ctx.configMap.StoragePath, projectId, branch))
	return err != nil
}

func (ctx *FSDriver) createProjectBranch(projectId string, branch string) error {
	if alreadyExist := ctx.hasProjectBranch(projectId, branch); alreadyExist {
		return errors.New("trying create project branch what is already exists")
	}

	path := filepath.Join(ctx.configMap.StoragePath, projectId, branch)

	if err := os.Mkdir(path, 0755); err != nil {
		return errors.Join(fmt.Errorf("failed on create project branch dir: `%s`", path), err)
	}

	return nil
}
