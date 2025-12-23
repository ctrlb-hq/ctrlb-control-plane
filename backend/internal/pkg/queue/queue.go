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
	AddAgent(id, hostname, ip, agentType string) error
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
	stopChan        chan struct{}
	wg              sync.WaitGroup
}

// NewQueue creates a new AgentQueue
func NewQueue(workerCount int, intervalSec int, queueRepository AgentQueueRepositoryInterface) AgentQueueInterface {
	// Buffer size: larger to handle bursts of agents becoming ready
	bufferSize := max(workerCount*10, 100)

	q := &AgentQueue{
		agents:          make(map[string]*AgentStatus),
		checkQueue:      make(chan string, bufferSize),
		workerCount:     workerCount,
		IntervalSecond:  intervalSec,
		QueueRepository: queueRepository,
		Metrics:         DefaultMetricsHelper{},
		stopChan:        make(chan struct{}),
	}
	q.startWorkers()
	q.startRetryScheduler()
	return q
}

// AddAgent adds a new agent to the queue
func (q *AgentQueue) AddAgent(id, hostname, ip, agentType string) error {
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
		Type:           agentType,
		CurrentStatus:  "unknown",
		RetryRemaining: 3,
		NextCheck:      time.Now(), // eligible immediately
		InFlight:       false,
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
		if err := q.AddAgent(agent.AgentID, agent.Hostname, agent.IP, agent.Type); err != nil {
			utils.Logger.Sugar().Errorf("Error adding agent [ID: %v] to queue", agent.AgentID)
		}
	}

	return nil
}

// Internal worker that handles agent check logic
func (q *AgentQueue) startWorkers() {
	for i := 0; i < q.workerCount; i++ {
		q.wg.Add(1)
		workerID := i + 1 // 1-indexed for better readability in logs
		go q.worker(workerID)
	}
}

// Retry scheduler that re-enqueues agents based on NextCheck
func (q *AgentQueue) startRetryScheduler() {
	q.wg.Add(1)
	go func() {
		defer q.wg.Done()
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-q.stopChan:
				return
			case <-ticker.C:
				now := time.Now()

				// Collect agents to check without holding lock for too long
				var agentsToCheck []string
				q.mutex.Lock()
				for id, agent := range q.agents {
					// Only enqueue if not already in-flight and check time has arrived
					if !agent.InFlight && (agent.NextCheck.Before(now) || agent.NextCheck.Equal(now)) {
						agentsToCheck = append(agentsToCheck, id)
					}
				}
				q.mutex.Unlock()

				// Now enqueue agents without holding the main lock
				for _, id := range agentsToCheck {
					q.mutex.Lock()
					agent, exists := q.agents[id]
					if exists && !agent.InFlight {
						select {
						case q.checkQueue <- id:
							// Mark as in-flight and schedule next check
							agent.InFlight = true
							agent.NextCheck = now.Add(time.Duration(q.IntervalSecond) * time.Second)
						default:
							utils.Logger.Sugar().Warnf("checkQueue full, skipping agent %s", id)
						}
					}
					q.mutex.Unlock()
				}
			}
		}
	}()
}

// Worker loop
func (q *AgentQueue) worker(workerID int) {
	defer q.wg.Done()
	utils.Logger.Sugar().Infof("Worker #%d started", workerID)

	for {
		select {
		case <-q.stopChan:
			utils.Logger.Sugar().Infof("Worker #%d shutting down", workerID)
			return
		case agentID, ok := <-q.checkQueue:
			if !ok {
				utils.Logger.Sugar().Infof("Worker #%d: checkQueue closed", workerID)
				return
			}

			// Process agent check with panic recovery
			func() {
				startTime := time.Now()
				utils.Logger.Sugar().Infof("Worker #%d picked up agent [ID:%s]", workerID, agentID)

				defer func() {
					if r := recover(); r != nil {
						utils.Logger.Sugar().Errorf("Worker #%d panic for agent [ID:%s]: %v", workerID, agentID, r)
						// Clear InFlight flag on panic
						q.mutex.Lock()
						if agent, exists := q.agents[agentID]; exists {
							agent.InFlight = false
						}
						q.mutex.Unlock()
					}
				}()

				q.mutex.RLock()
				agent, exists := q.agents[agentID]
				if !exists {
					q.mutex.RUnlock()
					utils.Logger.Sugar().Warnf("Worker #%d: agent [ID:%s] not found in queue", workerID, agentID)
					return
				}
				// Copy what we need for the check
				agentCopy := *agent
				q.mutex.RUnlock()

				err := q.checkAgentStatus(&agentCopy)

				// Prepare status update info before acquiring lock
				var newStatus string
				var shouldRemove bool
				var newRetryCount int

				q.mutex.Lock()
				// Re-check existence after acquiring lock
				agent, exists = q.agents[agentID]
				if !exists {
					q.mutex.Unlock()
					utils.Logger.Sugar().Warnf("Worker #%d: agent [ID:%s] was removed during check", workerID, agentID)
					return
				}

				// Clear the in-flight flag
				agent.InFlight = false

				if err != nil {
					agent.RetryRemaining--
					newRetryCount = agent.RetryRemaining
					if agent.RetryRemaining <= 0 {
						newStatus = "disconnected"
						agent.CurrentStatus = newStatus
						shouldRemove = true
						delete(q.agents, agentID) // Remove inline instead of calling RemoveAgent
					} else {
						newStatus = "unknown"
						agent.CurrentStatus = newStatus
					}
				} else {
					agent.RetryRemaining = 3
					newRetryCount = 3
					newStatus = "connected"
					agent.CurrentStatus = newStatus
				}
				q.mutex.Unlock()

				// Perform DB operations outside the lock
				if updateErr := q.QueueRepository.UpdateAgentStatus(agentID, newStatus); updateErr != nil {
					utils.Logger.Sugar().Errorf("Worker #%d: Failed to update agent status for [ID:%s]: %v", workerID, agentID, updateErr)
				}

				duration := time.Since(startTime)

				// Log results
				if err != nil {
					if shouldRemove {
						utils.Logger.Sugar().Infof("Worker #%d: Removed agent [ID:%s] after exhausting retries (took %v)", workerID, agentID, duration)
					}
					utils.Logger.Sugar().Errorf("Worker #%d: Error checking agent [ID:%s], attempts remaining: %v, error: %v (took %v)", workerID, agentID, newRetryCount, err, duration)
				} else {
					utils.Logger.Sugar().Infof("Worker #%d: Successfully checked agent [ID:%s], status: %s (took %v)", workerID, agentID, newStatus, duration)
				}
			}()
		}
	}
}

