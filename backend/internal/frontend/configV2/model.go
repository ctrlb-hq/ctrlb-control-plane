package frontendconfigV2

// ConfigUpsertRequest represents the structure for creating or updating a configuration
type ConfigUpsertRequest struct {
	Name        string `json:"name"`        // Configuration name
	Description string `json:"description"` // Brief description of the configuration
	Config      string `json:"config"`      // Configuration content (e.g., JSON or YAML)
	TargetAgent string `json:"targetAgent"` // Agent type the configuration targets
}

// ConfigUpsertRequest represents the structure for creating or updating a configuration
type ConfigSetUpsertRequest struct {
	Name        string            `json:"name"`                  // Configuration name
	Credentials map[string]string `json:"credentials,omitempty"` // Stored as JSON
}

type CreatePipelinesRequest struct {
	Name string `json:"name"` // Pipeline name
	Type string `json:"type"` // Pipeline type
}

type CreatePipelineComponentRequest struct {
	Name          string         `json:"name"`          // Component name
	Type          string         `json:"type"`          // Component type
	Config        map[string]any `json:"config"`        // Component configuration
	ComponentType string         `json:"componentType"` // Component type
	PipelineID    int            `json:"pipelineID"`    // Pipeline ID
}
