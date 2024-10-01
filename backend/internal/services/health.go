package services

import (
	"net/http"
	"sync"
	"time"

	"github.com/ctrlb-hq/all-father/internal/models"
)

type AgentQueue struct {
	agents      map[string]*models.AgentStatus
	mutex       sync.RWMutex
	checkQueue  chan string
	workerCount int
}

func NewQueue(workerCount int) *AgentQueue {
	q := &AgentQueue{
		agents:      make(map[string]*models.AgentStatus),
		checkQueue:  make(chan string, workerCount*2),
		workerCount: workerCount,
	}
	q.startWorkers()
	return q
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

func (q *AgentQueue) AddAgent(ID, Hostname, CurrentStatus string) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.agents[ID] = &models.AgentStatus{
		AgentID:        ID,
		Hostname:       Hostname,
		CurrentStatus:  CurrentStatus,
		RetryRemaining: 3,
	}
}

func (q *AgentQueue) RemoveAgent(id string) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	delete(q.agents, id)
}

func (q *AgentQueue) checkAgentStatus(agent *models.AgentStatus) {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(agent.Hostname + "/api/v1/status")

	q.mutex.Lock()
	defer q.mutex.Unlock()

	if err != nil {
		agent.RetryRemaining--
		if agent.RetryRemaining <= 0 {
			agent.CurrentStatus = "Down"
			//TODO: Updated DB and remove agent from Queue
		} else {
			//TODO: Update DB
			agent.CurrentStatus = "Retrying"
		}
	} else {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			if agent.CurrentStatus != "Online" {
				//TODO: Update DB
			}
			agent.CurrentStatus = "Online"
			agent.RetryRemaining = 3
		} else {
			//TODO: Update DB
			agent.CurrentStatus = "Error"
			agent.RetryRemaining--
		}
	}
}

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

func (q *AgentQueue) StartStatusCheck() {
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for range ticker.C {
			q.checkAllAgents()
		}
	}()
}
