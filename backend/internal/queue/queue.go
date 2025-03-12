package queue

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
)

// NewQueue creates a new AgentQueue with the specified number of workers.
func NewQueue(workerCount int, db *sql.DB) *AgentQueue {
	queueRepository := NewQueueRepository(db)
	q := &AgentQueue{
		agents:          make(map[string]*AgentStatus),
		checkQueue:      make(chan string, workerCount*2),
		workerCount:     workerCount,
		QueueRepository: queueRepository,
	}
	q.startWorkers()
	return q
}

// AddAgent adds a new agent to the queue.
func (q *AgentQueue) AddAgent(id, Hostname string) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	//TODO: Check if agent already exists
	q.agents[id] = &AgentStatus{
		AgentID:        id,
		Hostname:       Hostname,
		CurrentStatus:  "UNKNOWN",
		RetryRemaining: 3,
	}
	utils.Logger.Info(fmt.Sprintf("Successfully queued agent with ID: %s.", id))
}

// RemoveAgent removes an agent from the queue by ID.
func (q *AgentQueue) RemoveAgent(id string) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	delete(q.agents, id)
	utils.Logger.Info(fmt.Sprintf("Successfully removed agent with ID: %s.", id))
}

// StartStatusCheck starts a goroutine that checks the status of all agents at regular intervals.
func (q *AgentQueue) StartStatusCheck() {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			q.checkAllAgents()
		}
	}()
}

//FIXME: See if this shutsdown if receieves closing signal

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
			q.checkAgentStatus(agent)
		}
	}
}

func (q *AgentQueue) checkAgentStatus(agentStatus *AgentStatus) {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(agentStatus.Hostname + ":443" + "/agent/v1/status")

	q.mutex.Lock()
	defer q.mutex.Unlock()

	if err != nil {
		agentStatus.RetryRemaining--
		if agentStatus.RetryRemaining <= 0 {
			agentStatus.CurrentStatus = "DOWN"
			q.QueueRepository.UpdateStatusOnly(agentStatus.AgentID, agentStatus.CurrentStatus)
			q.RemoveAgent(agentStatus.AgentID)
		} else {
			agentStatus.CurrentStatus = "UNKNOWN"
			q.QueueRepository.UpdateStatusRetries(agentStatus.AgentID, agentStatus.RetryRemaining, agentStatus.CurrentStatus)
		}
	} else {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			agentStatus.CurrentStatus = "UP"
			agentStatus.RetryRemaining = 3

			var agentMetrics models.AgentMetrics

			decoder := json.NewDecoder(resp.Body)
			if err := decoder.Decode(&agentMetrics); err != nil {
				utils.Logger.Error(fmt.Sprintf("error decoding status response: %v", err))
				return
			}
			agentMetrics.AgentID = agentStatus.AgentID
			q.QueueRepository.UpdateMetrics(&agentMetrics)
		} else {
			agentStatus.CurrentStatus = "UNKNOWN"
			agentStatus.RetryRemaining--
			q.QueueRepository.UpdateStatusRetries(agentStatus.AgentID, agentStatus.RetryRemaining, agentStatus.CurrentStatus)
		}
	}
}
