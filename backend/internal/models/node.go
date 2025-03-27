package models

import (
	"encoding/json"
	"time"
)

type ComponentSchema struct {
	ID         int64           `db:"id"`
	Type       string          `db:"type"`        // receiver, exporter, processor, etc.
	Name       string          `db:"name"`        // e.g. otlp, batch, logging
	Subtype    string          `db:"subtype"`     // Optional: grpc, http, etc.
	SchemaJSON json.RawMessage `db:"schema_json"` // The full JSON schema
	Version    string          `db:"version"`     // Optional version tracking
	CreatedAt  time.Time       `db:"created_at"`
	UpdatedAt  time.Time       `db:"updated_at"`
}
