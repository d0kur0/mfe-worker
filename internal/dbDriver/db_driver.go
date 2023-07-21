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

// branches

func (d *DBDriver) CreateBranch(branch *Branch) (*Branch, error) {
	return branch, d.db.Create(branch).Error
}

func (d *DBDriver) UpdateBranch(branch *Branch) (*Branch, error) {
	return branch, d.db.Updates(branch).Error
}

func (d *DBDriver) DeleteBranch(branch *Branch) error {
	return d.db.Delete(branch).Error
}

func (d *DBDriver) GetBranch(projectId, name string) (*Branch, error) {
	var branch *Branch

	err := d.db.Model(Branch{}).Where(Branch{
		Name:      name,
		ProjectId: projectId,
	}).Preload("Revisions").First(&branch).Error

	if err != nil {
		return nil, err
	}

	return branch, err
}

// revisions

func (d *DBDriver) CreateRevision(revision *Revision) (*Revision, error) {
	return revision, d.db.Create(revision).Error
}

func (d *DBDriver) UpdateRevision(revision *Revision) (*Revision, error) {
	return revision, d.db.Updates(revision).Error
}

func (d *DBDriver) DeleteRevision(revision *Revision) error {
	return d.db.Delete(revision).Error
}

// builds

func (d *DBDriver) CreateBuild(build *Build) (*Build, error) {
	return build, d.db.Create(build).Error
}

func (d *DBDriver) UpdateBuild(build *Build) (*Build, error) {
	return build, d.db.Updates(&build).Error
}

func (d *DBDriver) DeleteBuild(build *Build) error {
	return d.db.Delete(build).Error
}

func (d *DBDriver) GetBranches(projectId string, pagination Pagination) (list []Branch, total int64, err error) {
	d.db.Model(Branch{}).Where(Branch{
		ProjectId: projectId,
	}).Limit(pagination.Limit).Preload("Revisions").Offset(pagination.Offset).Find(&list).Count(&total)

	return
}

func (d *DBDriver) GetRevisions(projectId, branch string, pagination Pagination) (list []Revision, total int64, err error) {
	type TmpList struct {
		*Revision
		Total uint
	}

	var tmpList []TmpList

	err = d.db.Raw(`
    SELECT 
			r.*, 
			(
				SELECT 
					count(*) 
				FROM 
					revisions r 
				WHERE 
					b.id = r.branch_id
			) as total 
		FROM 
			branches b 
			JOIN revisions r on b.id = r.branch_id 
		WHERE
			b.project_id = ? 
			AND b.name = ? 
		ORDER BY 
			b.id DESC 
		LIMIT 
			? OFFSET ?
  `, projectId, branch, pagination.Limit, pagination.Offset).Scan(&tmpList).Error

	for _, r := range tmpList {
		list = append(list, Revision{
			Model:    r.Model,
			Name:     r.Name,
			Build:    r.Build,
			BranchId: r.BranchId,
		})

		total = int64(r.Total)
	}

	return
}

func (d *DBDriver) GetBuilds(branch *Branch, revision string, pagination Pagination) (builds []Build, total int64, err error) {
	var rev *Revision
	for _, r := range branch.Revisions {
		if r.Name == revision {
			rev = &r
			break
		}
	}

	if rev == nil {
		return
	}

	err = d.db.Model(&Build{}).Where(&Build{RevisionId: rev.ID}).
		Limit(pagination.Limit).
		Offset(pagination.Offset).
		Preload("Files").
		Find(&builds).Error

	return
}

func NewDBDriver(configMap *configMap.ConfigMap) (*DBDriver, error) {
	db, err := gorm.Open(sqlite.Open(configMap.DBPath), &gorm.Config{})
	if err != nil {
		return nil, errors.Join(fmt.Errorf("failed on open sqlite db on path: %s", configMap.DBPath), err)
	}

	err = db.AutoMigrate(&Branch{}, &Revision{}, &BuildFiles{}, &Build{})
	if err != nil {
		return nil, errors.Join(errors.New("failed on auto migrate db models"), err)
	}

	return &DBDriver{
		db:        db,
		configMap: configMap,
	}, nil
}
