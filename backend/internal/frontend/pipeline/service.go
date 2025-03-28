package frontendpipeline

import (
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
)

type FrontendPipelineService struct {
	FrontendPipelineRepository *FrontendPipelineRepository
}

// NewFrontendPipelineService creates a new FrontendPipelineService
func NewFrontendPipelineService(frontendPipelineRepository *FrontendPipelineRepository) *FrontendPipelineService {
	return &FrontendPipelineService{
		FrontendPipelineRepository: frontendPipelineRepository,
	}
}

func (f *FrontendPipelineService) GetAllPipelines() ([]*Pipeline, error) {
	return f.FrontendPipelineRepository.GetAllPipelines()
}

func (f *FrontendPipelineService) GetPipelineInfo(pipelineId int) (*PipelineInfo, error) {
	if !f.FrontendPipelineRepository.PipelineExists(pipelineId) {
		return nil, utils.ErrPipelineDoesNotExists
	}

	return f.FrontendPipelineRepository.GetPipelineInfo(pipelineId)
}

func (f *FrontendPipelineService) CreatePipeline(createPipelineRequest CreatePipelineRequest) (string, error) {
	return f.FrontendPipelineRepository.CreatePipeline(createPipelineRequest)
}

func (f *FrontendPipelineService) DeletePipeline(pipelineId int) error {
	if !f.FrontendPipelineRepository.PipelineExists(pipelineId) {
		return utils.ErrPipelineDoesNotExists
	}

	return f.FrontendPipelineRepository.DeletePipeline(pipelineId)
}

func (f *FrontendPipelineService) GetAllAgentsAttachedToPipeline(pipelineId int) ([]models.AgentInfoHome, error) {
	if !f.FrontendPipelineRepository.PipelineExists(pipelineId) {
		return nil, utils.ErrPipelineDoesNotExists
	}

	return f.FrontendPipelineRepository.GetAllAgentsAttachedToPipeline(pipelineId)
}

func (f *FrontendPipelineService) DetachAgentFromPipeline(pipelineId int, agentId int) error {
	if !f.FrontendPipelineRepository.PipelineExists(pipelineId) {
		return utils.ErrPipelineDoesNotExists
	}

	return f.FrontendPipelineRepository.DetachAgentFromPipeline(pipelineId, agentId)
}

func (f *FrontendPipelineService) AttachAgentToPipeline(pipelineId int, agentId int) error {
	if !f.FrontendPipelineRepository.PipelineExists(pipelineId) {
		return utils.ErrPipelineDoesNotExists
	}

	return f.FrontendPipelineRepository.AttachAgentToPipeline(pipelineId, agentId)
}

func (f *FrontendPipelineService) GetPipelineGraph(pipelineId int) (*PipelineGraph, error) {
	if !f.FrontendPipelineRepository.PipelineExists(pipelineId) {
		return nil, utils.ErrPipelineDoesNotExists
	}

	return f.FrontendPipelineRepository.GetPipelineGraph(pipelineId)
}

func (f *FrontendPipelineService) SyncPipelineGraph(pipelineId int, pipelineGraph *PipelineGraph) error {
	if !f.FrontendPipelineRepository.PipelineExists(pipelineId) {
		return utils.ErrPipelineDoesNotExists
	}

	return f.FrontendPipelineRepository.SyncPipelineGraph(pipelineId, pipelineGraph.Nodes, pipelineGraph.Edges)
}
