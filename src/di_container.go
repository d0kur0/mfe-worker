package src

type DIContainer struct {
	queue     *Queue
	fsDriver  *FSDriver
	dbDriver  *DBDriver
	configMap *ConfigMap
}

func NewDIContainer(configMap *ConfigMap, queue *Queue, fsDriver *FSDriver, dbDriver *DBDriver) *DIContainer {
	return &DIContainer{
		queue:     queue,
		fsDriver:  fsDriver,
		dbDriver:  dbDriver,
		configMap: configMap,
	}
}
