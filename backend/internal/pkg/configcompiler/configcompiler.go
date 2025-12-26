package configcompiler

import (
	"fmt"
	"strconv"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/constants"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
)

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

	connectedComponents, err := analyzeGraphStructure(graph, "OTEL")
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("Failed to analyze graph structure: %v", err))
		return nil, err
	}

	receivers, processors, exporters, pipelines, err := buildOTELConfig(connectedComponents)
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("Failed to build OTEL pipelines: %v", err))
		return nil, err
	}

	utils.Logger.Info(fmt.Sprintf("Successfully compiled OTEL pipeline graph: receivers=%d processors=%d exporters=%d pipelines=%d",
		len(receivers), len(processors), len(exporters), len(pipelines)))

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

	if len(graph.Nodes) == 0 {
		return nil, fmt.Errorf("empty pipeline graph")
	}

	utils.Logger.Info(fmt.Sprintf("Building Fluent Bit pipeline from graph: nodes=%d edges=%d",
		len(graph.Nodes), len(graph.Edges)))

	inputs, filters, outputs, err := buildFluentBitConfig(graph)
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("Failed to build Fluent Bit pipeline: %v", err))
		return nil, err
	}

	utils.Logger.Info(fmt.Sprintf("Successfully compiled Fluent Bit pipeline: inputs=%d filters=%d outputs=%d",
		len(inputs), len(filters), len(outputs)))

	inputs = append(inputs, []any{constants.FluentBitPrometheusInput}...)

	outputs = append(outputs, []any{constants.FluentBitPrometheusOutput}...)

	finalConfig := map[string]any{
		"service": constants.FluentBitService,
		"pipeline": map[string]any{
			"inputs":  inputs,
			"filters": filters,
			"outputs": outputs,
		},
	}

	return &finalConfig, nil
}

// buildOTELConfig builds OTEL-specific configuration from connected components
func buildOTELConfig(connectedComponents [][]models.PipelineNodes) (map[string]any, map[string]any, map[string]any, Pipelines, error) {
	receiversConfig := make(map[string]any)
	processorsConfig := make(map[string]any)
	exportersConfig := make(map[string]any)
	pipelines := make(Pipelines)
	pipelineCounter := 1

	for _, componentNodes := range connectedComponents {
		utils.Logger.Debug(fmt.Sprintf("Building OTEL pipeline configuration: componentNodeCount=%d", len(componentNodes)))

		var commonSupportedSignals []string
		if len(componentNodes) > 0 {
			commonSupportedSignals = componentNodes[0].SupportedSignals
			for _, node := range componentNodes[1:] {
				commonSupportedSignals = intersectSupportedSignals(commonSupportedSignals, node.SupportedSignals)
			}
		}

		var receiverAliases, processorAliases, exporterAliases []string
		for _, node := range componentNodes {
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

// buildFluentBitConfig builds Fluent Bit configuration with proper tag-based routing
func buildFluentBitConfig(graph models.PipelineGraph) ([]any, []any, []any, error) {
	nodesByID := make(map[string]models.PipelineNodes)
	for _, node := range graph.Nodes {
		nodesByID[strconv.Itoa(node.ComponentID)] = node
	}

	outgoingEdges := make(map[string][]string) // source -> targets
	incomingEdges := make(map[string][]string) // target -> sources

	for _, edge := range graph.Edges {
		if _, exists := nodesByID[edge.Source]; !exists {
			return nil, nil, nil, fmt.Errorf("edge references non-existent source node: %s", edge.Source)
		}
		if _, exists := nodesByID[edge.Target]; !exists {
			return nil, nil, nil, fmt.Errorf("edge references non-existent target node: %s", edge.Target)
		}
		outgoingEdges[edge.Source] = append(outgoingEdges[edge.Source], edge.Target)
		incomingEdges[edge.Target] = append(incomingEdges[edge.Target], edge.Source)
	}

	var inputNodes, filterNodes, outputNodes []models.PipelineNodes
	for _, node := range graph.Nodes {
		switch node.ComponentRole {
		case "input":
			inputNodes = append(inputNodes, node)
		case "filter":
			filterNodes = append(filterNodes, node)
		case "output":
			outputNodes = append(outputNodes, node)
		default:
			return nil, nil, nil, fmt.Errorf("unknown component role: %s for node %s", node.ComponentRole, node.Name)
		}
	}

	if len(inputNodes) == 0 {
		return nil, nil, nil, fmt.Errorf("pipeline must have at least one input")
	}
	if len(outputNodes) == 0 {
		return nil, nil, nil, fmt.Errorf("pipeline must have at least one output")
	}

	inputTags := make(map[string]string) // nodeID -> tag
	for _, node := range inputNodes {
		nodeID := strconv.Itoa(node.ComponentID)
		alias := generateAlias(node)

		tag := getConfigString(node.Config, "tag")
		if tag == "" {
			pluginName := utils.TrimAfterUnderscore(node.ComponentName)
			tag = fmt.Sprintf("%s.%s", pluginName, alias)
		}
		inputTags[nodeID] = tag
	}

	// Trace back through graph to determine which input tags each filter/output should match
	nodeMatchTags := make(map[string][]string) // nodeID -> list of tags to match

	for _, node := range append(filterNodes, outputNodes...) {
		nodeID := strconv.Itoa(node.ComponentID)
		matchTags := findUpstreamInputTags(nodeID, inputTags, incomingEdges, nodesByID)
		nodeMatchTags[nodeID] = matchTags
	}

	var inputs []any
	for _, node := range inputNodes {
		nodeID := strconv.Itoa(node.ComponentID)
		config := buildFluentBitNodeConfig(node, inputTags[nodeID], nil)
		inputs = append(inputs, config)
	}

	var filters []any
	for _, node := range filterNodes {
		nodeID := strconv.Itoa(node.ComponentID)
		matchTags := nodeMatchTags[nodeID]
		config := buildFluentBitNodeConfig(node, "", matchTags)
		filters = append(filters, config)
	}

	var outputs []any
	for _, node := range outputNodes {
		nodeID := strconv.Itoa(node.ComponentID)
		matchTags := nodeMatchTags[nodeID]
		config := buildFluentBitNodeConfig(node, "", matchTags)
		outputs = append(outputs, config)
	}

	return inputs, filters, outputs, nil
}
