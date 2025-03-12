package frontendagent

// AgentInfoHome represents an agent with relevant details like type, version, and platform.
type AgentInfoHome struct {
	ID           string `json:"id"`           // Unique ID for the agent
	Name         string `json:"name"`         // Descriptive name for the agent
	Status       string `json:"status"`       // Current status of the agent (e.g., Disconnected, Connected)
	PipelineName string `json:"pipelineName"` // Pipeline the agent is associated with
	Version      string `json:"version"`      // Version of the agent
	LogRate      int    `json:"logRate"`      // Log rate of the agent
	MetricsRate  int    `json:"metricsRate"`  // Metrics rate of the agent
	TraceRate    int    `json:"traceRate"`    // Trace rate of the agent
}

type AgentInfoWithLabels struct {
	ID           string            `json:"id"`           // Unique ID for the agent
	Name         string            `json:"name"`         // Descriptive name for the agent
	Version      string            `json:"version"`      // Version of the agent
	PipelineID   string            `json:"pipelineID"`   // Pipeline the agent is associated with
	PipelineName string            `json:"pipelineName"` // Pipeline the agent is associated with
	Status       string            `json:"status"`       // Current status of the agent (e.g., Disconnected, Connected)
	Hostname     string            `json:"hostname"`     // Hostname where the agent is running
	Platform     string            `json:"platform"`     // Operating system platform (e.g., linux, windows)
	Labels       map[string]string `json:"labels"`       // Labels associated with the agent
}
