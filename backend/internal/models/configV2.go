package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// ConfigSet represents the `config_sets` table.
type ConfigSet struct {
	ID          int               `json:"id"`
	Name        string            `json:"name" validate:"required"`
	Version     string            `json:"version"`
	Credentials map[string]string `json:"credentials,omitempty"` // Stored as JSON
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// Extension represents the `extensions` table.
type Extension struct {
	ID            int            `json:"id"`
	ConfigSetID   int            `json:"config_set_id"`
	ExtensionName string         `json:"extension_name"`
	Enabled       bool           `json:"enabled"`
	Endpoint      string         `json:"endpoint"`
	Extra         map[string]any `json:"extra,omitempty"` // Store as JSON
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// Pipeline represents the `pipelines` table.
type Pipeline struct {
	ID          int       `json:"id"`
	ConfigSetID int       `json:"config_set_id"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PipelineComponent represents the `pipeline_components` table.
type PipelineComponent struct {
	ID         int            `json:"id"`
	PipelineID int            `json:"pipeline_id"`
	Section    string         `json:"section"` // 'source', 'processor', 'destination'
	Type       string         `json:"type"`
	Name       string         `json:"name"`
	Config     map[string]any `json:"config"` // Stored as JSON
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// MarshalExtra converts Extra (map) to JSON string for DB storage.
func (e *Extension) MarshalExtra() (string, error) {
	data, err := json.Marshal(e.Extra)
	if err != nil {
		return "", fmt.Errorf("failed to marshal Extra: %w", err)
	}
	return string(data), nil
}

// UnmarshalExtra converts JSON string from DB into Extra (map).
func (e *Extension) UnmarshalExtra(data string) error {
	var extra map[string]any
	err := json.Unmarshal([]byte(data), &extra)
	if err != nil {
		return fmt.Errorf("failed to unmarshal Extra: %w", err)
	}
	e.Extra = extra
	return nil
}
