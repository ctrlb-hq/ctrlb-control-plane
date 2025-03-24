package configcompiler

import (
	"encoding/json"
	"fmt"

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
			"telemetry": TelemetryService, // inject telemetry under service
		},
	}

	return &finalConfig, nil
}

func BuildPipelines(graph models.PipelineGraph) (map[string]any, map[string]any, map[string]any, Pipelines, error) {
	utils.Logger.Info(fmt.Sprintf("Building pipelines from graph: nodes=%d edges=%d",
		len(graph.Nodes), len(graph.Edges)))

	// Index nodes by ID
	nodeByID := map[int]models.PipelineComponent{}
	aliasByID := map[int]string{}

	for _, node := range graph.Nodes {
		alias := node.PluginName + "/" + node.Name
		nodeByID[node.ComponentID] = node
		aliasByID[node.ComponentID] = alias
	}

	// Build adjacency list
	adjList := map[int][]int{}
	for _, edge := range graph.Edges {
		adjList[edge.FromComponentID] = append(adjList[edge.FromComponentID], edge.ToComponentID)
		adjList[edge.ToComponentID] = append(adjList[edge.ToComponentID], edge.FromComponentID)
	}

	// Track visited nodes
	visited := map[int]bool{}
	components := [][]models.PipelineComponent{}

	// BFS to find connected components
	for id := range nodeByID {
		if visited[id] {
			continue
		}

		utils.Logger.Debug(fmt.Sprintf("Processing new component: startNodeId=%d", id))

		var queue []int
		var component []models.PipelineComponent

		queue = append(queue, id)
		visited[id] = true

		for len(queue) > 0 {
			curr := queue[0]
			queue = queue[1:]

			component = append(component, nodeByID[curr])

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
	receivers := map[string]any{}
	processors := map[string]any{}
	exporters := map[string]any{}
	pipelines := Pipelines{}

	for i, nodes := range components {
		utils.Logger.Debug(fmt.Sprintf("Building pipeline configuration: pipelineIndex=%d componentCount=%d",
			i, len(nodes)))

		var r, p, e []string

		for _, n := range nodes {
			alias := n.PluginName + "/" + n.Name
			var parsedConfig map[string]any
			if err := json.Unmarshal([]byte(n.Config), &parsedConfig); err != nil {
				utils.Logger.Error(fmt.Sprintf("Failed to parse component config: alias=%s error=%v",
					alias, err))
				return nil, nil, nil, nil, fmt.Errorf("invalid config for %s: %v", alias, err)
			}

			switch n.ComponentRole {
			case "receiver":
				r = append(r, alias)
				receivers[alias] = parsedConfig
			case "processor":
				p = append(p, alias)
				processors[alias] = parsedConfig
			case "exporter":
				e = append(e, alias)
				exporters[alias] = parsedConfig
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
