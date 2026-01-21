package configcompiler

import (
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
)

// =============================================================================
// BASIC SAMPLE GRAPH HELPERS
// =============================================================================

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
// REALISTIC OTEL PIPELINE HELPERS
// =============================================================================

// createRealisticOTELLogsPipeline creates a production-like OTEL logs pipeline
// Pattern: filelog receiver -> batch processor -> memory_limiter -> otlp exporter
func createRealisticOTELLogsPipeline() models.PipelineGraph {
	return models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{
				ComponentID:      1,
				Name:             "filelog_receiver",
				ComponentName:    "filelog_receiver",
				ComponentRole:    "receiver",
				SupportedSignals: []string{"logs"},
				Config: map[string]any{
					"include":           []string{"/var/log/containers/*.log"},
					"start_at":          "beginning",
					"include_file_path": true,
					"include_file_name": false,
					"operators": []map[string]any{
						{
							"type":   "json_parser",
							"field":  "body",
							"target": "attributes",
						},
					},
				},
			},
			{
				ComponentID:      2,
				Name:             "batch_processor",
				ComponentName:    "batch_processor",
				ComponentRole:    "processor",
				SupportedSignals: []string{"logs", "metrics", "traces"},
				Config: map[string]any{
					"send_batch_size":     8192,
					"send_batch_max_size": 0,
					"timeout":             "200ms",
				},
			},
			{
				ComponentID:      3,
				Name:             "memory_limiter",
				ComponentName:    "memorylimiter_processor",
				ComponentRole:    "processor",
				SupportedSignals: []string{"logs", "metrics", "traces"},
				Config: map[string]any{
					"check_interval":         "1s",
					"limit_mib":              400,
					"spike_limit_mib":        100,
					"limit_percentage":       0,
					"spike_limit_percentage": 0,
				},
			},
			{
				ComponentID:      4,
				Name:             "otlp_grpc_exporter",
				ComponentName:    "otlp_exporter",
				ComponentRole:    "exporter",
				SupportedSignals: []string{"logs", "metrics", "traces"},
				Config: map[string]any{
					"endpoint": "otel-collector.monitoring:4317",
					"tls": map[string]any{
						"insecure": true,
					},
					"retry_on_failure": map[string]any{
						"enabled":          true,
						"initial_interval": "5s",
						"max_interval":     "30s",
						"max_elapsed_time": "300s",
					},
				},
			},
		},
		Edges: []models.PipelineEdges{
			{Source: "1", Target: "2"},
			{Source: "2", Target: "3"},
			{Source: "3", Target: "4"},
		},
	}
}

// createRealisticOTELMetricsPipeline creates a production-like OTEL metrics pipeline
// Pattern: prometheus receiver -> filter processor -> batch -> remote_write exporter
func createRealisticOTELMetricsPipeline() models.PipelineGraph {
	return models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{
				ComponentID:      1,
				Name:             "prometheus_receiver",
				ComponentName:    "prometheus_receiver",
				ComponentRole:    "receiver",
				SupportedSignals: []string{"metrics"},
				Config: map[string]any{
					"config": map[string]any{
						"scrape_configs": []map[string]any{
							{
								"job_name":        "kubernetes-pods",
								"scrape_interval": "15s",
								"kubernetes_sd_configs": []map[string]any{
									{"role": "pod"},
								},
							},
						},
					},
				},
			},
			{
				ComponentID:      2,
				Name:             "filter_processor",
				ComponentName:    "filter_processor",
				ComponentRole:    "processor",
				SupportedSignals: []string{"metrics"},
				Config: map[string]any{
					"metrics": map[string]any{
						"include": map[string]any{
							"match_type":   "regexp",
							"metric_names": []string{"http_.*", "rpc_.*", "process_.*"},
						},
					},
				},
			},
			{
				ComponentID:      3,
				Name:             "batch_processor",
				ComponentName:    "batch_processor",
				ComponentRole:    "processor",
				SupportedSignals: []string{"logs", "metrics", "traces"},
				Config: map[string]any{
					"send_batch_size": 1000,
					"timeout":         "10s",
				},
			},
			{
				ComponentID:      4,
				Name:             "prometheus_remote_write",
				ComponentName:    "prometheusremotewrite_exporter",
				ComponentRole:    "exporter",
				SupportedSignals: []string{"metrics"},
				Config: map[string]any{
					"endpoint": "https://prometheus.example.com/api/v1/write",
					"tls": map[string]any{
						"ca_file":   "/etc/ssl/certs/ca.crt",
						"cert_file": "/etc/ssl/certs/client.crt",
						"key_file":  "/etc/ssl/certs/client.key",
					},
					"headers": map[string]any{
						"Authorization": "Bearer ${PROM_TOKEN}",
					},
				},
			},
		},
		Edges: []models.PipelineEdges{
			{Source: "1", Target: "2"},
			{Source: "2", Target: "3"},
			{Source: "3", Target: "4"},
		},
	}
}

