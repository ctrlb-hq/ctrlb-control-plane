package configcompiler

import (
	"fmt"
	"encoding/json"
	"testing"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/stretchr/testify/assert"
)

// Helper function to create a sample pipeline graph
func createSampleGraph() models.PipelineGraph {
	return models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{
				ComponentID:      1,
				Name:             "receiver_one",
				ComponentName:    "otlp_receiver",
				ComponentRole:    "receiver",
				SupportedSignals: []string{"metrics", "logs"},
				Config: map[string]any{
					"endpoint": "0.0.0.0:4317",
				},
			},
			{
				ComponentID:      2,
				Name:             "processor_batch",
				ComponentName:    "batch_processor",
				ComponentRole:    "processor",
				SupportedSignals: []string{"metrics", "logs"},
				Config: map[string]any{
					"timeout": "10s",
				},
			},
			{
				ComponentID:      3,
				Name:             "exporter_otlp",
				ComponentName:    "otlp_grpc_exporter",
				ComponentRole:    "exporter",
				SupportedSignals: []string{"metrics", "logs"},
				Config: map[string]any{
					"endpoint": "example.com:4317",
				},
			},
		},
		Edges: []models.PipelineEdges{
			{Source: "1", Target: "2"},
			{Source: "2", Target: "3"},
		},
	}
}

func createSampleGraph2() models.PipelineGraph {
	return models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{
				ComponentID:      1,
				Name:             "receiver_one",
				ComponentName:    "otlp_receiver",
				ComponentRole:    "receiver",
				SupportedSignals: []string{"traces", "logs"},
				Config: map[string]any{
					"protocols": map[string]any{
						"grpc": map[string]any{
							"endpoint": "0.0.0.0:4317",
						},
					},
				},
			},
			{
				ComponentID:      2,
				Name:             "processor_batch",
				ComponentName:    "batch_processor",
				ComponentRole:    "processor",
				SupportedSignals: []string{"traces"},
				Config: map[string]any{
					"timeout": "10s",
				},
			},
			{
				ComponentID:      3,
				Name:             "exporter_otlp",
				ComponentName:    "otlp_grpc_exporter",
				ComponentRole:    "exporter",
				SupportedSignals: []string{"traces", "logs"},
				Config: map[string]any{
					"protocols": map[string]any{
						"grpc": map[string]any{
							"endpoint": "example.com:4317",
						},
					},
				},
			},
			{
				ComponentID:      4,
				Name:             "receiver_two",
				ComponentName:    "otlp_receiver",
				ComponentRole:    "receiver",
				SupportedSignals: []string{"logs"},
				Config: map[string]any{
					"protocols": map[string]any{
						"grpc": map[string]any{
							"endpoint": "0.0.0.0:4318",
						},
					},
				},
			},
			{
				ComponentID:      5,
				Name:             "processor_batch_two",
				ComponentName:    "batch_processor",
				ComponentRole:    "processor",
				SupportedSignals: []string{"logs"},
				Config: map[string]any{
					"timeout": "5s",
				},
			},
		},
		Edges: []models.PipelineEdges{
			{Source: "1", Target: "2"},
			{Source: "2", Target: "3"},
			{Source: "4", Target: "5"},
			{Source: "5", Target: "3"},
		},
	}
}

// Helper function to create a sample FluentBit pipeline graph
func createSampleFluentBitGraph() models.PipelineGraph {
	return models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{
				ComponentID:      1,
				Name:             "tail_input",
				ComponentName:    "tail",
				ComponentRole:    "input",
				SupportedSignals: []string{"logs"},
				Config: map[string]any{
					"path": "/var/log/*.log",
				},
			},
			{
				ComponentID:      2,
				Name:             "grep_filter",
				ComponentName:    "grep",
				ComponentRole:    "filter",
				SupportedSignals: []string{"logs"},
				Config: map[string]any{
					"regex": "log error",
				},
			},
			{
				ComponentID:      3,
				Name:             "stdout_output",
				ComponentName:    "stdout",
				ComponentRole:    "output",
				SupportedSignals: []string{"logs"},
				Config: map[string]any{
					"format": "json",
				},
			},
		},
		Edges: []models.PipelineEdges{
			{Source: "1", Target: "2"},
			{Source: "2", Target: "3"},
		},
	}
}

// =============================================================================
// CompileGraph Tests
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

