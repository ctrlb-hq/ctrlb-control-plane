package agent

import (
	"errors"
	"testing"
	"time"

	frontendpipeline "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontend/pipeline"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/stretchr/testify/assert"
)

// Mock implementations

type MockAgentRepository struct {
	RegisterFunc func(req *models.AgentRegisterRequest) (*AgentRegisterResponse, error)
	ExistsFunc   func(hostname string) (bool, error)
}

func (m *MockAgentRepository) RegisterAgent(req *models.AgentRegisterRequest) (*AgentRegisterResponse, error) {
	return m.RegisterFunc(req)
}

func (m *MockAgentRepository) AgentExists(hostname string) (bool, error) {
	return m.ExistsFunc(hostname)
}

type MockAgentQueue struct {
	AddFunc            func(id, hostname, ip string) error
	RemoveFunc         func(id string) error
	RefreshFunc        func() error
	CheckAllAgentsFunc func()
}

func (m *MockAgentQueue) AddAgent(id, hostname, ip string) error {
	return m.AddFunc(id, hostname, ip)
}

func (m *MockAgentQueue) RemoveAgent(id string) error {
	return m.RemoveFunc(id)
}

func (m *MockAgentQueue) RefreshMonitoring() error {
	return m.RefreshFunc()
}

func (m *MockAgentQueue) StartStatusCheck() {
	// No-op for testing
}

func (m *MockAgentQueue) CheckAllAgents() {
	// No-op for testing
}

type MockFrontendPipeline struct {
	SyncFunc func(agentId string) error
}

func (m *MockFrontendPipeline) GetAllPipelines() ([]*frontendpipeline.Pipeline, error) {
	return nil, nil
}
func (m *MockFrontendPipeline) GetPipelineInfo(pipelineId int) (*frontendpipeline.PipelineInfo, error) {
	return nil, nil
}
func (m *MockFrontendPipeline) GetPipelineOverview(pipelineId int) (*frontendpipeline.PipelineInfoWithAgent, error) {
	return nil, nil
}
func (m *MockFrontendPipeline) CreatePipeline(createPipelineRequest models.CreatePipelineRequest) (string, error) {
	return "", nil
}
func (m *MockFrontendPipeline) DeletePipeline(pipelineId int) error {
	return nil
}
func (m *MockFrontendPipeline) GetAllAgentsAttachedToPipeline(pipelineId int) ([]models.AgentInfoHome, error) {
	return nil, nil
}
func (m *MockFrontendPipeline) DetachAgentFromPipeline(pipelineId int, agentId int) error {
	return nil
}
func (m *MockFrontendPipeline) AttachAgentToPipeline(pipelineId int, agentId int) error {
	return nil
}
func (m *MockFrontendPipeline) GetPipelineGraph(pipelineId int) (*models.PipelineGraph, error) {
	return nil, nil
}
func (m *MockFrontendPipeline) SyncPipelineGraph(pipelineId int, pipelineGraph models.PipelineGraph) error {
	return nil
}
func (m *MockFrontendPipeline) SyncConfig(agentId string) error {
	return m.SyncFunc(agentId)
}

func TestAgentService_RegisterAgent_Success(t *testing.T) {
	mockRepo := &MockAgentRepository{
		RegisterFunc: func(req *models.AgentRegisterRequest) (*AgentRegisterResponse, error) {
			return &AgentRegisterResponse{ID: 1, Config: map[string]any{"dummy": "value"}}, nil
		},
	}
	mockQueue := &MockAgentQueue{
		AddFunc: func(agentID string, hostname string, ip string) error {
			return nil
		},
		RemoveFunc:  func(id string) error { return nil },
		RefreshFunc: func() error { return nil },
	}
	mockFrontend := &MockFrontendPipeline{}

	svc := NewAgentService(mockRepo, mockQueue, mockFrontend)

	req := &models.AgentRegisterRequest{
		Platform: "linux",
		Hostname: "test-host",
		Version:  "v1.0",
		IP:       "127.0.0.1",
	}

	resp, err := svc.RegisterAgent(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int64(1), resp.ID)
	assert.NotEmpty(t, req.Name)
	assert.WithinDuration(t, time.Now(), time.Unix(req.RegisteredAt, 0), time.Second*2)
}

func TestAgentService_RegisterAgent_RepoError(t *testing.T) {
	mockRepo := &MockAgentRepository{
		RegisterFunc: func(req *models.AgentRegisterRequest) (*AgentRegisterResponse, error) {
			return nil, errors.New("db error")
		},
	}
	mockQueue := &MockAgentQueue{
		AddFunc: func(agentID string, hostname string, ip string) error {
			return nil
		},
		RemoveFunc:  func(id string) error { return nil },
		RefreshFunc: func() error { return nil },
	}
	mockFrontend := &MockFrontendPipeline{}

	svc := NewAgentService(mockRepo, mockQueue, mockFrontend)

	req := &models.AgentRegisterRequest{
		Platform: "linux",
		Hostname: "test-host",
		Version:  "v1.0",
		IP:       "127.0.0.1",
	}

	resp, err := svc.RegisterAgent(req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestAgentService_ConfigChangedPing_Success(t *testing.T) {
	mockRepo := &MockAgentRepository{}
	mockQueue := &MockAgentQueue{
		AddFunc:     func(agentID string, hostname string, ip string) error { return nil },
		RemoveFunc:  func(id string) error { return nil },
		RefreshFunc: func() error { return nil },
	}
	mockFrontend := &MockFrontendPipeline{
		SyncFunc: func(agentId string) error {
			return nil
		},
	}

	svc := NewAgentService(mockRepo, mockQueue, mockFrontend)

	err := svc.ConfigChangedPing("agent-id-123")
	assert.NoError(t, err)
}

func TestAgentService_ConfigChangedPing_Failure(t *testing.T) {
	mockRepo := &MockAgentRepository{}
	mockQueue := &MockAgentQueue{
		AddFunc:     func(agentID string, hostname string, ip string) error { return nil },
		RemoveFunc:  func(id string) error { return nil },
		RefreshFunc: func() error { return nil },
	}
	mockFrontend := &MockFrontendPipeline{
		SyncFunc: func(agentId string) error {
			return errors.New("sync failed")
		},
	}

	svc := NewAgentService(mockRepo, mockQueue, mockFrontend)

	err := svc.ConfigChangedPing("agent-id-123")
	assert.Error(t, err)
}
