package configcompiler

import (
	"fmt"
	"strconv"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/constants"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
)

// AgentType represents the type of agent (otel or fluent-bit)
type AgentType string

const (
	AgentTypeOTEL      AgentType = "otel"
	AgentTypeFluentBit AgentType = "fluent-bit"
)

// CompileGraph compiles a pipeline graph to the appropriate configuration format based on agent type
func CompileGraph(graph models.PipelineGraph, agentType AgentType) (*map[string]any, error) {
	switch agentType {
	case AgentTypeOTEL:
		return CompileGraphToJSON(graph)
	case AgentTypeFluentBit:
		return CompileGraphToFluentBit(graph)
	default:
		return nil, fmt.Errorf("unsupported agent type: %s", agentType)
	}
}

// CompileGraphToJSON compiles a pipeline graph to OTEL JSON format
func CompileGraphToJSON(graph models.PipelineGraph) (*map[string]any, error) {
	utils.Logger.Info("Starting OTEL pipeline graph compilation")

	// Get connected components using common logic
	connectedComponents, err := analyzeGraphStructure(graph, "OTEL")
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("Failed to analyze graph structure: %v", err))
		return nil, err
	}

	// Build OTEL-specific configuration
	receivers, processors, exporters, pipelines, err := buildOTELConfig(connectedComponents)
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("Failed to build OTEL pipelines: %v", err))
		return nil, err
	}

	utils.Logger.Info(fmt.Sprintf("Successfully compiled OTEL pipeline graph: receivers=%d processors=%d exporters=%d pipelines=%d",
		len(receivers), len(processors), len(exporters), len(pipelines)))

	// Construct the final OTEL config map
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

// CompileGraphToFluentBit compiles a pipeline graph to Fluent Bit YAML format
func CompileGraphToFluentBit(graph models.PipelineGraph) (*map[string]any, error) {
	utils.Logger.Info("Starting Fluent Bit pipeline graph compilation")

	// Get connected components using common logic
	connectedComponents, err := analyzeGraphStructure(graph, "Fluent Bit")
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("Failed to analyze graph structure: %v", err))
		return nil, err
	}

	// Build Fluent Bit-specific configuration
	inputs, filters, outputs, err := buildFluentBitConfig(connectedComponents)
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("Failed to build Fluent Bit pipeline: %v", err))
		return nil, err
	}

	utils.Logger.Info(fmt.Sprintf("Successfully compiled Fluent Bit pipeline: inputs=%d filters=%d outputs=%d",
		len(inputs), len(filters), len(outputs)))

	// Construct the final Fluent Bit config
	finalConfig := map[string]any{
		"service": map[string]any{
			"flush":       1,
			"log_level":   "info",
			"http_server": "on",
			"http_listen": "0.0.0.0",
			"http_port":   2020,
			"hot_reload":  "on",
		},
		"pipeline": map[string]any{
			"inputs":  inputs,
			"filters": filters,
			"outputs": outputs,
		},
	}

	return &finalConfig, nil
}

