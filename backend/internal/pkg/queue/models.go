package queue

import (
	"time"
)

// AgentStatus tracks the current status of an agent including retry attempts.
type AgentStatus struct {
	AgentID        string    `json:"agentId"`        // Unique ID of the agent
	Hostname       string    `json:"hostname"`       // Hostname where the agent is running
	IP             string    `json:"ip"`             // IP where the agent is running
	CurrentStatus  string    `json:"currentStatus"`  // Status of the agent (e.g., online, offline)
	RetryRemaining int       `json:"retryRemaining"` // Number of retry attempts left
	UpdatedAt      time.Time `json:"updatedAt"`      // Timestamp of the last status update
}

type AggregatedAgentMetrics struct {
	AgentID           string
	LogsRateSent      float64
	TracesRateSent    float64
	MetricsRateSent   float64
	DataSentBytes     float64
	DataReceivedBytes float64
	Status            string
	UpdatedAt         time.Time
}

type RealtimeAgentMetrics struct {
	AgentID           string
	LogsRateSent      float64
	TracesRateSent    float64
	MetricsRateSent   float64
	DataSentBytes     float64
	DataReceivedBytes float64
	CPUUtilization    float64
	MemoryUtilization float64
	Timestamp         time.Time
}
