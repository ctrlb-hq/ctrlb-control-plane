package models

import "time"

type AgentWithConfig struct {
	ID           string    `json:"id"`           // Unique ID for the agent
	Name         string    `json:"name"`         // Descriptive name for the agent
	Type         string    `json:"type"`         // Type/category of the agent (e.g., collector, forwarder)
	Version      string    `json:"version"`      // Version of the agent
	Hostname     string    `json:"hostname"`     // Hostname where the agent is running
	Platform     string    `json:"platform"`     // Operating system platform (e.g., linux, windows)
	Config       Config    `json:"config"`       // Associated configuration
	IsPipeline   bool      `json:"isPipeline"`   // Indicates if the agent is part of a data pipeline
	RegisteredAt time.Time `json:"registeredAt"` // Timestamp when the agent was registered
}

type AgentMetrics struct {
	AgentID            string  `json:"agentId"`            // Unique ID of the agent
	Status             string  `json:"status"`             // Current status (e.g., running, stopped)
	ExportedDataVolume float64 `json:"exportedDataVolume"` // Volume of data exported (in MB/GB)
	UptimeSeconds      float64 `json:"uptimeSeconds"`      // Uptime in seconds
	DroppedRecords     float64 `json:"droppedRecords"`     // Number of records dropped by the agent
}
