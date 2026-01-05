package agent

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/constants"
	frontendpipeline "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontend/pipeline"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/pkg/queue"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
)

type AgentRepositoryInterface interface {
	RegisterAgent(req *models.AgentRegisterRequest) (*AgentRegisterResponse, error)
	AgentExists(hostname string) (bool, error)
}

// MetricsRepositoryInterface defines the interface for storing agent metrics
type MetricsRepositoryInterface interface {
	UpdateAgentMetricsInDB(agg queue.AggregatedAgentMetrics, rt queue.RealtimeAgentMetrics) error
}

type AgentServiceInterface interface {
	RegisterAgent(req *models.AgentRegisterRequest) (*AgentRegisterResponse, error)
	ConfigChangedPing(agentID string) error
	ProcessHeartbeat(agentID string, req *HeartbeatRequest) error
}

// AgentService manages agent operations.
type AgentService struct {
	AgentRepository      AgentRepositoryInterface
	MetricsRepository    MetricsRepositoryInterface
	FrontendAgentService frontendpipeline.FrontendPipelineServiceInterface
}

// NewAgentService creates a new AgentService instance.
func NewAgentService(agentRepository AgentRepositoryInterface, metricsRepository MetricsRepositoryInterface, frontendPipelineService frontendpipeline.FrontendPipelineServiceInterface) *AgentService {
	return &AgentService{
		AgentRepository:      agentRepository,
		MetricsRepository:    metricsRepository,
		FrontendAgentService: frontendPipelineService,
	}
}

// RegisterAgent processes the registration of a new agent.
func (a *AgentService) RegisterAgent(req *models.AgentRegisterRequest) (*AgentRegisterResponse, error) {
	if string(req.Type) == "" {
		req.Type = models.AgentTypeOTEL
	}
	req.RegisteredAt = time.Now().Unix()

	hash := sha256.Sum256(fmt.Appendf(nil, "%s-%s-%s", req.Platform, req.Hostname, req.Version))
	req.Name = fmt.Sprintf("%s-agent-%s", req.Platform, hex.EncodeToString(hash[:6]))

	response, err := a.AgentRepository.RegisterAgent(req)
	if err != nil {
		return nil, err
	}

	if req.PipelineName != "" {
		var createDefaultPipelineReq models.CreatePipelineRequest
		createDefaultPipelineReq.Name = req.PipelineName
		createDefaultPipelineReq.AgentIDs = []int{int(response.ID)}
		createDefaultPipelineReq.CreatedBy = req.StartedBy
		createDefaultPipelineReq.Type = req.Type

		// Use appropriate default pipeline graph based on agent type
		if req.Type == models.AgentTypeFluentBit {
			createDefaultPipelineReq.PipelineGraph = constants.DefaultFluentBitPipelineGraph
		} else {
			createDefaultPipelineReq.PipelineGraph = constants.DefaultOTELPipelineGraph
		}

		_, err := a.FrontendAgentService.CreatePipeline(createDefaultPipelineReq)
		if err != nil {
			return nil, err
		}
	}

	utils.Logger.Info(fmt.Sprintf("Agent %d registered successfully", response.ID))

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

// ProcessHeartbeat updates agent metrics from a heartbeat request.
func (a *AgentService) ProcessHeartbeat(agentID string, req *HeartbeatRequest) error {

	utils.Logger.Info(fmt.Sprintf("Processing heartbeat for agent %s", agentID))
	utils.Logger.Info(fmt.Sprintf("Heartbeat request: %+v", req))
	agg := queue.AggregatedAgentMetrics{
		AgentID:           agentID,
		LogsRateSent:      req.LogsRateSent,
		TracesRateSent:    req.TracesRateSent,
		MetricsRateSent:   req.MetricsRateSent,
		DataSentBytes:     req.DataSentBytes,
		DataReceivedBytes: req.DataReceivedBytes,
		Status:            "connected",
		UpdatedAt:         time.Now().Unix(),
	}

	rt := queue.RealtimeAgentMetrics{
		AgentID:           agentID,
		LogsRateSent:      req.LogsRateSent,
		TracesRateSent:    req.TracesRateSent,
		MetricsRateSent:   req.MetricsRateSent,
		DataSentBytes:     req.DataSentBytes,
		DataReceivedBytes: req.DataReceivedBytes,
		CPUUtilization:    req.CPUUtilization,
		MemoryUtilization: req.MemoryUtilization,
		Timestamp:         time.Now().Unix(),
	}

	return a.MetricsRepository.UpdateAgentMetricsInDB(agg, rt)
}
