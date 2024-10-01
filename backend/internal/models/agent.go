package models

type Agent struct {
	ID       string `json:"-"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Version  string `json:"version"`
	Hostname string `json:"hostname"`
	Platform string `json:"platform"`
	Config   string `json:"config"`
}

type AgentInfo struct {
	AgentID            string `json:"agentId"`
	Status             string `json:"status"`
	ExportedDataVolume int    `json:"exportedDataVolume"`
	Uptime             int    `json:"uptime"`
	DroppedRecords     int    `json:"droppedRecords"`
}

type AgentStatus struct {
	AgentID        string
	Hostname       string
	CurrentStatus  string
	RetryRemaining int
}
