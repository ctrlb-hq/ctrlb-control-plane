package frontendpipeline

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

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
	SyncPipelineGraph(tx *sql.Tx, pipelineID int, graph models.PipelineGraph) error
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

	err := f.FrontendPipelineRepository.SyncPipelineGraph(nil, pipelineId, pipelineGraph)
	if err != nil {
		return err
	}

	attachedAgent, err := f.FrontendPipelineRepository.GetAllAgentsAttachedToPipeline(pipelineId)
	if err != nil {
		return err
	}

	return f.sendConfigToAgents(attachedAgent, pipelineGraph)
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

func (f *FrontendPipelineService) sendConfigToAgents(agents []models.AgentInfoHome, pipelineGraph models.PipelineGraph) error {
	config, err := configcompiler.CompileGraphToJSON(pipelineGraph)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshaling config: %v", err)
	}

	var failedAgents []string
	var lastErr error

	for _, agent := range agents {
		if err := f.sendConfigToSingleAgent(agent, jsonData); err != nil {
			failedAgents = append(failedAgents, fmt.Sprintf("Agent[ID:%v]", agent.ID))
			lastErr = err
			utils.Logger.Sugar().Errorf("Failed to send config to agent [ID:%v]: %v", agent.ID, err)
		}
	}

	if len(failedAgents) == len(agents) {
		return fmt.Errorf("failed to send configuration to all %d agent(s): last error: %v", len(agents), lastErr)
	} else if len(failedAgents) > 0 {
		return fmt.Errorf("partial failure: configuration failed for %d out of %d agent(s) [%v]; last error: %v",
			len(failedAgents), len(agents), failedAgents, lastErr)
	}

	return nil
}

func (f *FrontendPipelineService) sendConfigToSingleAgent(agent models.AgentInfoHome, jsonData []byte) error {
	// create a client with 10s timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	trySend := func(endpoint string) error {
		url := fmt.Sprintf("http://%s:443/agent/v1/config", endpoint)
		resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		switch {
		case resp.StatusCode >= 200 && resp.StatusCode < 300:
			return nil
		case resp.StatusCode == 500:
			return utils.ErrInvalidConfig
		default:
			return fmt.Errorf("request failed with status %d", resp.StatusCode)
		}
	}

	// Try with hostname first
	err := trySend(agent.Hostname)
	if err == nil || errors.Is(err, utils.ErrInvalidConfig) {
		return err
	}

	// Retry with IP if hostname failed due to network error
	utils.Logger.Sugar().Warnf("Hostname failed for agent [ID:%v], retrying with IP: %v", agent.ID, err)
	return trySend(agent.IP)
}
