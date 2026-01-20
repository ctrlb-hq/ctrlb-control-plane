package configcompiler

import (
	"fmt"
	"regexp"
	"strings"
)

var tailInputKnownKeys = map[string]bool{
	// Core parameters
	"path":                                  true,
	"tag":                                   true,
	"tag_regex":                             true,
	"exclude_path":                          true,
	"key":                                   true,
	"read_from_head":                        true,
	"read_newly_discovered_files_from_head": true,
	"refresh_interval":                      true,
	"rotate_wait":                           true,
	"ignore_older":                          true,
	"ignore_active_older_files":             true,
	"skip_long_lines":                       true,
	"skip_empty_lines":                      true,
	"exit_on_eof":                           true,

	// Buffer parameters
	"buffer_chunk_size": true,
	"buffer_max_size":   true,
	"mem_buf_limit":     true,

	// Database parameters
	"db":                  true,
	"db.sync":             true,
	"db.locking":          true,
	"db.journal_mode":     true,
	"db.compare_filename": true,

	// Parser parameters
	"parser": true,

	// Multiline core parameters
	"multiline.parser": true,

	// Old multiline parameters
	"multiline":        true,
	"multiline_flush":  true,
	"parser_firstline": true,
	// parser_N handled dynamically

	// Docker mode parameters
	"docker_mode":        true,
	"docker_mode_flush":  true,
	"docker_mode_parser": true,

	// Encoding parameters
	"unicode.encoding": true,
	"generic.encoding": true,

	// File monitoring parameters
	"inotify_watcher":              true,
	"file_cache_advise":            true,
	"static_batch_size":            true,
	"event_batch_size":             true,
	"progress_check_interval":      true,
	"progress_check_interval_nsec": true,
	"watcher_interval":             true,

	// Path/offset key parameters
	"path_key":   true,
	"offset_key": true,

	// Threading parameters
	"threaded":                    true,
	"thread.ring_buffer.capacity": true,
	"thread.ring_buffer.window":   true,

	// Long line handling
	"truncate_long_lines": true,
}

var (
	allowedDBSync          = []string{"extra", "full", "normal", "off"}
	allowedDBJournalMode   = []string{"delete", "truncate", "persist", "memory", "wal", "off"}
	allowedUnicodeEncoding = []string{"utf-16le", "utf-16be", "auto"}
	allowedGenericEncoding = []string{
		"shiftjis", "sjis", "cp932", "windows-31j",
		"gb18030",
		"gbk", "cp936",
		"uhc", "cp949", "windows-949",
		"big5", "cp950",
		"win1250", "cp1250",
		"win1251", "cp1251",
		"win1252", "cp1252",
		"win1253", "cp1253",
		"win1254", "cp1254",
		"win1255", "cp1255",
		"win1256", "cp1256",
		"win866", "cp866",
		"win874", "cp874",
	}
)

type TailInputValidator struct {
	errors ValidationErrors
}

func ValidateTailInputConfig(config map[string]any) ValidationErrors {
	v := &TailInputValidator{}

	v.checkUnknownKeys(config)

	v.validatePath(config)
	v.validateBufferSettings(config)
	v.validateDatabaseSettings(config)
	v.validateParserSettings(config)
	v.validateMultilineSettings(config)
	v.validateDockerModeSettings(config)
	v.validateEncodingSettings(config)
	v.validateFileMonitoringSettings(config)
	v.validateThreadingSettings(config)
	v.validateNumericFields(config)
	v.validateBooleanFields(config)
	v.validateTagSettings(config)

	return v.errors
}

func (v *TailInputValidator) addError(field, message string) {
	v.errors = append(v.errors, ValidationError{Field: field, Message: message})
}

func (v *TailInputValidator) checkUnknownKeys(config map[string]any) {
	parserNPattern := regexp.MustCompile(`^parser_\d+$`)

	for key := range config {
		normalizedKey := strings.ToLower(key)

		// Check for parser_N pattern (parser_1, parser_2, etc.)
		if parserNPattern.MatchString(normalizedKey) {
			continue
		}

		if !tailInputKnownKeys[normalizedKey] {
			v.addError(key, "unknown configuration key (may be unsupported)")
		}
	}
}

