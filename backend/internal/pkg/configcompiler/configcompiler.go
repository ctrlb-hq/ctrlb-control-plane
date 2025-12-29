package configcompiler

import (
	"fmt"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
)

// CompileGraph compiles a pipeline graph to the appropriate configuration format based on agent type
func CompileGraph(graph models.PipelineGraph, agentType AgentType) (*map[string]any, error) {
	switch agentType {
	case AgentTypeOTEL:
		return CompileGraphToOTEL(graph)
	case AgentTypeFluentBit:
		return CompileGraphToFluentBit(graph)
	default:
		return nil, fmt.Errorf("unsupported agent type: %s", agentType)
	}
}

func CompileGraphToOTEL(graph models.PipelineGraph) (*map[string]any, error) {
	return compileOTEL(graph)
}

func CompileGraphToFluentBit(graph models.PipelineGraph) (*map[string]any, error) {
	return compileFB(graph)
}
