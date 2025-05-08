package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

// MockAgentService for handler tests
type MockAgentService struct {
	RegisterAgentFunc     func(req *models.AgentRegisterRequest) (*AgentRegisterResponse, error)
	ConfigChangedPingFunc func(agentID string) error
}

func (m *MockAgentService) RegisterAgent(req *models.AgentRegisterRequest) (*AgentRegisterResponse, error) {
	return m.RegisterAgentFunc(req)
}

func (m *MockAgentService) ConfigChangedPing(agentID string) error {
	return m.ConfigChangedPingFunc(agentID)
}

func TestAgentHandler_RegisterAgent_Success(t *testing.T) {
	mockService := &MockAgentService{
		RegisterAgentFunc: func(req *models.AgentRegisterRequest) (*AgentRegisterResponse, error) {
			return &AgentRegisterResponse{ID: 1, Config: map[string]any{"dummy": "value"}}, nil
		},
	}

	handler := NewAgentHandler(mockService)

	body := &models.AgentRegisterRequest{
		Platform: "linux",
		Hostname: "test-host",
		Version:  "v1.0",
		IP:       "127.0.0.1",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/agents/register", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.RegisterAgent(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp AgentRegisterResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), resp.ID)
}

func TestAgentHandler_RegisterAgent_InvalidJSON(t *testing.T) {
	mockService := &MockAgentService{}

	handler := NewAgentHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/agents/register", bytes.NewBuffer([]byte("{invalid_json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.RegisterAgent(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAgentHandler_RegisterAgent_ServiceError(t *testing.T) {
	mockService := &MockAgentService{
		RegisterAgentFunc: func(req *models.AgentRegisterRequest) (*AgentRegisterResponse, error) {
			return nil, errors.New("registration failed")
		},
	}

	handler := NewAgentHandler(mockService)

	body := &models.AgentRegisterRequest{
		Platform: "linux",
		Hostname: "test-host",
		Version:  "v1.0",
		IP:       "127.0.0.1",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/agents/register", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.RegisterAgent(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestAgentHandler_ConfigChangedPing_Success(t *testing.T) {
	mockService := &MockAgentService{
		ConfigChangedPingFunc: func(agentID string) error {
			return nil
		},
	}

	handler := NewAgentHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/agents/1/config-changed", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler.ConfigChangedPing(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestAgentHandler_ConfigChangedPing_Failure(t *testing.T) {
	mockService := &MockAgentService{
		ConfigChangedPingFunc: func(agentID string) error {
			return errors.New("sync failed")
		},
	}

	handler := NewAgentHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/agents/1/config-changed", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler.ConfigChangedPing(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