func (v *TailInputValidator) validatePath(config map[string]any) {
	path, exists := getStringValue(config, "path")
	if !exists {
		v.addError("path", "required parameter is missing")
		return
	}

	if path == "" {
		v.addError("path", "cannot be empty")
		return
	}

	// Path can contain wildcards and multiple patterns separated by commas
	// Basic validation: should not contain invalid characters for file paths
	if strings.ContainsAny(path, "\n\r\t") {
		v.addError("path", "cannot contain newline or tab characters")
	}

	// Check exclude_path if present
	if excludePath, exists := getStringValue(config, "exclude_path"); exists {
		if excludePath == "" {
			v.addError("exclude_path", "cannot be empty string if specified")
		}
	}
}

// validateBufferSettings validates buffer-related parameters
func (v *TailInputValidator) validateBufferSettings(config map[string]any) {
	// buffer_chunk_size
	if bufferChunk, exists := getStringValue(config, "buffer_chunk_size"); exists {
		if !isValidUnitSize(bufferChunk) {
			v.addError("buffer_chunk_size", "must be a valid unit size (e.g., '32k', '1M', '256')")
		}
	}

	// buffer_max_size
	if bufferMax, exists := getStringValue(config, "buffer_max_size"); exists {
		if !isValidUnitSize(bufferMax) {
			v.addError("buffer_max_size", "must be a valid unit size (e.g., '32k', '1M', '256')")
		}
	}

	// mem_buf_limit
	if memBufLimit, exists := getStringValue(config, "mem_buf_limit"); exists {
		if !isValidUnitSize(memBufLimit) {
			v.addError("mem_buf_limit", "must be a valid unit size (e.g., '5M', '10M')")
		}
	}

	// Logical check: buffer_max_size should be >= buffer_chunk_size
	bufferChunk, chunkExists := getStringValue(config, "buffer_chunk_size")
	bufferMax, maxExists := getStringValue(config, "buffer_max_size")
	if chunkExists && maxExists {
		chunkBytes := parseUnitSize(bufferChunk)
		maxBytes := parseUnitSize(bufferMax)
		if chunkBytes > 0 && maxBytes > 0 && maxBytes < chunkBytes {
			v.addError("buffer_max_size", "should be greater than or equal to buffer_chunk_size")
		}
	}

	// static_batch_size
	if staticBatch, exists := getStringValue(config, "static_batch_size"); exists {
		if !isValidUnitSize(staticBatch) {
			v.addError("static_batch_size", "must be a valid unit size (e.g., '50M')")
		}
	}

	// event_batch_size
	if eventBatch, exists := getStringValue(config, "event_batch_size"); exists {
		if !isValidUnitSize(eventBatch) {
			v.addError("event_batch_size", "must be a valid unit size (e.g., '50M')")
		}
	}
}

func (v *TailInputValidator) validateDatabaseSettings(config map[string]any) {
	dbPath, dbExists := getStringValue(config, "db")

	// If db is not set, db.* options should not be used
	dbOptions := []string{"db.sync", "db.locking", "db.journal_mode", "db.compare_filename"}
	for _, opt := range dbOptions {
		if _, exists := config[opt]; exists && !dbExists {
			v.addError(opt, "requires 'db' parameter to be set")
		}
	}

	if dbExists {
		if dbPath == "" {
			v.addError("db", "cannot be empty string if specified")
		}

		// Validate db.sync
		if dbSync, exists := getStringValue(config, "db.sync"); exists {
			if !containsStringCI(allowedDBSync, dbSync) {
				v.addError("db.sync", fmt.Sprintf("must be one of: %s", strings.Join(allowedDBSync, ", ")))
			}
		}

		// Validate db.journal_mode
		if journalMode, exists := getStringValue(config, "db.journal_mode"); exists {
			if !containsStringCI(allowedDBJournalMode, journalMode) {
				v.addError("db.journal_mode", fmt.Sprintf("must be one of: %s", strings.Join(allowedDBJournalMode, ", ")))
			}
		}

		// Validate db.locking (boolean)
		if val, exists := config["db.locking"]; exists {
			if !isBooleanLike(val) {
				v.addError("db.locking", "must be a boolean value (true/false, on/off)")
			}
		}

		// Validate db.compare_filename (boolean)
		if val, exists := config["db.compare_filename"]; exists {
			if !isBooleanLike(val) {
				v.addError("db.compare_filename", "must be a boolean value (true/false, on/off)")
			}
		}
	}
}

