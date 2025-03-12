package agent

import (
	"errors"
	"time"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/constants"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/queue"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
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
func (a *AgentService) RegisterAgent(request AgentRegisterRequest) (interface{}, error) {
	var config *models.Config

	// Set default config ID based on agent type
	switch request.Type {
	case "fluent-bit":
		config, _ = a.AgentRepository.GetConfig(constants.DEFAULT_CONFIG_FB_ID)
	case "otel":
		config, _ = a.AgentRepository.GetConfig(constants.DEFAULT_CONFIG_OTEL_ID)
	default:
		return nil, errors.New("agent type not supported")
	}

	// Create a new agent instance
	agent := models.AgentWithConfig{
		ID:           utils.CreateNewUUID(),
		Name:         utils.GenerateAgentName(request.Type, request.Version, request.Hostname),
		Type:         request.Type,
		Version:      request.Version,
		Hostname:     request.Hostname,
		Platform:     request.Platform,
		Config:       *config,
		IsPipeline:   request.IsPipeline,
		RegisteredAt: time.Now(),
	}

	// Register the agent in the repository
	err := a.AgentRepository.RegisterAgent(&agent)
	if err != nil {
		return nil, err
	}

	a.AgentQueue.AddAgent(agent.ID, agent.Hostname)

	return agent, nil // Return the registered agent
}
