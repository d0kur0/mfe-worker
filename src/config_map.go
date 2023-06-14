package src

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
)

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

func (ctx *ConfigMap) ReadFromFileSystem() error {
	defaultPlacePath := ""

	for _, path := range configPlaces {
		if _, err := os.Stat(path); err == nil {
			defaultPlacePath = path
			break
		}
	}

	if len(defaultPlacePath) == 0 {
		log.Printf("trying to find config on default places (%s) and not found", configPlaces)
		log.Printf("trying create config file from template at: %s", configPlaces[0])

		configMapTemplate := ConfigMap{
			HttpBaseUrl: "[base http url: ex: http://localhost:3433]",
			DBPath:      "[path for save sqlite db file, ex: mf_worker.db]",
			GitlabUrl:   "[base gitlab instance url]",
			GitlabToken: "[gitlab token with access to read projects]",
			StoragePath: "[path to dir for store build assets]",
			Projects: []Project{{
				Branches:      []string{"[branches white list or empty array for pass all names]"},
				ProjectID:     "[project id of gitlab]",
				DistFiles:     []string{"[files what need to save after build and share]", "dist/app.js", "dist/app.css"},
				ProjectName:   "[project name (any value, not gitlab name)]",
				BuildCommands: []string{"[commands for build project after clone]", "npm run prebuild", "npm run build"},
			}},
		}

		configAsBytes, err := json.MarshalIndent(configMapTemplate, "", "\t")
		if err != nil {
			return errors.Join(errors.New("failed on stringify ConfigMap struct"), err)
		}

		if err := os.WriteFile(configPlaces[0], configAsBytes, 0644); err != nil {
			return errors.Join(fmt.Errorf("failed on write to file `%s`", configPlaces[0]), err)
		}

		log.Fatalf("empty configuration file by template was created at `%s`, fill correct values and restart", configPlaces[0])
	}

	configAsBytes, err := os.ReadFile(defaultPlacePath)
	if err != nil {
		return errors.Join(fmt.Errorf("failed on read config file `%s`, check access rights", defaultPlacePath), err)
	}

	if err := json.Unmarshal(configAsBytes, ctx); err != nil {
		return errors.Join(errors.New("failed on parse configuration file from JSON"), err)
	}

	return nil
}

func NewConfigMap() (*ConfigMap, error) {
	var configMap ConfigMap
	return &configMap, configMap.ReadFromFileSystem()
}
