package frontendpipeline

type Pipeline struct {
}

type PipelineInfo struct {
	ID        int
	Name      string
	CreatedBy string
	CreatedAt int
	UpdatedAt int
}

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
