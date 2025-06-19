package constants

import "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"

var TelemetryService = map[string]any{
	"metrics": map[string]any{
		"level": "detailed",
		"readers": []any{
			map[string]any{
				"pull": map[string]any{
					"exporter": map[string]any{
						"prometheus": map[string]any{
							"host": "0.0.0.0",
							"port": 8888,
						},
					},
				},
			},
		},
	},
}

var DefaultConfig = map[string]any{
	"receivers": map[string]any{
		"otlp": map[string]any{
			"protocols": map[string]any{
				"http": map[string]any{},
				"grpc": map[string]any{},
			},
		},
	},
	"processors": map[string]any{},
	"exporters": map[string]any{
		"debug": map[string]any{},
	},
	"service": map[string]any{
		"telemetry": TelemetryService,
		"pipelines": map[string]any{
			"logs/default": map[string]any{
				"receivers":  []any{"otlp"},
				"processors": []any{},
				"exporters":  []any{"debug"},
			},
		},
	},
}

var DefaultPipelineGraph = models.PipelineGraph{
	Nodes: []models.PipelineNodes{
		{
			ComponentID:   1,
			Name:          "Debug Exporter Configuration",
			ComponentName: "debug_exporter",
			ComponentRole: "exporter",
			SupportedSignals: []string{
				"traces",
				"metrics",
				"logs",
			},
			Config: map[string]any{
				"verbosity": "basic",
			},
		},
		{
			ComponentID:   2,
			Name:          "OTLP Receiver Configuration",
			ComponentName: "otlp_receiver",
			ComponentRole: "receiver",
			SupportedSignals: []string{
				"traces",
				"metrics",
				"logs",
			},
			Config: map[string]any{
				"protocols": map[string]any{
					"http": map[string]any{
						"endpoint": "0.0.0.0:4317",
					},
				},
			},
		},
	},
	Edges: []models.PipelineEdges{
		{
			Source: "2",
			Target: "1",
		},
	},
}
