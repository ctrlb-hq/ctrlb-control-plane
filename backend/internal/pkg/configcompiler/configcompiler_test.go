package configcompiler

import (
	"testing"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// Test CompileGraph with OTEL agent type
func TestCompileGraph_OTEL(t *testing.T) {
	graph := createSampleGraph()

	result, err := CompileGraph(graph, AgentTypeOTEL)

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

// Test CompileGraph with FluentBit agent type
func TestCompileGraph_FluentBit(t *testing.T) {
	graph := createSampleGraph()

	result, err := CompileGraph(graph, AgentTypeFluentBit)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, *result, "service")
	assert.Contains(t, *result, "pipeline")

	pipeline := (*result)["pipeline"].(map[string]any)
	assert.Contains(t, pipeline, "inputs")
	assert.Contains(t, pipeline, "filters")
	assert.Contains(t, pipeline, "outputs")

	// Verify inputs/outputs exist
	inputs := pipeline["inputs"].([]any)
	outputs := pipeline["outputs"].([]any)
	assert.NotEmpty(t, inputs)
	assert.NotEmpty(t, outputs)
}

// Test CompileGraph with unsupported agent type
func TestCompileGraph_UnsupportedAgentType(t *testing.T) {
	graph := createSampleGraph()

	result, err := CompileGraph(graph, AgentType("unknown"))

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unsupported agent type")
}

// Test CompileGraphToJSON_Success (existing test)
func TestCompileGraphToJSON_Success(t *testing.T) {
	graph := createSampleGraph()

	result, err := CompileGraphToJSON(graph)
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

// Test CompileGraphToJSON_EmptyGraph (existing test)
func TestCompileGraphToJSON_EmptyGraph(t *testing.T) {
	graph := models.PipelineGraph{}
	result, err := CompileGraphToJSON(graph)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "empty pipeline graph")
}

// Test CompileGraphToFluentBit_Success
func TestCompileGraphToFluentBit_Success(t *testing.T) {
	graph := createSampleGraph()

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

	// Verify outputs have required fields
	outputs := pipeline["outputs"].([]any)
	assert.NotEmpty(t, outputs)
	firstOutput := outputs[0].(map[string]any)
	assert.Contains(t, firstOutput, "name")
	assert.Contains(t, firstOutput, "match")
}

// Test CompileGraphToFluentBit_EmptyGraph
func TestCompileGraphToFluentBit_EmptyGraph(t *testing.T) {
	graph := models.PipelineGraph{}

	result, err := CompileGraphToFluentBit(graph)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "empty pipeline graph")
}

// Test analyzeGraphStructure with valid graph
func TestAnalyzeGraphStructure_Success(t *testing.T) {
	graph := createSampleGraph()

	components, err := analyzeGraphStructure(graph, "Test")

	assert.NoError(t, err)
	assert.NotNil(t, components)
	assert.Len(t, components, 1)    // All nodes are connected in one component
	assert.Len(t, components[0], 3) // 3 nodes in the component
}

// Test analyzeGraphStructure with empty graph
func TestAnalyzeGraphStructure_EmptyGraph(t *testing.T) {
	graph := models.PipelineGraph{}

	components, err := analyzeGraphStructure(graph, "Test")

	assert.Error(t, err)
	assert.Nil(t, components)
	assert.Contains(t, err.Error(), "empty pipeline graph")
}

// Test analyzeGraphStructure with invalid edge source
func TestAnalyzeGraphStructure_InvalidEdgeSource(t *testing.T) {
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
		Edges: []models.PipelineEdges{
			{Source: "999", Target: "1"}, // Invalid source
		},
	}

	components, err := analyzeGraphStructure(graph, "Test")

	assert.Error(t, err)
	assert.Nil(t, components)
	assert.Contains(t, err.Error(), "non-existent source node")
}

// Test analyzeGraphStructure with invalid edge target
func TestAnalyzeGraphStructure_InvalidEdgeTarget(t *testing.T) {
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
		Edges: []models.PipelineEdges{
			{Source: "1", Target: "999"}, // Invalid target
		},
	}

	components, err := analyzeGraphStructure(graph, "Test")

	assert.Error(t, err)
	assert.Nil(t, components)
	assert.Contains(t, err.Error(), "non-existent target node")
}

// Test analyzeGraphStructure with multiple disconnected components
func TestAnalyzeGraphStructure_MultipleComponents(t *testing.T) {
	graph := models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			// Component 1
			{ComponentID: 1, Name: "r1", ComponentName: "receiver1", ComponentRole: "receiver", SupportedSignals: []string{"logs"}, Config: map[string]any{}},
			{ComponentID: 2, Name: "e1", ComponentName: "exporter1", ComponentRole: "exporter", SupportedSignals: []string{"logs"}, Config: map[string]any{}},
			// Component 2
			{ComponentID: 3, Name: "r2", ComponentName: "receiver2", ComponentRole: "receiver", SupportedSignals: []string{"logs"}, Config: map[string]any{}},
			{ComponentID: 4, Name: "e2", ComponentName: "exporter2", ComponentRole: "exporter", SupportedSignals: []string{"logs"}, Config: map[string]any{}},
		},
		Edges: []models.PipelineEdges{
			{Source: "1", Target: "2"}, // Component 1
			{Source: "3", Target: "4"}, // Component 2
		},
	}

	components, err := analyzeGraphStructure(graph, "Test")

	assert.NoError(t, err)
	assert.Len(t, components, 2) // Two separate components
}