// =============================================================================
// OTEL Compiler Tests
// =============================================================================

func TestCompileGraphToOTEL_Success(t *testing.T) {
	graph := createSampleGraph()

	result, err := CompileGraphToOTEL(graph)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, *result, "receivers")
	assert.Contains(t, *result, "processors")
	assert.Contains(t, *result, "exporters")
	assert.Contains(t, *result, "service")

	service := (*result)["service"].(map[string]any)
	assert.Contains(t, service, "pipelines")
	assert.Contains(t, service, "telemetry")

	// Verify pipelines were created for both signals (metrics and logs)
	pipelines := service["pipelines"].(map[string]any)
	assert.GreaterOrEqual(t, len(pipelines), 1, "Should have at least 1 pipeline")
}

func TestCompileGraphToOTEL_MultiPathGraph(t *testing.T) {
	graph := createSampleGraph2()

	result, err := CompileGraphToOTEL(graph)

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

	// The graph has two separate paths that share an exporter:
	// Path 1: R1 -> P1 -> E1 (traces)
	// Path 2: R2 -> P2 -> E1 (logs)
	// These should create separate pipelines
	pipelines := service["pipelines"].(map[string]any)
	assert.GreaterOrEqual(t, len(pipelines), 2, "Should have at least 2 pipelines for different paths")
}

func TestCompileGraphToOTEL_EmptyGraph(t *testing.T) {
	graph := models.PipelineGraph{}
	result, err := CompileGraphToOTEL(graph)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "empty pipeline graph")
}

func TestCompileGraphToOTEL_NoReceivers(t *testing.T) {
	graph := models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{
				ComponentID:      1,
				Name:             "exporter",
				ComponentName:    "otlp_exporter",
				ComponentRole:    "exporter",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{},
			},
		},
		Edges: []models.PipelineEdges{},
	}

	result, err := CompileGraphToOTEL(graph)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "at least one receiver")
}

func TestCompileGraphToOTEL_NoExporters(t *testing.T) {
	graph := models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{
				ComponentID:      1,
				Name:             "receiver",
				ComponentName:    "otlp_receiver",
				ComponentRole:    "receiver",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{},
			},
		},
		Edges: []models.PipelineEdges{},
	}

	result, err := CompileGraphToOTEL(graph)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "at least one exporter")
}

func TestCompileGraphToOTEL_WithCycle(t *testing.T) {
	graph := models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{
				ComponentID:      1,
				Name:             "receiver",
				ComponentName:    "otlp_receiver",
				ComponentRole:    "receiver",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{},
			},
			{
				ComponentID:      2,
				Name:             "processor",
				ComponentName:    "batch_processor",
				ComponentRole:    "processor",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{},
			},
			{
				ComponentID:      3,
				Name:             "exporter",
				ComponentName:    "otlp_exporter",
				ComponentRole:    "exporter",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{},
			},
		},
		Edges: []models.PipelineEdges{
			{Source: "1", Target: "2"},
			{Source: "2", Target: "3"},
			{Source: "3", Target: "1"}, // Cycle
		},
	}

	result, err := CompileGraphToOTEL(graph)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "cycle detected")
}

func TestCompileGraphToOTEL_FanIn(t *testing.T) {
	// Two receivers feeding into one processor -> one exporter
	graph := models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{
				ComponentID:      1,
				Name:             "receiver_one",
				ComponentName:    "otlp_receiver",
				ComponentRole:    "receiver",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{"endpoint": "0.0.0.0:4317"},
			},
			{
				ComponentID:      2,
				Name:             "receiver_two",
				ComponentName:    "otlp_receiver",
				ComponentRole:    "receiver",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{"endpoint": "0.0.0.0:4318"},
			},
			{
				ComponentID:      3,
				Name:             "processor",
				ComponentName:    "batch_processor",
				ComponentRole:    "processor",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{},
			},
			{
				ComponentID:      4,
				Name:             "exporter",
				ComponentName:    "otlp_exporter",
				ComponentRole:    "exporter",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{},
			},
		},
		Edges: []models.PipelineEdges{
			{Source: "1", Target: "3"},
			{Source: "2", Target: "3"},
			{Source: "3", Target: "4"},
		},
	}

	result, err := CompileGraphToOTEL(graph)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	service := (*result)["service"].(map[string]any)
	pipelines := service["pipelines"].(map[string]any)

	// Should have one pipeline with multiple receivers
	assert.GreaterOrEqual(t, len(pipelines), 1)

	// Find a logs pipeline and verify it has 2 receivers
	for name, p := range pipelines {
		if name[:4] == "logs" {
			pipeline := p.(map[string]any)
			receivers := pipeline["receivers"].([]string)
			assert.Len(t, receivers, 2, "Fan-in pipeline should have 2 receivers")
		}
	}
}

