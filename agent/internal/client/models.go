package client

type AgentRequest struct {
	Type       string `json:"type"`       // The type of the agent (e.g., worker, collector)
	Version    string `json:"version"`    // The version of the agent
	Hostname   string `json:"hostname"`   // The hostname of the machine running the agent
	Platform   string `json:"platform"`   // The platform (e.g., OS) the agent is running on
	IsPipeline bool   `json:"isPipeline"` // Indicates if the agent is part of a data pipeline
}
