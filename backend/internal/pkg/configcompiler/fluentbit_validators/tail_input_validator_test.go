package configcompiler

import (
	"fmt"
	"testing"
)

func TestValidateTailInputConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]any
		expectError bool
		errorFields []string
	}{
		// Basic valid configurations
		{
			name: "minimal valid config",
			config: map[string]any{
				"path": "/var/log/syslog",
			},
			expectError: false,
		},
		{
			name: "full valid config",
			config: map[string]any{
				"path":              "/var/log/containers/*.log",
				"tag":               "kube.*",
				"read_from_head":    true,
				"refresh_interval":  60,
				"rotate_wait":       5,
				"skip_long_lines":   "on",
				"buffer_chunk_size": "32k",
				"buffer_max_size":   "256k",
				"mem_buf_limit":     "5M",
				"db":                "/var/log/flb_kube.db",
				"db.sync":           "normal",
				"db.locking":        true,
			},
			expectError: false,
		},

		// Path validation
		{
			name:        "missing path",
			config:      map[string]any{},
			expectError: true,
			errorFields: []string{"path"},
		},
		{
			name: "empty path",
			config: map[string]any{
				"path": "",
			},
			expectError: true,
			errorFields: []string{"path"},
		},
		{
			name: "path with wildcards",
			config: map[string]any{
				"path": "/var/log/*.log,/var/log/containers/*.log",
			},
			expectError: false,
		},

		// Buffer validation
		{
			name: "valid buffer sizes",
			config: map[string]any{
				"path":              "/var/log/syslog",
				"buffer_chunk_size": "64k",
				"buffer_max_size":   "128k",
			},
			expectError: false,
		},
		{
			name: "buffer_max_size smaller than buffer_chunk_size",
			config: map[string]any{
				"path":              "/var/log/syslog",
				"buffer_chunk_size": "256k",
				"buffer_max_size":   "32k",
			},
			expectError: true,
			errorFields: []string{"buffer_max_size"},
		},
		{
			name: "invalid buffer_chunk_size format",
			config: map[string]any{
				"path":              "/var/log/syslog",
				"buffer_chunk_size": "invalid",
			},
			expectError: true,
			errorFields: []string{"buffer_chunk_size"},
		},
		{
			name: "valid mem_buf_limit",
			config: map[string]any{
				"path":          "/var/log/syslog",
				"mem_buf_limit": "10M",
			},
			expectError: false,
		},

		// Database validation
		{
			name: "valid db config",
			config: map[string]any{
				"path":            "/var/log/syslog",
				"db":              "/path/to/tail.db",
				"db.sync":         "full",
				"db.journal_mode": "wal",
				"db.locking":      true,
			},
			expectError: false,
		},
		{
			name: "db.sync without db",
			config: map[string]any{
				"path":    "/var/log/syslog",
				"db.sync": "normal",
			},
			expectError: true,
			errorFields: []string{"db.sync"},
		},
		{
			name: "invalid db.sync value",
			config: map[string]any{
				"path":    "/var/log/syslog",
				"db":      "/path/to/tail.db",
				"db.sync": "invalid",
			},
			expectError: true,
			errorFields: []string{"db.sync"},
		},
		{
			name: "invalid db.journal_mode value",
			config: map[string]any{
				"path":            "/var/log/syslog",
				"db":              "/path/to/tail.db",
				"db.journal_mode": "invalid",
			},
			expectError: true,
			errorFields: []string{"db.journal_mode"},
		},
		{
			name: "all db options without db",
			config: map[string]any{
				"path":                "/var/log/syslog",
				"db.sync":             "normal",
				"db.locking":          true,
				"db.journal_mode":     "wal",
				"db.compare_filename": true,
			},
			expectError: true,
			errorFields: []string{"db.sync", "db.locking", "db.journal_mode", "db.compare_filename"},
		},

		// Multiline.parser validation (new style)
		{
			name: "valid multiline.parser",
			config: map[string]any{
				"path":             "/var/log/containers/*.log",
				"multiline.parser": "docker, cri",
			},
			expectError: false,
		},
		{
			name: "multiline.parser with old multiline",
			config: map[string]any{
				"path":             "/var/log/syslog",
				"multiline.parser": "docker",
				"multiline":        true,
			},
			expectError: true,
			errorFields: []string{"multiline.parser"},
		},
		{
			name: "multiline.parser with parser_firstline",
			config: map[string]any{
				"path":             "/var/log/syslog",
				"multiline.parser": "docker",
				"parser_firstline": "my_parser",
			},
			expectError: true,
			errorFields: []string{"multiline.parser"},
		},
		{
			name: "multiline.parser with docker_mode",
			config: map[string]any{
				"path":             "/var/log/syslog",
				"multiline.parser": "docker",
				"docker_mode":      true,
			},
			expectError: true,
			errorFields: []string{"multiline.parser", "docker_mode"},
		},

		// Old multiline validation
		{
			name: "valid old multiline config",
			config: map[string]any{
				"path":             "/var/log/java.log",
				"multiline":        true,
				"parser_firstline": "multiline_parser",
				"parser_1":         "parser_a",
				"parser_2":         "parser_b",
			},
			expectError: false,
		},
		{
			name: "multiline without parser_firstline",
			config: map[string]any{
				"path":      "/var/log/java.log",
				"multiline": true,
			},
			expectError: true,
			errorFields: []string{"multiline"},
		},
		{
			name: "multiline with docker_mode",
			config: map[string]any{
				"path":             "/var/log/java.log",
				"multiline":        true,
				"parser_firstline": "my_parser",
				"docker_mode":      true,
			},
			expectError: true,
			errorFields: []string{"multiline", "docker_mode"},
		},

		// Parser validation
		{
			name: "valid parser",
			config: map[string]any{
				"path":   "/var/log/syslog",
				"parser": "syslog",
			},
			expectError: false,
		},
		{
			name: "parser with multiline.parser",
			config: map[string]any{
				"path":             "/var/log/syslog",
				"parser":           "syslog",
				"multiline.parser": "docker",
			},
			expectError: true,
			errorFields: []string{"parser"},
		},
		{
			name: "parser with old multiline",
			config: map[string]any{
				"path":             "/var/log/syslog",
				"parser":           "syslog",
				"multiline":        true,
				"parser_firstline": "my_parser",
			},
			expectError: true,
			errorFields: []string{"parser"},
		},

		// Docker mode validation
		{
			name: "valid docker_mode config",
			config: map[string]any{
				"path":               "/var/log/containers/*.log",
				"docker_mode":        true,
				"docker_mode_flush":  4,
				"docker_mode_parser": "docker_parser",
			},
			expectError: false,
		},
		{
			name: "docker_mode_flush without docker_mode",
			config: map[string]any{
				"path":              "/var/log/syslog",
				"docker_mode_flush": 4,
			},
			expectError: true,
			errorFields: []string{"docker_mode_flush"},
		},
		{
			name: "docker_mode_parser without docker_mode",
			config: map[string]any{
				"path":               "/var/log/syslog",
				"docker_mode_parser": "parser",
			},
			expectError: true,
			errorFields: []string{"docker_mode_parser"},
		},

		// Encoding validation
		{
			name: "valid unicode.encoding UTF-16LE",
			config: map[string]any{
				"path":             "/var/log/windows.log",
				"unicode.encoding": "UTF-16LE",
			},
			expectError: false,
		},
		{
			name: "valid generic.encoding ShiftJIS",
			config: map[string]any{
				"path":             "/var/log/japanese.log",
				"generic.encoding": "ShiftJIS",
			},
			expectError: false,
		},
		{
			name: "both unicode.encoding and generic.encoding",
			config: map[string]any{
				"path":             "/var/log/syslog",
				"unicode.encoding": "UTF-16LE",
				"generic.encoding": "ShiftJIS",
			},
			expectError: true,
			errorFields: []string{"unicode.encoding", "generic.encoding"},
		},
		{
			name: "invalid unicode.encoding",
			config: map[string]any{
				"path":             "/var/log/syslog",
				"unicode.encoding": "UTF-32",
			},
			expectError: true,
			errorFields: []string{"unicode.encoding"},
		},
		{
			name: "invalid generic.encoding",
			config: map[string]any{
				"path":             "/var/log/syslog",
				"generic.encoding": "INVALID",
			},
			expectError: true,
			errorFields: []string{"generic.encoding"},
		},

		// Threading validation
		{
			name: "valid threading config",
			config: map[string]any{
				"path":                         "/var/log/syslog",
				"threaded":                     true,
				"thread.ring_buffer.capacity":  2048,
				"thread.ring_buffer.window":    10,
			},
			expectError: false,
		},
		{
			name: "thread.ring_buffer without threaded",
			config: map[string]any{
				"path":                        "/var/log/syslog",
				"thread.ring_buffer.capacity": 2048,
			},
			expectError: true,
			errorFields: []string{"thread.ring_buffer.capacity"},
		},
		{
			name: "thread.ring_buffer.window out of range",
			config: map[string]any{
				"path":                      "/var/log/syslog",
				"threaded":                  true,
				"thread.ring_buffer.window": 150,
			},
			expectError: true,
			errorFields: []string{"thread.ring_buffer.window"},
		},

		// Duration/interval validation
		{
			name: "valid ignore_older",
			config: map[string]any{
				"path":         "/var/log/syslog",
				"ignore_older": "7d",
			},
			expectError: false,
		},
		{
			name: "invalid ignore_older format",
			config: map[string]any{
				"path":         "/var/log/syslog",
				"ignore_older": "7days",
			},
			expectError: true,
			errorFields: []string{"ignore_older"},
		},
		{
			name: "valid progress_check_interval",
			config: map[string]any{
				"path":                    "/var/log/syslog",
				"progress_check_interval": "2s",
			},
			expectError: false,
		},

		// Tag validation
		{
			name: "valid tag with asterisk",
			config: map[string]any{
				"path": "/var/log/syslog",
				"tag":  "logs.*",
			},
			expectError: false,
		},
		{
			name: "valid tag_regex",
			config: map[string]any{
				"path":      "/var/log/containers/*.log",
				"tag_regex": `(?<pod_name>[a-z0-9-]+)_(?<namespace>[^_]+)\.log$`,
			},
			expectError: false,
		},
		{
			name: "invalid tag_regex",
			config: map[string]any{
				"path":      "/var/log/syslog",
				"tag_regex": "[invalid(regex",
			},
			expectError: true,
			errorFields: []string{"tag_regex"},
		},

		// Boolean field validation
		{
			name: "valid boolean fields",
			config: map[string]any{
				"path":            "/var/log/syslog",
				"read_from_head":  true,
				"skip_long_lines": "on",
				"skip_empty_lines": "off",
				"exit_on_eof":     false,
			},
			expectError: false,
		},
		{
			name: "invalid boolean value",
			config: map[string]any{
				"path":           "/var/log/syslog",
				"read_from_head": "enabled",
			},
			expectError: true,
			errorFields: []string{"read_from_head"},
		},

		// Unknown key detection
		{
			name: "unknown configuration key",
			config: map[string]any{
				"path":        "/var/log/syslog",
				"unknown_key": "value",
			},
			expectError: true,
			errorFields: []string{"unknown_key"},
		},

		// Real-world configurations
		{
			name: "kubernetes container logs config",
			config: map[string]any{
				"path":             "/var/log/containers/*.log",
				"tag":              "kube.<namespace_name>.<pod_name>.<container_name>",
				"multiline.parser": "docker, cri",
				"read_from_head":   false,
				"db":               "/var/log/flb_kube.db",
				"mem_buf_limit":    "5MB",
				"skip_long_lines":  true,
				"refresh_interval": 10,
			},
			expectError: false,
		},
		{
			name: "sumo logic config",
			config: map[string]any{
				"path":              "/var/log/syslog",
				"db":                "/var/db/fluentbit/pos/syslog.db",
				"db.sync":           "normal",
				"buffer_chunk_size": "1M",
				"buffer_max_size":   "5M",
				"read_from_head":    true,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateTailInputConfig(tt.config)

			if tt.expectError && !errs.HasErrors() {
				t.Errorf("expected errors but got none")
				return
			}

			if !tt.expectError && errs.HasErrors() {
				t.Errorf("expected no errors but got: %v", errs)
				return
			}

			if tt.expectError && len(tt.errorFields) > 0 {
				for _, field := range tt.errorFields {
					found := false
					for _, err := range errs {
						if err.Field == field {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected error for field %q but not found in errors: %v", field, errs)
					}
				}
			}
		})
	}
}

