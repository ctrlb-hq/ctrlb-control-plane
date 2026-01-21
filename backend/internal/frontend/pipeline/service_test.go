package frontendpipeline_test

import (
	"database/sql"
	"errors"
	"testing"

	frontendpipeline "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontend/pipeline"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mocks ---

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) PipelineExists(id int) bool {
	args := m.Called(id)
	return args.Bool(0)
}
func (m *MockRepo) GetAllPipelines() ([]*frontendpipeline.Pipeline, error) {
	args := m.Called()
	return args.Get(0).([]*frontendpipeline.Pipeline), args.Error(1)
}
func (m *MockRepo) GetPipelineInfo(id int) (*frontendpipeline.PipelineInfo, error) {
	args := m.Called(id)
	return args.Get(0).(*frontendpipeline.PipelineInfo), args.Error(1)
}
func (m *MockRepo) GetPipelineOverview(id int) (*frontendpipeline.PipelineInfoWithAgent, error) {
	args := m.Called(id)
	return args.Get(0).(*frontendpipeline.PipelineInfoWithAgent), args.Error(1)
}
func (m *MockRepo) CreatePipeline(req models.CreatePipelineRequest) (string, error) {
	args := m.Called(req)
	return args.String(0), args.Error(1)
}
func (m *MockRepo) DeletePipeline(id int) error {
	args := m.Called(id)
	return args.Error(0)
}
func (m *MockRepo) GetAllAgentsAttachedToPipeline(id int) ([]models.AgentInfoHome, error) {
	args := m.Called(id)
	return args.Get(0).([]models.AgentInfoHome), args.Error(1)
}
func (m *MockRepo) DetachAgentFromPipeline(pipelineId int, agentId int) error {
	args := m.Called(pipelineId, agentId)
	return args.Error(0)
}
func (m *MockRepo) AttachAgentToPipeline(pipelineId int, agentId int) error {
	args := m.Called(pipelineId, agentId)
	return args.Error(0)
}
func (m *MockRepo) GetPipelineGraph(pipelineId int) (*models.PipelineGraph, error) {
	args := m.Called(pipelineId)
	return args.Get(0).(*models.PipelineGraph), args.Error(1)
}
func (m *MockRepo) SyncPipelineGraph(tx *sql.Tx, pipelineID int, graph models.PipelineGraph, agentType models.AgentType) error {
	args := m.Called(tx, pipelineID, graph)
	return args.Error(0)
}
func (m *MockRepo) GetAgentInfo(agentId int) (*models.AgentInfoHome, error) {
	args := m.Called(agentId)
	return args.Get(0).(*models.AgentInfoHome), args.Error(1)
}
func (m *MockRepo) GetAgentPipelineId(agentId string) (*int, error) {
	args := m.Called(agentId)
	return args.Get(0).(*int), args.Error(1)
}
func (m *MockRepo) DetachAllAgentsFromPipeline(pipelineId int) error {
	args := m.Called(pipelineId)
	return args.Error(0)
}

type MockAgentService struct {
	mock.Mock
}

