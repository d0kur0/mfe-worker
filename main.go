package main

import "log"

func main() {
	var configMap ConfigMap

	if err := configMap.ReadFromFileSystem(); err != nil {
		log.Fatalf("failed init configuration: %s", err)
	}

	var artifactDriver = NewArtifactDriver(configMap)

	if err := artifactDriver.LookAtRequirements(); err != nil {
		log.Fatalf("failed on check requirements: %s", err)
	}
}