func TestIsValidUnitSize(t *testing.T) {
	validCases := []string{
		"32k", "32K", "32kb", "32KB",
		"1m", "1M", "1mb", "1MB",
		"256", "512",
		"1.5M", "2.5k",
		"50M",
	}

	for _, c := range validCases {
		if !isValidUnitSize(c) {
			t.Errorf("expected %q to be valid unit size", c)
		}
	}

	invalidCases := []string{
		"", "invalid", "32x", "MB", "k32",
	}

	for _, c := range invalidCases {
		if isValidUnitSize(c) {
			t.Errorf("expected %q to be invalid unit size", c)
		}
	}
}

func TestParseUnitSize(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"32k", 32 * 1024},
		{"1M", 1024 * 1024},
		{"1G", 1024 * 1024 * 1024},
		{"256", 256},
		{"1.5k", 1536},
	}

	for _, tt := range tests {
		result := parseUnitSize(tt.input)
		if result != tt.expected {
			t.Errorf("parseUnitSize(%q) = %d, expected %d", tt.input, result, tt.expected)
		}
	}
}

func TestIsValidIgnoreOlderDuration(t *testing.T) {
	validCases := []string{"5m", "1h", "7d", "30m", "24h"}

	for _, c := range validCases {
		if !isValidIgnoreOlderDuration(c) {
			t.Errorf("expected %q to be valid ignore_older duration", c)
		}
	}

	invalidCases := []string{"", "5", "5s", "7days", "1week"}

	for _, c := range invalidCases {
		if isValidIgnoreOlderDuration(c) {
			t.Errorf("expected %q to be invalid ignore_older duration", c)
		}
	}
}

func ExampleValidateTailInputConfig() {
	config := map[string]any{
		"path":             "/var/log/containers/*.log",
		"multiline.parser": "docker, cri",
		"read_from_head":   false,
		"db":               "/var/log/flb_kube.db",
		"mem_buf_limit":    "5M",
	}

	errs := ValidateTailInputConfig(config)
	if errs.HasErrors() {
		fmt.Printf("Validation errors: %v\n", errs)
	} else {
		fmt.Println("Configuration is valid")
	}
	// Output: Configuration is valid
}

func ExampleValidateTailInputConfig_withErrors() {
	config := map[string]any{
		"path":             "/var/log/syslog",
		"multiline.parser": "docker",
		"docker_mode":      true, // Conflict!
		"db.sync":          "normal", // Missing db!
	}

	errs := ValidateTailInputConfig(config)
	if errs.HasErrors() {
		fmt.Printf("Found %d validation errors\n", len(errs))
		for _, err := range errs {
			fmt.Printf("  - %s: %s\n", err.Field, err.Message)
		}
	}
}
