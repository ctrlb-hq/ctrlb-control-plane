package configcompiler

import (
	"testing"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Graph Utility Tests
// =============================================================================

func TestValidateGraph(t *testing.T) {
	t.Run("empty graph", func(t *testing.T) {
		graph := models.PipelineGraph{}
		err := ValidateGraph(graph)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty pipeline graph")
	})

	t.Run("invalid edge source", func(t *testing.T) {
		graph := models.PipelineGraph{
			Nodes: []models.PipelineNodes{
				{ComponentID: 1, Name: "node1"},
			},
			Edges: []models.PipelineEdges{
				{Source: "999", Target: "1"},
			},
		}
		err := ValidateGraph(graph)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "non-existent source")
	})

	t.Run("invalid edge target", func(t *testing.T) {
		graph := models.PipelineGraph{
			Nodes: []models.PipelineNodes{
				{ComponentID: 1, Name: "node1"},
			},
			Edges: []models.PipelineEdges{
				{Source: "1", Target: "999"},
			},
		}
		err := ValidateGraph(graph)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "non-existent target")
	})

	t.Run("valid graph", func(t *testing.T) {
		graph := createSampleGraph()
		err := ValidateGraph(graph)
		assert.NoError(t, err)
	})
}

func TestBuildGraphState(t *testing.T) {
	graph := createSampleGraph()
	state := BuildGraphState(graph)

	assert.Len(t, state.NodesByID, 3)
	assert.Contains(t, state.OutEdges, "1")
	assert.Contains(t, state.OutEdges["1"], "2")
	assert.Contains(t, state.InEdges, "2")
	assert.Contains(t, state.InEdges["2"], "1")
}

func TestDetectCycle(t *testing.T) {
	t.Run("no cycle - linear", func(t *testing.T) {
		graph := models.PipelineGraph{
			Nodes: []models.PipelineNodes{
				{ComponentID: 1, Name: "a"},
				{ComponentID: 2, Name: "b"},
				{ComponentID: 3, Name: "c"},
			},
			Edges: []models.PipelineEdges{
				{Source: "1", Target: "2"},
				{Source: "2", Target: "3"},
			},
		}
		state := BuildGraphState(graph)
		cycle := DetectCycle(state)
		assert.Nil(t, cycle)
	})

	t.Run("no cycle - diamond", func(t *testing.T) {
		graph := models.PipelineGraph{
			Nodes: []models.PipelineNodes{
				{ComponentID: 1, Name: "a"},
				{ComponentID: 2, Name: "b"},
				{ComponentID: 3, Name: "c"},
				{ComponentID: 4, Name: "d"},
			},
			Edges: []models.PipelineEdges{
				{Source: "1", Target: "2"},
				{Source: "1", Target: "3"},
				{Source: "2", Target: "4"},
				{Source: "3", Target: "4"},
			},
		}
		state := BuildGraphState(graph)
		cycle := DetectCycle(state)
		assert.Nil(t, cycle)
	})

	t.Run("self loop", func(t *testing.T) {
		graph := models.PipelineGraph{
			Nodes: []models.PipelineNodes{
				{ComponentID: 1, Name: "a"},
			},
			Edges: []models.PipelineEdges{
				{Source: "1", Target: "1"},
			},
		}
		state := BuildGraphState(graph)
		cycle := DetectCycle(state)
		assert.NotNil(t, cycle)
	})

	t.Run("simple cycle", func(t *testing.T) {
		graph := models.PipelineGraph{
			Nodes: []models.PipelineNodes{
				{ComponentID: 1, Name: "a"},
				{ComponentID: 2, Name: "b"},
				{ComponentID: 3, Name: "c"},
			},
			Edges: []models.PipelineEdges{
				{Source: "1", Target: "2"},
				{Source: "2", Target: "3"},
				{Source: "3", Target: "1"},
			},
		}
		state := BuildGraphState(graph)
		cycle := DetectCycle(state)
		assert.NotNil(t, cycle)
		assert.GreaterOrEqual(t, len(cycle), 2)
	})
}