func TestCompileGraphToOTEL_FanOut(t *testing.T) {
	// One receiver -> one processor -> two exporters
	graph := models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{
				ComponentID:      1,
				Name:             "receiver",
				ComponentName:    "otlp_receiver",
				ComponentRole:    "receiver",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{},
			},
			{
				ComponentID:      2,
				Name:             "processor",
				ComponentName:    "batch_processor",
				ComponentRole:    "processor",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{},
			},
			{
				ComponentID:      3,
				Name:             "exporter_one",
				ComponentName:    "otlp_exporter",
				ComponentRole:    "exporter",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{"endpoint": "backend1:4317"},
			},
			{
				ComponentID:      4,
				Name:             "exporter_two",
				ComponentName:    "otlp_exporter",
				ComponentRole:    "exporter",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{"endpoint": "backend2:4317"},
			},
		},
		Edges: []models.PipelineEdges{
			{Source: "1", Target: "2"},
			{Source: "2", Target: "3"},
			{Source: "2", Target: "4"},
		},
	}

	result, err := CompileGraphToOTEL(graph)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	service := (*result)["service"].(map[string]any)
	pipelines := service["pipelines"].(map[string]any)
	assert.GreaterOrEqual(t, len(pipelines), 1)

	// Find a logs pipeline and verify it has 2 exporters
	for name, p := range pipelines {
		if name[:4] == "logs" {
			pipeline := p.(map[string]any)
			exporters := pipeline["exporters"].([]string)
			assert.Len(t, exporters, 2, "Fan-out pipeline should have 2 exporters")
		}
	}
}

func TestCompileGraphToOTEL_DisconnectedComponents(t *testing.T) {
	// Two completely separate pipelines
	graph := models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			// Pipeline 1
			{
				ComponentID:      1,
				Name:             "receiver_one",
				ComponentName:    "otlp_receiver",
				ComponentRole:    "receiver",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{},
			},
			{
				ComponentID:      2,
				Name:             "exporter_one",
				ComponentName:    "otlp_exporter",
				ComponentRole:    "exporter",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{},
			},
			// Pipeline 2
			{
				ComponentID:      3,
				Name:             "receiver_two",
				ComponentName:    "otlp_receiver",
				ComponentRole:    "receiver",
				SupportedSignals: []string{"metrics"},
				Config:           map[string]any{},
			},
			{
				ComponentID:      4,
				Name:             "exporter_two",
				ComponentName:    "otlp_exporter",
				ComponentRole:    "exporter",
				SupportedSignals: []string{"metrics"},
				Config:           map[string]any{},
			},
		},
		Edges: []models.PipelineEdges{
			{Source: "1", Target: "2"},
			{Source: "3", Target: "4"},
		},
	}

	result, err := CompileGraphToOTEL(graph)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	service := (*result)["service"].(map[string]any)
	pipelines := service["pipelines"].(map[string]any)

	// Should have 2 separate pipelines
	assert.GreaterOrEqual(t, len(pipelines), 2, "Should have at least 2 pipelines for disconnected components")
}

// =============================================================================
// Fluent Bit Compiler Tests
// =============================================================================

func TestCompileGraphToFluentBit_Success(t *testing.T) {
	graph := createSampleFluentBitGraph()

	result, err := CompileGraphToFluentBit(graph)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify service section
	assert.Contains(t, *result, "service")
	service := (*result)["service"].(map[string]any)
	assert.Equal(t, 1, service["flush"])
	assert.Equal(t, "info", service["log_level"])
	assert.Equal(t, "on", service["http_server"])
	assert.Equal(t, "on", service["hot_reload"])

	// Verify pipeline section
	assert.Contains(t, *result, "pipeline")
	pipeline := (*result)["pipeline"].(map[string]any)
	assert.Contains(t, pipeline, "inputs")
	assert.Contains(t, pipeline, "filters")
	assert.Contains(t, pipeline, "outputs")

	// Verify inputs have required fields
	inputs := pipeline["inputs"].([]any)
	assert.NotEmpty(t, inputs)
	firstInput := inputs[0].(map[string]any)
	assert.Contains(t, firstInput, "name")
	assert.Contains(t, firstInput, "tag")
	assert.Contains(t, firstInput, "alias")

	// Verify outputs have required fields (skip prometheus output)
	outputs := pipeline["outputs"].([]any)
	assert.NotEmpty(t, outputs)
	firstOutput := outputs[0].(map[string]any)
	assert.Contains(t, firstOutput, "name")
	assert.Contains(t, firstOutput, "match")
}

