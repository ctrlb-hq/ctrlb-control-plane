package configcompiler

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type HTTPOutputValidator struct {
	errors ValidationErrors
}

var httpOutputKnownKeys = map[string]bool{
	"allow_duplicated_headers": true,
	"aws_auth":                 true,
	"aws_external_id":          true,
	"aws_profile":              true,
	"aws_region":               true,
	"aws_role_arn":             true,
	"aws_service":              true,
	"aws_sts_endpoint":         true,
	"body_key":                 true,
	"compress":                 true,
	"format":                   true,
	"gelf_full_message_key":    true,
	"gelf_host_key":            true,
	"gelf_level_key":           true,
	"gelf_short_message_key":   true,
	"gelf_timestamp_key":       true,
	"header":                   true,
	"header_tag":               true,
	"headers_key":              true,
	"host":                     true,
	"http.read_idle_timeout":   true,
	"http.response_timeout":    true,
	"http_method":              true,
	"http_passwd":              true,
	"http_user":                true,
	"json_date_format":         true,
	"json_date_key":            true,
	"log_response_payload":     true,
	"port":                     true,
	"proxy":                    true,
	"uri":                      true,
	"workers":                  true,
	// TLS/SSL common keys
	"tls":            true,
	"tls.verify":     true,
	"tls.debug":      true,
	"tls.ca_file":    true,
	"tls.ca_path":    true,
	"tls.crt_file":   true,
	"tls.key_file":   true,
	"tls.key_passwd": true,
	"tls.vhost":      true,
}

var (
	allowedFormats        = []string{"gelf", "json", "json_stream", "json_lines", "msgpack"}
	allowedCompression    = []string{"gzip", "snappy", "zstd"}
	allowedHTTPMethods    = []string{"POST", "PUT"}
	allowedJSONDateFormat = []string{"double", "epoch", "epoch_ms", "iso8601", "java_sql_timestamp"}
)

func ValidateHTTPOutputConfig(config map[string]any) ValidationErrors {
	v := &HTTPOutputValidator{}

	v.checkUnknownKeys(config)

	v.validateHost(config)
	v.validatePort(config)
	v.validateURI(config)
	v.validateHTTPMethod(config)
	v.validateFormat(config)
	v.validateCompression(config)
	v.validateJSONDateFormat(config)
	v.validateWorkers(config)
	v.validateBooleanFields(config)
	v.validateTimeoutFields(config)
	v.validateProxy(config)
	v.validateHeaders(config)
	v.validateBodyKey(config)
	v.validateHeadersKey(config)
	v.validateAWSAuth(config)
	v.validateBasicAuth(config)
	v.validateGELFFields(config)

	return v.errors
}

func (v *HTTPOutputValidator) addError(field, message string) {
	v.errors = append(v.errors, ValidationError{Field: field, Message: message})
}

func (v *HTTPOutputValidator) checkUnknownKeys(config map[string]any) {
	for key := range config {
		normalizedKey := strings.ToLower(key)
		if !httpOutputKnownKeys[normalizedKey] {
			// This is a warning - unknown keys might be valid in newer versions
			// but we flag them for awareness
			v.addError(key, "unknown configuration key (may be unsupported)")
		}
	}
}

func (v *HTTPOutputValidator) validateHost(config map[string]any) {
	host, exists := getStringValue(config, "host")
	if !exists {
		return // Default is 127.0.0.1
	}

	if host == "" {
		v.addError("host", "cannot be empty string")
		return
	}

	// Basic validation - should be a valid hostname or IP
	// Allow IPv4, IPv6 (in brackets), or hostname
	ipv4Pattern := `^(\d{1,3}\.){3}\d{1,3}$`
	ipv6Pattern := `^\[([0-9a-fA-F:]+)\]$`
	hostnamePattern := `^[a-zA-Z0-9]([a-zA-Z0-9\-]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]*[a-zA-Z0-9])?)*$`

	ipv4Match, _ := regexp.MatchString(ipv4Pattern, host)
	ipv6Match, _ := regexp.MatchString(ipv6Pattern, host)
	hostnameMatch, _ := regexp.MatchString(hostnamePattern, host)

	if !ipv4Match && !ipv6Match && !hostnameMatch {
		v.addError("host", "must be a valid IP address or hostname")
	}

	// Additional IPv4 octet validation
	if ipv4Match {
		parts := strings.Split(host, ".")
		for _, part := range parts {
			num, _ := strconv.Atoi(part)
			if num > 255 {
				v.addError("host", "invalid IPv4 address (octet > 255)")
				break
			}
		}
	}
}