// createRealisticOTELTracesPipeline creates a production-like OTEL traces pipeline with sampling
// Pattern: otlp receiver -> probabilistic sampler -> batch -> jaeger exporter
func createRealisticOTELTracesPipeline() models.PipelineGraph {
	return models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{
				ComponentID:      1,
				Name:             "otlp_receiver",
				ComponentName:    "otlp_receiver",
				ComponentRole:    "receiver",
				SupportedSignals: []string{"traces", "metrics", "logs"},
				Config: map[string]any{
					"protocols": map[string]any{
						"grpc": map[string]any{
							"endpoint":              "0.0.0.0:4317",
							"max_recv_msg_size_mib": 4,
						},
						"http": map[string]any{
							"endpoint": "0.0.0.0:4318",
							"cors": map[string]any{
								"allowed_origins": []string{"*"},
							},
						},
					},
				},
			},
			{
				ComponentID:      2,
				Name:             "probabilistic_sampler",
				ComponentName:    "probabilisticsampler_processor",
				ComponentRole:    "processor",
				SupportedSignals: []string{"traces"},
				Config: map[string]any{
					"sampling_percentage": 10.0,
					"hash_seed":           22,
				},
			},
			{
				ComponentID:      3,
				Name:             "batch_processor",
				ComponentName:    "batch_processor",
				ComponentRole:    "processor",
				SupportedSignals: []string{"logs", "metrics", "traces"},
				Config: map[string]any{
					"send_batch_size": 512,
					"timeout":         "5s",
				},
			},
			{
				ComponentID:      4,
				Name:             "jaeger_exporter",
				ComponentName:    "jaeger_exporter",
				ComponentRole:    "exporter",
				SupportedSignals: []string{"traces"},
				Config: map[string]any{
					"endpoint": "jaeger-collector.tracing:14250",
					"tls": map[string]any{
						"insecure": true,
					},
				},
			},
		},
		Edges: []models.PipelineEdges{
			{Source: "1", Target: "2"},
			{Source: "2", Target: "3"},
			{Source: "3", Target: "4"},
		},
	}
}

// createMultiSignalOTELPipeline creates a pipeline that handles all three signal types
// with a single unified path to a multi-signal exporter
func createMultiSignalOTELPipeline() models.PipelineGraph {
	return models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{
				ComponentID:      1,
				Name:             "otlp_receiver",
				ComponentName:    "otlp_receiver",
				ComponentRole:    "receiver",
				SupportedSignals: []string{"traces", "metrics", "logs"},
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
				Name:             "resource_processor",
				ComponentName:    "resource_processor",
				ComponentRole:    "processor",
				SupportedSignals: []string{"traces", "metrics", "logs"},
				Config: map[string]any{
					"attributes": []map[string]any{
						{
							"key":    "environment",
							"value":  "production",
							"action": "upsert",
						},
						{
							"key":    "service.namespace",
							"value":  "my-namespace",
							"action": "upsert",
						},
					},
				},
			},
			{
				ComponentID:      3,
				Name:             "batch_processor",
				ComponentName:    "batch_processor",
				ComponentRole:    "processor",
				SupportedSignals: []string{"traces", "metrics", "logs"},
				Config: map[string]any{
					"timeout": "10s",
				},
			},
			{
				ComponentID:      4,
				Name:             "otlp_exporter",
				ComponentName:    "otlp_exporter",
				ComponentRole:    "exporter",
				SupportedSignals: []string{"traces", "metrics", "logs"},
				Config: map[string]any{
					"endpoint": "otel-backend.monitoring:4317",
				},
			},
		},
		Edges: []models.PipelineEdges{
			{Source: "1", Target: "2"},
			{Source: "2", Target: "3"},
			{Source: "3", Target: "4"},
		},
	}
}

