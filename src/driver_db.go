package src

import (
	"errors"
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Artifact struct {
	gorm.Model
	Branch    string         `json:"branch"`
	ProjectId string         `json:"project_id"`
	Revision  string         `json:"revision"`
	Files     []ArtifactFile `json:"files"`
}

type ArtifactFile struct {
	gorm.Model
	WebPath    string `json:"web_path"`
	ArtifactID uint   `json:"-"`
}

type DBDriver struct {
	db        *gorm.DB
	configMap *ConfigMap
}

func (ctx *DBDriver) Save(artifact *Artifact) error {
	return ctx.db.Create(artifact).Error
}

func (ctx *DBDriver) GetList() (artifacts []Artifact, err error) {
	err = ctx.db.Model(&Artifact{}).Preload("Files").Find(&artifacts).Error
	return
}

func (ctx *DBDriver) CleanUp() error {
	return nil
}

func NewDBDriver(configMap *ConfigMap) (*DBDriver, error) {
	db, err := gorm.Open(sqlite.Open(configMap.DBPath), &gorm.Config{})
	if err != nil {
		return nil, errors.Join(fmt.Errorf("failed on open sqlite db on path: %s", configMap.DBPath), err)
	}

	err = db.AutoMigrate(&ArtifactFile{}, &Artifact{})
	if err != nil {
		return nil, errors.Join(errors.New("failed on auto migrate db models"), err)
	}

	return &DBDriver{
		db:        db,
		configMap: configMap,
	}, nil
}
