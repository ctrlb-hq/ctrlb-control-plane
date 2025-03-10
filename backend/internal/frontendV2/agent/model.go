package frontendagent

import (
	"time"
)

// Agent represents an agent with relevant details like type, version, and platform.
type Agent struct {
	ID           string    `json:"id"`           // Unique ID for the agent
	Name         string    `json:"name"`         // Descriptive name for the agent
	Type         string    `json:"type"`         // Type/category of the agent (e.g., collector, forwarder)
	Version      string    `json:"version"`      // Version of the agent
	Hostname     string    `json:"hostname"`     // Hostname where the agent is running
	Platform     string    `json:"platform"`     // Operating system platform (e.g., linux, windows)
	ConfigID     string    `json:"configId"`     // Associated configuration ID
	IsPipeline   bool      `json:"isPipeline"`   // Indicates if the agent is part of a data pipeline
	RegisteredAt time.Time `json:"registeredAt"` // Timestamp when the agent was registered
}
