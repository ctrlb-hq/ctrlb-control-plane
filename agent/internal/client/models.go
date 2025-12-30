package client

type AgentRequest struct {
	IP           string `json:"ip"`
	Version      string `json:"version"`       // The version of the agent
	Hostname     string `json:"hostname"`      // The hostname of the machine running the agent
	Platform     string `json:"platform"`      // The platform (e.g., OS) the agent is running on
	PipelineName string `json:"pipeline_name"` // The name of the pipeline
	StartedBy    string `json:"started_by"`    // The user who started the agent
	Type         string `json:"type"`          // The type of agent
}

type AgentResponse struct {
	ID     int64          `json:"id"`     // Unique ID for the agent
	Config map[string]any `json:"config"` // Associated configuration
}

// HeartbeatRequest represents metrics payload sent to backend
type HeartbeatRequest struct {
	LogsRateSent      float64 `json:"logs_rate_sent"`
	TracesRateSent    float64 `json:"traces_rate_sent"`
	MetricsRateSent   float64 `json:"metrics_rate_sent"`
	DataSentBytes     float64 `json:"data_sent_bytes"`
	DataReceivedBytes float64 `json:"data_received_bytes"`
	CPUUtilization    float64 `json:"cpu_utilization"`
	MemoryUtilization float64 `json:"memory_utilization"`
}