func TestTopologicalSort(t *testing.T) {
	t.Run("linear graph", func(t *testing.T) {
		graph := models.PipelineGraph{
			Nodes: []models.PipelineNodes{
				{ComponentID: 1, Name: "a"},
				{ComponentID: 2, Name: "b"},
				{ComponentID: 3, Name: "c"},
			},
			Edges: []models.PipelineEdges{
				{Source: "1", Target: "2"},
				{Source: "2", Target: "3"},
			},
		}
		state := BuildGraphState(graph)

		sorted, err := TopologicalSort([]string{"1", "2", "3"}, state)
		assert.NoError(t, err)
		assert.Equal(t, []string{"1", "2", "3"}, sorted)
	})

	t.Run("diamond graph", func(t *testing.T) {
		graph := models.PipelineGraph{
			Nodes: []models.PipelineNodes{
				{ComponentID: 1, Name: "a"},
				{ComponentID: 2, Name: "b"},
				{ComponentID: 3, Name: "c"},
				{ComponentID: 4, Name: "d"},
			},
			Edges: []models.PipelineEdges{
				{Source: "1", Target: "2"},
				{Source: "1", Target: "3"},
				{Source: "2", Target: "4"},
				{Source: "3", Target: "4"},
			},
		}
		state := BuildGraphState(graph)

		sorted, err := TopologicalSort([]string{"1", "2", "3", "4"}, state)
		assert.NoError(t, err)

		// 1 must come first, 4 must come last
		assert.Equal(t, "1", sorted[0])
		assert.Equal(t, "4", sorted[3])
	})

	t.Run("subset of nodes", func(t *testing.T) {
		graph := models.PipelineGraph{
			Nodes: []models.PipelineNodes{
				{ComponentID: 1, Name: "a"},
				{ComponentID: 2, Name: "b"},
				{ComponentID: 3, Name: "c"},
			},
			Edges: []models.PipelineEdges{
				{Source: "1", Target: "2"},
				{Source: "2", Target: "3"},
			},
		}
		state := BuildGraphState(graph)

		// Only sort nodes 2 and 3
		sorted, err := TopologicalSort([]string{"2", "3"}, state)
		assert.NoError(t, err)
		assert.Equal(t, []string{"2", "3"}, sorted)
	})
}

func TestIntersectSignals(t *testing.T) {
	t.Run("common signals", func(t *testing.T) {
		nodes := []models.PipelineNodes{
			{SupportedSignals: []string{"logs", "metrics", "traces"}},
			{SupportedSignals: []string{"logs", "metrics"}},
			{SupportedSignals: []string{"logs"}},
		}

		result := IntersectSignals(nodes)
		assert.Equal(t, []string{"logs"}, result)
	})

	t.Run("no common signals", func(t *testing.T) {
		nodes := []models.PipelineNodes{
			{SupportedSignals: []string{"logs"}},
			{SupportedSignals: []string{"metrics"}},
		}

		result := IntersectSignals(nodes)
		assert.Empty(t, result)
	})

	t.Run("identical signals", func(t *testing.T) {
		nodes := []models.PipelineNodes{
			{SupportedSignals: []string{"logs", "metrics"}},
			{SupportedSignals: []string{"logs", "metrics"}},
		}

		result := IntersectSignals(nodes)
		assert.Len(t, result, 2)
		assert.Contains(t, result, "logs")
		assert.Contains(t, result, "metrics")
	})

	t.Run("empty nodes", func(t *testing.T) {
		result := IntersectSignals([]models.PipelineNodes{})
		assert.Nil(t, result)
	})

	t.Run("single node", func(t *testing.T) {
		nodes := []models.PipelineNodes{
			{SupportedSignals: []string{"logs", "metrics"}},
		}

		result := IntersectSignals(nodes)
		assert.Len(t, result, 2)
	})
}

func TestGetNodesByRole(t *testing.T) {
	graph := models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{ComponentID: 1, ComponentRole: "receiver"},
			{ComponentID: 2, ComponentRole: "processor"},
			{ComponentID: 3, ComponentRole: "exporter"},
			{ComponentID: 4, ComponentRole: "input"},
			{ComponentID: 5, ComponentRole: "filter"},
			{ComponentID: 6, ComponentRole: "output"},
		},
		Edges: []models.PipelineEdges{},
	}

	state := BuildGraphState(graph)
	receivers, processors, exporters, inputs, filters, outputs := GetNodesByRole(state)

	assert.Len(t, receivers, 1)
	assert.Len(t, processors, 1)
	assert.Len(t, exporters, 1)
	assert.Len(t, inputs, 1)
	assert.Len(t, filters, 1)
	assert.Len(t, outputs, 1)
}