func TestCompileGraphToFluentBit_EmptyGraph(t *testing.T) {
	graph := models.PipelineGraph{}

	result, err := CompileGraphToFluentBit(graph)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "empty pipeline graph")
}

func TestCompileGraphToFluentBit_MissingInput(t *testing.T) {
	graph := models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{
				ComponentID:      1,
				Name:             "output",
				ComponentName:    "stdout",
				ComponentRole:    "output",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{},
			},
		},
		Edges: []models.PipelineEdges{},
	}

	result, err := CompileGraphToFluentBit(graph)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "at least one input")
}

func TestCompileGraphToFluentBit_MissingOutput(t *testing.T) {
	graph := models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{
				ComponentID:      1,
				Name:             "input",
				ComponentName:    "tail",
				ComponentRole:    "input",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{},
			},
		},
		Edges: []models.PipelineEdges{},
	}

	result, err := CompileGraphToFluentBit(graph)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "at least one output")
}

func TestCompileGraphToFluentBit_WithFilters(t *testing.T) {
	graph := models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{
				ComponentID:      1,
				Name:             "input",
				ComponentName:    "tail",
				ComponentRole:    "input",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{"path": "/var/log/*.log"},
			},
			{
				ComponentID:      2,
				Name:             "filter",
				ComponentName:    "grep",
				ComponentRole:    "filter",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{"regex": "error"},
			},
			{
				ComponentID:      3,
				Name:             "output",
				ComponentName:    "stdout",
				ComponentRole:    "output",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{},
			},
		},
		Edges: []models.PipelineEdges{
			{Source: "1", Target: "2"},
			{Source: "2", Target: "3"},
		},
	}

	result, err := CompileGraphToFluentBit(graph)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	pipeline := (*result)["pipeline"].(map[string]any)
	filters := pipeline["filters"].([]any)

	assert.Len(t, filters, 1)

	// Verify match was added to filter
	filter := filters[0].(map[string]any)
	assert.Contains(t, filter, "match")
	assert.NotEmpty(t, filter["match"])
}

func TestCompileGraphToFluentBit_WithCycle(t *testing.T) {
	graph := models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{
				ComponentID:      1,
				Name:             "input",
				ComponentName:    "tail",
				ComponentRole:    "input",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{},
			},
			{
				ComponentID:      2,
				Name:             "filter1",
				ComponentName:    "grep",
				ComponentRole:    "filter",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{},
			},
			{
				ComponentID:      3,
				Name:             "filter2",
				ComponentName:    "modify",
				ComponentRole:    "filter",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{},
			},
			{
				ComponentID:      4,
				Name:             "output",
				ComponentName:    "stdout",
				ComponentRole:    "output",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{},
			},
		},
		Edges: []models.PipelineEdges{
			{Source: "1", Target: "2"},
			{Source: "2", Target: "3"},
			{Source: "3", Target: "2"}, // Cycle
			{Source: "3", Target: "4"},
		},
	}

	result, err := CompileGraphToFluentBit(graph)

	// New implementation detects cycles
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "cycle detected")
}

