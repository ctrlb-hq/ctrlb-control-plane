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
	receivers, processors, exporters, pipelines, err := BuildPipelines(graph)
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

func BuildPipelines(graph models.PipelineGraph) (map[string]any, map[string]any, map[string]any, Pipelines, error) {
	if len(graph.Nodes) == 0 {
		return nil, nil, nil, nil, fmt.Errorf("empty pipeline graph")
	}

	utils.Logger.Info(fmt.Sprintf("Building pipelines from graph: nodes=%d edges=%d",
		len(graph.Nodes), len(graph.Edges)))

	// Index nodes by ID
	nodeByID := make(map[string]models.PipelineComponent)
	aliasByID := make(map[string]string)

	for _, node := range graph.Nodes {
		alias := node.ComponentName + "/" + node.Name
		nodeByID[strconv.Itoa(node.ComponentID)] = node
		aliasByID[strconv.Itoa(node.ComponentID)] = alias
	}

	// Build adjacency list
	adjList := make(map[string][]string)
	for _, edge := range graph.Edges {
		if _, exists := nodeByID[edge.Source]; !exists {
			return nil, nil, nil, nil, fmt.Errorf("edge references non-existent source node: %s", edge.Source)
		}
		if _, exists := nodeByID[edge.Target]; !exists {
			return nil, nil, nil, nil, fmt.Errorf("edge references non-existent target node: %s", edge.Target)
		}
		adjList[edge.Source] = append(adjList[edge.Source], edge.Target)
		adjList[edge.Target] = append(adjList[edge.Target], edge.Source)
	}

	// Track visited nodes
	visited := make(map[string]bool)
	components := make([][]models.PipelineComponent, 0)

	// BFS to find connected components
	for id := range nodeByID {
		if visited[id] {
			continue
		}

		utils.Logger.Debug(fmt.Sprintf("Processing new component: startNodeId=%s", id))

		queue := []string{id}
		var component []models.PipelineComponent
		visited[id] = true

		for len(queue) > 0 {
			curr := queue[0]
			queue = queue[1:]

			node, exists := nodeByID[curr]
			if !exists {
				return nil, nil, nil, nil, fmt.Errorf("invalid node reference: %s", curr)
			}
			component = append(component, node)

			for _, neighbor := range adjList[curr] {
				if !visited[neighbor] {
					visited[neighbor] = true
					queue = append(queue, neighbor)
				}
			}
		}

		components = append(components, component)
	}

	// Prepare the output maps
	receivers := make(map[string]any)
	processors := make(map[string]any)
	exporters := make(map[string]any)
	pipelines := make(Pipelines)

	for i, nodes := range components {
		utils.Logger.Debug(fmt.Sprintf("Building pipeline configuration: pipelineIndex=%d componentCount=%d",
			i, len(nodes)))

		var r, p, e []string

		for _, n := range nodes {
			alias := utils.TrimAfterUnderscore(n.ComponentName) + "/" + utils.ToCamelCase(n.Name)

			switch n.ComponentRole {
			case "receiver":
				r = append(r, alias)
				receivers[alias] = n.Config
			case "processor":
				p = append(p, alias)
				processors[alias] = n.Config
			case "exporter":
				e = append(e, alias)
				exporters[alias] = n.Config
			default:
				return nil, nil, nil, nil, fmt.Errorf("unknown component role: %s", n.ComponentRole)
			}
		}

		pipelineName := fmt.Sprintf("pipeline_%d", i+1)
		pipelines[pipelineName] = Pipeline{
			Receivers:  r,
			Processors: p,
			Exporters:  e,
		}
	}

	utils.Logger.Info(fmt.Sprintf("Successfully built pipeline configurations: receivers=%d processors=%d exporters=%d",
		len(receivers), len(processors), len(exporters)))

	return receivers, processors, exporters, pipelines, nil
}
