package frontendpipeline

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/pkg/configcompiler"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
)

type FrontendPipelineRepositoryInterface interface {
	PipelineExists(pipelineId int) bool
	GetAllPipelines() ([]*Pipeline, error)
	GetPipelineInfo(pipelineId int) (*PipelineInfo, error)
	GetPipelineOverview(pipelineId int) (*PipelineInfoWithAgent, error)
	CreatePipeline(createPipelineRequest models.CreatePipelineRequest) (string, error)
	DeletePipeline(pipelineId int) error
	GetAllAgentsAttachedToPipeline(pipelineId int) ([]models.AgentInfoHome, error)
	DetachAgentFromPipeline(pipelineId int, agentId int) error
	AttachAgentToPipeline(pipelineId int, agentId int) error
	GetPipelineGraph(pipelineId int) (*models.PipelineGraph, error)
	SyncPipelineGraph(tx *sql.Tx, pipelineID int, components []models.PipelineNodes, edges []models.PipelineEdges) error
	GetAgentInfo(agentId int) (*models.AgentInfoHome, error)
	GetAgentPipelineId(agentId string) (*int, error)
}

type FrontendPipelineServiceInterface interface {
	GetAllPipelines() ([]*Pipeline, error)
	GetPipelineInfo(pipelineId int) (*PipelineInfo, error)
	GetPipelineOverview(pipelineId int) (*PipelineInfoWithAgent, error)
	CreatePipeline(createPipelineRequest models.CreatePipelineRequest) (string, error)
	DeletePipeline(pipelineId int) error
	GetAllAgentsAttachedToPipeline(pipelineId int) ([]models.AgentInfoHome, error)
	DetachAgentFromPipeline(pipelineId int, agentId int) error
	AttachAgentToPipeline(pipelineId int, agentId int) error
	GetPipelineGraph(pipelineId int) (*models.PipelineGraph, error)
	SyncPipelineGraph(pipelineId int, pipelineGraph models.PipelineGraph) error
	SyncConfig(agentId string) error
}

type FrontendPipelineService struct {
	FrontendPipelineRepository FrontendPipelineRepositoryInterface
}

// NewFrontendPipelineService creates a new FrontendPipelineService
func NewFrontendPipelineService(frontendPipelineRepository FrontendPipelineRepositoryInterface) FrontendPipelineServiceInterface {
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

func (f *FrontendPipelineService) GetPipelineOverview(pipelineId int) (*PipelineInfoWithAgent, error) {
	if !f.FrontendPipelineRepository.PipelineExists(pipelineId) {
		return nil, utils.ErrPipelineDoesNotExists
	}

	return f.FrontendPipelineRepository.GetPipelineOverview(pipelineId)
}

func (f *FrontendPipelineService) CreatePipeline(createPipelineRequest models.CreatePipelineRequest) (string, error) {
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

	err := f.FrontendPipelineRepository.AttachAgentToPipeline(pipelineId, agentId)
	if err != nil {
		return err
	}

	graph, err := f.GetPipelineGraph(pipelineId)
	if err != nil {
		return err
	}

	agent, err := f.FrontendPipelineRepository.GetAgentInfo(agentId)
	if err != nil {
		return err
	}

	var agents []models.AgentInfoHome
	agents = append(agents, *agent)

	return f.sendConfigToAgents(agents, *graph)
}

func (f *FrontendPipelineService) GetPipelineGraph(pipelineId int) (*models.PipelineGraph, error) {
	if !f.FrontendPipelineRepository.PipelineExists(pipelineId) {
		return nil, utils.ErrPipelineDoesNotExists
	}

	return f.FrontendPipelineRepository.GetPipelineGraph(pipelineId)
}

func (f *FrontendPipelineService) SyncPipelineGraph(pipelineId int, pipelineGraph models.PipelineGraph) error {
	if !f.FrontendPipelineRepository.PipelineExists(pipelineId) {
		return utils.ErrPipelineDoesNotExists
	}

	err := f.FrontendPipelineRepository.SyncPipelineGraph(nil, pipelineId, pipelineGraph.Nodes, pipelineGraph.Edges)
	if err != nil {
		return err
	}

	attachedAgent, err := f.FrontendPipelineRepository.GetAllAgentsAttachedToPipeline(pipelineId)
	if err != nil {
		return err
	}

	return f.sendConfigToAgents(attachedAgent, pipelineGraph)
}

func (f *FrontendPipelineService) sendConfigToAgents(agents []models.AgentInfoHome, pipelineGraph models.PipelineGraph) error {
	config, err := configcompiler.CompileGraphToJSON(pipelineGraph)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshaling config: %v", err)
	}

	for _, agent := range agents {
		url := fmt.Sprintf("http://%s:443/agent/v1/config", agent.Hostname)
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			utils.Logger.Sugar().Errorf("Retry failed using Hostname for agent [ID:%v]: %v", agent.ID, err)
			url = fmt.Sprintf("http://%s:443/agent/v1/config", agent.IP)
			resp, err = http.Post(url, "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				utils.Logger.Sugar().Errorf("Retry failed using IP for agent [ID:%v]: %v", agent.ID, err)
				continue
			}
		}
		if resp != nil {
			defer resp.Body.Close()
		}

	}

	return nil
}

func (f *FrontendPipelineService) SyncConfig(agentId string) error {
	pipelineId, err := f.FrontendPipelineRepository.GetAgentPipelineId(agentId)
	if err != nil {
		return err
	}

	graph, err := f.GetPipelineGraph(*pipelineId)
	if err != nil {
		return err
	}

	agentIDInt, err := strconv.Atoi(agentId)
	if err != nil {
		return fmt.Errorf("error converting agent ID to int: %v", err)
	}

	agent, err := f.FrontendPipelineRepository.GetAgentInfo(agentIDInt)
	if err != nil {
		return err
	}

	var agents []models.AgentInfoHome
	agents = append(agents, *agent)

	return f.sendConfigToAgents(agents, *graph)
}
