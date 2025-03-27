package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func LoadSchemasFromDirectory(db *sql.DB, dir string, typeMapping map[string]string, signalMapping map[string][]string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read schema directory: %w", err)
	}

	insertQuery := `
	INSERT OR IGNORE INTO component_schemas (
		name, type, display_name, supported_signals, schema_json
	) VALUES (?, ?, ?, ?, ?);
	`

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		// Extract name from filename (without .json)
		name := strings.TrimSuffix(file.Name(), ".json")

		// Read schema file
		fullPath := filepath.Join(dir, file.Name())
		schemaBytes, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", fullPath, err)
		}

		// Parse to extract display_name (title)
		var schemaMap map[string]any
		if err := json.Unmarshal(schemaBytes, &schemaMap); err != nil {
			return fmt.Errorf("invalid JSON in %s: %w", fullPath, err)
		}

		displayName := name // fallback
		if title, ok := schemaMap["title"].(string); ok && title != "" {
			displayName = title
		}

		// Lookup type and supported signals
		componentType := typeMapping[name]
		signals := signalMapping[name]
		signalStr := strings.Join(signals, ",")

		// Execute insert
		_, err = db.Exec(insertQuery, name, componentType, displayName, signalStr, string(schemaBytes))
		if err != nil {
			return fmt.Errorf("failed to insert schema for %s: %w", name, err)
		}
	}

	return nil
}
