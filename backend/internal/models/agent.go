package models

import "time"

// AgentMetrics represents metrics related to an agent's performance.
type AgentMetrics struct {
	AgentID            string    `json:"agentId"`            // Unique ID of the agent
	Status             string    `json:"status"`             // Current status (e.g., running, stopped)
	ExportedDataVolume int       `json:"exportedDataVolume"` // Volume of data exported (in MB/GB)
	UptimeSeconds      int       `json:"uptimeSeconds"`      // Uptime in seconds
	DroppedRecords     int       `json:"droppedRecords"`     // Number of records dropped by the agent
	UpdatedAt          time.Time `json:"updatedAt"`          // Timestamp of the last metrics update
}
