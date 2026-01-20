package configcompiler

import (
	"fmt"
	"regexp"
	"strings"
)

// Rules that require exactly 2 parameters (KEY VALUE)
var modifyTwoParamRules = map[string]bool{
	"set":         true,
	"add":         true,
	"rename":      true,
	"hard_rename": true,
	"copy":        true,
	"hard_copy":   true,
}

// Rules that require exactly 1 parameter (KEY only)
var modifyOneParamRules = map[string]bool{
	"remove":          true,
	"remove_wildcard": true,
	"remove_regex":    true,
	"move_to_start":   true,
	"move_to_end":     true,
}

// Conditions that require exactly 1 parameter
var modifyOneParamConditions = map[string]bool{
	"key_exists":         true,
	"key_does_not_exist": true,
	"a_key_matches":      true,
	"no_key_matches":     true,
}

// Conditions that require exactly 2 parameters
var modifyTwoParamConditions = map[string]bool{
	"key_value_equals":                          true,
	"key_value_does_not_equal":                  true,
	"key_value_matches":                         true,
	"key_value_does_not_match":                  true,
	"matching_keys_have_matching_values":        true,
	"matching_keys_do_not_have_matching_values": true,
}

// All known keys for Modify filter
var modifyFilterKnownKeys map[string]bool

func init() {
	modifyFilterKnownKeys = make(map[string]bool)
	for k := range modifyTwoParamRules {
		modifyFilterKnownKeys[k] = true
	}
	for k := range modifyOneParamRules {
		modifyFilterKnownKeys[k] = true
	}
	for k := range modifyOneParamConditions {
		modifyFilterKnownKeys[k] = true
	}
	for k := range modifyTwoParamConditions {
		modifyFilterKnownKeys[k] = true
	}
	modifyFilterKnownKeys["condition"] = true // Generic condition key
}

func ValidateModifyFilterConfig(config map[string]any) ValidationErrors {
	var errors ValidationErrors

	hasRules := false

	for key, val := range config {
		lowerKey := strings.ToLower(key)

		// Check if it's a known key
		if !modifyFilterKnownKeys[lowerKey] {
			errors = append(errors, ValidationError{Field: key, Message: "unknown configuration key"})
			continue
		}

		// Validate two-param rules (KEY VALUE)
		if modifyTwoParamRules[lowerKey] {
			hasRules = true
			errors = append(errors, validateModifyEntries(key, val, 2)...)
		}

		// Validate one-param rules (KEY only)
		if modifyOneParamRules[lowerKey] {
			hasRules = true
			// For regex-based rules, validate regex syntax
			if lowerKey == "remove_regex" {
				errors = append(errors, validateModifyRegexEntries(key, val)...)
			} else {
				errors = append(errors, validateModifyEntries(key, val, 1)...)
			}
		}

		// Validate conditions
		if modifyOneParamConditions[lowerKey] {
			if lowerKey == "a_key_matches" || lowerKey == "no_key_matches" {
				errors = append(errors, validateModifyRegexEntries(key, val)...)
			} else {
				errors = append(errors, validateModifyEntries(key, val, 1)...)
			}
		}

		if modifyTwoParamConditions[lowerKey] {
			errors = append(errors, validateModifyConditionWithRegex(key, val, lowerKey)...)
		}

		// Handle generic "condition" key
		if lowerKey == "condition" {
			errors = append(errors, validateGenericCondition(key, val)...)
		}
	}

	if !hasRules {
		errors = append(errors, ValidationError{Field: "modify", Message: "at least one rule (Set, Add, Remove, Rename, etc.) is required"})
	}

	return errors
}

func validateModifyEntries(key string, val any, expectedParams int) ValidationErrors {
	var errors ValidationErrors
	entries := extractRules(val)

	for i, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			errors = append(errors, ValidationError{Field: key, Message: fmt.Sprintf("entry %d: cannot be empty", i+1)})
			continue
		}

		parts := strings.Fields(entry)
		if len(parts) < expectedParams {
			if expectedParams == 2 {
				errors = append(errors, ValidationError{Field: key, Message: fmt.Sprintf("entry %d: requires KEY and VALUE (space-separated)", i+1)})
			} else {
				errors = append(errors, ValidationError{Field: key, Message: fmt.Sprintf("entry %d: requires KEY parameter", i+1)})
			}
		}
	}

	return errors
}

func validateModifyRegexEntries(key string, val any) ValidationErrors {
	var errors ValidationErrors
	entries := extractRules(val)

	for i, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			errors = append(errors, ValidationError{Field: key, Message: fmt.Sprintf("entry %d: cannot be empty", i+1)})
			continue
		}

		if _, err := regexp.Compile(entry); err != nil {
			errors = append(errors, ValidationError{Field: key, Message: fmt.Sprintf("entry %d: invalid regex '%s': %v", i+1, entry, err)})
		}
	}

	return errors
}

func validateModifyConditionWithRegex(key string, val any, condType string) ValidationErrors {
	var errors ValidationErrors
	entries := extractRules(val)

	needsRegexValue := strings.Contains(condType, "matches")

	for i, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			errors = append(errors, ValidationError{Field: key, Message: fmt.Sprintf("entry %d: cannot be empty", i+1)})
			continue
		}

		parts := strings.SplitN(entry, " ", 2)
		if len(parts) < 2 {
			errors = append(errors, ValidationError{Field: key, Message: fmt.Sprintf("entry %d: requires KEY and VALUE parameters", i+1)})
			continue
		}

		// Validate regex in value if needed
		if needsRegexValue {
			if _, err := regexp.Compile(strings.TrimSpace(parts[1])); err != nil {
				errors = append(errors, ValidationError{Field: key, Message: fmt.Sprintf("entry %d: invalid regex in VALUE: %v", i+1, err)})
			}
		}
	}

	return errors
}

func validateGenericCondition(key string, val any) ValidationErrors {
	var errors ValidationErrors
	entries := extractRules(val)

	for i, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			errors = append(errors, ValidationError{Field: key, Message: fmt.Sprintf("entry %d: cannot be empty", i+1)})
			continue
		}

		parts := strings.Fields(entry)
		if len(parts) < 1 {
			errors = append(errors, ValidationError{Field: key, Message: fmt.Sprintf("entry %d: requires condition type", i+1)})
			continue
		}

		condType := strings.ToLower(parts[0])

		// Check if valid condition type
		if !modifyOneParamConditions[condType] && !modifyTwoParamConditions[condType] {
			errors = append(errors, ValidationError{Field: key, Message: fmt.Sprintf("entry %d: unknown condition type '%s'", i+1, parts[0])})
			continue
		}

		// Validate parameter count
		if modifyOneParamConditions[condType] && len(parts) < 2 {
			errors = append(errors, ValidationError{Field: key, Message: fmt.Sprintf("entry %d: condition '%s' requires 1 parameter", i+1, parts[0])})
		}
		if modifyTwoParamConditions[condType] && len(parts) < 3 {
			errors = append(errors, ValidationError{Field: key, Message: fmt.Sprintf("entry %d: condition '%s' requires 2 parameters", i+1, parts[0])})
		}
	}

	return errors
}
