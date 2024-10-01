package models

import (
	"time"
)

type Agent struct {
	IPAddress           string              `json:"ip_address"`
	Uptime              int                 `json:"uptime"`
	AgentIdentifier     string              `json:"agent_identifier"`
	AgentVersion        string              `json:"agent_version"`
	Port                int                 `json:"port"`
	AuthMethod          string              `json:"auth_method"`
	ConnectionStatus    ConnectionStatus    `json:"connection_status"`
	ResourceUtilization ResourceUtilization `json:"resource_utilization"`
	Tags                map[string]string   `json:"tags"`
	Environment         string              `json:"environment"`
	CurrentConfig       string              `json:"current_config"`
}

type ConnectionStatus struct {
	IsActive           bool      `json:"is_active"`
	ConnectionAttempts int       `json:"connection_attempts"`
	MaxAttempts        int       `json:"max_attempts"`
	LastAttemptTime    time.Time `json:"last_attempt_time"`
}

type ResourceUtilization struct {
	CPU     float64            `json:"cpu"`
	Memory  float64            `json:"memory"`
	Network NetworkUtilization `json:"network"`
}

type NetworkUtilization struct {
	BytesSent       int64 `json:"bytes_sent"`
	BytesReceived   int64 `json:"bytes_received"`
	RecordsSent     int64 `json:"records_sent"`
	RecordsReceived int64 `json:"records_received"`
}
