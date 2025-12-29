package configcompiler

import (
	"fmt"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/constants"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
)

// OTELPipelinePath represents a path from receivers through processors to exporters
type OTELPipelinePath struct {
	Receivers  []string // Node IDs
	Processors []string // Node IDs in topological order
	Exporters  []string // Node IDs
}

func compileOTEL(graph models.PipelineGraph) (*map[string]any, error) {
	utils.Logger.Info("Starting OTEL pipeline graph compilation")

	if err := ValidateGraph(graph); err != nil {
		return nil, err
	}

	state := BuildGraphState(graph)

	if cycle := DetectCycle(state); cycle != nil {
		return nil, fmt.Errorf("cycle detected in pipeline graph: %v", cycle)
	}

	receivers, processors, exporters, _, _, _ := GetNodesByRole(state)

	if len(receivers) == 0 {
		return nil, fmt.Errorf("pipeline must have at least one receiver")
	}
	if len(exporters) == 0 {
		return nil, fmt.Errorf("pipeline must have at least one exporter")
	}

	paths, err := findOTELPipelinePaths(state, receivers, processors, exporters)
	if err != nil {
		return nil, err
	}

	if len(paths) == 0 {
		return nil, fmt.Errorf("no valid pipeline paths found (receivers must connect to exporters)")
	}

	config, err := buildOTELConfigFromPaths(paths, state)
	if err != nil {
		return nil, err
	}

	utils.Logger.Info("Successfully compiled OTEL pipeline graph")

	return config, nil
}

// findOTELPipelinePaths finds all distinct paths from receivers to exporters
func findOTELPipelinePaths(state *GraphState, receivers, processors, exporters []string) ([]OTELPipelinePath, error) {
	processorSet := make(map[string]bool)
	for _, p := range processors {
		processorSet[p] = true
	}

	exporterSet := make(map[string]bool)
	for _, e := range exporters {
		exporterSet[e] = true
	}

	// For each receiver, find all exporters it can reach and the processors in between
	// We use a "reachability" approach: for each receiver, BFS/DFS to find connected exporters

	type pathInfo struct {
		receivers  map[string]bool
		processors map[string]bool
		exporters  map[string]bool
	}

	// Group receivers that share the same processor->exporter subgraph
	// This handles fan-in cases where multiple receivers feed the same pipeline

	// First, find which exporters each receiver can reach
	receiverToExporters := make(map[string]map[string]bool)
	receiverToProcessors := make(map[string]map[string]bool)

	for _, receiverID := range receivers {
		reachableExporters := make(map[string]bool)
		reachableProcessors := make(map[string]bool)

		// BFS from receiver
		visited := make(map[string]bool)
		queue := []string{receiverID}
		visited[receiverID] = true

		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]

			for _, next := range state.OutEdges[current] {
				if visited[next] {
					continue
				}
				visited[next] = true

				if exporterSet[next] {
					reachableExporters[next] = true
				} else if processorSet[next] {
					reachableProcessors[next] = true
					queue = append(queue, next)
				} else {
					// Another receiver? Skip it
					queue = append(queue, next)
				}
			}
		}

		receiverToExporters[receiverID] = reachableExporters
		receiverToProcessors[receiverID] = reachableProcessors
	}

	// Now group receivers that have identical processor+exporter sets
	// This creates merged pipelines for fan-in scenarios

	type pipelineKey struct {
		processors string
		exporters  string
	}

	pipelineGroups := make(map[pipelineKey]*pathInfo)

	for _, receiverID := range receivers {
		procs := receiverToProcessors[receiverID]
		exps := receiverToExporters[receiverID]

		if len(exps) == 0 {
			// This receiver doesn't reach any exporter - might be an error or disconnected
			continue
		}

		// Create a key from sorted processor and exporter IDs
		procKey := sortedKeyFromMap(procs)
		expKey := sortedKeyFromMap(exps)
		key := pipelineKey{processors: procKey, exporters: expKey}

		if _, exists := pipelineGroups[key]; !exists {
			pipelineGroups[key] = &pathInfo{
				receivers:  make(map[string]bool),
				processors: procs,
				exporters:  exps,
			}
		}
		pipelineGroups[key].receivers[receiverID] = true
	}

	// Convert groups to paths
	var paths []OTELPipelinePath
	for _, group := range pipelineGroups {
		// Get processor IDs and sort them topologically
		var procIDs []string
		for p := range group.processors {
			procIDs = append(procIDs, p)
		}

		sortedProcs, err := TopologicalSort(procIDs, state)
		if err != nil {
			return nil, fmt.Errorf("failed to sort processors: %v", err)
		}

		path := OTELPipelinePath{
			Receivers:  keysFromMap(group.receivers),
			Processors: sortedProcs,
			Exporters:  keysFromMap(group.exporters),
		}
		paths = append(paths, path)
	}

	return paths, nil
}

