package models

// Struct for pipeline component (Node)
type PipelineComponent struct {
	ComponentID   int    `json:"component_id"`
	Name          string `json:"name"`
	ComponentRole string `json:"component_role"`
	PluginName    string `json:"plugin_name"`
	Config        string `json:"config"`
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
