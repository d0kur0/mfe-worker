package main

import (
	"fmt"
	"os"
)

type ArtifactDriver struct {
	configMap ConfigMap
}

func NewArtifactDriver(configMap ConfigMap) *ArtifactDriver {
	return &ArtifactDriver{configMap}
}

func (a *ArtifactDriver) LookAtRequirements() error {
	if _, err := os.Stat(a.configMap.StoragePath); os.IsNotExist(err) {
		return fmt.Errorf(`artifact directory was not found (%s), create it first with correct access rights`, a.configMap.StoragePath)
	}

	return nil
}
