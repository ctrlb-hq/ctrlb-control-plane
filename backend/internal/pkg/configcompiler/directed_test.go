package configcompiler

import (
	"encoding/json"
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/stretchr/testify/require"
)

func linearGraph() models.PipelineGraph {
	return models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{ComponentID: 1, Name: "Source", ComponentRole: "receiver", ComponentName: "otlp_receiver", SupportedSignals: []string{"logs", "metrics"}, Config: map[string]any{}},
			{ComponentID: 2, Name: "First", ComponentRole: "processor", ComponentName: "memorylimiter_processor", SupportedSignals: []string{"logs", "metrics"}, Config: map[string]any{"limit_mib": 400}},
			{ComponentID: 3, Name: "Second", ComponentRole: "processor", ComponentName: "batch_processor", SupportedSignals: []string{"logs", "metrics"}, Config: map[string]any{"timeout": "5s"}},
			{ComponentID: 4, Name: "Destination", ComponentRole: "exporter", ComponentName: "otlphttp_exporter", SupportedSignals: []string{"logs", "metrics"}, Config: map[string]any{"endpoint": "https://example.com", "sending_queue": map[string]any{"enabled": false}}},
		},
		Edges: []models.PipelineEdges{{Source: "1", Target: "2"}, {Source: "2", Target: "3"}, {Source: "3", Target: "4"}},
	}
}

func compileBytes(t *testing.T, graph models.PipelineGraph) []byte {
	t.Helper()
	config, err := CompileGraphToJSON(graph)
	require.NoError(t, err)
	data, err := json.Marshal(config)
	require.NoError(t, err)
	return data
}

func TestDirectedProcessorOrderIsStableAcrossGraphShufflesAndDatabaseIDs(t *testing.T) {
	graph := linearGraph()
	expected := compileBytes(t, graph)
	rng := rand.New(rand.NewSource(42))
	for index := 0; index < 40; index++ {
		graph := linearGraph()
		rng.Shuffle(len(graph.Nodes), func(i, j int) { graph.Nodes[i], graph.Nodes[j] = graph.Nodes[j], graph.Nodes[i] })
		rng.Shuffle(len(graph.Edges), func(i, j int) { graph.Edges[i], graph.Edges[j] = graph.Edges[j], graph.Edges[i] })
		for n := range graph.Nodes {
			graph.Nodes[n].ComponentID += index * 10
		}
		for n := range graph.Edges {
			source, _ := strconv.Atoi(graph.Edges[n].Source)
			target, _ := strconv.Atoi(graph.Edges[n].Target)
			graph.Edges[n].Source = strconv.Itoa(source + index*10)
			graph.Edges[n].Target = strconv.Itoa(target + index*10)
		}
		require.Equal(t, string(expected), string(compileBytes(t, graph)))
	}
	config, err := CompileGraphToJSON(graph)
	require.NoError(t, err)
	for _, pipeline := range (*config)["service"].(map[string]any)["pipelines"].(Pipelines) {
		require.Len(t, pipeline.Processors, 2)
		require.True(t, strings.HasPrefix(pipeline.Processors[0], "memory_limiter/First_"))
		require.True(t, strings.HasPrefix(pipeline.Processors[1], "batch/Second_"))
	}
	require.Contains(t, string(expected), `"enabled":false`)
}

func TestSignalsFollowCompatiblePathsAndSupportFanInAndExporterFanOut(t *testing.T) {
	graph := linearGraph()
	graph.Nodes[0].ComponentName, graph.Nodes[0].SupportedSignals = "file_log_receiver", []string{"logs"}
	metrics := graph.Nodes[0]
	metrics.ComponentID, metrics.Name, metrics.ComponentName, metrics.SupportedSignals = 5, "Metrics", "kubelet_stats_receiver", []string{"metrics"}
	graph.Nodes = append(graph.Nodes, metrics)
	graph.Edges = append(graph.Edges, models.PipelineEdges{Source: "5", Target: "2"})
	otherExporter := graph.Nodes[3]
	otherExporter.ComponentID, otherExporter.Name = 6, "Another destination"
	graph.Nodes = append(graph.Nodes, otherExporter)
	graph.Edges = append(graph.Edges, models.PipelineEdges{Source: "3", Target: "6"})
	config, err := CompileGraphToJSON(graph)
	require.NoError(t, err)
	pipelines := (*config)["service"].(map[string]any)["pipelines"].(Pipelines)
	require.Len(t, pipelines, 2)
	for name, pipeline := range pipelines {
		require.Len(t, pipeline.Receivers, 1)
		require.Len(t, pipeline.Exporters, 2)
		if strings.HasPrefix(name, "logs/") {
			require.True(t, strings.HasPrefix(pipeline.Receivers[0], "file_log/"))
		} else {
			require.True(t, strings.HasPrefix(name, "metrics/"))
			require.True(t, strings.HasPrefix(pipeline.Receivers[0], "kubelet_stats/"))
		}
	}
	// Two receivers following the same signal and processor chain share a pipeline.
	graph.Nodes[4].ComponentName, graph.Nodes[4].SupportedSignals = "filelog_receiver", []string{"logs"}
	config, err = CompileGraphToJSON(graph)
	require.NoError(t, err)
	pipelines = (*config)["service"].(map[string]any)["pipelines"].(Pipelines)
	require.Len(t, pipelines, 1)
	for _, pipeline := range pipelines {
		require.Len(t, pipeline.Receivers, 2)
	}
}

