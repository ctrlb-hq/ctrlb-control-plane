package configcompiler

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/stretchr/testify/assert"
)

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
				Config:           map[string]any{"path": "/var/log/*.log"},
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
				Config:           map[string]any{"regex": "message error"},
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
				Config:           map[string]any{"path": "/var/log/*.log"},
			},
			{
				ComponentID:      2,
				Name:             "filter1",
				ComponentName:    "grep",
				ComponentRole:    "filter",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{"regex": "message error"},
			},
			{
				ComponentID:      3,
				Name:             "filter2",
				ComponentName:    "modify",
				ComponentRole:    "filter",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{"add": "processed true"},
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
				Config:           map[string]any{"regex": "message error"},
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

	// Should have 2 user inputs + 2 monitoring inputs (node_metrics + internal_metrics) = 4
	assert.Len(t, inputs, 4, "Should have 2 user inputs + 2 monitoring inputs")

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
	// No filters in this graph: key may be omitted entirely.
	if raw, ok := pipeline["filters"]; ok {
		filters := raw.([]any)
		assert.Empty(t, filters)
	} else {
		assert.False(t, ok)
	}
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
// COMPREHENSIVE FLUENT BIT CONFIG VALIDATION TESTS
// =============================================================================

func TestFBKubernetesLogsConfigValidation(t *testing.T) {
	graph := createRealisticFBKubernetesLogsPipeline()

	result, err := CompileGraphToFluentBit(graph)

	b, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("Generated FluentBit K8s Logs Config:\n%s\n", string(b))

	assert.NoError(t, err)
	assert.NotNil(t, result)

	config := *result

	// Validate top-level structure
	assert.Contains(t, config, "service", "Config must have service section")
	assert.Contains(t, config, "pipeline", "Config must have pipeline section")

	// Validate service section
	service := config["service"].(map[string]any)
	assert.Equal(t, 1, service["flush"], "Flush should be 1")
	assert.Equal(t, "info", service["log_level"], "Log level should be info")
	assert.Equal(t, "on", service["http_server"], "HTTP server should be on")
	assert.Equal(t, "on", service["hot_reload"], "Hot reload should be on")

	// Validate pipeline section
	pipeline := config["pipeline"].(map[string]any)
	assert.Contains(t, pipeline, "inputs", "Pipeline must have inputs")
	assert.Contains(t, pipeline, "filters", "Pipeline must have filters")
	assert.Contains(t, pipeline, "outputs", "Pipeline must have outputs")

	// Validate inputs
	inputs := pipeline["inputs"].([]any)
	// User input + 2 prometheus inputs (node_metrics + internal_metrics)
	assert.GreaterOrEqual(t, len(inputs), 3, "Should have at least 3 inputs (1 user + 2 monitoring)")

	// Find user input (tail) and validate
	var userInput map[string]any
	for _, inp := range inputs {
		input := inp.(map[string]any)
		if input["name"] == "tail" {
			userInput = input
			break
		}
	}
	assert.NotNil(t, userInput, "Should have tail input")
	assert.Contains(t, userInput, "path", "Tail input must have path")
	assert.Contains(t, userInput, "tag", "Tail input must have tag")
	assert.Contains(t, userInput, "alias", "Tail input must have alias")
	assert.Equal(t, "/var/log/containers/*.log", userInput["path"], "Path should match config")

	// Validate filters
	filters := pipeline["filters"].([]any)
	assert.Len(t, filters, 2, "Should have 2 filters (kubernetes + record_modifier)")

	// Validate kubernetes filter config
	filterTypes := make(map[string]bool)
	for _, f := range filters {
		filter := f.(map[string]any)
		assert.Contains(t, filter, "name", "Filter must have name")
		assert.Contains(t, filter, "match", "Filter must have match")
		assert.Contains(t, filter, "alias", "Filter must have alias")
		filterTypes[filter["name"].(string)] = true
	}
	assert.Contains(t, filterTypes, "kubernetes", "Should have kubernetes filter")
	// record_modifier_filter becomes "record" after TrimAfterUnderscore
	assert.Contains(t, filterTypes, "record", "Should have record modifier filter")

	// Validate outputs
	outputs := pipeline["outputs"].([]any)
	// User output + 1 prometheus output
	assert.GreaterOrEqual(t, len(outputs), 2, "Should have at least 2 outputs")

	// Find HTTP output and validate
	var httpOutput map[string]any
	for _, out := range outputs {
		output := out.(map[string]any)
		if output["name"] == "http" {
			httpOutput = output
			break
		}
	}
	assert.NotNil(t, httpOutput, "Should have http output")
	assert.Contains(t, httpOutput, "host", "HTTP output must have host")
	assert.Contains(t, httpOutput, "port", "HTTP output must have port")
	assert.Contains(t, httpOutput, "match", "HTTP output must have match")
	assert.Equal(t, "logs-collector.monitoring", httpOutput["host"], "Host should match config")
}

func TestFBSyslogConfigValidation(t *testing.T) {
	graph := createRealisticFBSyslogPipeline()

	result, err := CompileGraphToFluentBit(graph)

	b, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("Generated FluentBit Syslog Config:\n%s\n", string(b))

	assert.NoError(t, err)
	assert.NotNil(t, result)

	config := *result
	pipeline := config["pipeline"].(map[string]any)

	// Validate inputs
	inputs := pipeline["inputs"].([]any)
	var syslogInput map[string]any
	for _, inp := range inputs {
		input := inp.(map[string]any)
		if input["name"] == "syslog" {
			syslogInput = input
			break
		}
	}
	assert.NotNil(t, syslogInput, "Should have syslog input")
	assert.Equal(t, "tcp", syslogInput["mode"], "Syslog mode should be tcp")
	assert.Equal(t, 5140, syslogInput["port"], "Syslog port should be 5140")

	// Validate filters
	filters := pipeline["filters"].([]any)
	assert.Len(t, filters, 2, "Should have 2 filters (parser + grep)")

	hasParser := false
	hasGrep := false
	for _, f := range filters {
		filter := f.(map[string]any)
		name := filter["name"].(string)
		if name == "parser" {
			hasParser = true
			assert.Contains(t, filter, "key_name", "Parser filter must have key_name")
		}
		if name == "grep" {
			hasGrep = true
			assert.Contains(t, filter, "regex", "Grep filter must have regex")
		}
	}
	assert.True(t, hasParser, "Should have parser filter")
	assert.True(t, hasGrep, "Should have grep filter")

	// Validate Splunk output
	outputs := pipeline["outputs"].([]any)
	var splunkOutput map[string]any
	for _, out := range outputs {
		output := out.(map[string]any)
		if output["name"] == "splunk" {
			splunkOutput = output
			break
		}
	}
	assert.NotNil(t, splunkOutput, "Should have splunk output")
	assert.Contains(t, splunkOutput, "host", "Splunk output must have host")
	assert.Contains(t, splunkOutput, "splunk_token", "Splunk output must have token")
	assert.Equal(t, "splunk-hec.example.com", splunkOutput["host"], "Splunk host should match")
}

func TestFBMultiInputConfigValidation(t *testing.T) {
	graph := createRealisticFBMultiInputPipeline()

	result, err := CompileGraphToFluentBit(graph)

	b, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("Generated FluentBit Multi-Input Config:\n%s\n", string(b))

	assert.NoError(t, err)
	assert.NotNil(t, result)

	config := *result
	pipeline := config["pipeline"].(map[string]any)

	// Validate inputs - should have 3 user inputs + 2 monitoring inputs
	inputs := pipeline["inputs"].([]any)
	userInputCount := 0
	userTags := make([]string, 0)
	for _, inp := range inputs {
		input := inp.(map[string]any)
		if input["name"] == "tail" {
			userInputCount++
			if tag, ok := input["tag"].(string); ok {
				userTags = append(userTags, tag)
			}
		}
	}
	assert.Equal(t, 3, userInputCount, "Should have 3 tail inputs")

	// Validate each user input has a unique tag
	assert.Len(t, userTags, 3, "Each user input should have a tag")
	tagSet := make(map[string]bool)
	for _, tag := range userTags {
		assert.NotContains(t, tagSet, tag, "Tags should be unique: "+tag)
		tagSet[tag] = true
	}

	// Validate filters
	filters := pipeline["filters"].([]any)
	// User-provided tags have different prefixes (app, audit, nginx), so filter is expanded
	// Even though all inputs are "tail", the custom tags make them incompatible
	assert.Len(t, filters, 3, "Should have 3 filter instances (user tags with different prefixes cause expansion)")

	for i, f := range filters {
		filter := f.(map[string]any)
		assert.Equal(t, "modify", filter["name"], "Filter %d should be modify type", i)
		assert.Contains(t, filter, "match", "Filter %d must have match", i)
		assert.Contains(t, filter, "alias", "Filter %d must have alias", i)
	}

	// Validate outputs - should have 3 user outputs (expanded for different tags) + 1 monitoring output
	outputs := pipeline["outputs"].([]any)
	s3OutputCount := 0
	for _, out := range outputs {
		output := out.(map[string]any)
		if output["name"] == "s3" {
			s3OutputCount++
			assert.Contains(t, output, "region", "S3 output must have region")
			assert.Contains(t, output, "bucket", "S3 output must have bucket")
			assert.Contains(t, output, "s3_key_format", "S3 output must have key format")
			assert.Equal(t, "us-east-1", output["region"], "S3 region should match")
			assert.Equal(t, "my-logs-bucket", output["bucket"], "S3 bucket should match")
		}
	}
	assert.Equal(t, 3, s3OutputCount, "Should have 3 S3 output instances (one per input tag)")
}

// =============================================================================
// EDGE CASE AND COMPLEX TOPOLOGY TESTS - FLUENT BIT
// =============================================================================

func TestFBComplexFilterChain(t *testing.T) {
	// Test a complex filter chain with multiple transformations
	graph := models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{
				ComponentID:      1,
				Name:             "tail_input",
				ComponentName:    "tail_input",
				ComponentRole:    "input",
				SupportedSignals: []string{"logs"},
				Config: map[string]any{
					"path":   "/var/log/app.log",
					"parser": "json",
				},
			},
			{
				ComponentID:      2,
				Name:             "parser_filter",
				ComponentName:    "parser_filter",
				ComponentRole:    "filter",
				SupportedSignals: []string{"logs"},
				Config: map[string]any{
					"key_name": "message",
					"parser":   "json",
				},
			},
			{
				ComponentID:      3,
				Name:             "grep_filter",
				ComponentName:    "grep_filter",
				ComponentRole:    "filter",
				SupportedSignals: []string{"logs"},
				Config: map[string]any{
					"regex": "level error",
				},
			},
			{
				ComponentID:      4,
				Name:             "modify_filter",
				ComponentName:    "modify_filter",
				ComponentRole:    "filter",
				SupportedSignals: []string{"logs"},
				Config: map[string]any{
					"add": "processed true",
				},
			},
			{
				ComponentID:      5,
				Name:             "nest_filter",
				ComponentName:    "nest_filter",
				ComponentRole:    "filter",
				SupportedSignals: []string{"logs"},
				Config: map[string]any{
					"operation":     "nest",
					"wildcard":      "*",
					"nest_under":    "kubernetes",
					"remove_prefix": "kubernetes_",
				},
			},
			{
				ComponentID:      6,
				Name:             "stdout_output",
				ComponentName:    "stdout_output",
				ComponentRole:    "output",
				SupportedSignals: []string{"logs"},
				Config: map[string]any{
					"format": "json_lines",
				},
			},
		},
		Edges: []models.PipelineEdges{
			{Source: "1", Target: "2"},
			{Source: "2", Target: "3"},
			{Source: "3", Target: "4"},
			{Source: "4", Target: "5"},
			{Source: "5", Target: "6"},
		},
	}

	result, err := CompileGraphToFluentBit(graph)

	b, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("Generated FluentBit Complex Filter Chain Config:\n%s\n", string(b))

	assert.NoError(t, err)
	assert.NotNil(t, result)

	config := *result
	pipeline := config["pipeline"].(map[string]any)

	// Validate filter count
	filters := pipeline["filters"].([]any)
	assert.Len(t, filters, 4, "Should have 4 filters")

	// Validate each filter has proper structure
	for i, f := range filters {
		filter := f.(map[string]any)
		assert.Contains(t, filter, "name", "Filter %d must have name", i)
		assert.Contains(t, filter, "match", "Filter %d must have match", i)
		assert.Contains(t, filter, "alias", "Filter %d must have alias", i)
	}

	// Verify input has tag
	inputs := pipeline["inputs"].([]any)
	var tailInput map[string]any
	for _, inp := range inputs {
		input := inp.(map[string]any)
		if input["name"] == "tail" {
			tailInput = input
			break
		}
	}
	assert.NotNil(t, tailInput, "Should have tail input")
	assert.Contains(t, tailInput, "tag", "Tail input must have tag")
}

