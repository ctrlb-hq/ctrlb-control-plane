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

type UserRegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