// createMultiExporterOTELPipeline creates a pipeline where data fans out to multiple
// exporters that all support the same signals (common real-world pattern)
func createMultiExporterOTELPipeline() models.PipelineGraph {
	return models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{
				ComponentID:      1,
				Name:             "otlp_receiver",
				ComponentName:    "otlp_receiver",
				ComponentRole:    "receiver",
				SupportedSignals: []string{"traces", "metrics", "logs"},
				Config: map[string]any{
					"protocols": map[string]any{
						"grpc": map[string]any{"endpoint": "0.0.0.0:4317"},
					},
				},
			},
			{
				ComponentID:      2,
				Name:             "batch_processor",
				ComponentName:    "batch_processor",
				ComponentRole:    "processor",
				SupportedSignals: []string{"traces", "metrics", "logs"},
				Config: map[string]any{
					"timeout": "10s",
				},
			},
			{
				ComponentID:      3,
				Name:             "primary_exporter",
				ComponentName:    "otlp_exporter",
				ComponentRole:    "exporter",
				SupportedSignals: []string{"traces", "metrics", "logs"},
				Config: map[string]any{
					"endpoint": "primary-backend:4317",
				},
			},
			{
				ComponentID:      4,
				Name:             "backup_exporter",
				ComponentName:    "otlp_exporter",
				ComponentRole:    "exporter",
				SupportedSignals: []string{"traces", "metrics", "logs"},
				Config: map[string]any{
					"endpoint": "backup-backend:4317",
				},
			},
			{
				ComponentID:      5,
				Name:             "archive_exporter",
				ComponentName:    "otlp_exporter",
				ComponentRole:    "exporter",
				SupportedSignals: []string{"traces", "metrics", "logs"},
				Config: map[string]any{
					"endpoint": "archive-backend:4317",
				},
			},
		},
		Edges: []models.PipelineEdges{
			{Source: "1", Target: "2"},
			{Source: "2", Target: "3"},
			{Source: "2", Target: "4"},
			{Source: "2", Target: "5"},
		},
	}
}

// =============================================================================
// REALISTIC FLUENT BIT PIPELINE HELPERS
// =============================================================================

// createRealisticFBKubernetesLogsPipeline creates a production-like FluentBit k8s logs pipeline
// Pattern: tail input -> kubernetes filter -> parser -> http output
func createRealisticFBKubernetesLogsPipeline() models.PipelineGraph {
	return models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{
				ComponentID:      1,
				Name:             "kubernetes_logs",
				ComponentName:    "tail_input",
				ComponentRole:    "input",
				SupportedSignals: []string{"logs"},
				Config: map[string]any{
					"path":               "/var/log/containers/*.log",
					"path_key":           "filename",
					"exclude_path":       "/var/log/containers/*fluent*.log",
					"read_from_head":     false,
					"refresh_interval":   10,
					"rotate_wait":        30,
					"skip_long_lines":    "on",
					"db":                 "/var/log/flb_kube.db",
					"db.sync":            "normal",
					"mem_buf_limit":      "5MB",
					"docker_mode":        "on",
					"docker_mode_flush":  4,
					"docker_mode_parser": "firstline",
				},
			},
			{
				ComponentID:      2,
				Name:             "kubernetes_filter",
				ComponentName:    "kubernetes_filter",
				ComponentRole:    "filter",
				SupportedSignals: []string{"logs"},
				Config: map[string]any{
					"kube_url":            "https://kubernetes.default.svc:443",
					"kube_ca_file":        "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
					"kube_token_file":     "/var/run/secrets/kubernetes.io/serviceaccount/token",
					"merge_log":           "on",
					"merge_log_key":       "log_processed",
					"k8s-logging.parser":  "on",
					"k8s-logging.exclude": "on",
					"labels":              "on",
					"annotations":         "off",
					"keep_log":            "off",
					"buffer_size":         "32k",
				},
			},
			{
				ComponentID:      3,
				Name:             "record_modifier",
				ComponentName:    "record_modifier_filter",
				ComponentRole:    "filter",
				SupportedSignals: []string{"logs"},
				Config: map[string]any{
					"record":     "cluster my-cluster",
					"remove_key": "stream",
				},
			},
			{
				ComponentID:      4,
				Name:             "http_output",
				ComponentName:    "http_output",
				ComponentRole:    "output",
				SupportedSignals: []string{"logs"},
				Config: map[string]any{
					"host":             "logs-collector.monitoring",
					"port":             8080,
					"uri":              "/api/v1/logs",
					"format":           "json",
					"json_date_key":    "timestamp",
					"json_date_format": "iso8601",
					"header":           "Content-Type application/json",
					"tls":              "off",
				},
			},
		},
		Edges: []models.PipelineEdges{
			{Source: "1", Target: "2"},
			{Source: "2", Target: "3"},
			{Source: "3", Target: "4"},
		},
	}
}

