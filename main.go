package main

import (
	"fmt"
	"github.com/xanzy/go-gitlab"
	"log"
	"mfe-worker/internal/configMap"
	"mfe-worker/internal/dbDriver"
	"mfe-worker/internal/depsInjection"
	"mfe-worker/internal/fsDriver"
	"mfe-worker/internal/http"
	"mfe-worker/internal/queue"
)

func main() {
	configMapInstance, err := configMap.NewConfigMap()
	if err != nil {
		log.Fatalf("failed init configuration: %s", err)
	}

	fsDriverInstance, err := fsDriver.NewFSDriver(configMapInstance)
	if err != nil {
		log.Fatalf("failed on init fsDriver: %s", err)
	}

	dbDriverInstance, err := dbDriver.NewDBDriver(configMapInstance)
	if err != nil {
		log.Fatalf("failed on init dbDriver: %s", err)
	}

	queue := queue.NewQueue(configMapInstance)
	queue.StartQueueWorker()

	gitlabClientArgs := gitlab.WithBaseURL(fmt.Sprintf("%s/api/v4", configMapInstance.GitlabUrl))
	gitlabClient, err := gitlab.NewClient(configMapInstance.GitlabToken, gitlabClientArgs)

	diContainer := depsInjection.NewDIContainer(configMapInstance, queue, fsDriverInstance, dbDriverInstance, gitlabClient)

	httpServer, err := http.NewHttpServer(diContainer)
	if err != nil {
		log.Fatalf("failed on init httpServer: %s", err)
	}

	err = httpServer.SetupHttpHandlers()
	if err != nil {
		log.Fatalf("failed on SetupHttpHandlers: %s", err)
	}
}