func (v *TailInputValidator) validateParserSettings(config map[string]any) {
	parser, parserExists := getStringValue(config, "parser")
	_, mlParserExists := getStringValue(config, "multiline.parser")
	dockerMode := getBoolValue(config, "docker_mode")
	oldMultiline := getBoolValue(config, "multiline")

	// parser cannot be empty if specified
	if parserExists && parser == "" {
		v.addError("parser", "cannot be empty string if specified")
	}

	// Conflict: parser should not be used with multiline.parser
	if parserExists && mlParserExists {
		v.addError("parser", "should not be used together with multiline.parser (use multiline.parser alone)")
	}

	// Conflict: parser should not be used with old multiline
	if parserExists && oldMultiline {
		v.addError("parser", "should not be used when multiline is enabled")
	}

	// Conflict: parser should not be used with docker_mode
	if parserExists && dockerMode {
		v.addError("parser", "should not be used when docker_mode is enabled")
	}
}

func (v *TailInputValidator) validateMultilineSettings(config map[string]any) {
	multilineParser, mlParserExists := getStringValue(config, "multiline.parser")
	oldMultiline := getBoolValue(config, "multiline")
	dockerMode := getBoolValue(config, "docker_mode")

	// Old multiline options
	_, parserFirstlineExists := config["parser_firstline"]
	_, multilineFlushExists := config["multiline_flush"]

	// Check for parser_N keys
	hasParserN := false
	for key := range config {
		if matched, _ := regexp.MatchString(`^parser_\d+$`, strings.ToLower(key)); matched {
			hasParserN = true
			break
		}
	}

	// multiline.parser validation
	if mlParserExists {
		if multilineParser == "" {
			v.addError("multiline.parser", "cannot be empty string if specified")
		}

		// Conflicts with old multiline configuration
		if oldMultiline {
			v.addError("multiline.parser", "cannot be used with old 'multiline' option")
		}
		if parserFirstlineExists {
			v.addError("multiline.parser", "cannot be used with 'parser_firstline' (old multiline config)")
		}
		if hasParserN {
			v.addError("multiline.parser", "cannot be used with 'parser_N' options (old multiline config)")
		}
		if multilineFlushExists {
			v.addError("multiline.parser", "cannot be used with 'multiline_flush' (old multiline config)")
		}
		if dockerMode {
			v.addError("multiline.parser", "cannot be used with 'docker_mode'")
		}
	}

	// Old multiline validation
	if oldMultiline {
		// parser_firstline is required when multiline is enabled
		if !parserFirstlineExists {
			v.addError("multiline", "requires 'parser_firstline' to be set")
		}

		// Cannot be used with docker_mode
		if dockerMode {
			v.addError("multiline", "cannot be used at the same time as docker_mode")
		}
	}

	// parser_firstline requires multiline to be enabled
	if parserFirstlineExists && !oldMultiline && !mlParserExists {
		v.addError("parser_firstline", "requires 'multiline' to be enabled or use 'multiline.parser' instead")
	}

	// multiline_flush validation
	if multilineFlushExists {
		if flush, exists := getNumericValue(config, "multiline_flush"); exists {
			if flush < 0 {
				v.addError("multiline_flush", "must be a non-negative number")
			}
		}
	}
}

