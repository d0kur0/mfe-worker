package configMap

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
)

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

		configAsBytes, err := json.MarshalIndent(ConfigTemplate, "", "\t")
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
