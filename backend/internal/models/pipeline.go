package models

type CreatePipelineRequest struct {
	Name          string        `json:"name"`
	CreatedBy     string        `json:"created_by"`
	AgentIDs      []int         `json:"agent_ids"`
	PipelineGraph PipelineGraph `json:"pipeline_graph"`
}

// Struct for pipeline component (Node)
type PipelineNodes struct {
	ComponentID      int            `json:"component_id"`
	Name             string         `json:"name"`
	ComponentRole    string         `json:"component_role"`
	ComponentName    string         `json:"component_name"`
	Config           map[string]any `json:"config"`
	SupportedSignals []string       `json:"supported_signals"`
}

// Struct for dependency/edge
type PipelineEdges struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

// Struct for API response
type PipelineGraph struct {
	Nodes []PipelineNodes `json:"nodes"`
	Edges []PipelineEdges `json:"edges"`
}
