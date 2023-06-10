package src

import (
	"errors"
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type ImageStatus uint

const (
	ImageStatusQueued     ImageStatus = iota
	ImageStatusReady                  = iota
	ImageStatusInProgress             = iota
)

type Image struct {
	gorm.Model
	Files     []ImageFile `json:"files"`
	Branch    string      `json:"branch"`
	Status    ImageStatus `json:"status"`
	Revision  string      `json:"revision"`
	ProjectId string      `json:"project_id"`
}

type ImageFile struct {
	gorm.Model
	WebPath string `json:"web_path"`
	ImageId uint   `json:"-"`
}

type DBDriver struct {
	db        *gorm.DB
	configMap *ConfigMap
}

func (ctx *DBDriver) Save(image *Image) (*Image, error) {
	result := ctx.db.Create(&image)
	return image, result.Error
}

func (ctx *DBDriver) GetList() (artifacts []Image, err error) {
	err = ctx.db.Model(&Image{}).Preload("Files").Find(&artifacts).Error
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

	err = db.AutoMigrate(&ImageFile{}, &Image{})
	if err != nil {
		return nil, errors.Join(errors.New("failed on auto migrate db models"), err)
	}

	return &DBDriver{
		db:        db,
		configMap: configMap,
	}, nil
}
