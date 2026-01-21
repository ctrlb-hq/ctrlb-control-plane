package configcompiler

import (
	"fmt"
	"testing"
)

func TestValidateHTTPOutputConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]any
		expectError bool
		errorFields []string // Fields expected to have errors
	}{
		{
			name:        "empty config (all defaults)",
			config:      map[string]any{},
			expectError: false,
		},
		{
			name: "valid minimal config",
			config: map[string]any{
				"host": "192.168.1.1",
				"port": 8080,
				"uri":  "/api/logs",
			},
			expectError: false,
		},
		{
			name: "valid full config",
			config: map[string]any{
				"host":                     "logs.example.com",
				"port":                     443,
				"uri":                      "/v1/logs",
				"format":                   "json_lines",
				"http_method":              "POST",
				"compress":                 "gzip",
				"json_date_format":         "iso8601",
				"json_date_key":            "timestamp",
				"workers":                  4,
				"allow_duplicated_headers": true,
				"log_response_payload":     false,
				"tls":                      true,
				"http.response_timeout":    "30s",
			},
			expectError: false,
		},
		{
			name: "invalid port - too high",
			config: map[string]any{
				"port": 70000,
			},
			expectError: true,
			errorFields: []string{"port"},
		},
		{
			name: "invalid port - negative",
			config: map[string]any{
				"port": -1,
			},
			expectError: true,
			errorFields: []string{"port"},
		},
		{
			name: "invalid format",
			config: map[string]any{
				"format": "xml",
			},
			expectError: true,
			errorFields: []string{"format"},
		},
		{
			name: "invalid http_method",
			config: map[string]any{
				"http_method": "DELETE",
			},
			expectError: true,
			errorFields: []string{"http_method"},
		},
		{
			name: "invalid compression",
			config: map[string]any{
				"compress": "lz4",
			},
			expectError: true,
			errorFields: []string{"compress"},
		},
		{
			name: "uri missing leading slash",
			config: map[string]any{
				"uri": "api/logs",
			},
			expectError: true,
			errorFields: []string{"uri"},
		},
		{
			name: "invalid host",
			config: map[string]any{
				"host": "invalid host with spaces",
			},
			expectError: true,
			errorFields: []string{"host"},
		},
		{
			name: "invalid IPv4",
			config: map[string]any{
				"host": "192.168.1.300",
			},
			expectError: true,
			errorFields: []string{"host"},
		},
		{
			name: "body_key without dollar prefix",
			config: map[string]any{
				"body_key": "message",
			},
			expectError: true,
			errorFields: []string{"body_key"},
		},
		{
			name: "body_key without headers_key",
			config: map[string]any{
				"body_key": "$message",
			},
			expectError: true,
			errorFields: []string{"body_key"},
		},
		{
			name: "valid body_key with headers_key",
			config: map[string]any{
				"body_key":    "$message",
				"headers_key": "$headers",
			},
			expectError: false,
		},
		{
			name: "headers_key without dollar prefix",
			config: map[string]any{
				"headers_key": "headers",
			},
			expectError: true,
			errorFields: []string{"headers_key"},
		},
		{
			name: "aws fields without aws_auth",
			config: map[string]any{
				"aws_region":  "us-east-1",
				"aws_service": "es",
			},
			expectError: true,
			errorFields: []string{"aws_region", "aws_service"},
		},
		{
			name: "aws_auth enabled but missing required fields",
			config: map[string]any{
				"aws_auth": true,
			},
			expectError: true,
			errorFields: []string{"aws_region", "aws_service"},
		},
		{
			name: "valid aws_auth config",
			config: map[string]any{
				"aws_auth":    true,
				"aws_region":  "us-east-1",
				"aws_service": "es",
			},
			expectError: false,
		},
		{
			name: "aws_external_id without aws_role_arn",
			config: map[string]any{
				"aws_auth":        true,
				"aws_region":      "us-east-1",
				"aws_service":     "es",
				"aws_external_id": "external-123",
			},
			expectError: true,
			errorFields: []string{"aws_external_id"},
		},
		{
			name: "http_passwd without http_user",
			config: map[string]any{
				"http_passwd": "secret",
			},
			expectError: true,
			errorFields: []string{"http_passwd"},
		},
		{
			name: "http_user without http_passwd (warning)",
			config: map[string]any{
				"http_user": "admin",
			},
			expectError: true,
			errorFields: []string{"http_user"},
		},
		{
			name: "valid basic auth",
			config: map[string]any{
				"http_user":   "admin",
				"http_passwd": "secret",
			},
			expectError: false,
		},
		{
			name: "gelf fields without gelf format",
			config: map[string]any{
				"format":                 "json",
				"gelf_short_message_key": "message",
			},
			expectError: true,
			errorFields: []string{"format"},
		},
		{
			name: "valid gelf config",
			config: map[string]any{
				"format":                 "gelf",
				"gelf_short_message_key": "message",
				"gelf_host_key":          "hostname",
			},
			expectError: false,
		},
		{
			name: "invalid proxy - https",
			config: map[string]any{
				"proxy": "https://proxy.example.com:8080",
			},
			expectError: true,
			errorFields: []string{"proxy"},
		},
		{
			name: "valid proxy",
			config: map[string]any{
				"proxy": "http://proxy.example.com:8080",
			},
			expectError: false,
		},
		{
			name: "invalid timeout format",
			config: map[string]any{
				"http.response_timeout": "sixty seconds",
			},
			expectError: true,
			errorFields: []string{"http.response_timeout"},
		},
		{
			name: "valid timeout formats",
			config: map[string]any{
				"http.response_timeout":  "60s",
				"http.read_idle_timeout": "5m",
			},
			expectError: false,
		},
		{
			name: "invalid json_date_format",
			config: map[string]any{
				"json_date_format": "rfc3339",
			},
			expectError: true,
			errorFields: []string{"json_date_format"},
		},
		{
			name: "unknown configuration key",
			config: map[string]any{
				"unknown_key": "value",
			},
			expectError: true,
			errorFields: []string{"unknown_key"},
		},
		{
			name: "valid headers as string",
			config: map[string]any{
				"header": "X-Custom-Header my-value",
			},
			expectError: false,
		},
		{
			name: "valid headers as array",
			config: map[string]any{
				"header": []any{
					"X-Key-A Value_A",
					"X-Key-B Value_B",
				},
			},
			expectError: false,
		},
		{
			name: "invalid header format - missing value",
			config: map[string]any{
				"header": "X-Custom-Header",
			},
			expectError: true,
			errorFields: []string{"header"},
		},
		{
			name: "boolean field with string value",
			config: map[string]any{
				"tls":        "on",
				"tls.verify": "false",
			},
			expectError: false,
		},
		{
			name: "invalid boolean value",
			config: map[string]any{
				"tls": "enabled",
			},
			expectError: true,
			errorFields: []string{"tls"},
		},
		{
			name: "workers unusually high (warning)",
			config: map[string]any{
				"workers": 100,
			},
			expectError: true,
			errorFields: []string{"workers"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateHTTPOutputConfig(tt.config)

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

func ExampleValidateHTTPOutputConfig() {
	config := map[string]any{
		"host":             "logs.example.com",
		"port":             443,
		"uri":              "/v1/logs",
		"format":           "json_lines",
		"http_method":      "POST",
		"compress":         "gzip",
		"json_date_format": "iso8601",
		"tls":              true,
	}

	errs := ValidateHTTPOutputConfig(config)
	if errs.HasErrors() {
		fmt.Printf("Validation errors: %v\n", errs)
	} else {
		fmt.Println("Configuration is valid")
	}
	// Output: Configuration is valid
}

func ExampleValidateHTTPOutputConfig_withErrors() {
	config := map[string]any{
		"host":        "invalid host",
		"port":        99999,
		"format":      "xml",
		"http_method": "DELETE",
	}

	errs := ValidateHTTPOutputConfig(config)
	if errs.HasErrors() {
		fmt.Printf("Found %d validation errors\n", len(errs))
		for _, err := range errs {
			fmt.Printf("  - %s: %s\n", err.Field, err.Message)
		}
	}
}
