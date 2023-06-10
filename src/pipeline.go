package src

type Pipeline struct {
	configMap *ConfigMap
	dbDriver  *DBDriver
	fsDriver  *FSDriver
}

func (ctx *Pipeline) Run() {
	// build app
}

func NewPipeline(dbDriver *DBDriver, fsDriver *FSDriver, configMap *ConfigMap) *Pipeline {
	return &Pipeline{
		dbDriver:  dbDriver,
		fsDriver:  fsDriver,
		configMap: configMap,
	}
}
