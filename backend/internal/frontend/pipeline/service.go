package frontendpipeline

import "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/pkg/queue"

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
	if err := f.FrontendPipelineRepository.VerifyPipelineExists(pipelineId); err != nil {
		return nil, err
	}

	return f.FrontendPipelineRepository.GetAllAgentsAttachedToPipeline(pipelineId)
}

func (f *FrontendPipelineService) DetachAgentFromPipeline(pipelineId int, agentId int) error {
	if err := f.FrontendPipelineRepository.VerifyPipelineExists(pipelineId); err != nil {
		return err
	}

	return f.FrontendPipelineRepository.DetachAgentFromPipeline(pipelineId, agentId)
}

func (f *FrontendPipelineService) AttachAgentToPipeline(pipelineId int, agentId int) error {
	if err := f.FrontendPipelineRepository.VerifyPipelineExists(pipelineId); err != nil {
		return err
	}
	return f.FrontendPipelineRepository.AttachAgentToPipeline(pipelineId, agentId)
}

func (f *FrontendPipelineService) GetPipelineGraph(pipelineId int) (*PipelineGraph, error) {
	if err := f.FrontendPipelineRepository.VerifyPipelineExists(pipelineId); err != nil {
		return nil, err
	}
	return f.FrontendPipelineRepository.GetPipelineGraph(pipelineId)
}

func (f *FrontendPipelineService) SyncPipelineGraph(pipelineId int, pipelineGraph *PipelineGraph) error {
	if err := f.FrontendPipelineRepository.VerifyPipelineExists(pipelineId); err != nil {
		return err
	}
	return f.FrontendPipelineRepository.SyncPipelineGraph(pipelineId, pipelineGraph.Nodes, pipelineGraph.Edges)
}
