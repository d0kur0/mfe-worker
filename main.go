package main

import (
	"log"
	"mfe-worker/src"
)

func main() {
	configMap, err := src.NewConfigMap()
	if err != nil {
		log.Fatalf("failed init configuration: %s", err)
	}

	fsDriver, err := src.NewFSDriver(configMap)
	if err != nil {
		log.Fatalf("failed on init fsDriver: %s", err)
	}

	dbDriver, err := src.NewDBDriver(configMap)
	if err != nil {
		log.Fatalf("failed on init dbDriver: %s", err)
	}

	queue := src.NewQueue(configMap)
	queue.StartQueueWorker()

	diContainer := src.NewDIContainer(configMap, queue, fsDriver, dbDriver)

	httpServer, err := src.NewHttpServer(diContainer)
	if err != nil {
		log.Fatalf("failed on init httpServer: %s", err)
	}

	err = httpServer.SetupHttpHandlers()
	if err != nil {
		log.Fatalf("failed on SetupHttpHandlers: %s", err)
	}
}
