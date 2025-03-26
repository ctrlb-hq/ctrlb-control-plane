package models

import "time"

type AgentRegisterRequest struct {
	Name         string `json:"_"`             // The name of the agent
	Version      string `json:"version"`       // The version of the agent
	Hostname     string `json:"hostname"`      // The hostname of the machine running the agent
	Platform     string `json:"platform"`      // The platform (e.g., OS) the agent is running on
	Type         string `json:"type"`          // The type of agent
	RegisteredAt int64  `json:"registered_at"` // The Unix timestamp when the agent was registered
}

// AgentWithConfig represents an agent with its configuration details.
type AgentWithConfig struct {
	ID           string    `json:"id"`           // Unique ID for the agent
	Name         string    `json:"name"`         // Descriptive name for the agent
	Type         string    `json:"type"`         // Type/category of the agent (e.g., collector, forwarder)
	Version      string    `json:"version"`      // Version of the agent
	Hostname     string    `json:"hostname"`     // Hostname where the agent is running
	Platform     string    `json:"platform"`     // Operating system platform (e.g., linux, windows)
	IsPipeline   bool      `json:"isPipeline"`   // Indicates if the agent is part of a data pipeline
	RegisteredAt time.Time `json:"registeredAt"` // Timestamp when the agent was registered
}

// AgentMetrics represents metrics related to an agent's performance.
type AgentMetrics struct {
	AgentID            string    `json:"agentId"`            // Unique ID of the agent
	Status             string    `json:"status"`             // Current status (e.g., running, stopped)
	ExportedDataVolume int       `json:"exportedDataVolume"` // Volume of data exported (in MB/GB)
	UptimeSeconds      int       `json:"uptimeSeconds"`      // Uptime in seconds
	DroppedRecords     int       `json:"droppedRecords"`     // Number of records dropped by the agent
	UpdatedAt          time.Time `json:"updatedAt"`          // Timestamp of the last metrics update
}
