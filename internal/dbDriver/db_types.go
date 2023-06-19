package dbDriver

import "time"

type ImageStatus uint

const (
	ImageStatusQueued     ImageStatus = iota
	ImageStatusReady                  = iota
	ImageStatusInProgress             = iota
)

type Model struct {
	ID        uint       `gorm:"primary_key" json:"-"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" json:"-"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type Pagination struct {
	Limit  int
	Offset int
}

type Image struct {
	Model
	Files     []ImageFile `json:"files"`
	Branch    string      `json:"branch"`
	Status    ImageStatus `json:"status"`
	Revision  string      `json:"revision"`
	ProjectId string      `json:"project_id"`
}

type ImageFile struct {
	Model
	Path    string `json:"path"`
	WebPath string `json:"web_path"`
	ImageId uint   `json:"-"`
}

type ExtendedImage struct {
	Image
	RevCount uint
}

type BranchInfo struct {
	Name     string `json:"name"`
	RevCount uint   `json:"revisions_count"`
}
