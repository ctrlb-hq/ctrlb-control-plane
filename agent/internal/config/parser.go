package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func SaveToYAML(inputString map[string]any, filePath string) error {
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("config already exists, but unable to remove at %s: %v", filePath, err)
	}

	yamlData, err := yaml.Marshal(inputString)
	if err != nil {
		return fmt.Errorf("could not marshal to YAML: %v", err)
	}

	if err := os.WriteFile(filePath, yamlData, 0644); err != nil {
		return fmt.Errorf("could not write YAML file: %v", err)
	}

	return nil
}
