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

// --- Mock Implementations ---

type mockRepo struct {
	statusLog []string
	mu        sync.Mutex
}

func (r *mockRepo) RefreshMonitoring() ([]queue.AgentStatus, error) {
	return nil, nil
}

func (r *mockRepo) UpdateAgentStatus(agentID string, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.statusLog = append(r.statusLog, status)
	return nil
}

func (r *mockRepo) UpdateAgentMetricsInDB(_ queue.AggregatedAgentMetrics, _ queue.RealtimeAgentMetrics) error {
	return nil
}

type mockMetricsHelper struct {
	fetchErrCount int
	callCount     int
}

func (m *mockMetricsHelper) Fetch(url string) (map[string]*io_prometheus_client.MetricFamily, error) {
	m.callCount++
	if m.callCount <= m.fetchErrCount {
		return nil, errors.New("simulated fetch error")
	}
	return map[string]*io_prometheus_client.MetricFamily{}, nil
}

func (m *mockMetricsHelper) ExtractValue(map[string]*io_prometheus_client.MetricFamily, string) float64 {
	return 0
}

// --- Test Case ---

func TestAgentQueue_RetryScheduler_RemovesAfterFailures(t *testing.T) {
	repo := &mockRepo{}
	helper := &mockMetricsHelper{
		fetchErrCount: 999, // always fail
	}

	// Create the AgentQueue with 1 worker and 1s retry interval
	q := queue.NewQueue(1, 1, repo).(*queue.AgentQueue)
	q.Metrics = helper

	// Add agent
	err := q.AddAgent("agent-1", "agent1.test.local", "192.168.1.100")
	assert.NoError(t, err)

	// Poll for agent removal within timeout
	timeout := time.After(6 * time.Second)
	tick := time.Tick(200 * time.Millisecond)

REMOVAL_WAIT:
	for {
		select {
		case <-timeout:
			t.Fatal("Agent was not removed within timeout")
		case <-tick:
			_, exists := q.GetAgent("agent-1")
			if !exists {
				break REMOVAL_WAIT
			}
		}
	}

	// Safely copy log without deadlocking
	repo.mu.Lock()
	statuses := make([]string, len(repo.statusLog))
	copy(statuses, repo.statusLog)
	repo.mu.Unlock()

	assert.Equal(t, []string{"unknown", "unknown", "disconnected"}, statuses, "Unexpected agent status sequence")
}
