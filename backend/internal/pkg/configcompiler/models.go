package configcompiler

import "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"

type Pipeline struct {
	Receivers  []string `json:"receivers"`
	Processors []string `json:"processors"`
	Exporters  []string `json:"exporters"`
}

type Pipelines map[string]Pipeline


// fluentBitNode represents a node with computed routing information
type FluentBitNode struct {
	Node         models.PipelineNodes
	Alias        string
	Tag          string   // For inputs: the tag they emit
	MatchTags    []string // For filters/outputs: tags they should match
	MatchPattern string   // Computed match pattern
}