package agent

import (
	"fmt"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/pkg/queue"
)

// AgentService manages agent operations.
type AgentService struct {
	AgentRepository *AgentRepository  // Repository for agent data
	AgentQueue      *queue.AgentQueue // Queue for agent tasks
}

// NewAgentService creates a new AgentService instance.
func NewAgentService(agentRepository *AgentRepository, agentQueue *queue.AgentQueue) *AgentService {
	return &AgentService{
		AgentRepository: agentRepository,
		AgentQueue:      agentQueue,
	}
}

// RegisterAgent processes the registration of a new agent.
func (a *AgentService) RegisterAgent(request *AgentRegisterRequest) (*AgentRegisterResponse, error) {

	// Register the agent in the repository
	response, err := a.AgentRepository.RegisterAgent(request)
	if err != nil {
		return nil, err
	}

	a.AgentQueue.AddAgent(fmt.Sprint(response.ID), request.Hostname)

	return response, nil // Return the registered agent
}