func (m *MockAgentService) StopAgent(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// --- Tests ---

func TestGetAllPipelines_Service(t *testing.T) {
	mockRepo := new(MockRepo)
	mockAgentService := new(MockAgentService)
	service := frontendpipeline.NewFrontendPipelineService(mockRepo, mockAgentService)

	expected := []*frontendpipeline.Pipeline{{ID: 1, Name: "TestPipeline"}}
	mockRepo.On("GetAllPipelines").Return(expected, nil)

	pipelines, err := service.GetAllPipelines()
	assert.NoError(t, err)
	assert.Equal(t, expected, pipelines)
}

func TestGetPipelineInfo_Service_Exists(t *testing.T) {
	mockRepo := new(MockRepo)
	mockAgentService := new(MockAgentService)
	service := frontendpipeline.NewFrontendPipelineService(mockRepo, mockAgentService)

	mockRepo.On("PipelineExists", 1).Return(true)
	expected := &frontendpipeline.PipelineInfo{ID: 1, Name: "TestPipeline"}
	mockRepo.On("GetPipelineInfo", 1).Return(expected, nil)

	info, err := service.GetPipelineInfo(1)
	assert.NoError(t, err)
	assert.Equal(t, expected, info)
}

func TestGetPipelineInfo_Service_NotExists(t *testing.T) {
	mockRepo := new(MockRepo)
	mockAgentService := new(MockAgentService)
	service := frontendpipeline.NewFrontendPipelineService(mockRepo, mockAgentService)

	mockRepo.On("PipelineExists", 404).Return(false)

	info, err := service.GetPipelineInfo(404)
	assert.Error(t, err)
	assert.Nil(t, info)
	assert.Equal(t, utils.ErrPipelineDoesNotExists, err)
}

func TestDeletePipeline_Service_StopsAndDetachesAgents(t *testing.T) {
	mockRepo := new(MockRepo)
	mockAgentService := new(MockAgentService)
	service := frontendpipeline.NewFrontendPipelineService(mockRepo, mockAgentService)

	pipelineID := 10
	agents := []models.AgentInfoHome{
		{ID: 1, Name: "agent-1"},
		{ID: 2, Name: "agent-2"},
	}

	mockRepo.On("PipelineExists", pipelineID).Return(true)
	mockRepo.On("GetAllAgentsAttachedToPipeline", pipelineID).Return(agents, nil)
	mockAgentService.On("StopAgent", "1").Return(nil)
	mockAgentService.On("StopAgent", "2").Return(nil)
	mockRepo.On("DetachAllAgentsFromPipeline", pipelineID).Return(nil)
	mockRepo.On("DeletePipeline", pipelineID).Return(nil)

	err := service.DeletePipeline(pipelineID)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockAgentService.AssertExpectations(t)
}

func TestDeletePipeline_Service_StopFailure(t *testing.T) {
	mockRepo := new(MockRepo)
	mockAgentService := new(MockAgentService)
	service := frontendpipeline.NewFrontendPipelineService(mockRepo, mockAgentService)

	pipelineID := 11
	agents := []models.AgentInfoHome{
		{ID: 3, Name: "agent-3"},
	}

	mockRepo.On("PipelineExists", pipelineID).Return(true)
	mockRepo.On("GetAllAgentsAttachedToPipeline", pipelineID).Return(agents, nil)
	mockAgentService.On("StopAgent", "3").Return(errors.New("stop failed"))
	mockRepo.On("DetachAllAgentsFromPipeline", pipelineID).Return(nil)

	err := service.DeletePipeline(pipelineID)
	assert.Error(t, err)
	mockRepo.AssertNotCalled(t, "DeletePipeline", pipelineID)
	mockRepo.AssertExpectations(t)
	mockAgentService.AssertExpectations(t)
}

func TestDeletePipeline_Service_DetachFailure(t *testing.T) {
	mockRepo := new(MockRepo)
	mockAgentService := new(MockAgentService)
	service := frontendpipeline.NewFrontendPipelineService(mockRepo, mockAgentService)

	pipelineID := 12
	agents := []models.AgentInfoHome{
		{ID: 4, Name: "agent-4"},
	}

	mockRepo.On("PipelineExists", pipelineID).Return(true)
	mockRepo.On("GetAllAgentsAttachedToPipeline", pipelineID).Return(agents, nil)
	mockAgentService.On("StopAgent", "4").Return(nil)
	mockRepo.On("DetachAllAgentsFromPipeline", pipelineID).Return(errors.New("detach failed"))

	err := service.DeletePipeline(pipelineID)
	assert.Error(t, err)
	mockRepo.AssertNotCalled(t, "DeletePipeline", pipelineID)
	mockRepo.AssertExpectations(t)
	mockAgentService.AssertExpectations(t)
}