// checkAgentStatus fetches Prometheus metrics from agent using the appropriate strategy
func (q *AgentQueue) checkAgentStatus(agent *AgentStatus) error {
	// Get the appropriate metrics extraction strategy for this agent type
	strategy := GetMetricsStrategy(agent.Type)

	// Build endpoints using the strategy's port
	port := strategy.GetPort()
	endpoints := []string{
		fmt.Sprintf("http://%s:%d/metrics", agent.Hostname, port),
		fmt.Sprintf("http://%s:%d/metrics", agent.IP, port),
	}

	// Fetch metrics from the agent (try hostname first, then IP)
	var (
		metrics map[string]*io_prometheus_client.MetricFamily
		err     error
	)
	for _, url := range endpoints {
		metrics, err = q.Metrics.Fetch(url)
		if err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("fetch metrics failed (hostname & IP): %w", err)
	}

	// Extract metrics using the strategy
	metricsData := strategy.ExtractMetrics(metrics, q.Metrics)

	// Build the aggregated (DB) metrics
	agg := AggregatedAgentMetrics{
		AgentID:           agent.AgentID,
		LogsRateSent:      metricsData.LogsRateSent,
		TracesRateSent:    metricsData.TracesRateSent,
		MetricsRateSent:   metricsData.MetricsRateSent,
		DataSentBytes:     metricsData.DataSentBytes,
		DataReceivedBytes: metricsData.DataReceivedBytes,
		Status:            "connected",
		UpdatedAt:         time.Now().Unix(),
	}

	// Build the realtime metrics
	rt := RealtimeAgentMetrics{
		AgentID:           agent.AgentID,
		LogsRateSent:      metricsData.LogsRateSent,
		TracesRateSent:    metricsData.TracesRateSent,
		MetricsRateSent:   metricsData.MetricsRateSent,
		DataSentBytes:     metricsData.DataSentBytes,
		DataReceivedBytes: metricsData.DataReceivedBytes,
		CPUUtilization:    metricsData.CPUUtilization,
		MemoryUtilization: metricsData.MemoryUtilization,
		Timestamp:         time.Now().Unix(),
	}

	return q.QueueRepository.UpdateAgentMetricsInDB(agg, rt)
}

// GetAgent returns agent by ID — test helper
// Returns a copy to prevent external modifications
func (q *AgentQueue) GetAgent(id string) (*AgentStatus, bool) {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	agent, exists := q.agents[id]
	if !exists {
		return nil, false
	}
	// Return a copy to prevent external modifications
	agentCopy := *agent
	return &agentCopy, true
}

// Shutdown gracefully stops all workers and the scheduler
func (q *AgentQueue) Shutdown(timeout time.Duration) error {
	utils.Logger.Info("Shutting down AgentQueue...")

	// Signal all goroutines to stop
	close(q.stopChan)

	// Wait for graceful shutdown with timeout
	done := make(chan struct{})
	go func() {
		q.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		utils.Logger.Info("AgentQueue shutdown completed successfully")
		return nil
	case <-time.After(timeout):
		utils.Logger.Error("AgentQueue shutdown timeout exceeded")
		return fmt.Errorf("shutdown timeout exceeded")
	}
}
