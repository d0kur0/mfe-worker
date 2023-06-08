package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
)

var configPlaces = [...]string{".mfe-worker.json", "~/.mfe-worker.json"}

type Project struct {
	// gitlab project id
	ProjectID string `json:"projectID"`
	// branches white list, is array empty - track all
	Branches []string `json:"branches"`
}

type ConfigMap struct {
	// base url for gitlab instance
	GitlabUrl string `json:"gitlabUrl"`
	// token with access to read projects
	GitlabToken string `json:"gitlabToken"`
	// path for store build artifacts
	StoragePath string `json:"storagePath"`
	// projects list
	Projects []Project `json:"projects"`
}

func (c *ConfigMap) ReadFromFileSystem() error {
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
			GitlabUrl:   "[base gitlab instance url]",
			GitlabToken: "[gitlab token with access to read projects]",
			StoragePath: "[path to dir for store build assets]",
			Projects: []Project{{
				ProjectID: "[project id of gitlab]",
				Branches:  []string{"[branches white list or empty array for pass all names]"},
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

	if err := json.Unmarshal(configAsBytes, c); err != nil {
		return errors.Join(errors.New("failed on parse configuration file from JSON"), err)
	}

	return nil
}
