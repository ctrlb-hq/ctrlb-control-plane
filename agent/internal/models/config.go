package models

import "time"

type Config struct {
	ID          string    `json:"id"`          // Unique ID of the config
	Name        string    `json:"name"`        // Name for config
	Description string    `json:"description"` // Brief description of the config
	Config      string    `json:"config"`      // Configuration data (usually in JSON or YAML format)
	TargetAgent string    `json:"targetAgent"` // Type of agent this configuration is applicable to
	CreatedAt   time.Time `json:"createdAt"`   // Timestamp when the config was created
	UpdatedAt   time.Time `json:"updatedAt"`   // Timestamp when the config was last updated
}

type ConfigUpdateRequest struct {
	Name        string `json:"name"`        // Configuration name
	Description string `json:"description"` // Brief description of the configuration
	Config      string `json:"config"`      // Configuration content (e.g., JSON or YAML)
	TargetAgent string `json:"targetAgent"` // Agent type the configuration targets
}