// validateDockerModeSettings validates Docker mode parameters
func (v *TailInputValidator) validateDockerModeSettings(config map[string]any) {
	dockerMode := getBoolValue(config, "docker_mode")
	oldMultiline := getBoolValue(config, "multiline")
	_, mlParserExists := config["multiline.parser"]

	// Docker mode options
	_, dockerModeFlushExists := config["docker_mode_flush"]
	_, dockerModeParserExists := config["docker_mode_parser"]

	// docker_mode cannot be used with multiline
	if dockerMode && oldMultiline {
		v.addError("docker_mode", "cannot be used at the same time as multiline")
	}

	// docker_mode cannot be used with multiline.parser
	if dockerMode && mlParserExists {
		v.addError("docker_mode", "cannot be used with multiline.parser")
	}

	// docker_mode_flush and docker_mode_parser require docker_mode
	if dockerModeFlushExists && !dockerMode {
		v.addError("docker_mode_flush", "requires 'docker_mode' to be enabled")
	}

	if dockerModeParserExists && !dockerMode {
		v.addError("docker_mode_parser", "requires 'docker_mode' to be enabled")
	}

	// Validate docker_mode_flush value
	if dockerModeFlushExists {
		if flush, exists := getNumericValue(config, "docker_mode_flush"); exists {
			if flush < 0 {
				v.addError("docker_mode_flush", "must be a non-negative number")
			}
		}
	}

	// docker_mode_parser should not be empty if specified
	if dockerModeParserExists {
		if parser, _ := getStringValue(config, "docker_mode_parser"); parser == "" {
			v.addError("docker_mode_parser", "cannot be empty string if specified")
		}
	}
}

func (v *TailInputValidator) validateEncodingSettings(config map[string]any) {
	unicodeEnc, unicodeExists := getStringValue(config, "unicode.encoding")
	genericEnc, genericExists := getStringValue(config, "generic.encoding")

	// Cannot use both unicode.encoding and generic.encoding
	if unicodeExists && genericExists {
		v.addError("unicode.encoding", "cannot be used together with 'generic.encoding'")
		v.addError("generic.encoding", "cannot be used together with 'unicode.encoding'")
	}

	// Validate unicode.encoding values
	if unicodeExists {
		if unicodeEnc == "" {
			v.addError("unicode.encoding", "cannot be empty string if specified")
		} else if !containsStringCI(allowedUnicodeEncoding, unicodeEnc) {
			v.addError("unicode.encoding", fmt.Sprintf("must be one of: %s", strings.Join(allowedUnicodeEncoding, ", ")))
		}
	}

	// Validate generic.encoding values
	if genericExists {
		if genericEnc == "" {
			v.addError("generic.encoding", "cannot be empty string if specified")
		} else if !containsStringCI(allowedGenericEncoding, genericEnc) {
			v.addError("generic.encoding", "must be a valid encoding (ShiftJIS, GBK, GB18030, Big5, UHC, Win1250-1256, Win866, Win874, or their aliases)")
		}
	}
}

