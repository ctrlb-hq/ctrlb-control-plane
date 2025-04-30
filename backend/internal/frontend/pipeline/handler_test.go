package frontendpipeline_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	frontendpipeline "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontend/pipeline"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) GetAllPipelines() ([]*frontendpipeline.Pipeline, error) {
	args := m.Called()
	return args.Get(0).([]*frontendpipeline.Pipeline), args.Error(1)
}
func (m *MockService) GetPipelineInfo(id int) (*frontendpipeline.PipelineInfo, error) {
	args := m.Called(id)
	return args.Get(0).(*frontendpipeline.PipelineInfo), args.Error(1)
}
func (m *MockService) GetPipelineOverview(id int) (*frontendpipeline.PipelineInfoWithAgent, error) {
	args := m.Called(id)
	return args.Get(0).(*frontendpipeline.PipelineInfoWithAgent), args.Error(1)
}
func (m *MockService) CreatePipeline(req models.CreatePipelineRequest) (string, error) {
	args := m.Called(req)
	return args.String(0), args.Error(1)
}
func (m *MockService) DeletePipeline(id int) error {
	args := m.Called(id)
	return args.Error(0)
}
func (m *MockService) GetAllAgentsAttachedToPipeline(id int) ([]models.AgentInfoHome, error) {
	args := m.Called(id)
	return args.Get(0).([]models.AgentInfoHome), args.Error(1)
}
func (m *MockService) DetachAgentFromPipeline(pipelineId, agentId int) error {
	args := m.Called(pipelineId, agentId)
	return args.Error(0)
}
func (m *MockService) AttachAgentToPipeline(pipelineId, agentId int) error {
	args := m.Called(pipelineId, agentId)
	return args.Error(0)
}
func (m *MockService) GetPipelineGraph(id int) (*models.PipelineGraph, error) {
	args := m.Called(id)
	return args.Get(0).(*models.PipelineGraph), args.Error(1)
}
func (m *MockService) SyncPipelineGraph(id int, graph models.PipelineGraph) error {
	args := m.Called(id, graph)
	return args.Error(0)
}
func (m *MockService) SyncConfig(agentId string) error {
	args := m.Called(agentId)
	return args.Error(0)
}

func TestGetAllPipelinesHandler(t *testing.T) {
	mockSvc := new(MockService)
	handler := frontendpipeline.NewFrontendPipelineHandler(mockSvc)

	mockSvc.On("GetAllPipelines").Return([]*frontendpipeline.Pipeline{
		{ID: 1, Name: "TestPipeline"},
	}, nil)

	req := httptest.NewRequest("GET", "/pipelines", nil)
	w := httptest.NewRecorder()

	handler.GetAllPipelines(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestCreatePipelineHandler_Valid(t *testing.T) {
	mockSvc := new(MockService)
	handler := frontendpipeline.NewFrontendPipelineHandler(mockSvc)

	reqBody := models.CreatePipelineRequest{
		Name:      "PipelineX",
		CreatedBy: "admin",
		AgentIDs:  []int{1},
		PipelineGraph: models.PipelineGraph{
			Nodes: []models.PipelineNodes{{
				ComponentID:   1,
				Name:          "receiver_otlp",
				ComponentRole: "receiver",
				ComponentName: "otlp",
				Config:        map[string]any{"endpoint": "localhost:4317"},
				SupportedSignals: []string{"traces"},
			}, {
				ComponentID:   2,
				Name:          "exporter_debug",
				ComponentRole: "exporter",
				ComponentName: "debug",
				Config:        map[string]any{},
				SupportedSignals: []string{"traces"},
			}},
			Edges: []models.PipelineEdges{{
				Source: "1",
				Target: "2",
			}},
		},
	}

	mockSvc.On("CreatePipeline", reqBody).Return("123", nil)

	jsonData, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/pipelines", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreatePipeline(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}


func TestCreatePipelineHandler_InvalidJSON(t *testing.T) {
	mockSvc := new(MockService)
	handler := frontendpipeline.NewFrontendPipelineHandler(mockSvc)

	req := httptest.NewRequest("POST", "/pipelines", bytes.NewReader([]byte("{invalid json")))
	w := httptest.NewRecorder()

	handler.CreatePipeline(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetPipelineInfoHandler(t *testing.T) {
	mockSvc := new(MockService)
	handler := frontendpipeline.NewFrontendPipelineHandler(mockSvc)

	expected := &frontendpipeline.PipelineInfo{ID: 1, Name: "TestPipeline"}
	mockSvc.On("GetPipelineInfo", 1).Return(expected, nil)

	req := httptest.NewRequest("GET", "/pipelines/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	handler.GetPipelineInfo(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetPipelineInfoHandler_InvalidID(t *testing.T) {
	mockSvc := new(MockService)
	handler := frontendpipeline.NewFrontendPipelineHandler(mockSvc)

	req := httptest.NewRequest("GET", "/pipelines/abc", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})
	w := httptest.NewRecorder()

	handler.GetPipelineInfo(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeletePipelineHandler(t *testing.T) {
	mockSvc := new(MockService)
	handler := frontendpipeline.NewFrontendPipelineHandler(mockSvc)

	mockSvc.On("DeletePipeline", 1).Return(nil)

	req := httptest.NewRequest("DELETE", "/pipelines/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	handler.DeletePipeline(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetAllAgentsAttachedToPipelineHandler(t *testing.T) {
	mockSvc := new(MockService)
	handler := frontendpipeline.NewFrontendPipelineHandler(mockSvc)

	mockSvc.On("GetAllAgentsAttachedToPipeline", 1).Return([]models.AgentInfoHome{}, nil)

	req := httptest.NewRequest("GET", "/pipelines/1/agents", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	handler.GetAllAgentsAttachedToPipeline(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDetachAgentFromPipelineHandler(t *testing.T) {
	mockSvc := new(MockService)
	handler := frontendpipeline.NewFrontendPipelineHandler(mockSvc)

	mockSvc.On("DetachAgentFromPipeline", 1, 2).Return(nil)

	req := httptest.NewRequest("DELETE", "/pipelines/1/agents/2", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1", "agent_id": "2"})
	w := httptest.NewRecorder()

	handler.DetachAgentFromPipeline(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAttachAgentToPipelineHandler(t *testing.T) {
	mockSvc := new(MockService)
	handler := frontendpipeline.NewFrontendPipelineHandler(mockSvc)

	mockSvc.On("AttachAgentToPipeline", 1, 2).Return(nil)

	req := httptest.NewRequest("POST", "/pipelines/1/agents/2", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1", "agent_id": "2"})
	w := httptest.NewRecorder()

	handler.AttachAgentToPipeline(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetPipelineGraphHandler(t *testing.T) {
	mockSvc := new(MockService)
	handler := frontendpipeline.NewFrontendPipelineHandler(mockSvc)

	mockSvc.On("GetPipelineGraph", 1).Return(&models.PipelineGraph{}, nil)

	req := httptest.NewRequest("GET", "/pipelines/1/graph", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	handler.GetPipelineGraph(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSyncPipelineGraphHandler(t *testing.T) {
	mockSvc := new(MockService)
	handler := frontendpipeline.NewFrontendPipelineHandler(mockSvc)

	graph := models.PipelineGraph{}
	mockSvc.On("SyncPipelineGraph", 1, graph).Return(nil)

	body, _ := json.Marshal(graph)
	req := httptest.NewRequest("POST", "/pipelines/1/graph", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	handler.SyncPipelineGraph(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