func (v *HTTPOutputValidator) validatePort(config map[string]any) {
	port, exists := getNumericValue(config, "port")
	if !exists {
		return // Default is 80
	}

	if port < 1 || port > 65535 {
		v.addError("port", "must be between 1 and 65535")
	}
}

func (v *HTTPOutputValidator) validateURI(config map[string]any) {
	uri, exists := getStringValue(config, "uri")
	if !exists {
		return // Default is /
	}

	if uri == "" {
		v.addError("uri", "cannot be empty string (use '/' for root)")
		return
	}

	if !strings.HasPrefix(uri, "/") {
		v.addError("uri", "must start with '/'")
	}

	// Check for invalid characters in URI
	if strings.ContainsAny(uri, " \t\n\r") {
		v.addError("uri", "cannot contain whitespace characters")
	}
}

func (v *HTTPOutputValidator) validateHTTPMethod(config map[string]any) {
	method, exists := getStringValue(config, "http_method")
	if !exists {
		return // Default is POST
	}

	upperMethod := strings.ToUpper(method)
	if !containsString(allowedHTTPMethods, upperMethod) {
		v.addError("http_method", fmt.Sprintf("must be one of: %s", strings.Join(allowedHTTPMethods, ", ")))
	}
}

func (v *HTTPOutputValidator) validateFormat(config map[string]any) {
	format, exists := getStringValue(config, "format")
	if !exists {
		return // Default is json
	}

	lowerFormat := strings.ToLower(format)
	if !containsString(allowedFormats, lowerFormat) {
		v.addError("format", fmt.Sprintf("must be one of: %s", strings.Join(allowedFormats, ", ")))
	}
}

func (v *HTTPOutputValidator) validateCompression(config map[string]any) {
	compress, exists := getStringValue(config, "compress")
	if !exists {
		return // Default is none
	}

	lowerCompress := strings.ToLower(compress)
	if !containsString(allowedCompression, lowerCompress) {
		v.addError("compress", fmt.Sprintf("must be one of: %s", strings.Join(allowedCompression, ", ")))
	}
}

func (v *HTTPOutputValidator) validateJSONDateFormat(config map[string]any) {
	dateFormat, exists := getStringValue(config, "json_date_format")
	if !exists {
		return // Default is double
	}

	lowerFormat := strings.ToLower(dateFormat)
	if !containsString(allowedJSONDateFormat, lowerFormat) {
		v.addError("json_date_format", fmt.Sprintf("must be one of: %s", strings.Join(allowedJSONDateFormat, ", ")))
	}
}

func (v *HTTPOutputValidator) validateWorkers(config map[string]any) {
	workers, exists := getNumericValue(config, "workers")
	if !exists {
		return // Default is 2
	}

	if workers < 0 {
		v.addError("workers", "must be a non-negative integer")
	}

	// Warn if unusually high
	if workers > 64 {
		v.addError("workers", "unusually high value (> 64), verify this is intentional")
	}
}

func (v *HTTPOutputValidator) validateBooleanFields(config map[string]any) {
	boolFields := []string{
		"allow_duplicated_headers",
		"aws_auth",
		"log_response_payload",
		"tls",
		"tls.verify",
	}

	for _, field := range boolFields {
		if val, exists := config[field]; exists {
			if !isBooleanLike(val) {
				v.addError(field, "must be a boolean value (true/false, on/off, yes/no)")
			}
		}
	}
}

func (v *HTTPOutputValidator) validateTimeoutFields(config map[string]any) {
	timeoutFields := []string{"http.read_idle_timeout", "http.response_timeout"}

	for _, field := range timeoutFields {
		if val, exists := getStringValue(config, field); exists {
			if !isValidDuration(val) {
				v.addError(field, "must be a valid duration (e.g., '60s', '5m', '1h')")
			}
		}
	}
}

func (v *HTTPOutputValidator) validateProxy(config map[string]any) {
	proxy, exists := getStringValue(config, "proxy")
	if !exists {
		return
	}

	if proxy == "" {
		v.addError("proxy", "cannot be empty string")
		return
	}

	// Parse and validate proxy URL
	parsed, err := url.Parse(proxy)
	if err != nil {
		v.addError("proxy", fmt.Sprintf("invalid URL format: %v", err))
		return
	}

	if parsed.Scheme != "http" {
		v.addError("proxy", "only 'http' scheme is supported (not https)")
	}

	if parsed.Host == "" {
		v.addError("proxy", "must include host (expected format: http://HOST:PORT)")
	}
}

