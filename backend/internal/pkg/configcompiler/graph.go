package configcompiler

import (
	"fmt"
	"strconv"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
)

// GraphState holds parsed graph data for compilation
type GraphState struct {
	NodesByID map[string]models.PipelineNodes
	InEdges   map[string][]string // target -> sources
	OutEdges  map[string][]string // source -> targets
}

// ValidateGraph performs common validation for both compilers
func ValidateGraph(graph models.PipelineGraph) error {
	if len(graph.Nodes) == 0 {
		return fmt.Errorf("empty pipeline graph")
	}

	nodeIDs := make(map[string]bool)
	for _, node := range graph.Nodes {
		nodeIDs[strconv.Itoa(node.ComponentID)] = true
	}

	for _, edge := range graph.Edges {
		if !nodeIDs[edge.Source] {
			return fmt.Errorf("edge references non-existent source node: %s", edge.Source)
		}
		if !nodeIDs[edge.Target] {
			return fmt.Errorf("edge references non-existent target node: %s", edge.Target)
		}
	}

	return nil
}

// BuildGraphState creates the adjacency lists and node map from a graph
func BuildGraphState(graph models.PipelineGraph) *GraphState {
	state := &GraphState{
		NodesByID: make(map[string]models.PipelineNodes),
		InEdges:   make(map[string][]string),
		OutEdges:  make(map[string][]string),
	}

	for _, node := range graph.Nodes {
		nodeID := strconv.Itoa(node.ComponentID)
		state.NodesByID[nodeID] = node
	}

	for _, edge := range graph.Edges {
		state.OutEdges[edge.Source] = append(state.OutEdges[edge.Source], edge.Target)
		state.InEdges[edge.Target] = append(state.InEdges[edge.Target], edge.Source)
	}

	return state
}

// DetectCycle performs DFS to detect cycles in the directed graph
// Returns the cycle path if found, nil otherwise
func DetectCycle(state *GraphState) []string {
	// Node states: 0 = unvisited, 1 = visiting (in current DFS path), 2 = visited
	visitState := make(map[string]int)
	parent := make(map[string]string)

	var dfs func(nodeID string) []string
	dfs = func(nodeID string) []string {
		visitState[nodeID] = 1 // Mark as visiting

		for _, neighborID := range state.OutEdges[nodeID] {
			if visitState[neighborID] == 1 {
				// Back edge found - cycle detected
				cycle := []string{neighborID}
				current := nodeID
				for current != neighborID {
					cycle = append([]string{current}, cycle...)
					current = parent[current]
				}
				cycle = append([]string{neighborID}, cycle...)
				return cycle
			}

			if visitState[neighborID] == 0 {
				parent[neighborID] = nodeID
				if cycle := dfs(neighborID); cycle != nil {
					return cycle
				}
			}
		}

		visitState[nodeID] = 2 // Mark as visited
		return nil
	}

	// Try DFS from each unvisited node
	for nodeID := range state.NodesByID {
		if visitState[nodeID] == 0 {
			if cycle := dfs(nodeID); cycle != nil {
				return cycle
			}
		}
	}

	return nil
}

// TopologicalSort returns nodes in dependency order (sources before targets)
// Only sorts the provided nodeIDs, not the entire graph
func TopologicalSort(nodeIDs []string, state *GraphState) ([]string, error) {
	// Build a set for quick lookup
	nodeSet := make(map[string]bool)
	for _, id := range nodeIDs {
		nodeSet[id] = true
	}

	// Compute in-degrees within the subset
	inDegree := make(map[string]int)
	for _, id := range nodeIDs {
		inDegree[id] = 0
	}

	for _, id := range nodeIDs {
		for _, targetID := range state.OutEdges[id] {
			if nodeSet[targetID] {
				inDegree[targetID]++
			}
		}
	}

	// Kahn's algorithm
	queue := []string{}
	for _, id := range nodeIDs {
		if inDegree[id] == 0 {
			queue = append(queue, id)
		}
	}

	var result []string
	for len(queue) > 0 {
		nodeID := queue[0]
		queue = queue[1:]
		result = append(result, nodeID)

		for _, targetID := range state.OutEdges[nodeID] {
			if nodeSet[targetID] {
				inDegree[targetID]--
				if inDegree[targetID] == 0 {
					queue = append(queue, targetID)
				}
			}
		}
	}

	if len(result) != len(nodeIDs) {
		return nil, fmt.Errorf("cycle detected in subgraph")
	}

	return result, nil
}

