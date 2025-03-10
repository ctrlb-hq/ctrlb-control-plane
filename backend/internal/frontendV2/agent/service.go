package frontendagent

import (
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/queue"
)

type FrontendAgentService struct {
	FrontendAgentRepository *FrontendAgentRepository
	AgentQueue              *queue.AgentQueue
}

// NewFrontendAgentService creates a new FrontendAgentService
func NewFrontendAgentService(frontendAgentRepository *FrontendAgentRepository, agentQueue *queue.AgentQueue) *FrontendAgentService {
	return &FrontendAgentService{
		FrontendAgentRepository: frontendAgentRepository,
		AgentQueue:              agentQueue,
	}
}
