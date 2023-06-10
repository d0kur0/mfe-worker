package src

import "github.com/xanzy/go-gitlab"

type DIContainer struct {
	queue        *Queue
	fsDriver     *FSDriver
	dbDriver     *DBDriver
	configMap    *ConfigMap
	gitlabClient *gitlab.Client
}

func NewDIContainer(configMap *ConfigMap, queue *Queue, fsDriver *FSDriver, dbDriver *DBDriver, gitlabClient *gitlab.Client) *DIContainer {
	return &DIContainer{
		queue:        queue,
		fsDriver:     fsDriver,
		dbDriver:     dbDriver,
		configMap:    configMap,
		gitlabClient: gitlabClient,
	}
}
