package queue_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/pkg/queue"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

type mockRepo struct {
	mu         sync.Mutex
	statusLog  []string
	metricsSet bool
}

func (m *mockRepo) RefreshMonitoring() ([]queue.AgentStatus, error) {
	return nil, nil
}

func (m *mockRepo) UpdateAgentMetricsInDB(agg queue.AggregatedAgentMetrics, rt queue.RealtimeAgentMetrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metricsSet = true
	return nil
}

func (m *mockRepo) UpdateAgentStatus(agentID string, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.statusLog = append(m.statusLog, status)
	return nil
}

type mockMetricsHelper struct {
	fetchErrCount int
}

func (m *mockMetricsHelper) Fetch(url string) (map[string]*io_prometheus_client.MetricFamily, error) {
	if m.fetchErrCount > 0 {
		m.fetchErrCount--
		return nil, errors.New("mock fetch error")
	}

	val := 10.0
	return map[string]*io_prometheus_client.MetricFamily{
		"otelcol_exporter_sent_log_records": {
			Metric: []*io_prometheus_client.Metric{{Counter: &io_prometheus_client.Counter{Value: &val}}},
		},
		"otelcol_exporter_sent_spans": {
			Metric: []*io_prometheus_client.Metric{{Counter: &io_prometheus_client.Counter{Value: &val}}},
		},
		"otelcol_exporter_sent_metric_points": {
			Metric: []*io_prometheus_client.Metric{{Counter: &io_prometheus_client.Counter{Value: &val}}},
		},
		"otelcol_exporter_sent_bytes": {
			Metric: []*io_prometheus_client.Metric{{Counter: &io_prometheus_client.Counter{Value: &val}}},
		},
		"otelcol_receiver_accepted_bytes": {
			Metric: []*io_prometheus_client.Metric{{Counter: &io_prometheus_client.Counter{Value: &val}}},
		},
		"otelcol_process_cpu_seconds_total": {
			Metric: []*io_prometheus_client.Metric{{Counter: &io_prometheus_client.Counter{Value: &val}}},
		},
		"otelcol_process_memory_rss": {
			Metric: []*io_prometheus_client.Metric{{Gauge: &io_prometheus_client.Gauge{Value: &val}}},
		},
	}, nil
}

func (m *mockMetricsHelper) ExtractValue(metrics map[string]*io_prometheus_client.MetricFamily, name string) float64 {
	if mf, ok := metrics[name]; ok {
		for _, metric := range mf.Metric {
			if metric.GetGauge() != nil {
				return metric.GetGauge().GetValue()
			}
			if metric.GetCounter() != nil {
				return metric.GetCounter().GetValue()
			}
		}
	}
	return 0
}

func TestAgentQueue_CheckAgentStatus_RetriesAndRemoves(t *testing.T) {
	repo := &mockRepo{}
	helper := &mockMetricsHelper{fetchErrCount: 3}

	q := queue.NewQueue(1, 1, repo).(*queue.AgentQueue)
	q.Metrics = helper

	err := q.AddAgent("agent-1", "localhost", "127.0.0.1")
	assert.NoError(t, err)

	// Simulate 3 failed checks
	for i := 0; i < 3; i++ {
		q.CheckAllAgents()
		time.Sleep(50 * time.Millisecond)
	}

	_, exists := q.GetAgent("agent-1")
	assert.False(t, exists, "Agent should be removed after 3 failures")

	repo.mu.Lock()
	defer repo.mu.Unlock()
	assert.Equal(t, []string{"unknown", "unknown", "disconnected"}, repo.statusLog)
}

func TestAgentQueue_CheckAgentStatus_Success(t *testing.T) {
	repo := &mockRepo{}
	helper := &mockMetricsHelper{}

	q := queue.NewQueue(1, 1, repo).(*queue.AgentQueue)
	q.Metrics = helper

	err := q.AddAgent("agent-2", "localhost", "127.0.0.1")
	assert.NoError(t, err)

	q.CheckAllAgents()
	time.Sleep(50 * time.Millisecond)

	agent, exists := q.GetAgent("agent-2")
	assert.True(t, exists, "Agent should still exist after successful check")
	assert.Equal(t, "connected", agent.CurrentStatus)

	repo.mu.Lock()
	assert.True(t, repo.metricsSet)
	repo.mu.Unlock()
}
