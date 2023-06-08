package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type FSDriver struct {
	configMap ConfigMap
}

func NewFSDriver(configMap ConfigMap) *FSDriver {
	return &FSDriver{configMap}
}

func (a *FSDriver) LookAtRequirements() error {
	if _, err := os.Stat(a.configMap.StoragePath); os.IsNotExist(err) {
		return fmt.Errorf(`artifact directory was not found (%s), create it first with correct access rights`, a.configMap.StoragePath)
	}

	return nil
}

func (a *FSDriver) hasProject(projectId string) bool {
	_, err := os.Stat(filepath.Join(a.configMap.StoragePath, projectId))
	return err != nil
}

func (a *FSDriver) createProject(projectId string) error {
	if alreadyExist := a.hasProject(projectId); alreadyExist {
		return errors.New("trying create project what is already exists")
	}

	path := filepath.Join(a.configMap.StoragePath, projectId)

	if err := os.Mkdir(path, 0755); err != nil {
		return errors.Join(fmt.Errorf("failed on create project dir: `%s`", path), err)
	}

	return nil
}

func (a *FSDriver) hasProjectBranch(projectId string, branch string) bool {
	_, err := os.Stat(filepath.Join(a.configMap.StoragePath, projectId, branch))
	return err != nil
}

func (a *FSDriver) createProjectBranch(projectId string, branch string) error {
	if alreadyExist := a.hasProjectBranch(projectId, branch); alreadyExist {
		return errors.New("trying create project branch what is already exists")
	}

	path := filepath.Join(a.configMap.StoragePath, projectId, branch)

	if err := os.Mkdir(path, 0755); err != nil {
		return errors.Join(fmt.Errorf("failed on create project branch dir: `%s`", path), err)
	}

	return nil
}
