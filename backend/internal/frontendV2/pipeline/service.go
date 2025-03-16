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

func (f *FrontendPipelineService) GetAllPipelines() ([]*Pipeline, error) {
	return f.FrontendPipelineRepository.GetAllPipelines()
}

func (f *FrontendPipelineService) GetPipelineInfo(pipelineId int) (*PipelineInfo, error) {
	return f.FrontendPipelineRepository.GetPipelineInfo(pipelineId)
}

func (f *FrontendPipelineService) DeletePipeline(pipelineId int) error {
	return f.FrontendPipelineRepository.DeletePipeline(pipelineId)
}

func (f *FrontendPipelineService) GetAllAgentsAttachedToPipeline(pipelineId int) ([]AgentInfoHome, error) {
	return f.FrontendPipelineRepository.GetAllAgentsAttachedToPipeline(pipelineId)
}
