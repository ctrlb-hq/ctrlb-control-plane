package agent

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
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
func (a *AgentService) RegisterAgent(req *models.AgentRegisterRequest) (*AgentRegisterResponse, error) {
	if req.Type == "" {
		req.Type = "OTEL"
	}
	req.RegisteredAt = time.Now().Unix()

	hash := sha256.Sum256(fmt.Appendf(nil, "%s-%s-%s", req.Platform, req.Hostname, req.Version))
	req.Name = fmt.Sprintf("%s-agent-%s", req.Platform, hex.EncodeToString(hash[:6]))

	// Register the agent in the repository
	response, err := a.AgentRepository.RegisterAgent(req)
	if err != nil {
		return nil, err
	}

	a.AgentQueue.AddAgent(fmt.Sprint(response.ID), req.Hostname, req.IP)

	return response, nil // Return the registered agent
}
