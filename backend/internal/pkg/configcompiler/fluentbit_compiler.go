package configcompiler

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/constants"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	validators "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/pkg/configcompiler/fluentbit_validators"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
)

func compileFB(graph models.PipelineGraph) (*map[string]any, error) {
	utils.Logger.Info("Starting Fluent Bit pipeline graph compilation")

	if err := ValidateGraph(graph); err != nil {
		return nil, err
	}

	state := BuildGraphState(graph)

	if cycle := DetectCycle(state); cycle != nil {
		return nil, fmt.Errorf("cycle detected in pipeline graph: %v", cycle)
	}

	_, _, _, inputs, filters, outputs := GetNodesByRole(state)

	if len(inputs) == 0 {
		return nil, fmt.Errorf("pipeline must have at least one input")
	}
	if len(outputs) == 0 {
		return nil, fmt.Errorf("pipeline must have at least one output")
	}

	instances, err := buildFBNodeInstances(state, inputs, filters, outputs)
	if err != nil {
		return nil, err
	}

	config := buildFBConfigFromInstances(instances)

	utils.Logger.Info(fmt.Sprintf("Successfully compiled Fluent Bit pipeline: inputs=%d filters=%d outputs=%d",
		len(inputs), len(filters), len(outputs)))

	return config, nil
}

// buildFBNodeInstances builds all node instances with proper tag routing
func buildFBNodeInstances(state *GraphState, inputs, filters, outputs []string) ([]FBNodeInstance, error) {
	var instances []FBNodeInstance

	// Step 1: Create input instances and assign tags
	inputTags := make(map[string]FBTagState) // nodeID -> tag state

	for _, nodeID := range inputs {
		node := state.NodesByID[nodeID]
		alias := GenerateFBAlias(node)

		if err := validateFluentBitNodeConfig(node); err != nil {
			return nil, err
		}

		// Check if user provided a tag in config
		tag := getConfigString(node.Config, "tag")
		if tag == "" {
			tag = GenerateFBTag(node, alias)
		}

		inputTags[nodeID] = FBTagState{
			Tag:        tag,
			SourceNode: nodeID,
		}

		// Build config for input
		config := copyConfig(node.Config)
		// Flatten nested objects for Fluent Bit compatibility
		config = flattenNestedConfig(config)
		config["name"] = GetPluginName(node)
		config["alias"] = alias
		config["tag"] = tag

		compactConfig(config)

		instances = append(instances, FBNodeInstance{
			OriginalNodeID: nodeID,
			InstanceIndex:  0,
			Alias:          alias,
			Tag:            tag,
			Config:         config,
			Role:           "input",
		})
	}

	// Step 2: Process filters and outputs in topological order
	// Combine filters and outputs for sorting
	nonInputs := append(filters, outputs...)

	sortedNonInputs, err := TopologicalSort(nonInputs, state)
	if err != nil {
		return nil, fmt.Errorf("failed to sort filters/outputs: %v", err)
	}

	// Track what tags each node instance emits (for downstream nodes)
	// Key: originalNodeID, Value: list of tags emitted by all instances of that node
	nodeEmittedTags := make(map[string][]FBTagState)

	// Initialize with input tags
	for nodeID, tagState := range inputTags {
		nodeEmittedTags[nodeID] = []FBTagState{tagState}
	}

	// Step 3: Process each non-input node
	for _, nodeID := range sortedNonInputs {
		node := state.NodesByID[nodeID]

		if err := validateFluentBitNodeConfig(node); err != nil {
			return nil, err
		}

		// Find all upstream tags
		upstreamTags := findUpstreamTags(nodeID, state, nodeEmittedTags)

		if len(upstreamTags) == 0 {
			// No upstream connections - this shouldn't happen in a valid graph
			// but handle gracefully by using wildcard
			upstreamTags = []FBTagState{{Tag: "*", SourceNode: ""}}
		}

		// Check if tags are compatible (can use single match pattern)
		compatible, matchPattern := computeFBMatchPattern(upstreamTags)

		if compatible {
			// Single instance with combined match pattern
			alias := GenerateFBAlias(node)

			config := copyConfig(node.Config)
			// Flatten nested objects for Fluent Bit compatibility
			config = flattenNestedConfig(config)
			config["name"] = GetPluginName(node)
			config["alias"] = alias
			if _, hasMatch := config["match"]; !hasMatch {
				if _, hasMatchRegex := config["match_regex"]; !hasMatchRegex {
					config["match"] = matchPattern
				}
			}

			instances = append(instances, FBNodeInstance{
				OriginalNodeID: nodeID,
				InstanceIndex:  0,
				Alias:          alias,
				MatchPattern:   matchPattern,
				Config:         config,
				Role:           node.ComponentRole,
			})

			// This node emits the same tags it receives (unless it modifies them)
			// For simplicity, we pass through the same tags
			nodeEmittedTags[nodeID] = upstreamTags

		} else {
			// Incompatible tags - create separate instances
			var emittedTags []FBTagState

			for i, tagState := range upstreamTags {
				alias := fmt.Sprintf("%s_inst%d", GenerateFBAlias(node), i+1)
				pattern := tagState.Tag + "*"

				config := copyConfig(node.Config)
				// Flatten nested objects for Fluent Bit compatibility
				config = flattenNestedConfig(config)
				config["name"] = GetPluginName(node)
				config["alias"] = alias
				if _, hasMatch := config["match"]; !hasMatch {
					if _, hasMatchRegex := config["match_regex"]; !hasMatchRegex {
						config["match"] = pattern
					}
				}

				instances = append(instances, FBNodeInstance{
					OriginalNodeID: nodeID,
					InstanceIndex:  i + 1,
					Alias:          alias,
					MatchPattern:   pattern,
					Config:         config,
					Role:           node.ComponentRole,
				})

				// Each instance emits the tag it matches
				emittedTags = append(emittedTags, tagState)
			}

			nodeEmittedTags[nodeID] = emittedTags
		}
	}

	return instances, nil
}

