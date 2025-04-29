package queue

import (
	"database/sql"
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
	StartStatusCheck()
}

type AgentQueue struct {
	agents          map[string]*AgentStatus
	mutex           sync.RWMutex
	checkQueue      chan string
	workerCount     int
	IntervalMins    int
	QueueRepository *QueueRepository
}

// NewQueue creates a new AgentQueue with the specified number of workers.
func NewQueue(workerCount int, intervalMins int, db *sql.DB) *AgentQueue {
	queueRepository := NewQueueRepository(db)
	q := &AgentQueue{
		agents:          make(map[string]*AgentStatus),
		checkQueue:      make(chan string, workerCount*2),
		workerCount:     workerCount,
		IntervalMins:    intervalMins,
		QueueRepository: queueRepository,
	}
	q.startWorkers()
	return q
}

// AddAgent adds a new agent to the queue.
func (q *AgentQueue) AddAgent(id, Hostname, IP string) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	if _, exists := q.agents[id]; exists {
		utils.Logger.Error(fmt.Sprintf("Agent with ID: %s already being monitored.", id))
		return fmt.Errorf("agent with ID: %s already being monitored", id)
	}
	q.agents[id] = &AgentStatus{
		AgentID:        id,
		Hostname:       Hostname,
		IP:             IP,
		CurrentStatus:  "unknown",
		RetryRemaining: 3,
	}
	utils.Logger.Info(fmt.Sprintf("Successfully queued agent with ID: %s.", id))
	return nil
}

// RemoveAgent removes an agent from the queue by ID.
func (q *AgentQueue) RemoveAgent(id string) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	delete(q.agents, id)
	utils.Logger.Info(fmt.Sprintf("Successfully removed agent with ID: %s.", id))
	return nil
}

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

// StartStatusCheck starts a goroutine that checks the status of all agents at regular intervals.
func (q *AgentQueue) StartStatusCheck() {
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for range ticker.C {
			q.checkAllAgents()
		}
	}()
}

// Internal methods
func (q *AgentQueue) checkAllAgents() {
	q.mutex.RLock()
	agentIDs := make([]string, 0, len(q.agents))
	for id := range q.agents {
		agentIDs = append(agentIDs, id)
	}
	q.mutex.RUnlock()

	for _, id := range agentIDs {
		q.checkQueue <- id
	}
}

func (q *AgentQueue) startWorkers() {
	for i := 0; i < q.workerCount; i++ {
		go q.worker()
	}
}

func (q *AgentQueue) worker() {
	for agentID := range q.checkQueue {
		q.mutex.RLock()
		agent, exists := q.agents[agentID]
		q.mutex.RUnlock()

		if exists {
			if err := q.checkAgentStatus(agent); err != nil {
				q.mutex.Lock()

				agent.RetryRemaining--
				if agent.RetryRemaining <= 0 {
					agent.CurrentStatus = "disconnected"
				} else {
					agent.CurrentStatus = "unknown"
				}

				q.QueueRepository.UpdateAgentStatus(agent.AgentID, agent.CurrentStatus)
				utils.Logger.Sugar().Errorf("Error checking status of agent [ID:%s], Attempts remaining: %v", agent.AgentID, agent.RetryRemaining)
				q.mutex.Unlock()
			} else {
				q.mutex.Lock()
				agent.RetryRemaining = 3
				agent.CurrentStatus = "connected"
				q.mutex.Unlock()
			}
			if agent.RetryRemaining <= 0 {
				q.RemoveAgent(agentID)
			}
		}
	}
}

func (q *AgentQueue) checkAgentStatus(agent *AgentStatus) error {
	endpoints := []string{
		fmt.Sprintf("http://%s:8888/metrics", agent.Hostname),
		fmt.Sprintf("http://%s:8888/metrics", agent.IP),
	}

	var metrics map[string]*io_prometheus_client.MetricFamily
	var err error

	for _, url := range endpoints {
		metrics, err = fetchMetrics(url)
		if err == nil {
			break
		}
	}

	if err != nil {
		return fmt.Errorf("failed to fetch metrics from both hostname and IP: %w", err)
	}

	agg := AggregatedAgentMetrics{
		AgentID:           agent.AgentID,
		LogsRateSent:      extractMetricValue(metrics, "otelcol_exporter_sent_log_records"),
		TracesRateSent:    extractMetricValue(metrics, "otelcol_exporter_sent_spans"),
		MetricsRateSent:   extractMetricValue(metrics, "otelcol_exporter_sent_metric_points"),
		DataSentBytes:     extractMetricValue(metrics, "otelcol_exporter_sent_bytes"),
		DataReceivedBytes: extractMetricValue(metrics, "otelcol_receiver_accepted_bytes"),
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
		CPUUtilization:    extractMetricValue(metrics, "otelcol_process_cpu_seconds_total"),
		MemoryUtilization: extractMetricValue(metrics, "otelcol_process_memory_rss"),
		Timestamp:         time.Now().Unix(),
	}

	return q.QueueRepository.UpdateAgentMetricsInDB(agg, rt)
}
