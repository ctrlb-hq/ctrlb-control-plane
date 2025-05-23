package models

import "time"

type AgentRegisterRequest struct {
	Name         string `json:"_"`             // The name of the agent
	Version      string `json:"version"`       // The version of the agent
	Hostname     string `json:"hostname"`      // The hostname of the machine running the agent
	IP           string `json:"ip"`            // IP address of machine running the agent
	Platform     string `json:"platform"`      // The platform (e.g., OS) the agent is running on
	Type         string `json:"type"`          // The type of agent
	PipelineName string `json:"pipeline_name"` // The name of the pipeline the agent is associated with
	StartedBy    string `json:"started_by"`    // The user who started the agent
	RegisteredAt int64  `json:"registered_at"` // The Unix timestamp when the agent was registered
}

// AgentMetrics represents metrics related to an agent's performance.
type AgentMetrics struct {
	AgentID            string    `json:"agent_id"`             // Unique ID of the agent
	Status             string    `json:"status"`               // Current status (e.g., running, stopped)
	ExportedDataVolume int       `json:"exported_data_volume"` // Volume of data exported (in MB/GB)
	UptimeSeconds      int       `json:"uptime_seconds"`       // Uptime in seconds
	DroppedRecords     int       `json:"dropped_records"`      // Number of records dropped by the agent
	UpdatedAt          time.Time `json:"updated_at"`           // Timestamp of the last metrics update
}

// AgentInfoHome represents an agent with relevant details like type, version, and platform.
type AgentInfoHome struct {
	ID           int64  `json:"id"`            // Unique ID for the agent
	Name         string `json:"name"`          // Descriptive name for the agent
	Status       string `json:"status"`        // Current status of the agent (e.g., Disconnected, Connected)
	PipelineName string `json:"pipeline_name"` // Pipeline the agent is associated with
	Version      string `json:"version"`       // Version of the agent
	LogRate      int    `json:"log_rate"`      // Log rate of the agent
	MetricsRate  int    `json:"metrics_rate"`  // Metrics rate of the agent
	TraceRate    int    `json:"trace_rate"`    // Trace rate of the agent
	Hostname     string `json:"-"`
	IP           string `json:"_"`
}
