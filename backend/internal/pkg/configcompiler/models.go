package configcompiler

type AgentType string

const (
	AgentTypeOTEL      AgentType = "otel"
	AgentTypeFluentBit AgentType = "fluent-bit"
	AgentTypeVector    AgentType = "vector"
)

// Pipeline represents an OTEL pipeline definition
type Pipeline struct {
	Receivers  []string `json:"receivers"`
	Processors []string `json:"processors,omitempty"`
	Exporters  []string `json:"exporters"`
}

// Pipelines is a map of pipeline names to pipeline definitions
type Pipelines map[string]Pipeline

// FBTagState tracks tag information through the pipeline
type FBTagState struct {
	Tag        string // The tag value
	SourceNode string // Node ID that emits this tag
}

// FBNodeInstance represents a potentially duplicated node instance
type FBNodeInstance struct {
	OriginalNodeID string
	InstanceIndex  int            // 0 for first/only instance, 1+ for duplicates
	Alias          string         // Unique alias for this instance
	Tag            string         // For inputs: the tag they emit
	MatchPattern   string         // For filters/outputs: the match pattern
	Config         map[string]any // The node config
	Role           string         // input, filter, or output
}