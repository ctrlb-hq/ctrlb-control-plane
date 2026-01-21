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

var DefaultConfigOTEL = map[string]any{
	"receivers": map[string]any{
		"otlp": map[string]any{
			"protocols": map[string]any{
				"grpc": map[string]any{
					"endpoint": "0.0.0.0:4317",
				},
				"http": map[string]any{
					"endpoint": "0.0.0.0:4318",
				},
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

var FluentBitService = map[string]any{
	"flush":       1,
	"log_level":   "info",
	"http_server": "on",
	"http_listen": "0.0.0.0",
	"http_port":   2020,
	"hot_reload":  "on",
}

var FluentBitNodeMetricsInput = map[string]any{
	"name":            "node_exporter_metrics",
	"tag":             "ctrlb_agent_node_metrics",
	"scrape_interval": 2,
}

var FluentBitInternalMetricsInput = map[string]any{
	"name":            "fluentbit_metrics",
	"tag":             "ctrlb_agent_internal_metrics",
	"scrape_interval": 2,
}

var FluentBitPrometheusOutput = map[string]any{
	"name":  "prometheus_exporter",
	"match": "ctrlb_agent_*_metrics",
	"host":  "0.0.0.0",
	"port":  2021,
}

// FluentBitDefaultParsers is an inline YAML-compatible representation of common
// This lets us ship a runnable default config without requiring an external parsers_file on the agent host.
var FluentBitDefaultParsers = []any{
	map[string]any{
		"name":        "apache",
		"format":      "regex",
		"regex":       `^(?<host>[^ ]*) [^ ]* (?<user>[^ ]*) \[(?<time>[^\]]*)\] "(?<method>\S+)(?: +(?<path>[^\"]*?)(?: +\S*)?)?" (?<code>[^ ]*) (?<size>[^ ]*)(?: "(?<referer>[^\"]*)" "(?<agent>[^\"]*)")?$`,
		"time_key":    "time",
		"time_format": "%d/%b/%Y:%H:%M:%S %z",
	},
	map[string]any{
		"name":        "apache2",
		"format":      "regex",
		"regex":       `^(?<host>[^ ]*) [^ ]* (?<user>[^ ]*) \[(?<time>[^\]]*)\] "(?<method>\S+)(?: +(?<path>[^ ]*) +\S*)?" (?<code>[^ ]*) (?<size>[^ ]*)(?: "(?<referer>[^\"]*)" "(?<agent>.*)")?$`,
		"time_key":    "time",
		"time_format": "%d/%b/%Y:%H:%M:%S %z",
	},
	map[string]any{
		"name":   "apache_error",
		"format": "regex",
		"regex":  `^\[[^ ]* (?<time>[^\]]*)\] \[(?<level>[^\]]*)\](?: \[pid (?<pid>[^\]]*)\])?( \[client (?<client>[^\]]*)\])? (?<message>.*)$`,
	},
	map[string]any{
		"name":        "nginx",
		"format":      "regex",
		"regex":       `^(?<remote>[^ ]*) (?<host>[^ ]*) (?<user>[^ ]*) \[(?<time>[^\]]*)\] "(?<method>\S+)(?: +(?<path>[^\"]*?)(?: +\S*)?)?" (?<code>[^ ]*) (?<size>[^ ]*)(?: "(?<referer>[^\"]*)" "(?<agent>[^\"]*)")`,
		"time_key":    "time",
		"time_format": "%d/%b/%Y:%H:%M:%S %z",
	},
	map[string]any{
		"name":        "k8s-nginx-ingress",
		"format":      "regex",
		"regex":       `^(?<host>[^ ]*) - (?<user>[^ ]*) \[(?<time>[^\]]*)\] "(?<method>\S+)(?: +(?<path>[^\"]*?)(?: +\S*)?)?" (?<code>[^ ]*) (?<size>[^ ]*) "(?<referer>[^\"]*)" "(?<agent>[^\"]*)" (?<request_length>[^ ]*) (?<request_time>[^ ]*) \[(?<proxy_upstream_name>[^ ]*)\] (\[(?<proxy_alternative_upstream_name>[^ ]*)\] )?(?<upstream_addr>[^ ]*) (?<upstream_response_length>[^ ]*) (?<upstream_response_time>[^ ]*) (?<upstream_status>[^ ]*) (?<reg_id>[^ ]*).*$`,
		"time_key":    "time",
		"time_format": "%d/%b/%Y:%H:%M:%S %z",
	},
	map[string]any{
		"name":        "json",
		"format":      "json",
		"time_key":    "timestamp",
		"time_format": "%Y-%m-%dT%H:%M:%S.%L%z",
	},
	map[string]any{
		"name":   "logfmt",
		"format": "logfmt",
	},
	map[string]any{
		"name":        "docker",
		"format":      "json",
		"time_key":    "time",
		"time_format": "%Y-%m-%dT%H:%M:%S.%L",
		"time_keep":   true,
	},
	map[string]any{
		"name":        "docker-daemon",
		"format":      "regex",
		"regex":       `time="(?<time>[^ ]*)" level=(?<level>[^ ]*) msg="(?<msg>[^ ].*)"`,
		"time_key":    "time",
		"time_format": "%Y-%m-%dT%H:%M:%S.%L",
		"time_keep":   true,
	},
	map[string]any{
		"name":        "syslog-rfc5424",
		"format":      "regex",
		"regex":       `^\<(?<pri>[0-9]{1,5})\>1 (?<time>[^ ]+) (?<host>[^ ]+) (?<ident>[^ ]+) (?<pid>[-0-9]+) (?<msgid>[^ ]+) (?<extradata>(\[(.*?)\]|-)) (?<message>.+)$`,
		"time_key":    "time",
		"time_format": "%Y-%m-%dT%H:%M:%S.%L%z",
		"time_keep":   true,
	},
	map[string]any{
		"name":        "syslog-rfc3164-local",
		"format":      "regex",
		"regex":       `^\<(?<pri>[0-9]+)\>(?<time>[^ ]* {1,2}[^ ]* [^ ]*) (?<ident>[a-zA-Z0-9_\/\.\-]*)(?:\[(?<pid>[0-9]+)\])?(?:[^\:]*\:)? *(?<message>.*)$`,
		"time_key":    "time",
		"time_format": "%b %d %H:%M:%S",
		"time_keep":   true,
	},
	map[string]any{
		"name":        "syslog-rfc3164",
		"format":      "regex",
		"regex":       `/^\<(?<pri>[0-9]+)\>(?<time>[^ ]* {1,2}[^ ]* [^ ]*) (?<host>[^ ]*) (?<ident>[a-zA-Z0-9_\/\.\-]*)(?:\[(?<pid>[0-9]+)\])?(?:[^\:]*\:)? *(?<message>.*)$/`,
		"time_key":    "time",
		"time_format": "%b %d %H:%M:%S",
		"time_keep":   true,
	},
	map[string]any{
		"name":        "mongodb",
		"format":      "regex",
		"regex":       `^(?<time>[^ ]*)\s+(?<severity>\w)\s+(?<component>[^ ]+)\s+\[(?<context>[^\]]+)]\s+(?<message>.*?) *(?<ms>(\d+))?(:?ms)?$`,
		"time_key":    "time",
		"time_format": "%Y-%m-%dT%H:%M:%S.%L",
		"time_keep":   true,
	},
	map[string]any{
		"name":        "envoy",
		"format":      "regex",
		"regex":       `^\[(?<start_time>[^\]]*)\] "(?<method>\S+)(?: +(?<path>[^\"]*?)(?: +\S*)?)? (?<protocol>\S+)" (?<code>[^ ]*) (?<response_flags>[^ ]*) (?<bytes_received>[^ ]*) (?<bytes_sent>[^ ]*) (?<duration>[^ ]*) (?<x_envoy_upstream_service_time>[^ ]*) "(?<x_forwarded_for>[^ ]*)" "(?<user_agent>[^\"]*)" "(?<request_id>[^\"]*)" "(?<authority>[^ ]*)" "(?<upstream_host>[^ ]*)"`,
		"time_key":    "start_time",
		"time_format": "%Y-%m-%dT%H:%M:%S.%L%z",
		"time_keep":   true,
	},
	map[string]any{
		"name":        "istio-envoy-proxy",
		"format":      "regex",
		"regex":       `^\[(?<start_time>[^\]]*)\] "(?<method>\S+)(?: +(?<path>[^\"]*?)(?: +\S*)?)? (?<protocol>\S+)" (?<response_code>[^ ]*) (?<response_flags>[^ ]*) (?<response_code_details>[^ ]*) (?<connection_termination_details>[^ ]*) (?<upstream_transport_failure_reason>[^ ]*) (?<bytes_received>[^ ]*) (?<bytes_sent>[^ ]*) (?<duration>[^ ]*) (?<x_envoy_upstream_service_time>[^ ]*) "(?<x_forwarded_for>[^ ]*)" "(?<user_agent>[^\"]*)" "(?<x_request_id>[^\"]*)" (?<authority>[^ ]*)" "(?<upstream_host>[^ ]*)" (?<upstream_cluster>[^ ]*) (?<upstream_local_address>[^ ]*) (?<downstream_local_address>[^ ]*) (?<downstream_remote_address>[^ ]*) (?<requested_server_name>[^ ]*) (?<route_name>[^  ]*)`,
		"time_key":    "start_time",
		"time_format": "%Y-%m-%dT%H:%M:%S.%L%z",
		"time_keep":   true,
	},
	map[string]any{
		"name":        "cri",
		"format":      "regex",
		"regex":       `^(?<time>[^ ]+) (?<stream>stdout|stderr) (?<logtag>[^ ]*) (?<message>.*)$`,
		"time_key":    "time",
		"time_format": "%Y-%m-%dT%H:%M:%S.%L%z",
		"time_keep":   true,
	},
	map[string]any{
		"name":   "kube-custom",
		"format": "regex",
		"regex":  `(?<tag>[^.]+)?\.?(?<pod_name>[a-z0-9](?:[-a-z0-9]*[a-z0-9])?(?:\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*)_(?<namespace_name>[^_]+)_(?<container_name>.+)-(?<docker_id>[a-z0-9]{64})\.log$`,
	},
	map[string]any{
		"name":        "kmsg-netfilter-log",
		"format":      "regex",
		"regex":       `^\<(?<pri>[0-9]{1,5})\>1 (?<time>[^ ]+) (?<host>[^ ]+) kernel - - - \[[0-9\.]*\] (?<logprefix>[^ ]*)\s?IN=(?<in>[^ ]*) OUT=(?<out>[^ ]*) MAC=(?<macsrc>[0-9a-f]{2}:[0-9a-f]{2}:[0-9a-f]{2}:[0-9a-f]{2}:[0-9a-f]{2}:[0-9a-f]{2}):(?<macdst>[0-9a-f]{2}:[0-9a-f]{2}:[0-9a-f]{2}:[0-9a-f]{2}:[0-9a-f]{2}:[0-9a-f]{2}):(?<ethtype>[0-9a-f]{2}:[0-9a-f]{2}) SRC=(?<saddr>[^ ]*) DST=(?<daddr>[^ ]*) LEN=(?<len>[^ ]*) TOS=(?<tos>[^ ]*) PREC=(?<prec>[^ ]*) TTL=(?<ttl>[^ ]*) ID=(?<id>[^ ]*) (D*F*)\s*PROTO=(?<proto>[^ ]*)\s?((SPT=)?(?<sport>[0-9]*))\s?((DPT=)?(?<dport>[0-9]*))\s?((LEN=)?(?<protolen>[0-9]*))\s?((WINDOW=)?(?<window>[0-9]*))\s?((RES=)?(?<res>0?x?[0-9]*))\s?(?<flag>[^ ]*)\s?((URGP=)?(?<urgp>[0-9]*))`,
		"time_key":    "time",
		"time_format": "%Y-%m-%dT%H:%M:%S.%L%z",
	},
}

var DefaultConfigFluentBit = map[string]any{
	"service": FluentBitService,
	"parsers": FluentBitDefaultParsers,
	"pipeline": map[string]any{
		"inputs": []any{
			FluentBitNodeMetricsInput,
			FluentBitInternalMetricsInput,
			map[string]any{
				"name":           "tail",
				"tag":            "auth_tail_input",
				"path":           "/var/log/auth.log",
				"path_key":       "filename",
				"read_from_head": false,
			},
		},
		"outputs": []any{FluentBitPrometheusOutput,
			map[string]any{
				"name":    "stdout",
				"format":  "json_lines",
				"workers": 1,
				"match":   "auth_tail_input",
			},
		},
	},
}

var DefaultOTELPipelineGraph = models.PipelineGraph{
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

var DefaultFluentBitPipelineGraph = models.PipelineGraph{
	Nodes: []models.PipelineNodes{
		{
			ComponentID:   1,
			Name:          "Fluent Bit Tail Input Configuration",
			ComponentName: "tail_input",
			ComponentRole: "input",
			SupportedSignals: []string{
				"logs",
			},
			Config: map[string]any{
				"path":           "/var/log/auth.log",
				"path_key":       "filename",
				"read_from_head": false,
			},
		},
		{
			ComponentID:   2,
			Name:          "Fluent Bit Stdout Output Configuration",
			ComponentName: "stdout_output",
			ComponentRole: "output",
			SupportedSignals: []string{
				"logs",
			},
			Config: map[string]any{
				"format": "json_lines",
			},
		},
	},
	Edges: []models.PipelineEdges{
		{
			Source: "1",
			Target: "2",
		},
	},
}
