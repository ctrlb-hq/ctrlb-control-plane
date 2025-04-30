package frontendagent_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	frontendagent "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontend/agent"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockFrontendAgentService struct {
	mock.Mock
}

func (m *MockFrontendAgentService) GetAllAgents() ([]models.AgentInfoHome, error) {
	args := m.Called()
	return args.Get(0).([]models.AgentInfoHome), args.Error(1)
}

func (m *MockFrontendAgentService) GetAllUnmanagedAgents() ([]frontendagent.UnmanagedAgents, error) {
	args := m.Called()
	return args.Get(0).([]frontendagent.UnmanagedAgents), args.Error(1)
}

func (m *MockFrontendAgentService) GetAgent(id string) (*frontendagent.AgentInfoWithLabels, error) {
	args := m.Called(id)
	return args.Get(0).(*frontendagent.AgentInfoWithLabels), args.Error(1)
}

func (m *MockFrontendAgentService) DeleteAgent(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockFrontendAgentService) StartAgent(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockFrontendAgentService) StopAgent(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockFrontendAgentService) RestartMonitoring(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockFrontendAgentService) GetHealthMetricsForGraph(id string) (*[]frontendagent.AgentMetrics, error) {
	args := m.Called(id)
	return args.Get(0).(*[]frontendagent.AgentMetrics), args.Error(1)
}

func (m *MockFrontendAgentService) GetRateMetricsForGraph(id string) (*[]frontendagent.AgentMetrics, error) {
	args := m.Called(id)
	return args.Get(0).(*[]frontendagent.AgentMetrics), args.Error(1)
}

func (m *MockFrontendAgentService) AddLabels(id string, labels map[string]string) error {
	args := m.Called(id, labels)
	return args.Error(0)
}

func (m *MockFrontendAgentService) GetLatestAgentSince(since string) (*frontendagent.LatestAgentResponse, error) {
	args := m.Called(since)
	return args.Get(0).(*frontendagent.LatestAgentResponse), args.Error(1)
}

func TestGetAllAgentsHandler(t *testing.T) {
	mockService := new(MockFrontendAgentService)
	handler := frontendagent.NewFrontendAgentHandler(mockService)

	mockService.On("GetAllAgents").Return([]models.AgentInfoHome{
		{ID: 1},
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/agents", nil)
	w := httptest.NewRecorder()
	handler.GetAllAgents(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestAddLabelsHandler(t *testing.T) {
	mockService := new(MockFrontendAgentService)
	handler := frontendagent.NewFrontendAgentHandler(mockService)

	body := map[string]string{"env": "prod"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/agent/labels", bytes.NewBuffer(jsonBody))
	req = muxSetVars(req, map[string]string{"id": "agent-1"})

	w := httptest.NewRecorder()
	mockService.On("AddLabels", "agent-1", body).Return(nil)

	handler.AddLabels(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// muxSetVars helps set route vars for testing
func muxSetVars(r *http.Request, vars map[string]string) *http.Request {
	return mux.SetURLVars(r, vars)
}
