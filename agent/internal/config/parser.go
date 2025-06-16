package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func stripEnabledRecursive(input any) any {
	switch v := input.(type) {
	case map[string]any:
		cleaned := make(map[string]any)
		for k, val := range v {
			if k == "enabled" {
				continue
			}
			cleanedVal := stripEnabledRecursive(val)
			cleaned[k] = cleanedVal
		}
		return cleaned

	case []any:
		for i, val := range v {
			v[i] = stripEnabledRecursive(val)
		}
		return v

	default:
		return v
	}
}

func SaveToYAML(inputString map[string]any, filePath string) error {
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("config already exists, but unable to remove at %s: %v", filePath, err)
	}

	cleaned := stripEnabledRecursive(inputString)

	yamlData, err := yaml.Marshal(cleaned)
	if err != nil {
		return fmt.Errorf("could not marshal to YAML: %v", err)
	}

	if err := os.WriteFile(filePath, yamlData, 0644); err != nil {
		return fmt.Errorf("could not write YAML file: %v", err)
	}

	return nil
}