func TestCompileGraphToFluentBit_MultipleInputsSameOutput_Incompatible(t *testing.T) {
	graph := models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{
				ComponentID:      1,
				Name:             "tail_input",
				ComponentName:    "tail",
				ComponentRole:    "input",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{"path": "/var/log/app.log"},
			},
			{
				ComponentID:      2,
				Name:             "stdin_input",
				ComponentName:    "stdin",
				ComponentRole:    "input",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{},
			},
			{
				ComponentID:      3,
				Name:             "grep_filter",
				ComponentName:    "grep",
				ComponentRole:    "filter",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{"regex": "error"},
			},
			{
				ComponentID:      4,
				Name:             "stdout_output",
				ComponentName:    "stdout",
				ComponentRole:    "output",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{"format": "json"},
			},
		},
		Edges: []models.PipelineEdges{
			{Source: "1", Target: "3"},
			{Source: "2", Target: "3"},
			{Source: "3", Target: "4"},
		},
	}

	result, err := CompileGraphToFluentBit(graph)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	pipeline := (*result)["pipeline"].(map[string]any)
	inputs := pipeline["inputs"].([]any)
	filters := pipeline["filters"].([]any)
	outputs := pipeline["outputs"].([]any)

	// Should have 2 user inputs + 1 prometheus input = 3
	assert.Len(t, inputs, 3, "Should have 2 user inputs + 1 prometheus input")

	// Since inputs have different plugin prefixes (tail vs stdin),
	// filter and output should be expanded into 2 instances each
	assert.Len(t, filters, 2, "Should have 2 filter instances for incompatible inputs")

	// 2 user output instances + 1 prometheus output = 3
	assert.Len(t, outputs, 3, "Should have 2 user output instances + 1 prometheus output")

	// Verify each filter has a specific match pattern (not wildcard)
	for _, f := range filters {
		filterMap := f.(map[string]any)
		match := filterMap["match"].(string)
		assert.NotEqual(t, "*", match, "Filter should not use wildcard")
	}

	// Verify user outputs have specific match patterns
	userOutputCount := 0
	for _, o := range outputs {
		outputMap := o.(map[string]any)
		if outputMap["name"] == "stdout" {
			userOutputCount++
			match := outputMap["match"].(string)
			assert.NotEqual(t, "*", match, "User output should not use wildcard")
		}
	}
	assert.Equal(t, 2, userOutputCount, "Should have 2 stdout output instances")
}

func TestCompileGraphToFluentBit_MultipleInputsSameType(t *testing.T) {
	graph := models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{
				ComponentID:      1,
				Name:             "tail_app",
				ComponentName:    "tail",
				ComponentRole:    "input",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{"path": "/var/log/app.log"},
			},
			{
				ComponentID:      2,
				Name:             "tail_system",
				ComponentName:    "tail",
				ComponentRole:    "input",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{"path": "/var/log/syslog"},
			},
			{
				ComponentID:      3,
				Name:             "stdout_output",
				ComponentName:    "stdout",
				ComponentRole:    "output",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{},
			},
		},
		Edges: []models.PipelineEdges{
			{Source: "1", Target: "3"},
			{Source: "2", Target: "3"},
		},
	}

	result, err := CompileGraphToFluentBit(graph)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	pipeline := (*result)["pipeline"].(map[string]any)
	inputs := pipeline["inputs"].([]any)
	outputs := pipeline["outputs"].([]any)

	// 2 user inputs + 1 prometheus input
	assert.GreaterOrEqual(t, len(inputs), 2)

	// Since both inputs are "tail" (same plugin), output should NOT be expanded
	// Output count = 1 user output + 1 prometheus output
	assert.Len(t, outputs, 2, "Should have 1 user output + 1 prometheus output")

	// Verify output match pattern starts with "tail"
	firstOutput := outputs[0].(map[string]any)
	matchPattern := firstOutput["match"].(string)
	assert.True(t, len(matchPattern) >= 4 && matchPattern[0:4] == "tail",
		"Expected match pattern to start with 'tail', got: %s", matchPattern)
}

func TestCompileGraphToFluentBit_DirectInputToOutput(t *testing.T) {
	graph := models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{
				ComponentID:      1,
				Name:             "tail_input",
				ComponentName:    "tail",
				ComponentRole:    "input",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{"path": "/var/log/*.log"},
			},
			{
				ComponentID:      2,
				Name:             "stdout_output",
				ComponentName:    "stdout",
				ComponentRole:    "output",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{"format": "json"},
			},
		},
		Edges: []models.PipelineEdges{
			{Source: "1", Target: "2"},
		},
	}

	result, err := CompileGraphToFluentBit(graph)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	pipeline := (*result)["pipeline"].(map[string]any)
	filters := pipeline["filters"].([]any)

	// No filters in this graph
	assert.Empty(t, filters)
}

