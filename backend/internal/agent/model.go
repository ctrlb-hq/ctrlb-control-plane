package agent

type AgentRegisterResponse struct {
	ID     int64          `json:"id"`     // The unique ID assigned to the agent
	Config map[string]any `json:"config"` // The configuration settings for the agent
}
