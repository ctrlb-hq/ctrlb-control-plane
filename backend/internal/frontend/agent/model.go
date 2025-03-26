package frontendagent

type UnmanagedAgents struct {
	ID       string `json:"id"`       // Unique ID for the agent
	Name     string `json:"name"`     // Descriptive name for the agent
	Type     string `json:"type"`     // Type of the agent
	Version  string `json:"version"`  // Version of the agent
	Hostname string `json:"hostname"` // Hostname where the agent is running
	Platform string `json:"platform"` // Operating system platform (e.g., linux, windows)
}

type AgentInfoWithLabels struct {
	ID           string            `json:"id"`            // Unique ID for the agent
	Name         string            `json:"name"`          // Descriptive name for the agent
	Version      string            `json:"version"`       // Version of the agent
	PipelineID   string            `json:"pipeline_id"`   // Pipeline the agent is associated with
	PipelineName string            `json:"pipeline_name"` // Pipeline the agent is associated with
	Status       string            `json:"status"`        // Current status of the agent (e.g., Disconnected, Connected)
	Hostname     string            `json:"hostname"`      // Hostname where the agent is running
	IP           string            `json:"ip"`            // IP where the agent is running
	Platform     string            `json:"platform"`      // Operating system platform (e.g., linux, windows)
	Labels       map[string]string `json:"labels"`        // Labels associated with the agent
}

type AgentMetrics struct {
	MetricName string      `json:"metric_name"`
	DataPoints []DataPoint `json:"data_points"`
}

type DataPoint struct {
	Timestamp int64   `json:"timestamp"` // Unix timestamp for efficiency
	Value     float64 `json:"value"`
}
