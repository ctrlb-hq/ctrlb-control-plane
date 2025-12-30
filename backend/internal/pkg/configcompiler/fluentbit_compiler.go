package configcompiler

import (
	"fmt"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/constants"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
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
		config["name"] = GetPluginName(node)
		config["alias"] = alias
		config["tag"] = tag

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

	config := map[string]any{
		"service": constants.FluentBitService,
		"pipeline": map[string]any{
			"inputs":  inputs,
			"filters": filters,
			"outputs": outputs,
		},
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

func getConfigString(config map[string]any, key string) string {
	if v, ok := config[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
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
