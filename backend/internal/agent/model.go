package agent

type AgentRegisterResponse struct {
	ID     int64          `json:"id"`     // The unique ID assigned to the agent
	Config map[string]any `json:"config"` // The configuration settings for the agent
}

type HeartbeatRequest struct {
	LogsRateSent      float64 `json:"logs_rate_sent"`
	TracesRateSent    float64 `json:"traces_rate_sent"`
	MetricsRateSent   float64 `json:"metrics_rate_sent"`
	DataSentBytes     float64 `json:"data_sent_bytes"`
	DataReceivedBytes float64 `json:"data_received_bytes"`
	CPUUtilization    float64 `json:"cpu_utilization"`
	MemoryUtilization float64 `json:"memory_utilization"`
}