// analyzeGraphStructure performs common graph analysis: validates structure and finds connected components
func analyzeGraphStructure(graph models.PipelineGraph, agentTypeName string) ([][]models.PipelineNodes, error) {
	if len(graph.Nodes) == 0 {
		return nil, fmt.Errorf("empty pipeline graph")
	}

	utils.Logger.Info(fmt.Sprintf("Building %s pipeline from graph: nodes=%d edges=%d",
		agentTypeName, len(graph.Nodes), len(graph.Edges)))

	// Index nodes by ID for quick lookup
	nodesByID := make(map[string]models.PipelineNodes)
	for _, node := range graph.Nodes {
		nodesByID[strconv.Itoa(node.ComponentID)] = node
	}

	// Build the adjacency list for graph traversal and validate edges
	adjacencyList := make(map[string][]string)
	for _, edge := range graph.Edges {
		if _, exists := nodesByID[edge.Source]; !exists {
			return nil, fmt.Errorf("edge references non-existent source node: %s", edge.Source)
		}
		if _, exists := nodesByID[edge.Target]; !exists {
			return nil, fmt.Errorf("edge references non-existent target node: %s", edge.Target)
		}
		adjacencyList[edge.Source] = append(adjacencyList[edge.Source], edge.Target)
		adjacencyList[edge.Target] = append(adjacencyList[edge.Target], edge.Source)
	}

	// Find connected components using BFS (Breadth-First Search)
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

		// BFS traversal to find all nodes in this connected component
		for len(nodeQueue) > 0 {
			currentNodeID := nodeQueue[0]
			nodeQueue = nodeQueue[1:]

			node, exists := nodesByID[currentNodeID]
			if !exists {
				return nil, fmt.Errorf("invalid node reference: %s", currentNodeID)
			}
			currentComponent = append(currentComponent, node)

			// Add unvisited neighbors to the queue
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

// buildOTELConfig builds OTEL-specific configuration from connected components
func buildOTELConfig(connectedComponents [][]models.PipelineNodes) (map[string]any, map[string]any, map[string]any, Pipelines, error) {
	receiversConfig := make(map[string]any)
	processorsConfig := make(map[string]any)
	exportersConfig := make(map[string]any)
	pipelines := make(Pipelines)
	pipelineCounter := 1

	// Process each connected component to build OTEL pipelines
	for _, componentNodes := range connectedComponents {
		utils.Logger.Debug(fmt.Sprintf("Building OTEL pipeline configuration: componentNodeCount=%d", len(componentNodes)))

		// Compute common supported signals for all nodes in this component
		var commonSupportedSignals []string
		if len(componentNodes) > 0 {
			commonSupportedSignals = componentNodes[0].SupportedSignals
			for _, node := range componentNodes[1:] {
				commonSupportedSignals = intersectSupportedSignals(commonSupportedSignals, node.SupportedSignals)
			}
		}

		// Build role-specific alias lists and configs
		var receiverAliases, processorAliases, exporterAliases []string
		for _, node := range componentNodes {
			// Use meaningful alias formatting
			alias := fmt.Sprintf("%s/%s_%s", utils.TrimAfterUnderscore(node.ComponentName), utils.ToCamelCase(node.Name), utils.HashFromConfig(node.Config))
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

		// Create separate pipeline per supported signal
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

	utils.Logger.Info(fmt.Sprintf("Successfully built OTEL pipeline configurations: receivers=%d processors=%d exporters=%d",
		len(receiversConfig), len(processorsConfig), len(exportersConfig)))

	return receiversConfig, processorsConfig, exportersConfig, pipelines, nil
}

// buildFluentBitConfig builds Fluent Bit-specific configuration from connected components
func buildFluentBitConfig(connectedComponents [][]models.PipelineNodes) ([]any, []any, []any, error) {
	var inputs []any
	var filters []any
	var outputs []any

	// Process each connected component to build Fluent Bit pipeline
	for componentIdx, componentNodes := range connectedComponents {
		utils.Logger.Debug(fmt.Sprintf("Building Fluent Bit pipeline configuration: component=%d nodeCount=%d",
			componentIdx, len(componentNodes)))

		// Validate that component has required inputs and outputs
		hasInput := false
		hasOutput := false
		for _, node := range componentNodes {
			if node.ComponentRole == "receiver" {
				hasInput = true
			}
			if node.ComponentRole == "exporter" {
				hasOutput = true
			}
		}

		if !hasInput {
			return nil, nil, nil, fmt.Errorf("component %d missing input (receiver)", componentIdx)
		}
		if !hasOutput {
			return nil, nil, nil, fmt.Errorf("component %d missing output (exporter)", componentIdx)
		}

		// Build Fluent Bit components for this connected component
		for _, node := range componentNodes {
			// Create a copy of config to avoid mutating the original
			config := make(map[string]any)
			for k, v := range node.Config {
				config[k] = v
			}

			// Add name if not already present
			if _, hasName := config["name"]; !hasName {
				config["name"] = utils.TrimAfterUnderscore(node.ComponentName)
			}

			// Add alias for identification
			alias := fmt.Sprintf("%s_%s", utils.ToCamelCase(node.Name), utils.HashFromConfig(node.Config))
			if _, hasAlias := config["alias"]; !hasAlias {
				config["alias"] = alias
			}

			// For Fluent Bit, ensure tag/match fields are set appropriately
			switch node.ComponentRole {
			case "receiver":
				// Inputs should have a tag field
				if _, hasTag := config["tag"]; !hasTag {
					config["tag"] = fmt.Sprintf("%s.%s", node.ComponentName, alias)
				}
				inputs = append(inputs, config)

			case "processor":
				// Filters should have a match field (default to * for all)
				if _, hasMatch := config["match"]; !hasMatch {
					config["match"] = "*"
				}
				filters = append(filters, config)

			case "exporter":
				// Outputs should have a match field (default to * for all)
				if _, hasMatch := config["match"]; !hasMatch {
					config["match"] = "*"
				}
				outputs = append(outputs, config)

			default:
				return nil, nil, nil, fmt.Errorf("unknown component role: %s", node.ComponentRole)
			}
		}
	}

	if len(inputs) == 0 {
		return nil, nil, nil, fmt.Errorf("pipeline must have at least one input")
	}
	if len(outputs) == 0 {
		return nil, nil, nil, fmt.Errorf("pipeline must have at least one output")
	}

	utils.Logger.Info(fmt.Sprintf("Successfully built Fluent Bit pipeline configurations: inputs=%d filters=%d outputs=%d",
		len(inputs), len(filters), len(outputs)))

	return inputs, filters, outputs, nil
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

// Legacy functions for backward compatibility - delegates to new implementation
func buildPipelines(graph models.PipelineGraph) (map[string]any, map[string]any, map[string]any, Pipelines, error) {
	connectedComponents, err := analyzeGraphStructure(graph, "OTEL")
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return buildOTELConfig(connectedComponents)
}

func buildFluentBitPipeline(graph models.PipelineGraph) ([]any, []any, []any, error) {
	connectedComponents, err := analyzeGraphStructure(graph, "Fluent Bit")
	if err != nil {
		return nil, nil, nil, err
	}
	return buildFluentBitConfig(connectedComponents)
}
