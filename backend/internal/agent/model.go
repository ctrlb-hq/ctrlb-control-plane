package agent

// AgentRegisterRequest represents the request payload for registering a new agent.
type AgentRegisterRequest struct {
	Version      string `json:"version"`       // The version of the agent
	Hostname     string `json:"hostname"`      // The hostname of the machine running the agent
	Platform     string `json:"platform"`      // The platform (e.g., OS) the agent is running on
	Type         string `json:"type"`          // The type of agent
	RegisteredAt int64  `json:"registered_at"` // The Unix timestamp when the agent was registered
}

type AgentRegisterResponse struct {
	ID     int64                  `json:"id"`     // The unique ID assigned to the agent
	Config map[string]interface{} `json:"config"` // The configuration settings for the agent
}
