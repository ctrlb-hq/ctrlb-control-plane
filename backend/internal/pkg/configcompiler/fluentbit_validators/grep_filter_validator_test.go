package configcompiler

import (
	"testing"
)

func TestValidateGrepFilterConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]any
		expectError bool
		errorFields []string
	}{
		{
			name:        "valid single regex",
			config:      map[string]any{"regex": "log aa"},
			expectError: false,
		},
		{
			name:        "valid single exclude",
			config:      map[string]any{"exclude": "log error"},
			expectError: false,
		},
		{
			name: "valid multiple regex with logical_op",
			config: map[string]any{
				"regex":      []any{"value something", "value error"},
				"logical_op": "or",
			},
			expectError: false,
		},
		{
			name: "valid nested key regex",
			config: map[string]any{
				"exclude": "$kubernetes['labels']['app'] myapp",
			},
			expectError: false,
		},
		{
			name:        "missing regex and exclude",
			config:      map[string]any{},
			expectError: true,
			errorFields: []string{"regex/exclude"},
		},
		{
			name:        "invalid regex syntax",
			config:      map[string]any{"regex": "log [invalid("},
			expectError: true,
			errorFields: []string{"regex"},
		},
		{
			name:        "missing KEY in rule",
			config:      map[string]any{"regex": "onlyvalue"},
			expectError: true,
			errorFields: []string{"regex"},
		},
		{
			name: "invalid logical_op value",
			config: map[string]any{
				"regex":      "log aa",
				"logical_op": "xor",
			},
			expectError: true,
			errorFields: []string{"logical_op"},
		},
		{
			name: "mixing regex and exclude with non-legacy logical_op",
			config: map[string]any{
				"regex":      "log aa",
				"exclude":    "log bb",
				"logical_op": "and",
			},
			expectError: true,
			errorFields: []string{"logical_op"},
		},
		{
			name: "mixing regex and exclude with legacy is ok",
			config: map[string]any{
				"regex":      "log aa",
				"exclude":    "log bb",
				"logical_op": "legacy",
			},
			expectError: false,
		},
		{
			name:        "unknown key",
			config:      map[string]any{"regex": "log aa", "unknown": "value"},
			expectError: true,
			errorFields: []string{"unknown"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateGrepFilterConfig(tt.config)
			if tt.expectError && !errs.HasErrors() {
				t.Errorf("expected errors but got none")
			}
			if !tt.expectError && errs.HasErrors() {
				t.Errorf("expected no errors but got: %v", errs)
			}
			for _, field := range tt.errorFields {
				found := false
				for _, err := range errs {
					if err.Field == field {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error for field %q", field)
				}
			}
		})
	}
}