// findUpstreamTags finds all tags that flow into a node
func findUpstreamTags(nodeID string, state *GraphState, nodeEmittedTags map[string][]FBTagState) []FBTagState {
	seen := make(map[string]bool) // Dedupe tags
	var result []FBTagState

	for _, sourceID := range state.InEdges[nodeID] {
		emittedTags, exists := nodeEmittedTags[sourceID]
		if !exists {
			continue
		}

		for _, tagState := range emittedTags {
			if !seen[tagState.Tag] {
				seen[tagState.Tag] = true
				result = append(result, tagState)
			}
		}
	}

	return result
}

// computeFBMatchPattern determines if tags can share a match pattern
// Returns (compatible, pattern)
func computeFBMatchPattern(tags []FBTagState) (bool, string) {
	if len(tags) == 0 {
		return true, "*"
	}

	if len(tags) == 1 {
		return true, tags[0].Tag + "*"
	}

	// Extract tag strings
	tagStrings := make([]string, len(tags))
	for i, t := range tags {
		tagStrings[i] = t.Tag
	}

	// Check if all tags have the same plugin prefix
	firstPrefix := ExtractTagPrefix(tagStrings[0])
	allSamePrefix := true
	for _, tag := range tagStrings[1:] {
		if ExtractTagPrefix(tag) != firstPrefix {
			allSamePrefix = false
			break
		}
	}

	if allSamePrefix {
		// Try to find a common prefix
		commonPrefix := longestCommonPrefix(tagStrings)
		if len(commonPrefix) >= 3 {
			return true, commonPrefix + "*"
		}
		// Fall back to plugin prefix
		return true, firstPrefix + "*"
	}

	// Different plugin prefixes - incompatible
	return false, ""
}

// buildFBConfigFromInstances converts instances to final config format
func buildFBConfigFromInstances(instances []FBNodeInstance) *map[string]any {
	var inputs, filters, outputs []any

	for _, inst := range instances {
		switch inst.Role {
		case "input":
			inputs = append(inputs, inst.Config)
		case "filter":
			filters = append(filters, inst.Config)
		case "output":
			outputs = append(outputs, inst.Config)
		}
	}

	// Add prometheus monitoring
	inputs = append(inputs, constants.FluentBitNodeMetricsInput, constants.FluentBitInternalMetricsInput)
	outputs = append(outputs, constants.FluentBitPrometheusOutput)

	pipeline := map[string]any{
		"inputs":  inputs,
		"outputs": outputs,
	}
	// Only add filters if non-empty (omit key entirely rather than null)
	if len(filters) > 0 {
		pipeline["filters"] = filters
	}

	config := map[string]any{
		"service":  constants.FluentBitService,
		"parsers":  constants.FluentBitDefaultParsers,
		"pipeline": pipeline,
	}

	return &config
}

// Helper functions

