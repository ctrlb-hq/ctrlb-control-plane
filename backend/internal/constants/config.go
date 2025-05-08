package constants

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
