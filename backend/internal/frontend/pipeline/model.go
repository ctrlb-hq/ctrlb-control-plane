package frontendpipeline

type CreatePipelineRequest struct {
	Name          string        `json:"name"`
	AgentsID      []int         `json:"agent_ids"`
	PipelineGraph PipelineGraph `json:"pipeline_graph"`
}

type Pipeline struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Agents        int    `json:"agents"`
	IncomingBytes int    `json:"incoming_bytes"`
	OutgoingBytes int    `json:"outgoing_bytes"`
	UpdatedAt     int    `json:"updatedAt"`
}

type PipelineInfo struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	CreatedBy string `json:"created_by"`
	CreatedAt int    `json:"created_at"`
	UpdatedAt int    `json:"updated_at"`
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
