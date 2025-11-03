package database

func GetComponentTypeMap() map[string]string {
	return map[string]string{
		// Receivers
		"otlp_receiver":                  "receiver",
		"hostmetrics_receiver":           "receiver",
		"awscloudwatchmetrics_receiver":  "receiver",
		"awscloudwatch_receiver":         "receiver",
		"azuremonitor_receiver":          "receiver",
		"filelog_receiver":               "receiver",
		"googlecloudmonitoring_receiver": "receiver",

		// Processors
		"batch_processor":                "processor",
		"memorylimiter_processor":        "processor",
		"probabilisticsampler_processor": "processor",
		"attributes_processor":           "processor",
		"filter_processor":               "processor",
		"tailsampling_processor":         "processor",

		// Exporters
		"otlp_grpc_exporter":  "exporter",
		"otlphttp_exporter":   "exporter",
		"debug_exporter":      "exporter",
		"kafka_exporter":      "exporter",
		"prometheus_exporter": "exporter",
	}
}

func GetSignalSupportMap() map[string][]string {
	return map[string][]string{
		// Receivers
		"otlp_receiver":                  {"traces", "metrics", "logs"},
		"hostmetrics_receiver":           {"metrics"},
		"awscloudwatchmetrics_receiver":  {"metrics"},
		"awscloudwatch_receiver":         {"logs"},
		"azuremonitor_receiver":          {"metrics"},
		"filelog_receiver":               {"logs"},
		"googlecloudmonitoring_receiver": {"metrics"},

		// Processors
		"batch_processor":                {"traces", "metrics", "logs"},
		"memorylimiter_processor":        {"traces", "metrics", "logs"},
		"probabilisticsampler_processor": {"traces"},
		"attributes_processor":           {"traces", "metrics", "logs"},
		"filter_processor":               {"traces", "metrics", "logs"},
		"tailsampling_processor":         {"traces"},

		// Exporters
		"otlp_grpc_exporter":  {"traces", "metrics", "logs"},
		"otlphttp_exporter":   {"traces", "metrics", "logs"},
		"debug_exporter":      {"traces", "metrics", "logs"},
		"kafka_exporter":      {"traces", "metrics", "logs"},
		"prometheus_exporter": {"metrics"},
	}
}
