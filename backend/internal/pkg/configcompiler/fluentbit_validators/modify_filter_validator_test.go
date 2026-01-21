package configcompiler
import "testing"

func TestValidateModifyFilterConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]any
		expectError bool
		errorFields []string
	}{
		{
			name:        "valid set rule",
			config:      map[string]any{"set": "key1 value1"},
			expectError: false,
		},
		{
			name:        "valid add rule",
			config:      map[string]any{"add": []any{"Service1 VALUE1", "Service2 VALUE2"}},
			expectError: false,
		},
		{
			name:        "valid rename rule",
			config:      map[string]any{"rename": "OldKey NewKey"},
			expectError: false,
		},
		{
			name:        "valid remove rule",
			config:      map[string]any{"remove": "unwanted_key"},
			expectError: false,
		},
		{
			name:        "valid remove_wildcard",
			config:      map[string]any{"remove_wildcard": "Mem.*"},
			expectError: false,
		},
		{
			name:        "valid remove_regex",
			config:      map[string]any{"remove_regex": "^temp_.*$"},
			expectError: false,
		},
		{
			name:        "invalid remove_regex",
			config:      map[string]any{"remove_regex": "[invalid("},
			expectError: true,
			errorFields: []string{"remove_regex"},
		},
		{
			name: "valid complex config",
			config: map[string]any{
				"set":             []any{"key1 value1", "key2 value2"},
				"rename":          "old new",
				"remove_wildcard": "temp_",
			},
			expectError: false,
		},
		{
			name:        "missing rules",
			config:      map[string]any{},
			expectError: true,
			errorFields: []string{"modify"},
		},
		{
			name:        "set missing value",
			config:      map[string]any{"set": "keyonly"},
			expectError: true,
			errorFields: []string{"set"},
		},
		{
			name:        "rename missing second param",
			config:      map[string]any{"rename": "oldkey"},
			expectError: true,
			errorFields: []string{"rename"},
		},
		{
			name:        "valid condition key_exists",
			config:      map[string]any{"set": "k v", "key_exists": "mykey"},
			expectError: false,
		},
		{
			name:        "valid condition key_value_equals",
			config:      map[string]any{"set": "k v", "key_value_equals": "mykey myvalue"},
			expectError: false,
		},
		{
			name:        "valid condition key_value_matches with regex",
			config:      map[string]any{"set": "k v", "key_value_matches": "mykey ^prefix.*"},
			expectError: false,
		},
		{
			name:        "invalid regex in key_value_matches",
			config:      map[string]any{"set": "k v", "key_value_matches": "mykey [invalid("},
			expectError: true,
			errorFields: []string{"key_value_matches"},
		},
		{
			name:        "valid a_key_matches condition",
			config:      map[string]any{"set": "k v", "a_key_matches": "^log_.*"},
			expectError: false,
		},
		{
			name:        "invalid a_key_matches regex",
			config:      map[string]any{"set": "k v", "a_key_matches": "[bad("},
			expectError: true,
			errorFields: []string{"a_key_matches"},
		},
		{
			name: "valid generic condition format",
			config: map[string]any{
				"set":       "k v",
				"condition": []any{"Key_Exists mykey", "Key_Value_Equals status active"},
			},
			expectError: false,
		},
		{
			name: "invalid generic condition - unknown type",
			config: map[string]any{
				"set":       "k v",
				"condition": "Unknown_Condition param",
			},
			expectError: true,
			errorFields: []string{"condition"},
		},
		{
			name: "generic condition missing params",
			config: map[string]any{
				"set":       "k v",
				"condition": "Key_Value_Equals onlyone",
			},
			expectError: true,
			errorFields: []string{"condition"},
		},
		{
			name:        "unknown configuration key",
			config:      map[string]any{"set": "k v", "unknown_rule": "value"},
			expectError: true,
			errorFields: []string{"unknown_rule"},
		},
		{
			name: "real world example",
			config: map[string]any{
				"add":             []any{"Service1 SOMEVALUE", "Service2 SOMEVALUE2"},
				"rename":          []any{"Mem.free MEMFREE", "Mem.used MEMUSED"},
				"remove_wildcard": "Swap",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateModifyFilterConfig(tt.config)
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
