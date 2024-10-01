package models

type AgentRequest struct {
	AgentID string `json:"agentId"`
}

type AgentRegisterRequest struct {
	Type     string `json:"type"`
	Version  string `json:"version"`
	Hostname string `json:"hostname"`
	Platform string `json:"platform"`
	Config   string `json:"config"`
}

type ConfigUpdateRequest struct {
	AgentID string `json:"agentId"`
	Config  string `json:"config"`
}
