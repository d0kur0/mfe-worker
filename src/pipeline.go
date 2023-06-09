package src

type Pipeline struct {
	dbDriver *DBDriver
	fsDriver *FSDriver
}

func (ctx *Pipeline) Run() {

}

func NewPipeline(dbDriver *DBDriver, fsDriver *FSDriver) *Pipeline {
	return &Pipeline{
		dbDriver: dbDriver,
		fsDriver: fsDriver,
	}
}
