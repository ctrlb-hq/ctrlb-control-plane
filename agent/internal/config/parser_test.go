package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSaveToYAML_Success(t *testing.T) {
	filePath := "test_config.yaml"
	config := map[string]any{
		"log_level": "debug",
		"enabled":   true,
	}

	err := SaveToYAML(config, filePath)
	assert.NoError(t, err)

	// Check file exists
	_, err = os.Stat(filePath)
	assert.NoError(t, err)

	// Cleanup
	_ = os.Remove(filePath)
}

func TestSaveToYAML_InvalidPath(t *testing.T) {
	// Assuming this is an invalid directory
	filePath := "/invalid/path/test_config.yaml"
	config := map[string]any{
		"log_level": "debug",
	}

	err := SaveToYAML(config, filePath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not write YAML")
}
