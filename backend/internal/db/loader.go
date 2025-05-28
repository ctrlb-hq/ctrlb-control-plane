package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"
)

func LoadSchemasFromDirectory(
	db *sql.DB,
	schemasFS fs.FS,
	uiSchemasFS fs.FS,
	typeMapping map[string]string,
	signalMapping map[string][]string,
) error {
	schemaFiles, err := fs.ReadDir(schemasFS, ".")
	if err != nil {
		return fmt.Errorf("failed to read schema directory: %w", err)
	}

	insertQuery := `
	INSERT OR IGNORE INTO component_schemas (
		name, type, display_name, supported_signals, schema_json, ui_schema_json
	) VALUES (?, ?, ?, ?, ?, ?);
	`

	for _, file := range schemaFiles {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		name := strings.TrimSuffix(file.Name(), ".json")

		// Read schema JSON
		schemaBytes, err := fs.ReadFile(schemasFS, file.Name())
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", file.Name(), err)
		}

		// Read and parse schema JSON
		var schemaMap map[string]any
		if err := json.Unmarshal(schemaBytes, &schemaMap); err != nil {
			return fmt.Errorf("invalid JSON in %s: %w", file.Name(), err)
		}

		// Determine display name
		displayName := name
		if title, ok := schemaMap["title"].(string); ok && title != "" {
			displayName = title
		}

		// Load corresponding UI schema
		uiSchemaBytes, err := fs.ReadFile(uiSchemasFS, file.Name())
		if err != nil {
			return fmt.Errorf("failed to read UI schema for %s: %w", file.Name(), err)
		}

		componentType := typeMapping[name]
		signals := signalMapping[name]
		signalStr := strings.Join(signals, ",")

		// Insert into DB
		_, err = db.Exec(insertQuery, name, componentType, displayName, signalStr, string(schemaBytes), string(uiSchemaBytes))
		if err != nil {
			return fmt.Errorf("failed to insert schema for %s: %w", name, err)
		}
	}

	return nil
}
