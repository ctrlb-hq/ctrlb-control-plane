package models

import "time"

// Config represents the configuration settings for an agent.
type Config struct {
	ID          string    `json:"id"`          // Unique ID of the config
	Name        string    `json:"name"`        // Name for the config
	Description string    `json:"description"` // Brief description of the config
	Config      string    `json:"config"`      // Configuration data (usually in JSON or YAML format)
	TargetAgent string    `json:"targetAgent"` // Type of agent this configuration is applicable to
	CreatedAt   time.Time `json:"createdAt"`   // Timestamp when the config was created
	UpdatedAt   time.Time `json:"updatedAt"`   // Timestamp when the config was last updated
}
