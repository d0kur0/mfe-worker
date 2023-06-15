package dbDriver

import "gorm.io/gorm"

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