func TestFBOutputExpansion(t *testing.T) {
	// Test that outputs are correctly expanded when receiving from incompatible inputs
	graph := models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{
				ComponentID:      1,
				Name:             "syslog_input",
				ComponentName:    "syslog_input",
				ComponentRole:    "input",
				SupportedSignals: []string{"logs"},
				Config: map[string]any{
					"mode": "tcp",
					"port": 5140,
				},
			},
			{
				ComponentID:      2,
				Name:             "forward_input",
				ComponentName:    "forward_input",
				ComponentRole:    "input",
				SupportedSignals: []string{"logs"},
				Config: map[string]any{
					"port": 24224,
				},
			},
			{
				ComponentID:      3,
				Name:             "es_output",
				ComponentName:    "es_output",
				ComponentRole:    "output",
				SupportedSignals: []string{"logs"},
				Config: map[string]any{
					"host":  "elasticsearch.example.com",
					"port":  9200,
					"index": "logs",
				},
			},
		},
		Edges: []models.PipelineEdges{
			{Source: "1", Target: "3"},
			{Source: "2", Target: "3"},
		},
	}

	result, err := CompileGraphToFluentBit(graph)

	b, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("Generated FluentBit Output Expansion Config:\n%s\n", string(b))

	assert.NoError(t, err)
	assert.NotNil(t, result)

	config := *result
	pipeline := config["pipeline"].(map[string]any)

	// Validate inputs
	inputs := pipeline["inputs"].([]any)
	userInputCount := 0
	for _, inp := range inputs {
		input := inp.(map[string]any)
		name := input["name"].(string)
		if name == "syslog" || name == "forward" {
			userInputCount++
			assert.Contains(t, input, "tag", "Input must have tag")
		}
	}
	assert.Equal(t, 2, userInputCount, "Should have 2 user inputs")

	// Validate outputs - since syslog and forward have different prefixes,
	// output should be expanded to 2 instances
	outputs := pipeline["outputs"].([]any)
	esOutputCount := 0
	for _, out := range outputs {
		output := out.(map[string]any)
		if output["name"] == "es" {
			esOutputCount++
			assert.Contains(t, output, "match", "ES output must have match")
			match := output["match"].(string)
			assert.NotEqual(t, "*", match, "Match should not be wildcard")
		}
	}
	assert.Equal(t, 2, esOutputCount, "Should have 2 ES output instances for incompatible inputs")
}
