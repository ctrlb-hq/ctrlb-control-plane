package frontendconfigV2

// ConfigUpsertRequest represents the structure for creating or updating a configuration
type ConfigUpsertRequest struct {
	Name        string `json:"name"`        // Configuration name
	Description string `json:"description"` // Brief description of the configuration
	Config      string `json:"config"`      // Configuration content (e.g., JSON or YAML)
	TargetAgent string `json:"targetAgent"` // Agent type the configuration targets
}