// createRealisticFBSyslogPipeline creates a syslog ingestion pipeline
// Pattern: syslog input -> parser filter -> grep filter -> splunk output
func createRealisticFBSyslogPipeline() models.PipelineGraph {
	return models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{
				ComponentID:      1,
				Name:             "syslog_tcp_input",
				ComponentName:    "syslog_input",
				ComponentRole:    "input",
				SupportedSignals: []string{"logs"},
				Config: map[string]any{
					"mode":                "tcp",
					"listen":              "0.0.0.0",
					"port":                5140,
					"parser":              "syslog-rfc5424",
					"buffer_max_size":     "64KB",
					"receive_buffer_size": "1MB",
				},
			},
			{
				ComponentID:      2,
				Name:             "parser_filter",
				ComponentName:    "parser_filter",
				ComponentRole:    "filter",
				SupportedSignals: []string{"logs"},
				Config: map[string]any{
					"key_name":     "message",
					"parser":       "json",
					"preserve_key": "on",
					"reserve_data": "on",
				},
			},
			{
				ComponentID:      3,
				Name:             "grep_filter",
				ComponentName:    "grep_filter",
				ComponentRole:    "filter",
				SupportedSignals: []string{"logs"},
				Config: map[string]any{
					"regex": "severity (error|warning|critical)",
				},
			},
			{
				ComponentID:      4,
				Name:             "splunk_output",
				ComponentName:    "splunk_output",
				ComponentRole:    "output",
				SupportedSignals: []string{"logs"},
				Config: map[string]any{
					"host":             "splunk-hec.example.com",
					"port":             8088,
					"splunk_token":     "${SPLUNK_HEC_TOKEN}",
					"tls":              "on",
					"tls.verify":       "on",
					"splunk_send_raw":  "off",
					"event_host":       "${HOSTNAME}",
					"event_source":     "syslog",
					"event_sourcetype": "syslog",
					"event_index":      "main",
					"http_buffer_size": "512KB",
					"compress":         "gzip",
				},
			},
		},
		Edges: []models.PipelineEdges{
			{Source: "1", Target: "2"},
			{Source: "2", Target: "3"},
			{Source: "3", Target: "4"},
		},
	}
}

// createRealisticFBMultiInputPipeline creates a pipeline with multiple input sources
// converging to a single output (S3)
func createRealisticFBMultiInputPipeline() models.PipelineGraph {
	return models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{
				ComponentID:      1,
				Name:             "app_logs",
				ComponentName:    "tail_input",
				ComponentRole:    "input",
				SupportedSignals: []string{"logs"},
				Config: map[string]any{
					"path":           "/var/log/app/*.log",
					"tag":            "app.logs",
					"read_from_head": true,
					"parser":         "json",
				},
			},
			{
				ComponentID:      2,
				Name:             "audit_logs",
				ComponentName:    "tail_input",
				ComponentRole:    "input",
				SupportedSignals: []string{"logs"},
				Config: map[string]any{
					"path":           "/var/log/audit/*.log",
					"tag":            "audit.logs",
					"read_from_head": true,
					"parser":         "json",
				},
			},
			{
				ComponentID:      3,
				Name:             "access_logs",
				ComponentName:    "tail_input",
				ComponentRole:    "input",
				SupportedSignals: []string{"logs"},
				Config: map[string]any{
					"path":           "/var/log/nginx/access.log",
					"tag":            "nginx.access",
					"read_from_head": false,
					"parser":         "nginx",
				},
			},
			{
				ComponentID:      4,
				Name:             "timestamp_modifier",
				ComponentName:    "modify_filter",
				ComponentRole:    "filter",
				SupportedSignals: []string{"logs"},
				Config: map[string]any{
					"add":  "processed_at ${FLUENT_TIME}",
					"copy": "log message_backup",
				},
			},
			{
				ComponentID:      5,
				Name:             "s3_output",
				ComponentName:    "s3_output",
				ComponentRole:    "output",
				SupportedSignals: []string{"logs"},
				Config: map[string]any{
					"region":                       "us-east-1",
					"bucket":                       "my-logs-bucket",
					"s3_key_format":                "/logs/$TAG/%Y/%m/%d/%H/%M/%S",
					"s3_key_format_tag_delimiters": ".",
					"total_file_size":              "100M",
					"upload_timeout":               "10m",
					"use_put_object":               "on",
					"compression":                  "gzip",
					"content_type":                 "application/gzip",
					"store_dir":                    "/tmp/fluent-bit/s3",
				},
			},
		},
		Edges: []models.PipelineEdges{
			{Source: "1", Target: "4"},
			{Source: "2", Target: "4"},
			{Source: "3", Target: "4"},
			{Source: "4", Target: "5"},
		},
	}
}
