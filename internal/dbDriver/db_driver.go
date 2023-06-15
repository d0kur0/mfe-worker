package dbDriver

import (
	"errors"
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"mfe-worker/internal/configMap"
)

type DBDriver struct {
	db        *gorm.DB
	configMap *configMap.ConfigMap
}

func (d *DBDriver) Save(image *Image) error {
	return d.db.Create(&image).Error
}

func (d *DBDriver) Update(image *Image) error {
	return d.db.Updates(&image).Error
}

func (d *DBDriver) GetList() (images []Image, err error) {
	err = d.db.Model(&Image{}).Preload("Files").Find(&images).Error
	return
}

func (d *DBDriver) CleanUp() error {
	return nil
}

func (d *DBDriver) IsRevisionExists(projectId string, branch string, revision string) bool {
	var hasImageWithSameRevision bool

	d.db.
		Model(&Image{}).
		Select("count(*) > 0").
		Where("revision = ? AND project_id = ? AND branch = ?", revision, projectId, branch).
		Find(&hasImageWithSameRevision)

	return hasImageWithSameRevision
}

func NewDBDriver(configMap *configMap.ConfigMap) (*DBDriver, error) {
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