// buildOTELConfigFromPaths builds the final OTEL config from pipeline paths
func buildOTELConfigFromPaths(paths []OTELPipelinePath, state *GraphState) (*map[string]any, error) {
	receiversConfig := make(map[string]any)
	processorsConfig := make(map[string]any)
	exportersConfig := make(map[string]any)
	pipelines := make(map[string]any)

	// Track aliases to avoid duplicates
	nodeAliases := make(map[string]string) // nodeID -> alias

	// Helper to get or create alias for a node
	getAlias := func(nodeID string) string {
		if alias, exists := nodeAliases[nodeID]; exists {
			return alias
		}
		node := state.NodesByID[nodeID]
		alias := GenerateOTELAlias(node)
		nodeAliases[nodeID] = alias
		return alias
	}

	pipelineCounter := 1

	for _, path := range paths {
		// Collect all nodes in this path to compute signal intersection
		var pathNodes []models.PipelineNodes
		for _, id := range path.Receivers {
			pathNodes = append(pathNodes, state.NodesByID[id])
		}
		for _, id := range path.Processors {
			pathNodes = append(pathNodes, state.NodesByID[id])
		}
		for _, id := range path.Exporters {
			pathNodes = append(pathNodes, state.NodesByID[id])
		}

		signals := IntersectSignals(pathNodes)
		if len(signals) == 0 {
			return nil, fmt.Errorf("no common signals found for pipeline path")
		}

		// Build receiver aliases and configs
		var receiverAliases []string
		for _, nodeID := range path.Receivers {
			alias := getAlias(nodeID)
			receiverAliases = append(receiverAliases, alias)
			if _, exists := receiversConfig[alias]; !exists {
				receiversConfig[alias] = state.NodesByID[nodeID].Config
			}
		}

		// Build processor aliases and configs (in order)
		var processorAliases []string
		for _, nodeID := range path.Processors {
			alias := getAlias(nodeID)
			processorAliases = append(processorAliases, alias)
			if _, exists := processorsConfig[alias]; !exists {
				processorsConfig[alias] = state.NodesByID[nodeID].Config
			}
		}

		// Build exporter aliases and configs
		var exporterAliases []string
		for _, nodeID := range path.Exporters {
			alias := getAlias(nodeID)
			exporterAliases = append(exporterAliases, alias)
			if _, exists := exportersConfig[alias]; !exists {
				exportersConfig[alias] = state.NodesByID[nodeID].Config
			}
		}

		// Create a pipeline for each signal type
		for _, signal := range signals {
			pipelineName := fmt.Sprintf("%s/pipeline_%d", signal, pipelineCounter)

			pipelineDef := map[string]any{
				"receivers": receiverAliases,
				"exporters": exporterAliases,
			}

			if len(processorAliases) > 0 {
				pipelineDef["processors"] = processorAliases
			}

			pipelines[pipelineName] = pipelineDef
		}

		pipelineCounter++
	}

	finalConfig := map[string]any{
		"receivers":  receiversConfig,
		"processors": processorsConfig,
		"exporters":  exportersConfig,
		"service": map[string]any{
			"pipelines": pipelines,
			"telemetry": constants.TelemetryService,
		},
	}

	return &finalConfig, nil
}

// Helper functions

func sortedKeyFromMap(m map[string]bool) string {
	keys := keysFromMap(m)
	// Simple sort for deterministic keys
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	result := ""
	for _, k := range keys {
		result += k + ","
	}
	return result
}

func keysFromMap(m map[string]bool) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}