// GetNodesByRole returns node IDs grouped by their component role
func GetNodesByRole(state *GraphState) (receivers, processors, exporters, inputs, filters, outputs []string) {
	for nodeID, node := range state.NodesByID {
		switch node.ComponentRole {
		case "receiver":
			receivers = append(receivers, nodeID)
		case "processor":
			processors = append(processors, nodeID)
		case "exporter":
			exporters = append(exporters, nodeID)
		case "input":
			inputs = append(inputs, nodeID)
		case "filter":
			filters = append(filters, nodeID)
		case "output":
			outputs = append(outputs, nodeID)
		}
	}
	return
}

// IntersectSignals returns the intersection of supported signals across nodes
func IntersectSignals(nodes []models.PipelineNodes) []string {
	if len(nodes) == 0 {
		return nil
	}

	// Start with first node's signals
	signalCount := make(map[string]int)
	for _, signal := range nodes[0].SupportedSignals {
		signalCount[signal] = 1
	}

	// Count occurrences across all nodes
	for _, node := range nodes[1:] {
		for _, signal := range node.SupportedSignals {
			if _, exists := signalCount[signal]; exists {
				signalCount[signal]++
			}
		}
	}

	// Collect signals that appear in all nodes
	var result []string
	for signal, count := range signalCount {
		if count == len(nodes) {
			result = append(result, signal)
		}
	}

	return result
}

// otelComponentRoles are the valid component roles for OTEL agents
var otelComponentRoles = map[string]bool{
	"receiver":  true,
	"processor": true,
	"exporter":  true,
}

// fluentBitComponentRoles are the valid component roles for Fluent Bit agents
var fluentBitComponentRoles = map[string]bool{
	"input":  true,
	"filter": true,
	"output": true,
}

// ValidateGraphForAgentType checks if a pipeline graph is compatible with a specific agent type
// Returns nil if compatible, error with details if not
func ValidateGraphForAgentType(graph models.PipelineGraph, agentType string) error {
	if len(graph.Nodes) == 0 {
		return nil // Empty graph is compatible with any agent
	}

	var expectedRoles map[string]bool
	var agentTypeName string

	switch agentType {
	case string(AgentTypeOTEL):
		expectedRoles = otelComponentRoles
		agentTypeName = "OTEL"
	case string(AgentTypeFluentBit):
		expectedRoles = fluentBitComponentRoles
		agentTypeName = "Fluent Bit"
	default:
		return fmt.Errorf("unknown agent type: %s", agentType)
	}

	var incompatibleComponents []string
	for _, node := range graph.Nodes {
		if !expectedRoles[node.ComponentRole] {
			incompatibleComponents = append(incompatibleComponents,
				fmt.Sprintf("%s (role: %s)", node.ComponentName, node.ComponentRole))
		}
	}

	if len(incompatibleComponents) > 0 {
		return fmt.Errorf("pipeline contains components incompatible with %s agent: %v",
			agentTypeName, incompatibleComponents)
	}

	return nil
}

// IsFluentBitCompatibleGraph checks if a graph can be used with a Fluent Bit agent
func IsFluentBitCompatibleGraph(graph models.PipelineGraph) bool {
	return ValidateGraphForAgentType(graph, string(AgentTypeFluentBit)) == nil
}

// IsOTELCompatibleGraph checks if a graph can be used with an OTEL agent
func IsOTELCompatibleGraph(graph models.PipelineGraph) bool {
	return ValidateGraphForAgentType(graph, string(AgentTypeOTEL)) == nil
}