func (v *TailInputValidator) validateFileMonitoringSettings(config map[string]any) {
	// inotify_watcher (boolean)
	if val, exists := config["inotify_watcher"]; exists {
		if !isBooleanLike(val) {
			v.addError("inotify_watcher", "must be a boolean value (true/false)")
		}
	}

	// file_cache_advise (on/off)
	if val, exists := config["file_cache_advise"]; exists {
		if !isBooleanLike(val) {
			v.addError("file_cache_advise", "must be a boolean value (on/off)")
		}
	}

	// refresh_interval
	if val, exists := getNumericValue(config, "refresh_interval"); exists {
		if val < 0 {
			v.addError("refresh_interval", "must be a non-negative number (seconds)")
		}
	}

	// rotate_wait
	if val, exists := getNumericValue(config, "rotate_wait"); exists {
		if val < 0 {
			v.addError("rotate_wait", "must be a non-negative number (seconds)")
		}
	}

	// ignore_older (duration format: m, h, d)
	if ignoreOlder, exists := getStringValue(config, "ignore_older"); exists {
		if !isValidIgnoreOlderDuration(ignoreOlder) {
			v.addError("ignore_older", "must be a valid duration (e.g., '5m', '1h', '7d')")
		}
	}

	// progress_check_interval (duration)
	if interval, exists := getStringValue(config, "progress_check_interval"); exists {
		if !isValidDuration(interval) {
			v.addError("progress_check_interval", "must be a valid duration (e.g., '2s')")
		}
	}

	// progress_check_interval_nsec (non-negative integer)
	if val, exists := getNumericValue(config, "progress_check_interval_nsec"); exists {
		if val < 0 {
			v.addError("progress_check_interval_nsec", "must be a non-negative integer")
		}
	}

	// watcher_interval (duration)
	if interval, exists := getStringValue(config, "watcher_interval"); exists {
		if !isValidDuration(interval) {
			v.addError("watcher_interval", "must be a valid duration (e.g., '2s')")
		}
	}
}

func (v *TailInputValidator) validateThreadingSettings(config map[string]any) {
	threaded := getBoolValue(config, "threaded")

	// Validate threaded boolean
	if val, exists := config["threaded"]; exists {
		if !isBooleanLike(val) {
			v.addError("threaded", "must be a boolean value (true/false)")
		}
	}

	// thread.ring_buffer.capacity
	if val, exists := getNumericValue(config, "thread.ring_buffer.capacity"); exists {
		if val <= 0 {
			v.addError("thread.ring_buffer.capacity", "must be a positive integer")
		}
		if !threaded {
			v.addError("thread.ring_buffer.capacity", "requires 'threaded' to be enabled")
		}
	}

	// thread.ring_buffer.window (1-100 percentage)
	if val, exists := getNumericValue(config, "thread.ring_buffer.window"); exists {
		if val < 1 || val > 100 {
			v.addError("thread.ring_buffer.window", "must be a percentage between 1 and 100")
		}
		if !threaded {
			v.addError("thread.ring_buffer.window", "requires 'threaded' to be enabled")
		}
	}
}

func (v *TailInputValidator) validateNumericFields(config map[string]any) {
	// docker_mode_flush (seconds, default 4)
	if val, exists := getNumericValue(config, "docker_mode_flush"); exists {
		if val < 0 {
			v.addError("docker_mode_flush", "must be a non-negative number")
		}
	}

	// multiline_flush (seconds, default 4)
	if val, exists := getNumericValue(config, "multiline_flush"); exists {
		if val < 0 {
			v.addError("multiline_flush", "must be a non-negative number")
		}
	}
}

func (v *TailInputValidator) validateBooleanFields(config map[string]any) {
	boolFields := []string{
		"read_from_head",
		"read_newly_discovered_files_from_head",
		"skip_long_lines",
		"skip_empty_lines",
		"exit_on_eof",
		"ignore_active_older_files",
		"truncate_long_lines",
	}

	for _, field := range boolFields {
		if val, exists := config[field]; exists {
			if !isBooleanLike(val) {
				v.addError(field, "must be a boolean value (true/false, on/off)")
			}
		}
	}
}

func (v *TailInputValidator) validateTagSettings(config map[string]any) {
	// tag validation
	if tag, exists := getStringValue(config, "tag"); exists {
		if tag == "" {
			v.addError("tag", "cannot be empty string if specified")
		}
	}

	// tag_regex validation
	if tagRegex, exists := getStringValue(config, "tag_regex"); exists {
		if tagRegex == "" {
			v.addError("tag_regex", "cannot be empty string if specified")
		} else {
			// Validate regex syntax
			_, err := regexp.Compile(tagRegex)
			if err != nil {
				v.addError("tag_regex", fmt.Sprintf("invalid regular expression: %v", err))
			}
		}
	}
}
