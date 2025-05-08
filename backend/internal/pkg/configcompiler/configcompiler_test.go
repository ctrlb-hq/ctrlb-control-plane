package configcompiler

import (
	"testing"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestCompileGraphToJSON_Success(t *testing.T) {
	graph := models.PipelineGraph{
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
				ComponentName:    "otlp_exporter",
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

func TestCompileGraphToJSON_EmptyGraph(t *testing.T) {
	graph := models.PipelineGraph{}
	result, err := CompileGraphToJSON(graph)
	assert.Error(t, err)
	assert.Nil(t, result)
}