func TestInvalidDirectedGraphsAreRejected(t *testing.T) {
	tests := map[string]func(*models.PipelineGraph){
		"cycle": func(g *models.PipelineGraph) {
			g.Edges = append(g.Edges, models.PipelineEdges{Source: "3", Target: "2"})
		},
		"missing endpoint": func(g *models.PipelineGraph) { g.Edges[0].Target = "999" },
		"exporter as source": func(g *models.PipelineGraph) {
			g.Edges = append(g.Edges, models.PipelineEdges{Source: "4", Target: "2"})
		},
		"receiver as target": func(g *models.PipelineGraph) {
			g.Edges = append(g.Edges, models.PipelineEdges{Source: "3", Target: "1"})
		},
		"duplicate node ID":    func(g *models.PipelineGraph) { g.Nodes[1].ComponentID = 1 },
		"duplicate edge":       func(g *models.PipelineGraph) { g.Edges = append(g.Edges, g.Edges[0]) },
		"dead end":             func(g *models.PipelineGraph) { g.Edges = g.Edges[:2] },
		"no compatible signal": func(g *models.PipelineGraph) { g.Nodes[3].SupportedSignals = []string{"traces"} },
		"unknown signal":       func(g *models.PipelineGraph) { g.Nodes[0].SupportedSignals = []string{"profiles"} },
		"branching processor paths": func(g *models.PipelineGraph) {
			g.Edges = append(g.Edges, models.PipelineEdges{Source: "1", Target: "3"})
		},
		"processor and exporter branch": func(g *models.PipelineGraph) {
			g.Edges = append(g.Edges, models.PipelineEdges{Source: "2", Target: "4"})
		},
		"unused component": func(g *models.PipelineGraph) {
			node := g.Nodes[3]
			node.ComponentID, node.Name = 9, "Unused"
			g.Nodes = append(g.Nodes, node)
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			graph := linearGraph()
			mutate(&graph)
			_, err := CompileGraphToJSON(graph)
			require.Error(t, err)
		})
	}
}

func TestComponentNamesPreserveCanonicalUnderscoresAndLegacyCatalogNames(t *testing.T) {
	for name, expected := range map[string]string{
		"memorylimiter_processor": "memory_limiter", "memory_limiter_processor": "memory_limiter",
		"probabilisticsampler_processor": "probabilistic_sampler", "tailsampling_processor": "tail_sampling",
		"k8s_attributes_processor": "k8s_attributes", "k8sattributes_processor": "k8s_attributes",
	} {
		require.Equal(t, expected, componentType(name, "processor"))
	}
	require.Equal(t, "otlp", componentType("otlp_grpc_exporter", "exporter"))
	require.Equal(t, "otlp_http", componentType("otlp_http_exporter", "exporter"))
	require.Equal(t, "file_log", componentType("filelog_receiver", "receiver"))
}

func TestFileStorageIsDerivedAndInvalidStorageDoesNotPanic(t *testing.T) {
	graph := linearGraph()
	graph.Nodes[0].ComponentName, graph.Nodes[0].Config = "filelog_receiver", map[string]any{"storage": "file_storage"}
	config, err := CompileGraphToJSON(graph)
	require.NoError(t, err)
	require.Contains(t, (*config)["extensions"], "file_storage")
	require.Equal(t, []string{"file_storage"}, (*config)["service"].(map[string]any)["extensions"])
	for _, storage := range []any{map[string]any{}, "unsupported_storage", 5} {
		graph.Nodes[0].Config["storage"] = storage
		_, err := CompileGraphToJSON(graph)
		require.Error(t, err)
	}
}