func copyConfig(src map[string]any) map[string]any {
	dst := make(map[string]any)
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// flattenNestedConfig flattens nested objects into dotted keys for Fluent Bit compatibility.
// JSON Forms interprets schema properties with dots (e.g., "unicode.encoding") as nested paths,
// creating { unicode: { encoding: "value" } } instead of { "unicode.encoding": "value" }.
// This function converts nested structures back to the flat dotted-key format that Fluent Bit expects.
func flattenNestedConfig(config map[string]any) map[string]any {
	result := make(map[string]any)
	flattenRecursive(config, "", result)
	return result
}

func flattenRecursive(config map[string]any, prefix string, result map[string]any) {
	for k, v := range config {
		newKey := k
		if prefix != "" {
			newKey = prefix + "." + k
		}

		if nested, ok := v.(map[string]any); ok {
			// Recursively flatten nested maps
			flattenRecursive(nested, newKey, result)
		} else {
			result[newKey] = v
		}
	}
}

// compactConfig removes empty values that should not be emitted into Fluent Bit configs.
// This makes UI hide/show behave like "not configured" in the resulting config.
func compactConfig(config map[string]any) {
	// First, clean up conditional fields that don't apply
	cleanupConditionalFields(config)

	for k, v := range config {
		if v == nil {
			delete(config, k)
			continue
		}
		switch vv := v.(type) {
		case string:
			if strings.TrimSpace(vv) == "" {
				delete(config, k)
			}
		case []any:
			if len(vv) == 0 {
				delete(config, k)
			}
		case []string:
			if len(vv) == 0 {
				delete(config, k)
			}
		}
	}
}

// cleanupConditionalFields removes fields that depend on other fields being set.
// For example, db.* fields should only be present if "db" is set.
// This handles cases where the frontend sends default values for hidden fields.
func cleanupConditionalFields(config map[string]any) {
	// db.* fields require "db" to be set
	dbPath, _ := config["db"].(string)
	if strings.TrimSpace(dbPath) == "" {
		delete(config, "db")
		delete(config, "db.sync")
		delete(config, "db.locking")
		delete(config, "db.journal_mode")
		delete(config, "db.compare_filename")
	}

	// docker_mode_* fields require "docker_mode" to be true
	dockerMode := getBoolFromConfig(config, "docker_mode")
	if !dockerMode {
		delete(config, "docker_mode")
		delete(config, "docker_mode_flush")
		delete(config, "docker_mode_parser")
	}

	// multiline_flush, parser_firstline, parser_N require "multiline" to be "on" or true
	multilineEnabled := isMultilineEnabled(config)
	if !multilineEnabled {
		delete(config, "multiline")
		delete(config, "multiline_flush")
		delete(config, "parser_firstline")
		// Remove parser_N fields
		for k := range config {
			if strings.HasPrefix(strings.ToLower(k), "parser_") && k != "parser" {
				// Check if it's parser_N pattern (not parser_firstline)
				if matched, _ := regexp.MatchString(`^parser_\d+$`, strings.ToLower(k)); matched {
					delete(config, k)
				}
			}
		}
	}

	// thread.ring_buffer.* require "threaded" to be true
	threaded := getBoolFromConfig(config, "threaded")
	if !threaded {
		delete(config, "threaded")
		delete(config, "thread.ring_buffer.capacity")
		delete(config, "thread.ring_buffer.window")
	}
}

func getBoolFromConfig(config map[string]any, key string) bool {
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
	}
	return false
}

func isMultilineEnabled(config map[string]any) bool {
	val, exists := config["multiline"]
	if !exists {
		return false
	}

	switch v := val.(type) {
	case bool:
		return v
	case string:
		lower := strings.ToLower(v)
		return lower == "on" || lower == "true" || lower == "yes" || lower == "1"
	}
	return false
}

func getConfigString(config map[string]any, key string) string {
	if v, ok := config[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func validateFluentBitNodeConfig(node models.PipelineNodes) error {
	pluginName := GetPluginName(node)
	configCopy := copyConfig(node.Config)
	// Flatten nested objects (e.g., { unicode: { encoding: "..." } } -> { "unicode.encoding": "..." })
	configCopy = flattenNestedConfig(configCopy)
	compactConfig(configCopy)

	switch node.ComponentRole {
	case "input":
		if pluginName == "tail" {
			if errs := validators.ValidateTailInputConfig(configCopy); errs.HasErrors() {
				return fmt.Errorf("invalid tail input config for node %q: %w", node.Name, errs)
			}
		}
	case "filter":
		switch pluginName {
		case "grep":
			if errs := validators.ValidateGrepFilterConfig(configCopy); errs.HasErrors() {
				return fmt.Errorf("invalid grep filter config for node %q: %w", node.Name, errs)
			}
		case "modify":
			if errs := validators.ValidateModifyFilterConfig(configCopy); errs.HasErrors() {
				return fmt.Errorf("invalid modify filter config for node %q: %w", node.Name, errs)
			}
		}
	case "output":
		if pluginName == "http" {
			if errs := validators.ValidateHTTPOutputConfig(configCopy); errs.HasErrors() {
				return fmt.Errorf("invalid http output config for node %q: %w", node.Name, errs)
			}
		}
	}

	return nil
}

func longestCommonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	prefix := strs[0]
	for _, s := range strs[1:] {
		for len(prefix) > 0 && !hasPrefix(s, prefix) {
			prefix = prefix[:len(prefix)-1]
		}
		if prefix == "" {
			break
		}
	}
	return prefix
}

func hasPrefix(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	return s[:len(prefix)] == prefix
}
