package frontendpipeline_test

import (
	"database/sql"
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
func (m *MockRepo) SyncPipelineGraph(tx *sql.Tx, pipelineID int, graph models.PipelineGraph) error {
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

// --- Tests ---

func TestGetAllPipelines_Service(t *testing.T) {
	mockRepo := new(MockRepo)
	service := frontendpipeline.NewFrontendPipelineService(mockRepo)

	expected := []*frontendpipeline.Pipeline{{ID: 1, Name: "TestPipeline"}}
	mockRepo.On("GetAllPipelines").Return(expected, nil)

	pipelines, err := service.GetAllPipelines()
	assert.NoError(t, err)
	assert.Equal(t, expected, pipelines)
}

func TestGetPipelineInfo_Service_Exists(t *testing.T) {
	mockRepo := new(MockRepo)
	service := frontendpipeline.NewFrontendPipelineService(mockRepo)

	mockRepo.On("PipelineExists", 1).Return(true)
	expected := &frontendpipeline.PipelineInfo{ID: 1, Name: "TestPipeline"}
	mockRepo.On("GetPipelineInfo", 1).Return(expected, nil)

	info, err := service.GetPipelineInfo(1)
	assert.NoError(t, err)
	assert.Equal(t, expected, info)
}

func TestGetPipelineInfo_Service_NotExists(t *testing.T) {
	mockRepo := new(MockRepo)
	service := frontendpipeline.NewFrontendPipelineService(mockRepo)

	mockRepo.On("PipelineExists", 404).Return(false)

	info, err := service.GetPipelineInfo(404)
	assert.Error(t, err)
	assert.Nil(t, info)
	assert.Equal(t, utils.ErrPipelineDoesNotExists, err)
}
