package api_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	handlers "github.com/ctrlb-hq/ctrlb-collector/agent/internal/api/handlers"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/core/operators"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockOperator struct {
	mock.Mock
}

func (m *MockOperator) StartAgent() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockOperator) StopAgent() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockOperator) GracefulShutdown() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockOperator) UpdateCurrentConfig(cfg map[string]any) error {
	args := m.Called(cfg)
	return args.Error(0)
}

func TestStartAgent_Success(t *testing.T) {
	mockOp := new(MockOperator)
	mockOp.On("StartAgent").Return(nil)
	h := handlers.NewOperatorHandler(&operators.OperatorService{Operator: mockOp})

	r := httptest.NewRequest(http.MethodPost, "/agent/v1/start", nil)
	w := httptest.NewRecorder()

	h.StartAgent(w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockOp.AssertExpectations(t)
}

func TestStartAgent_Failure(t *testing.T) {
	mockOp := new(MockOperator)
	mockOp.On("StartAgent").Return(errors.New("start failed"))
	h := handlers.NewOperatorHandler(&operators.OperatorService{Operator: mockOp})

	r := httptest.NewRequest(http.MethodPost, "/agent/v1/start", nil)
	w := httptest.NewRecorder()

	h.StartAgent(w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	mockOp.AssertExpectations(t)
}

func TestUpdateCurrentConfig_Success(t *testing.T) {
	mockOp := new(MockOperator)
	mockOp.On("UpdateCurrentConfig", mock.Anything).Return(nil)
	h := handlers.NewOperatorHandler(&operators.OperatorService{Operator: mockOp})

	body := bytes.NewBufferString(`{"log_level": "debug"}`)
	r := httptest.NewRequest(http.MethodPost, "/agent/v1/config", body)
	w := httptest.NewRecorder()

	h.UpdateCurrentConfig(w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockOp.AssertExpectations(t)
}

func TestUpdateCurrentConfig_InvalidJSON(t *testing.T) {
	mockOp := new(MockOperator)
	h := handlers.NewOperatorHandler(&operators.OperatorService{Operator: mockOp})

	body := bytes.NewBufferString(`not-a-json`)
	r := httptest.NewRequest(http.MethodPost, "/agent/v1/config", body)
	w := httptest.NewRecorder()

	h.UpdateCurrentConfig(w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestUpdateCurrentConfig_Failure(t *testing.T) {
	mockOp := new(MockOperator)
	mockOp.On("UpdateCurrentConfig", mock.Anything).Return(errors.New("config update failed"))
	h := handlers.NewOperatorHandler(&operators.OperatorService{Operator: mockOp})

	body := bytes.NewBufferString(`{"log_level": "debug"}`)
	r := httptest.NewRequest(http.MethodPost, "/agent/v1/config", body)
	w := httptest.NewRecorder()

	h.UpdateCurrentConfig(w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	mockOp.AssertExpectations(t)
}

func TestStopAgent_Success(t *testing.T) {
	mockOp := new(MockOperator)
	mockOp.On("StopAgent").Return(nil)
	h := handlers.NewOperatorHandler(&operators.OperatorService{Operator: mockOp})

	r := httptest.NewRequest(http.MethodPost, "/agent/v1/stop", nil)
	w := httptest.NewRecorder()

	h.StopAgent(w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockOp.AssertExpectations(t)
}

func TestStopAgent_Failure(t *testing.T) {
	mockOp := new(MockOperator)
	mockOp.On("StopAgent").Return(errors.New("stop failed"))
	h := handlers.NewOperatorHandler(&operators.OperatorService{Operator: mockOp})

	r := httptest.NewRequest(http.MethodPost, "/agent/v1/stop", nil)
	w := httptest.NewRecorder()

	h.StopAgent(w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	mockOp.AssertExpectations(t)
}
