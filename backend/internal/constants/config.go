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
