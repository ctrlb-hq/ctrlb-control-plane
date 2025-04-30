package queue

import (
	"fmt"
	"sync"
	"time"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
	io_prometheus_client "github.com/prometheus/client_model/go"
)

// AgentQueueInterface defines the operations supported by an AgentQueue.
type AgentQueueInterface interface {
	AddAgent(id, hostname, ip string) error
	RemoveAgent(id string) error
	RefreshMonitoring() error
}

type AgentQueueRepositoryInterface interface {
	RefreshMonitoring() ([]AgentStatus, error)
	UpdateAgentMetricsInDB(agg AggregatedAgentMetrics, rt RealtimeAgentMetrics) error
	UpdateAgentStatus(agentID string, status string) error
}

// AgentQueue handles agent monitoring and retry logic
type AgentQueue struct {
	agents          map[string]*AgentStatus
	mutex           sync.RWMutex
	checkQueue      chan string
	workerCount     int
	IntervalSecond  int
	QueueRepository AgentQueueRepositoryInterface
	Metrics         MetricsHelper
}

// NewQueue creates a new AgentQueue
func NewQueue(workerCount int, intervalSec int, queueRepository AgentQueueRepositoryInterface) AgentQueueInterface {
	q := &AgentQueue{
		agents:          make(map[string]*AgentStatus),
		checkQueue:      make(chan string, workerCount*2),
		workerCount:     workerCount,
		IntervalSecond:  intervalSec,
		QueueRepository: queueRepository,
	}
	q.startWorkers()
	q.startRetryScheduler()
	return q
}

// AddAgent adds a new agent to the queue
func (q *AgentQueue) AddAgent(id, hostname, ip string) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	if _, exists := q.agents[id]; exists {
		utils.Logger.Error(fmt.Sprintf("Agent with ID: %s already being monitored.", id))
		return fmt.Errorf("agent with ID: %s already being monitored", id)
	}
	q.agents[id] = &AgentStatus{
		AgentID:        id,
		Hostname:       hostname,
		IP:             ip,
		CurrentStatus:  "unknown",
		RetryRemaining: 3,
		NextCheck:      time.Now(), // eligible immediately
	}
	utils.Logger.Info(fmt.Sprintf("Successfully queued agent with ID: %s.", id))
	return nil
}

// RemoveAgent removes an agent from the queue
func (q *AgentQueue) RemoveAgent(id string) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	delete(q.agents, id)
	utils.Logger.Info(fmt.Sprintf("Successfully removed agent with ID: %s.", id))
	return nil
}

// RefreshMonitoring re-queues all existing agents (used at boot)
func (q *AgentQueue) RefreshMonitoring() error {
	agents, err := q.QueueRepository.RefreshMonitoring()
	if err != nil {
		utils.Logger.Sugar().Errorf("Failed to get existing agents: %v", err)
		return err
	}

	for _, agent := range agents {
		if err := q.AddAgent(agent.AgentID, agent.Hostname, agent.IP); err != nil {
			utils.Logger.Sugar().Errorf("Error adding agent [ID: %v] to queue", agent.AgentID)
		}
	}

	return nil
}

// Internal worker that handles agent check logic
func (q *AgentQueue) startWorkers() {
	for i := 0; i < q.workerCount; i++ {
		go q.worker()
	}
}

// Retry scheduler that re-enqueues agents based on NextCheck
func (q *AgentQueue) startRetryScheduler() {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			now := time.Now()
			q.mutex.RLock()
			for id, agent := range q.agents {
				if agent.NextCheck.Before(now) || agent.NextCheck.Equal(now) {
					select {
					case q.checkQueue <- id:
						// throttle next check
						agent.NextCheck = now.Add(time.Duration(q.IntervalSecond) * time.Second)
					default:
						utils.Logger.Sugar().Errorf("checkQueue full, skipping agent %s", id)
					}
				}
			}
			q.mutex.RUnlock()
		}
	}()
}

// Worker loop
func (q *AgentQueue) worker() {
	for agentID := range q.checkQueue {
		q.mutex.RLock()
		agent, exists := q.agents[agentID]
		q.mutex.RUnlock()

		if !exists {
			continue
		}

		if err := q.checkAgentStatus(agent); err != nil {
			q.mutex.Lock()
			agent.RetryRemaining--
			if agent.RetryRemaining <= 0 {
				agent.CurrentStatus = "disconnected"
			} else {
				agent.CurrentStatus = "unknown"
			}
			_ = q.QueueRepository.UpdateAgentStatus(agent.AgentID, agent.CurrentStatus)
			q.mutex.Unlock()

			utils.Logger.Sugar().Errorf("Error checking status of agent [ID:%s], Attempts remaining: %v", agent.AgentID, agent.RetryRemaining)
		} else {
			q.mutex.Lock()
			agent.RetryRemaining = 3
			agent.CurrentStatus = "connected"
			_ = q.QueueRepository.UpdateAgentStatus(agent.AgentID, "connected")
			q.mutex.Unlock()
		}

		if agent.RetryRemaining <= 0 {
			_ = q.RemoveAgent(agentID)
		}
	}
}

// checkAgentStatus fetches Prometheus metrics from agent
func (q *AgentQueue) checkAgentStatus(agent *AgentStatus) error {
	endpoints := []string{
		fmt.Sprintf("http://%s:8888/metrics", agent.Hostname),
		fmt.Sprintf("http://%s:8888/metrics", agent.IP),
	}

	var metrics map[string]*io_prometheus_client.MetricFamily
	var err error

	for _, url := range endpoints {
		metrics, err = q.Metrics.Fetch(url)
		if err == nil {
			break
		}
	}

	if err != nil {
		return fmt.Errorf("failed to fetch metrics from both hostname and IP: %w", err)
	}

	agg := AggregatedAgentMetrics{
		AgentID:           agent.AgentID,
		LogsRateSent:      q.Metrics.ExtractValue(metrics, "otelcol_exporter_sent_log_records"),
		TracesRateSent:    q.Metrics.ExtractValue(metrics, "otelcol_exporter_sent_spans"),
		MetricsRateSent:   q.Metrics.ExtractValue(metrics, "otelcol_exporter_sent_metric_points"),
		DataSentBytes:     q.Metrics.ExtractValue(metrics, "otelcol_exporter_sent_bytes"),
		DataReceivedBytes: q.Metrics.ExtractValue(metrics, "otelcol_receiver_accepted_bytes"),
		Status:            "connected",
		UpdatedAt:         time.Now().Unix(),
	}

	rt := RealtimeAgentMetrics{
		AgentID:           agent.AgentID,
		LogsRateSent:      agg.LogsRateSent,
		TracesRateSent:    agg.TracesRateSent,
		MetricsRateSent:   agg.MetricsRateSent,
		DataSentBytes:     agg.DataSentBytes,
		DataReceivedBytes: agg.DataReceivedBytes,
		CPUUtilization:    q.Metrics.ExtractValue(metrics, "otelcol_process_cpu_seconds_total"),
		MemoryUtilization: q.Metrics.ExtractValue(metrics, "otelcol_process_memory_rss"),
		Timestamp:         time.Now().Unix(),
	}

	return q.QueueRepository.UpdateAgentMetricsInDB(agg, rt)
}

// GetAgent returns agent by ID — test helper
func (q *AgentQueue) GetAgent(id string) (*AgentStatus, bool) {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	agent, exists := q.agents[id]
	return agent, exists
}