func TestCompileGraphToFluentBit_UserProvidedTag(t *testing.T) {
	graph := models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{
				ComponentID:      1,
				Name:             "tail_input",
				ComponentName:    "tail",
				ComponentRole:    "input",
				SupportedSignals: []string{"logs"},
				Config: map[string]any{
					"path": "/var/log/*.log",
					"tag":  "custom.mytag",
				},
			},
			{
				ComponentID:      2,
				Name:             "stdout_output",
				ComponentName:    "stdout",
				ComponentRole:    "output",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{},
			},
		},
		Edges: []models.PipelineEdges{
			{Source: "1", Target: "2"},
		},
	}

	result, err := CompileGraphToFluentBit(graph)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	pipeline := (*result)["pipeline"].(map[string]any)
	inputs := pipeline["inputs"].([]any)

	// Find the user input (not prometheus)
	var userInput map[string]any
	for _, inp := range inputs {
		input := inp.(map[string]any)
		if input["name"] == "tail" {
			userInput = input
			break
		}
	}

	assert.NotNil(t, userInput)
	assert.Equal(t, "custom.mytag", userInput["tag"], "Should preserve user-provided tag")
}

// =============================================================================
// Graph Utility Tests
// =============================================================================

func TestValidateGraph(t *testing.T) {
	t.Run("empty graph", func(t *testing.T) {
		graph := models.PipelineGraph{}
		err := ValidateGraph(graph)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty pipeline graph")
	})

	t.Run("invalid edge source", func(t *testing.T) {
		graph := models.PipelineGraph{
			Nodes: []models.PipelineNodes{
				{ComponentID: 1, Name: "node1"},
			},
			Edges: []models.PipelineEdges{
				{Source: "999", Target: "1"},
			},
		}
		err := ValidateGraph(graph)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "non-existent source")
	})

	t.Run("invalid edge target", func(t *testing.T) {
		graph := models.PipelineGraph{
			Nodes: []models.PipelineNodes{
				{ComponentID: 1, Name: "node1"},
			},
			Edges: []models.PipelineEdges{
				{Source: "1", Target: "999"},
			},
		}
		err := ValidateGraph(graph)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "non-existent target")
	})

	t.Run("valid graph", func(t *testing.T) {
		graph := createSampleGraph()
		err := ValidateGraph(graph)
		assert.NoError(t, err)
	})
}

func TestBuildGraphState(t *testing.T) {
	graph := createSampleGraph()
	state := BuildGraphState(graph)

	assert.Len(t, state.NodesByID, 3)
	assert.Contains(t, state.OutEdges, "1")
	assert.Contains(t, state.OutEdges["1"], "2")
	assert.Contains(t, state.InEdges, "2")
	assert.Contains(t, state.InEdges["2"], "1")
}

func TestDetectCycle(t *testing.T) {
	t.Run("no cycle - linear", func(t *testing.T) {
		graph := models.PipelineGraph{
			Nodes: []models.PipelineNodes{
				{ComponentID: 1, Name: "a"},
				{ComponentID: 2, Name: "b"},
				{ComponentID: 3, Name: "c"},
			},
			Edges: []models.PipelineEdges{
				{Source: "1", Target: "2"},
				{Source: "2", Target: "3"},
			},
		}
		state := BuildGraphState(graph)
		cycle := DetectCycle(state)
		assert.Nil(t, cycle)
	})

	t.Run("no cycle - diamond", func(t *testing.T) {
		graph := models.PipelineGraph{
			Nodes: []models.PipelineNodes{
				{ComponentID: 1, Name: "a"},
				{ComponentID: 2, Name: "b"},
				{ComponentID: 3, Name: "c"},
				{ComponentID: 4, Name: "d"},
			},
			Edges: []models.PipelineEdges{
				{Source: "1", Target: "2"},
				{Source: "1", Target: "3"},
				{Source: "2", Target: "4"},
				{Source: "3", Target: "4"},
			},
		}
		state := BuildGraphState(graph)
		cycle := DetectCycle(state)
		assert.Nil(t, cycle)
	})

	t.Run("self loop", func(t *testing.T) {
		graph := models.PipelineGraph{
			Nodes: []models.PipelineNodes{
				{ComponentID: 1, Name: "a"},
			},
			Edges: []models.PipelineEdges{
				{Source: "1", Target: "1"},
			},
		}
		state := BuildGraphState(graph)
		cycle := DetectCycle(state)
		assert.NotNil(t, cycle)
	})

	t.Run("simple cycle", func(t *testing.T) {
		graph := models.PipelineGraph{
			Nodes: []models.PipelineNodes{
				{ComponentID: 1, Name: "a"},
				{ComponentID: 2, Name: "b"},
				{ComponentID: 3, Name: "c"},
			},
			Edges: []models.PipelineEdges{
				{Source: "1", Target: "2"},
				{Source: "2", Target: "3"},
				{Source: "3", Target: "1"},
			},
		}
		state := BuildGraphState(graph)
		cycle := DetectCycle(state)
		assert.NotNil(t, cycle)
		assert.GreaterOrEqual(t, len(cycle), 2)
	})
}

