package dbDriver

import "time"

type BuildStatus uint

const (
	BuildStatusReady      BuildStatus = iota
	BuildStatusInProgress             = iota
)

type Model struct {
	ID        uint       `gorm:"primary_key" json:"-"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type Pagination struct {
	Limit  int
	Offset int
}

type Branch struct {
	Model
	Name      string     `json:"name"`
	ProjectId string     `json:"project_id"`
	Revisions []Revision `json:"revisions,omitempty"`
}

type Revision struct {
	Model
	Name     string `json:"name"`
	Build    *Build `json:"build,omitempty"`
	BranchId uint   `json:"branch_id"`
}

type Build struct {
	Model
	Files      []BuildFiles `json:"files,omitempty"`
	Status     BuildStatus  `json:"status,omitempty"`
	RevisionId uint         `gorm:"index:unique_revision,unique" json:"revision_id"`
}

type BuildFiles struct {
	Model
	Path    string `json:"path"`
	WebPath string `json:"web_path"`
	BuildId uint   `json:"build_id"`
}
