package configcompiler

import (
	"fmt"
	"strings"
)

var grepFilterKnownKeys = map[string]bool{
	"regex":      true,
	"exclude":    true,
	"logical_op": true,
}

var allowedLogicalOp = []string{"and", "or", "legacy"}

func ValidateGrepFilterConfig(config map[string]any) ValidationErrors {
	var errors ValidationErrors

	// Check for unknown keys
	for key := range config {
		if !grepFilterKnownKeys[strings.ToLower(key)] {
			errors = append(errors, ValidationError{Field: key, Message: "unknown configuration key"})
		}
	}

	// At least one of regex or exclude must be present
	hasRegex := hasKey(config, "regex")
	hasExclude := hasKey(config, "exclude")

	if !hasRegex && !hasExclude {
		errors = append(errors, ValidationError{Field: "regex/exclude", Message: "at least one 'regex' or 'exclude' rule is required"})
	}

	// Validate regex entries
	if hasRegex {
		errors = append(errors, validateGrepRules(config, "regex")...)
	}

	// Validate exclude entries
	if hasExclude {
		errors = append(errors, validateGrepRules(config, "exclude")...)
	}

	// Validate logical_op
	logicalOp, hasLogicalOp := getStringValue(config, "logical_op")
	if hasLogicalOp {
		if logicalOp == "" {
			errors = append(errors, ValidationError{Field: "logical_op", Message: "cannot be empty if specified"})
		} else if !containsStringCI(allowedLogicalOp, logicalOp) {
			errors = append(errors, ValidationError{Field: "logical_op", Message: fmt.Sprintf("must be one of: %s", strings.Join(allowedLogicalOp, ", "))})
		}

		// When logical_op is set (and not legacy), mixing regex and exclude is an error
		if logicalOp != "" && !strings.EqualFold(logicalOp, "legacy") {
			if hasRegex && hasExclude {
				errors = append(errors, ValidationError{Field: "logical_op", Message: "when set to 'and' or 'or', cannot mix 'regex' and 'exclude' rules"})
			}
		}
	}

	return errors
}

func validateGrepRules(config map[string]any, key string) ValidationErrors {
	var errors ValidationErrors

	val, exists := config[key]
	if !exists {
		return errors
	}

	rules := extractRules(val)

	for i, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			errors = append(errors, ValidationError{Field: key, Message: fmt.Sprintf("rule %d: cannot be empty", i+1)})
			continue
		}

		// Rule format: KEY REGEX (space-separated)
		parts := strings.SplitN(rule, " ", 2)
		if len(parts) < 2 {
			errors = append(errors, ValidationError{Field: key, Message: fmt.Sprintf("rule %d: must be in 'KEY REGEX' format (space-separated)", i+1)})
			continue
		}

		keyPart := strings.TrimSpace(parts[0])
		regexPart := strings.TrimSpace(parts[1])

		if keyPart == "" {
			errors = append(errors, ValidationError{Field: key, Message: fmt.Sprintf("rule %d: KEY cannot be empty", i+1)})
		}

		if regexPart == "" {
			errors = append(errors, ValidationError{Field: key, Message: fmt.Sprintf("rule %d: REGEX cannot be empty", i+1)})
		}
		// else {
		// 	// Validate regex syntax
		// 	if _, err := regexp.Compile(regexPart); err != nil {
		// 		errors = append(errors, ValidationError{Field: key, Message: fmt.Sprintf("rule %d: invalid regex '%s': %v", i+1, regexPart, err)})
		// 	}
		// }
	}

	return errors
}
