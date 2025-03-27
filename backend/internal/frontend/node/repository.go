package frontendnode

import (
	"database/sql"
	"encoding/json"
	"strings"
)

type FrontendNodeRepository struct {
	db *sql.DB
}

func NewFrontendNodeRepository(db *sql.DB) *FrontendNodeRepository {
	return &FrontendNodeRepository{
		db: db,
	}
}

func (f *FrontendNodeRepository) GetComponents(componentType string) (*[]ComponentInfo, error) {
	var query string
	var args []any

	query = "SELECT name, display_name, supported_signals, type FROM component_schemas"
	if componentType != "" {
		query += " WHERE type = ?"
		args = append(args, componentType)
	}

	rows, err := f.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var components []ComponentInfo

	for rows.Next() {
		var name, displayName, supportedSignalsRaw, typ string
		if err := rows.Scan(&name, &displayName, &supportedSignalsRaw, &typ); err != nil {
			return nil, err
		}

		// Convert comma-separated string to slice
		supportedSignals := strings.Split(supportedSignalsRaw, ",")
		for i := range supportedSignals {
			supportedSignals[i] = strings.TrimSpace(supportedSignals[i])
		}

		components = append(components, ComponentInfo{
			Name:             name,
			DisplayName:      displayName,
			Type:             typ,
			SupportedSignals: supportedSignals,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &components, nil
}

func (f *FrontendNodeRepository) GetComponentSchemaByName(componentName string) (any, error) {
	query := "SELECT schema_json FROM component_schemas WHERE name = ?"

	var rawSchema string
	err := f.db.QueryRow(query, componentName).Scan(&rawSchema)
	if err != nil {
		return nil, err
	}

	var schema any
	if err := json.Unmarshal([]byte(rawSchema), &schema); err != nil {
		return nil, err
	}

	return schema, nil
}