// Test buildOTELConfig with valid components
func TestBuildOTELConfig_Success(t *testing.T) {
	components := [][]models.PipelineNodes{
		{
			{
				ComponentID:      1,
				Name:             "receiver_one",
				ComponentName:    "otlp_receiver",
				ComponentRole:    "receiver",
				SupportedSignals: []string{"metrics", "logs"},
				Config:           map[string]any{"endpoint": "0.0.0.0:4317"},
			},
			{
				ComponentID:      2,
				Name:             "exporter_one",
				ComponentName:    "otlp_grpc_exporter",
				ComponentRole:    "exporter",
				SupportedSignals: []string{"metrics", "logs"},
				Config:           map[string]any{"endpoint": "example.com:4317"},
			},
		},
	}

	receivers, processors, exporters, pipelines, err := buildOTELConfig(components)

	assert.NoError(t, err)
	assert.NotNil(t, receivers)
	assert.NotNil(t, processors)
	assert.NotNil(t, exporters)
	assert.NotNil(t, pipelines)
	assert.NotEmpty(t, receivers)
	assert.NotEmpty(t, exporters)
	assert.NotEmpty(t, pipelines)

	// Verify pipelines were created for both signals
	assert.Len(t, pipelines, 2) // metrics/pipeline_1 and logs/pipeline_1
}

// Test buildOTELConfig with no supported signals
func TestBuildOTELConfig_NoSupportedSignals(t *testing.T) {
	components := [][]models.PipelineNodes{
		{
			{
				ComponentID:      1,
				Name:             "receiver",
				ComponentName:    "receiver",
				ComponentRole:    "receiver",
				SupportedSignals: []string{}, // No signals
				Config:           map[string]any{},
			},
		},
	}

	_, _, _, _, err := buildOTELConfig(components)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no supported signals")
}

// Test buildFluentBitConfig with valid components
func TestBuildFluentBitConfig_Success(t *testing.T) {
	components := [][]models.PipelineNodes{
		{
			{
				ComponentID:      1,
				Name:             "tail_input",
				ComponentName:    "tail",
				ComponentRole:    "receiver",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{"path": "/var/log/*.log"},
			},
			{
				ComponentID:      2,
				Name:             "stdout_output",
				ComponentName:    "stdout",
				ComponentRole:    "exporter",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{"format": "json"},
			},
		},
	}

	inputs, filters, outputs, err := buildFluentBitConfig(components)

	assert.NoError(t, err)
	assert.NotEmpty(t, inputs)
	assert.NotEmpty(t, outputs)
	assert.Empty(t, filters) // No filters in this test
	assert.Len(t, inputs, 1)
	assert.Len(t, outputs, 1)

	// Verify tag was added to input
	input := inputs[0].(map[string]any)
	assert.Contains(t, input, "tag")
	assert.Contains(t, input, "name")
	assert.Contains(t, input, "alias")

	// Verify match was added to output
	output := outputs[0].(map[string]any)
	assert.Contains(t, output, "match")
	assert.Equal(t, "*", output["match"])
}

// Test buildFluentBitConfig missing input
func TestBuildFluentBitConfig_MissingInput(t *testing.T) {
	components := [][]models.PipelineNodes{
		{
			{
				ComponentID:      1,
				Name:             "output",
				ComponentName:    "stdout",
				ComponentRole:    "exporter", // Only exporter, no receiver
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{},
			},
		},
	}

	_, _, _, err := buildFluentBitConfig(components)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing input")
}

// Test buildFluentBitConfig missing output
func TestBuildFluentBitConfig_MissingOutput(t *testing.T) {
	components := [][]models.PipelineNodes{
		{
			{
				ComponentID:      1,
				Name:             "input",
				ComponentName:    "tail",
				ComponentRole:    "receiver", // Only receiver, no exporter
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{},
			},
		},
	}

	_, _, _, err := buildFluentBitConfig(components)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing output")
}

// Test buildFluentBitConfig with filters
func TestBuildFluentBitConfig_WithFilters(t *testing.T) {
	components := [][]models.PipelineNodes{
		{
			{
				ComponentID:      1,
				Name:             "input",
				ComponentName:    "tail",
				ComponentRole:    "receiver",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{"path": "/var/log/*.log"},
			},
			{
				ComponentID:      2,
				Name:             "filter",
				ComponentName:    "grep",
				ComponentRole:    "processor",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{"regex": "error"},
			},
			{
				ComponentID:      3,
				Name:             "output",
				ComponentName:    "stdout",
				ComponentRole:    "exporter",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{},
			},
		},
	}

	inputs, filters, outputs, err := buildFluentBitConfig(components)

	assert.NoError(t, err)
	assert.Len(t, inputs, 1)
	assert.Len(t, filters, 1)
	assert.Len(t, outputs, 1)

	// Verify match was added to filter
	filter := filters[0].(map[string]any)
	require.Contains(t, filter, "match")
	assert.Equal(t, "*", filter["match"])
}

// Test intersectSupportedSignals
func TestIntersectSupportedSignals(t *testing.T) {
	testCases := []struct {
		name     string
		first    []string
		second   []string
		expected []string
	}{
		{
			name:     "Common signals",
			first:    []string{"logs", "metrics", "traces"},
			second:   []string{"logs", "metrics"},
			expected: []string{"logs", "metrics"},
		},
		{
			name:     "No common signals",
			first:    []string{"logs"},
			second:   []string{"metrics"},
			expected: []string{},
		},
		{
			name:     "Identical signals",
			first:    []string{"logs", "metrics"},
			second:   []string{"logs", "metrics"},
			expected: []string{"logs", "metrics"},
		},
		{
			name:     "Empty first",
			first:    []string{},
			second:   []string{"logs"},
			expected: []string{},
		},
		{
			name:     "Empty second",
			first:    []string{"logs"},
			second:   []string{},
			expected: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := intersectSupportedSignals(tc.first, tc.second)
			assert.Equal(t, tc.expected, result)
		})
	}
}