func (v *HTTPOutputValidator) validateHeaders(config map[string]any) {
	header, exists := config["header"]
	if !exists {
		return
	}

	switch h := header.(type) {
	case string:
		// Single header - should be "Key Value" format
		if !isValidHeaderFormat(h) {
			v.addError("header", "must be in 'Key Value' format")
		}
	case []any:
		// Multiple headers
		for i, item := range h {
			if str, ok := item.(string); ok {
				if !isValidHeaderFormat(str) {
					v.addError("header", fmt.Sprintf("item %d must be in 'Key Value' format", i))
				}
			} else {
				v.addError("header", fmt.Sprintf("item %d must be a string", i))
			}
		}
	case []string:
		for i, str := range h {
			if !isValidHeaderFormat(str) {
				v.addError("header", fmt.Sprintf("item %d must be in 'Key Value' format", i))
			}
		}
	default:
		v.addError("header", "must be a string or array of strings")
	}
}

func (v *HTTPOutputValidator) validateBodyKey(config map[string]any) {
	bodyKey, exists := getStringValue(config, "body_key")
	if !exists {
		return
	}

	if !strings.HasPrefix(bodyKey, "$") {
		v.addError("body_key", "must be prefixed with '$'")
	}

	// If body_key is present, headers_key should also be present (for content-type)
	if _, headersKeyExists := config["headers_key"]; !headersKeyExists {
		v.addError("body_key", "when body_key is set, headers_key should also be set to specify content-type")
	}
}

func (v *HTTPOutputValidator) validateHeadersKey(config map[string]any) {
	headersKey, exists := getStringValue(config, "headers_key")
	if !exists {
		return
	}

	if !strings.HasPrefix(headersKey, "$") {
		v.addError("headers_key", "must be prefixed with '$'")
	}
}

func (v *HTTPOutputValidator) validateAWSAuth(config map[string]any) {
	awsAuth := getBoolValue(config, "aws_auth")
	if !awsAuth {
		// Check if AWS-specific fields are set without aws_auth enabled
		awsFields := []string{"aws_region", "aws_service", "aws_role_arn", "aws_external_id", "aws_profile", "aws_sts_endpoint"}
		for _, field := range awsFields {
			if _, exists := config[field]; exists {
				v.addError(field, "aws_auth must be enabled to use this option")
			}
		}
		return
	}

	// When aws_auth is enabled, aws_region and aws_service are typically required
	if _, exists := config["aws_region"]; !exists {
		v.addError("aws_region", "required when aws_auth is enabled")
	}

	if _, exists := config["aws_service"]; !exists {
		v.addError("aws_service", "required when aws_auth is enabled")
	}

	// If aws_role_arn is set, validate its format
	if roleArn, exists := getStringValue(config, "aws_role_arn"); exists {
		if !strings.HasPrefix(roleArn, "arn:aws:iam::") {
			v.addError("aws_role_arn", "must be a valid AWS IAM Role ARN (arn:aws:iam::...)")
		}
	}

	// If aws_external_id is set, aws_role_arn should also be set
	if _, extIDExists := config["aws_external_id"]; extIDExists {
		if _, roleExists := config["aws_role_arn"]; !roleExists {
			v.addError("aws_external_id", "aws_role_arn must be set when using aws_external_id")
		}
	}
}

func (v *HTTPOutputValidator) validateBasicAuth(config map[string]any) {
	_, userExists := config["http_user"]
	_, passwdExists := config["http_passwd"]

	// If password is set, user must also be set
	if passwdExists && !userExists {
		v.addError("http_passwd", "http_user must be set when using http_passwd")
	}

	// Warning: if only user is set without password
	if userExists && !passwdExists {
		v.addError("http_user", "http_passwd should typically be set when using http_user")
	}
}

func (v *HTTPOutputValidator) validateGELFFields(config map[string]any) {
	gelfFields := []string{
		"gelf_full_message_key",
		"gelf_host_key",
		"gelf_level_key",
		"gelf_short_message_key",
		"gelf_timestamp_key",
	}

	hasGELFFields := false
	for _, field := range gelfFields {
		if _, exists := config[field]; exists {
			hasGELFFields = true
			break
		}
	}

	if hasGELFFields {
		// Check if format is set to gelf
		format, _ := getStringValue(config, "format")
		if strings.ToLower(format) != "gelf" {
			v.addError("format", "should be 'gelf' when using gelf_* configuration options")
		}
	}
}
