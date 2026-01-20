package configcompiler

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type ValidationErrors []ValidationError

func (errs ValidationErrors) Error() string {
	if len(errs) == 0 {
		return ""
	}
	var msgs []string
	for _, e := range errs {
		msgs = append(msgs, e.Error())
	}
	return strings.Join(msgs, "; ")
}

func (errs ValidationErrors) HasErrors() bool {
	return len(errs) > 0
}

func getStringValue(config map[string]any, key string) (string, bool) {
	val, exists := config[key]
	if !exists {
		return "", false
	}

	switch v := val.(type) {
	case string:
		return v, true
	case int, int64, float64:
		return fmt.Sprintf("%v", v), true
	default:
		return "", true // exists but not convertible
	}
}

func getNumericValue(config map[string]any, key string) (int64, bool) {
	val, exists := config[key]
	if !exists {
		return 0, false
	}

	switch v := val.(type) {
	case int:
		return int64(v), true
	case int64:
		return v, true
	case float64:
		return int64(v), true
	case string:
		if num, err := strconv.ParseInt(v, 10, 64); err == nil {
			return num, true
		}
	}
	return 0, true // exists but not convertible
}

func getBoolValue(config map[string]any, key string) bool {
	val, exists := config[key]
	if !exists {
		return false
	}

	switch v := val.(type) {
	case bool:
		return v
	case string:
		lower := strings.ToLower(v)
		return lower == "true" || lower == "on" || lower == "yes" || lower == "1"
	case int, int64, float64:
		return fmt.Sprintf("%v", v) == "1"
	}
	return false
}

func isBooleanLike(val any) bool {
	switch v := val.(type) {
	case bool:
		return true
	case string:
		lower := strings.ToLower(v)
		validBools := []string{"true", "false", "on", "off", "yes", "no", "1", "0"}
		return containsString(validBools, lower)
	case int, int64, float64:
		s := fmt.Sprintf("%v", v)
		return s == "0" || s == "1"
	}
	return false
}

func isValidDuration(val string) bool {
	if val == "" {
		return false
	}

	// Pattern for duration: number followed by time unit (s, m, h, d)
	pattern := `^(\d+(\.\d+)?)(s|ms|m|h|d)?$`
	matched, _ := regexp.MatchString(pattern, val)
	return matched
}

func isValidHeaderFormat(header string) bool {
	// Header should be "Key Value" with at least one space
	parts := strings.SplitN(header, " ", 2)
	if len(parts) < 2 {
		return false
	}

	// Key should not be empty and should not contain invalid characters
	key := parts[0]
	if key == "" || strings.ContainsAny(key, " \t\n\r:") {
		return false
	}

	return true
}

func containsString(slice []string, str string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, str) {
			return true
		}
	}
	return false
}

func containsStringCI(slice []string, str string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, str) {
			return true
		}
	}
	return false
}

// extractRules extracts rules from various input types (string, []string, []any)
func extractRules(val any) []string {
	switch v := val.(type) {
	case string:
		return []string{v}
	case []string:
		return v
	case []any:
		var rules []string
		for _, item := range v {
			if s, ok := item.(string); ok {
				rules = append(rules, s)
			}
		}
		return rules
	}
	return nil
}

// hasKey checks if a key exists in config (case-insensitive)
func hasKey(config map[string]any, key string) bool {
	for k := range config {
		if strings.EqualFold(k, key) {
			return true
		}
	}
	return false
}

// isValidUnitSize validates Fluent Bit unit size format (e.g., "32k", "1M", "256")
func isValidUnitSize(val string) bool {
	if val == "" {
		return false
	}

	// Pattern: number optionally followed by unit (k, K, m, M, g, G)
	pattern := `^(\d+(\.\d+)?)\s*([kKmMgG])?[bB]?$`
	matched, _ := regexp.MatchString(pattern, strings.TrimSpace(val))
	return matched
}

// parseUnitSize converts unit size string to bytes (approximate, for comparison)
func parseUnitSize(val string) int64 {
	val = strings.TrimSpace(strings.ToLower(val))
	val = strings.TrimSuffix(val, "b")

	multiplier := int64(1)
	if strings.HasSuffix(val, "k") {
		multiplier = 1024
		val = strings.TrimSuffix(val, "k")
	} else if strings.HasSuffix(val, "m") {
		multiplier = 1024 * 1024
		val = strings.TrimSuffix(val, "m")
	} else if strings.HasSuffix(val, "g") {
		multiplier = 1024 * 1024 * 1024
		val = strings.TrimSuffix(val, "g")
	}

	num, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0
	}

	return int64(num * float64(multiplier))
}

// isValidIgnoreOlderDuration validates the ignore_older duration format (m, h, d)
func isValidIgnoreOlderDuration(val string) bool {
	if val == "" {
		return false
	}

	// Pattern: number followed by m, h, or d
	pattern := `^(\d+(\.\d+)?)\s*[mhd]$`
	matched, _ := regexp.MatchString(pattern, strings.TrimSpace(strings.ToLower(val)))
	return matched
}