func TestTopologicalSort(t *testing.T) {
	t.Run("linear graph", func(t *testing.T) {
		graph := models.PipelineGraph{
			Nodes: []models.PipelineNodes{
				{ComponentID: 1, Name: "a"},
				{ComponentID: 2, Name: "b"},
				{ComponentID: 3, Name: "c"},
			},
			Edges: []models.PipelineEdges{
				{Source: "1", Target: "2"},
				{Source: "2", Target: "3"},
			},
		}
		state := BuildGraphState(graph)

		sorted, err := TopologicalSort([]string{"1", "2", "3"}, state)
		assert.NoError(t, err)
		assert.Equal(t, []string{"1", "2", "3"}, sorted)
	})

	t.Run("diamond graph", func(t *testing.T) {
		graph := models.PipelineGraph{
			Nodes: []models.PipelineNodes{
				{ComponentID: 1, Name: "a"},
				{ComponentID: 2, Name: "b"},
				{ComponentID: 3, Name: "c"},
				{ComponentID: 4, Name: "d"},
			},
			Edges: []models.PipelineEdges{
				{Source: "1", Target: "2"},
				{Source: "1", Target: "3"},
				{Source: "2", Target: "4"},
				{Source: "3", Target: "4"},
			},
		}
		state := BuildGraphState(graph)

		sorted, err := TopologicalSort([]string{"1", "2", "3", "4"}, state)
		assert.NoError(t, err)

		// 1 must come first, 4 must come last
		assert.Equal(t, "1", sorted[0])
		assert.Equal(t, "4", sorted[3])
	})

	t.Run("subset of nodes", func(t *testing.T) {
		graph := models.PipelineGraph{
			Nodes: []models.PipelineNodes{
				{ComponentID: 1, Name: "a"},
				{ComponentID: 2, Name: "b"},
				{ComponentID: 3, Name: "c"},
			},
			Edges: []models.PipelineEdges{
				{Source: "1", Target: "2"},
				{Source: "2", Target: "3"},
			},
		}
		state := BuildGraphState(graph)

		// Only sort nodes 2 and 3
		sorted, err := TopologicalSort([]string{"2", "3"}, state)
		assert.NoError(t, err)
		assert.Equal(t, []string{"2", "3"}, sorted)
	})
}

func TestIntersectSignals(t *testing.T) {
	t.Run("common signals", func(t *testing.T) {
		nodes := []models.PipelineNodes{
			{SupportedSignals: []string{"logs", "metrics", "traces"}},
			{SupportedSignals: []string{"logs", "metrics"}},
			{SupportedSignals: []string{"logs"}},
		}

		result := IntersectSignals(nodes)
		assert.Equal(t, []string{"logs"}, result)
	})

	t.Run("no common signals", func(t *testing.T) {
		nodes := []models.PipelineNodes{
			{SupportedSignals: []string{"logs"}},
			{SupportedSignals: []string{"metrics"}},
		}

		result := IntersectSignals(nodes)
		assert.Empty(t, result)
	})

	t.Run("identical signals", func(t *testing.T) {
		nodes := []models.PipelineNodes{
			{SupportedSignals: []string{"logs", "metrics"}},
			{SupportedSignals: []string{"logs", "metrics"}},
		}

		result := IntersectSignals(nodes)
		assert.Len(t, result, 2)
		assert.Contains(t, result, "logs")
		assert.Contains(t, result, "metrics")
	})

	t.Run("empty nodes", func(t *testing.T) {
		result := IntersectSignals([]models.PipelineNodes{})
		assert.Nil(t, result)
	})

	t.Run("single node", func(t *testing.T) {
		nodes := []models.PipelineNodes{
			{SupportedSignals: []string{"logs", "metrics"}},
		}

		result := IntersectSignals(nodes)
		assert.Len(t, result, 2)
	})
}

