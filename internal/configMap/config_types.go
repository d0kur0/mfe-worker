package configMap

var configPlaces = [...]string{".mfe-worker.json", "~/.mfe-worker.json"}

type Project struct {
	Branches      []string `json:"branches"`
	ProjectID     string   `json:"project_id"`
	DistFiles     []string `json:"dist_files"`
	ProjectName   string   `json:"project_name"`
	BuildCommands []string `json:"build_commands"`
}

type ConfigMap struct {
	HttpBaseUrl string    `json:"http_base_url"`
	DBPath      string    `json:"db_path"`
	Projects    []Project `json:"projects"`
	GitlabUrl   string    `json:"gitlab_url"`
	GitlabToken string    `json:"gitlab_token"`
	StoragePath string    `json:"storage_path"`
}
