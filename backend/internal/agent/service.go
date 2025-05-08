package agent

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	frontendpipeline "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontend/pipeline"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/pkg/queue"
)

type AgentRepositoryInterface interface {
	RegisterAgent(req *models.AgentRegisterRequest) (*AgentRegisterResponse, error)
	AgentExists(hostname string) (bool, error)
}

type AgentServiceInterface interface {
	RegisterAgent(req *models.AgentRegisterRequest) (*AgentRegisterResponse, error)
	ConfigChangedPing(agentID string) error
}

// AgentService manages agent operations.
type AgentService struct {
	AgentRepository      AgentRepositoryInterface
	AgentQueue           queue.AgentQueueInterface
	FrontendAgentService frontendpipeline.FrontendPipelineServiceInterface
}

// NewAgentService creates a new AgentService instance.
func NewAgentService(agentRepository AgentRepositoryInterface, agentQueue queue.AgentQueueInterface, frontendPipelineService frontendpipeline.FrontendPipelineServiceInterface) *AgentService {
	return &AgentService{
		AgentRepository:      agentRepository,
		AgentQueue:           agentQueue,
		FrontendAgentService: frontendPipelineService,
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

	response, err := a.AgentRepository.RegisterAgent(req)
	if err != nil {
		return nil, err
	}

	err = a.AgentQueue.AddAgent(fmt.Sprint(response.ID), req.Hostname, req.IP)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// ConfigChangedPing notifies frontend to sync config.
func (a *AgentService) ConfigChangedPing(agentID string) error {
	err := a.FrontendAgentService.SyncConfig(agentID)
	if err != nil {
		return err
	}
	return nil
}

func (a *AgentService) ConfigChangedPing(agentID string) error {
	err := a.FrontendAgentService.SyncConfig(agentID)
	if err != nil {
		return err
	}
	return nil
}
