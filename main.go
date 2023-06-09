package main

import (
	"log"
	"mfe-worker/src"
)

func main() {
	var configMap src.ConfigMap

	if err := configMap.ReadFromFileSystem(); err != nil {
		log.Fatalf("failed init configuration: %s", err)
	}

	fsDriver, err := src.NewFSDriver(&configMap)
	if err != nil {
		log.Fatalf("failed on init fsDriver: %s", err)
	}

	dbDriver, err := src.NewDBDriver(&configMap)
	if err != nil {
		log.Fatalf("failed on init dbDriver: %s", err)
	}

	httpServer, err := src.NewHttpServer(&configMap, src.NewPipeline(dbDriver, fsDriver))
	if err != nil {
		log.Fatalf("failed on init httpServer: %s", err)
	}

	_ = fsDriver
	_ = dbDriver
	_ = httpServer
}
