package configcompiler

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// CompileGraph Tests - Main Entry Point
// =============================================================================

func TestCompileGraph_OTEL(t *testing.T) {
	graph := createSampleGraph()

	result, err := CompileGraph(graph, AgentTypeOTEL)

	b, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(b))

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, *result, "receivers")
	assert.Contains(t, *result, "processors")
	assert.Contains(t, *result, "exporters")
	assert.Contains(t, *result, "service")

	service := (*result)["service"].(map[string]any)
	assert.Contains(t, service, "pipelines")
	assert.Contains(t, service, "telemetry")
}

func TestCompileGraph_FluentBit(t *testing.T) {
	graph := createSampleFluentBitGraph()

	result, err := CompileGraph(graph, AgentTypeFluentBit)

	b, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(b))

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, *result, "service")
	assert.Contains(t, *result, "pipeline")

	pipeline := (*result)["pipeline"].(map[string]any)
	assert.Contains(t, pipeline, "inputs")
	assert.Contains(t, pipeline, "filters")
	assert.Contains(t, pipeline, "outputs")

	inputs := pipeline["inputs"].([]any)
	outputs := pipeline["outputs"].([]any)
	assert.NotEmpty(t, inputs)
	assert.NotEmpty(t, outputs)
}

func TestCompileGraph_UnsupportedAgentType(t *testing.T) {
	graph := createSampleGraph()

	result, err := CompileGraph(graph, AgentType("unknown"))

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unsupported agent type")
}
