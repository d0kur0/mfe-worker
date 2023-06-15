package depsInjection

import (
	"github.com/xanzy/go-gitlab"
	"mfe-worker/internal/configMap"
	"mfe-worker/internal/dbDriver"
	"mfe-worker/internal/fsDriver"
	"mfe-worker/internal/queue"
)

type DIContainer struct {
	Queue        *queue.Queue
	FSDriver     *fsDriver.FSDriver
	DBDriver     *dbDriver.DBDriver
	ConfigMap    *configMap.ConfigMap
	GitlabClient *gitlab.Client
}

func NewDIContainer(configMap *configMap.ConfigMap, queue *queue.Queue, fsDriver *fsDriver.FSDriver, dbDriver *dbDriver.DBDriver, gitlabClient *gitlab.Client) *DIContainer {
	return &DIContainer{
		Queue:        queue,
		FSDriver:     fsDriver,
		DBDriver:     dbDriver,
		ConfigMap:    configMap,
		GitlabClient: gitlabClient,
	}
}
