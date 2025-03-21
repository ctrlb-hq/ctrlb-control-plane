package client

type AgentRequest struct {
	Version  string `json:"version"`  // The version of the agent
	Hostname string `json:"hostname"` // The hostname of the machine running the agent
	Platform string `json:"platform"` // The platform (e.g., OS) the agent is running on
}

type AgentResponse struct {
	ID     int64          `json:"id"`     // Unique ID for the agent
	Config map[string]any `json:"config"` // Associated configuration
}
