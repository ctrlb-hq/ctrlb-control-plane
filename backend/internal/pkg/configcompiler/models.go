package configcompiler

var Receivers map[string]any
var Processors map[string]any
var Exporters map[string]any

type Pipeline struct {
	Receivers  []string
	Processors []string
	Exporters  []string
}

type Pipelines map[string]Pipeline

var TelemetryService = map[string]any{
	"telemetry": map[string]any{
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
	},
}