func TestGetNodesByRole(t *testing.T) {
	graph := models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{ComponentID: 1, ComponentRole: "receiver"},
			{ComponentID: 2, ComponentRole: "processor"},
			{ComponentID: 3, ComponentRole: "exporter"},
			{ComponentID: 4, ComponentRole: "input"},
			{ComponentID: 5, ComponentRole: "filter"},
			{ComponentID: 6, ComponentRole: "output"},
		},
		Edges: []models.PipelineEdges{},
	}

	state := BuildGraphState(graph)
	receivers, processors, exporters, inputs, filters, outputs := GetNodesByRole(state)

	assert.Len(t, receivers, 1)
	assert.Len(t, processors, 1)
	assert.Len(t, exporters, 1)
	assert.Len(t, inputs, 1)
	assert.Len(t, filters, 1)
	assert.Len(t, outputs, 1)
}

// =============================================================================
// Alias Generation Tests
// =============================================================================

func TestGenerateOTELAlias(t *testing.T) {
	node := models.PipelineNodes{
		Name:          "my_receiver",
		ComponentName: "otlp_receiver",
		Config:        map[string]any{"endpoint": "0.0.0.0:4317"},
	}

	alias := GenerateOTELAlias(node)

	// Verify the alias has the expected format: type/name_hash
	assert.Contains(t, alias, "otlp/")
	assert.Contains(t, alias, "my_receiver") // The actual name format
	assert.Contains(t, alias, "_")           // Should have underscore before hash
}

func TestGenerateFBAlias(t *testing.T) {
	node := models.PipelineNodes{
		Name:          "my_input",
		ComponentName: "tail",
		Config:        map[string]any{"path": "/var/log/*.log"},
	}

	alias := GenerateFBAlias(node)

	// Verify the alias has the expected format: name_hash
	assert.Contains(t, alias, "my_input") // The actual name format
	assert.NotContains(t, alias, "/")     // FB aliases don't have type prefix
	// Should have format: name_hash
	assert.Regexp(t, `^my_input_[a-f0-9]+$`, alias)
}

func TestGenerateFBTag(t *testing.T) {
	node := models.PipelineNodes{
		ComponentName: "tail",
	}

	tag := GenerateFBTag(node, "myInput_abc123")

	assert.Equal(t, "tail.myInput_abc123", tag)
}

func TestExtractTagPrefix(t *testing.T) {
	tests := []struct {
		tag      string
		expected string
	}{
		{"tail.myInput_abc", "tail"},
		{"stdin.input", "stdin"},
		{"notag", "notag"},
		{"a.b.c", "a"},
	}

	for _, tt := range tests {
		result := ExtractTagPrefix(tt.tag)
		assert.Equal(t, tt.expected, result)
	}
}

// =============================================================================
// FB Match Pattern Tests
// =============================================================================

func TestComputeFBMatchPattern(t *testing.T) {
	t.Run("no tags", func(t *testing.T) {
		compatible, pattern := computeFBMatchPattern([]FBTagState{})
		assert.True(t, compatible)
		assert.Equal(t, "*", pattern)
	})

	t.Run("single tag", func(t *testing.T) {
		tags := []FBTagState{{Tag: "tail.myInput_abc123"}}
		compatible, pattern := computeFBMatchPattern(tags)
		assert.True(t, compatible)
		assert.Equal(t, "tail.myInput_abc123*", pattern)
	})

	t.Run("multiple tags same plugin with common prefix", func(t *testing.T) {
		tags := []FBTagState{
			{Tag: "tail.input1_abc"},
			{Tag: "tail.input2_def"},
		}
		compatible, pattern := computeFBMatchPattern(tags)
		assert.True(t, compatible)
		assert.True(t, pattern == "tail.input*" || pattern == "tail*")
	})

	t.Run("multiple tags different plugins", func(t *testing.T) {
		tags := []FBTagState{
			{Tag: "tail.input1_abc"},
			{Tag: "stdin.input2_def"},
		}
		compatible, pattern := computeFBMatchPattern(tags)
		assert.False(t, compatible)
		assert.Equal(t, "", pattern)
	})
}