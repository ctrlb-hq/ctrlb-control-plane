package frontendpipeline

type Pipeline struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Agents        int    `json:"agents"`
	IncomingBytes int    `json:"incomingBytes"`
	OutgoingBytes int    `json:"outgoingBytes"`
	UpdatedAt     int    `json:"updatedAt"`
}

type PipelineInfo struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	CreatedBy string `json:"createdBy"`
	CreatedAt int    `json:"createdAt"`
	UpdatedAt int    `json:"updatedAt"`
}

// Struct for pipeline component (Node)
type PipelineComponent struct {
	ComponentID   int    `json:"component_id"`
	Name          string `json:"name"`
	ComponentRole string `json:"component_role"`
	PluginName    string `json:"plugin_name"`
}

// Struct for dependency/edge
type PipelineEdge struct {
	FromComponentID int `json:"from_component_id"`
	ToComponentID   int `json:"to_component_id"`
}

// Struct for API response
type PipelineGraph struct {
	Nodes []PipelineComponent `json:"nodes"`
	Edges []PipelineEdge      `json:"edges"`
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
