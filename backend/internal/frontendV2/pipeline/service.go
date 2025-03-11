package frontendpipeline

import "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/queue"

type FrontendPipelineService struct {
	FrontendPipelineRepository *FrontendPipelineRepository
	AgentQueue                 *queue.AgentQueue
}

// NewFrontendPipelineService creates a new FrontendPipelineService
func NewFrontendPipelineService(frontendPipelineRepository *FrontendPipelineRepository, agentQueue *queue.AgentQueue) *FrontendPipelineService {
	return &FrontendPipelineService{
		FrontendPipelineRepository: frontendPipelineRepository,
		AgentQueue:                 agentQueue,
	}
}
