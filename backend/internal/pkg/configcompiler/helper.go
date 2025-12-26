package configcompiler

import (
	"fmt"
	"strconv"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
)

// analyzeGraphStructure performs common graph analysis: validates structure and finds connected components
func analyzeGraphStructure(graph models.PipelineGraph, agentTypeName string) ([][]models.PipelineNodes, error) {
	if len(graph.Nodes) == 0 {
		return nil, fmt.Errorf("empty pipeline graph")
	}

	utils.Logger.Info(fmt.Sprintf("Building %s pipeline from graph: nodes=%d edges=%d",
		agentTypeName, len(graph.Nodes), len(graph.Edges)))

	nodesByID := make(map[string]models.PipelineNodes)
	for _, node := range graph.Nodes {
		nodesByID[strconv.Itoa(node.ComponentID)] = node
	}

	adjacencyList := make(map[string][]string)
	for _, edge := range graph.Edges {
		if _, exists := nodesByID[edge.Source]; !exists {
			return nil, fmt.Errorf("edge references non-existent source node: %s", edge.Source)
		}
		if _, exists := nodesByID[edge.Target]; !exists {
			return nil, fmt.Errorf("edge references non-existent target node: %s", edge.Target)
		}
		adjacencyList[edge.Source] = append(adjacencyList[edge.Source], edge.Target)
	}

	// Find connected components using BFS
	visitedNodes := make(map[string]bool)
	var connectedComponents [][]models.PipelineNodes

	for nodeID := range nodesByID {
		if visitedNodes[nodeID] {
			continue
		}

		utils.Logger.Debug(fmt.Sprintf("Processing new %s component: startNodeId=%s", agentTypeName, nodeID))
		nodeQueue := []string{nodeID}
		var currentComponent []models.PipelineNodes
		visitedNodes[nodeID] = true

		for len(nodeQueue) > 0 {
			currentNodeID := nodeQueue[0]
			nodeQueue = nodeQueue[1:]

			node, exists := nodesByID[currentNodeID]
			if !exists {
				return nil, fmt.Errorf("invalid node reference: %s", currentNodeID)
			}
			currentComponent = append(currentComponent, node)

			for _, neighborID := range adjacencyList[currentNodeID] {
				if !visitedNodes[neighborID] {
					visitedNodes[neighborID] = true
					nodeQueue = append(nodeQueue, neighborID)
				}
			}
		}
		connectedComponents = append(connectedComponents, currentComponent)
	}

	return connectedComponents, nil
}

// findUpstreamInputTags traces back from a node to find all input tags it should receive
func findUpstreamInputTags(
	startNodeID string,
	inputTags map[string]string,
	incomingEdges map[string][]string,
	nodesByID map[string]models.PipelineNodes,
) []string {
	visited := make(map[string]bool)
	var tags []string
	tagSet := make(map[string]bool) // dedupe tags

	var traverse func(nodeID string)
	traverse = func(nodeID string) {
		if visited[nodeID] {
			return
		}
		visited[nodeID] = true

		// If this is an input node, collect its tag
		if tag, isInput := inputTags[nodeID]; isInput {
			if !tagSet[tag] {
				tagSet[tag] = true
				tags = append(tags, tag)
			}
			return
		}

		// Otherwise, traverse upstream
		for _, sourceID := range incomingEdges[nodeID] {
			traverse(sourceID)
		}
	}

	// Start traversal from immediate upstream nodes
	for _, sourceID := range incomingEdges[startNodeID] {
		traverse(sourceID)
	}

	return tags
}

// buildFluentBitNodeConfig creates the config map for a Fluent Bit node
func buildFluentBitNodeConfig(node models.PipelineNodes, tag string, matchTags []string) map[string]any {
	config := make(map[string]any)

	// Copy existing config
	for k, v := range node.Config {
		config[k] = v
	}

	// Set plugin name
	if _, hasName := config["name"]; !hasName {
		config["name"] = utils.TrimAfterUnderscore(node.ComponentName)
	}

	// Set alias for identification/debugging
	alias := generateAlias(node)
	if _, hasAlias := config["alias"]; !hasAlias {
		config["alias"] = alias
	}

	// Set tag for inputs
	if tag != "" && node.ComponentRole == "input" {
		if _, hasTag := config["tag"]; !hasTag {
			config["tag"] = tag
		}
	}

	// Set match pattern for filters and outputs
	if node.ComponentRole == "filter" || node.ComponentRole == "output" {
		if _, hasMatch := config["match"]; !hasMatch {
			if _, hasMatchRegex := config["match_regex"]; !hasMatchRegex {
				config["match"] = computeMatchPattern(matchTags)
			}
		}
	}

	return config
}

// computeMatchPattern generates an optimal match pattern for the given tags
func computeMatchPattern(tags []string) string {
	if len(tags) == 0 {
		// No specific upstream inputs found, match all
		return "*"
	}

	if len(tags) == 1 {
		// Single tag, match exactly (with wildcard for sub-tags)
		return tags[0] + "*"
	}

	// Multiple tags - try to find common prefix
	prefix := longestCommonPrefix(tags)
	if prefix != "" && len(prefix) >= 3 {
		// Use prefix-based wildcard matching
		return prefix + "*"
	}

	// For Fluent Bit, if we can't find a common prefix and have multiple distinct
	// tag patterns, we need to use regex or accept that we match broader
	// In practice, you might want to use match_regex here for complex cases
	// For now, we'll use the first tag's prefix or wildcard
	if len(tags) > 0 {
		// Extract plugin prefix from first tag (e.g., "tail" from "tail.myAlias")
		firstTag := tags[0]
		for i, c := range firstTag {
			if c == '.' {
				return firstTag[:i+1] + "*"
			}
		}
		return firstTag + "*"
	}

	return "*"
}

// longestCommonPrefix finds the longest common prefix among strings
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

// generateAlias creates a unique alias for a node
func generateAlias(node models.PipelineNodes) string {
	return fmt.Sprintf("%s_%s", utils.ToCamelCase(node.Name), utils.HashFromConfig(node.Config))
}

// getConfigString safely extracts a string value from config
func getConfigString(config map[string]any, key string) string {
	if v, ok := config[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// intersectSupportedSignals returns the intersection of two string slices
func intersectSupportedSignals(firstSignals, secondSignals []string) []string {
	intersection := []string{}
	signalSet := make(map[string]bool)

	for _, signal := range firstSignals {
		signalSet[signal] = true
	}

	for _, signal := range secondSignals {
		if signalSet[signal] {
			intersection = append(intersection, signal)
		}
	}

	return intersection
}