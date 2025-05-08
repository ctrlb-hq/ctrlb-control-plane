package configcompiler

import (
	"fmt"
	"strconv"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/constants"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
)

func CompileGraphToJSON(graph models.PipelineGraph) (*map[string]any, error) {
	utils.Logger.Info("Starting pipeline graph compilation")
	receivers, processors, exporters, pipelines, err := buildPipelines(graph)
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("Failed to build pipelines: %v", err))
		return nil, err
	}

	utils.Logger.Info(fmt.Sprintf("Successfully compiled pipeline graph: receivers=%d processors=%d exporters=%d pipelines=%d",
		len(receivers), len(processors), len(exporters), len(pipelines)))

	// Construct the final config map
	finalConfig := map[string]any{
		"receivers":  receivers,
		"processors": processors,
		"exporters":  exporters,
		"service": map[string]any{
			"pipelines": pipelines,
			"telemetry": constants.TelemetryService,
		},
	}

	return &finalConfig, nil
}

// intersectSupportedSignals returns the intersection of two string slices.
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

func buildPipelines(graph models.PipelineGraph) (map[string]any, map[string]any, map[string]any, Pipelines, error) {
	if len(graph.Nodes) == 0 {
		return nil, nil, nil, nil, fmt.Errorf("empty pipeline graph")
	}

	utils.Logger.Info(fmt.Sprintf("Building pipelines from graph: nodes=%d edges=%d",
		len(graph.Nodes), len(graph.Edges)))

	// Index nodes by ID
	nodesByID := make(map[string]models.PipelineNodes)
	for _, node := range graph.Nodes {
		nodesByID[strconv.Itoa(node.ComponentID)] = node
	}

	// Build the adjacency list for graph traversal
	adjacencyList := make(map[string][]string)
	for _, edge := range graph.Edges {
		if _, exists := nodesByID[edge.Source]; !exists {
			return nil, nil, nil, nil, fmt.Errorf("edge references non-existent source node: %s", edge.Source)
		}
		if _, exists := nodesByID[edge.Target]; !exists {
			return nil, nil, nil, nil, fmt.Errorf("edge references non-existent target node: %s", edge.Target)
		}
		adjacencyList[edge.Source] = append(adjacencyList[edge.Source], edge.Target)
		adjacencyList[edge.Target] = append(adjacencyList[edge.Target], edge.Source)
	}

	// Find connected components using BFS
	visitedNodes := make(map[string]bool)
	var connectedComponents [][]models.PipelineNodes

	for nodeID := range nodesByID {
		if visitedNodes[nodeID] {
			continue
		}

		utils.Logger.Debug(fmt.Sprintf("Processing new component: startNodeId=%s", nodeID))
		nodeQueue := []string{nodeID}
		var currentComponent []models.PipelineNodes
		visitedNodes[nodeID] = true

		for len(nodeQueue) > 0 {
			currentNodeID := nodeQueue[0]
			nodeQueue = nodeQueue[1:]

			node, exists := nodesByID[currentNodeID]
			if !exists {
				return nil, nil, nil, nil, fmt.Errorf("invalid node reference: %s", currentNodeID)
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

	// Prepare maps for pipeline configurations
	receiversConfig := make(map[string]any)
	processorsConfig := make(map[string]any)
	exportersConfig := make(map[string]any)
	pipelines := make(Pipelines)
	pipelineCounter := 1 // Global counter for unique naming across pipelines

	// Process each connected component to build pipelines.
	for _, componentNodes := range connectedComponents {
		utils.Logger.Debug(fmt.Sprintf("Building pipeline configuration: componentNodeCount=%d", len(componentNodes)))

		// Compute common supported signals for all nodes in this component.
		var commonSupportedSignals []string
		if len(componentNodes) > 0 {
			commonSupportedSignals = componentNodes[0].SupportedSignals
			for _, node := range componentNodes[1:] {
				commonSupportedSignals = intersectSupportedSignals(commonSupportedSignals, node.SupportedSignals)
			}
		}

		// Build role-specific alias lists.
		var receiverAliases, processorAliases, exporterAliases []string
		for _, node := range componentNodes {
			// Use meaningful alias formatting: Trims component name after underscore and converts node name to CamelCase.
			alias := utils.TrimAfterUnderscore(node.ComponentName) + "/" + utils.ToCamelCase(node.Name)
			switch node.ComponentRole {
			case "receiver":
				receiverAliases = append(receiverAliases, alias)
				receiversConfig[alias] = node.Config
			case "processor":
				processorAliases = append(processorAliases, alias)
				processorsConfig[alias] = node.Config
			case "exporter":
				exporterAliases = append(exporterAliases, alias)
				exportersConfig[alias] = node.Config
			default:
				return nil, nil, nil, nil, fmt.Errorf("unknown component role: %s", node.ComponentRole)
			}
		}

		// Create separate pipeline per supported signal if multiple signals exist;
		// if none exist, fall back to "undefined".
		if len(commonSupportedSignals) > 0 {
			for _, signal := range commonSupportedSignals {
				pipelineName := fmt.Sprintf("%s/pipeline_%d", signal, pipelineCounter)
				pipelines[pipelineName] = Pipeline{
					Receivers:  receiverAliases,
					Processors: processorAliases,
					Exporters:  exporterAliases,
				}
				pipelineCounter++
			}
		} else {
			return nil, nil, nil, nil, fmt.Errorf("no supported signals found for component: %s", componentNodes[0].Name)
		}
	}

	utils.Logger.Info(fmt.Sprintf("Successfully built pipeline configurations: receivers=%d processors=%d exporters=%d",
		len(receiversConfig), len(processorsConfig), len(exportersConfig)))

	return receiversConfig, processorsConfig, exportersConfig, pipelines, nil
}
