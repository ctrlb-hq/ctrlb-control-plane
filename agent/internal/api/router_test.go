package api_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/api"
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

func TestNewRouter_CallsAreWired(t *testing.T) {
	mockOperator := new(MockOperator)
	mockOperator.On("StartAgent").Return(nil)
	mockOperator.On("StopAgent").Return(nil)
	mockOperator.On("GracefulShutdown").Return(nil)
	mockOperator.On("UpdateCurrentConfig", mock.Anything).Return(nil)

	service := &operators.OperatorService{Operator: mockOperator}

	// ✅ THIS is the actual function you want to test for coverage
	router := api.NewRouter(service)

	routes := []struct {
		path   string
		method string
	}{
		{"/agent/v1/start", "POST"},
		{"/agent/v1/stop", "POST"},
		{"/agent/v1/shutdown", "POST"},
		{"/agent/v1/config", "POST"},
	}

	for _, route := range routes {
		var body *strings.Reader
		if route.path == "/agent/v1/config" {
			body = strings.NewReader(`{"log_level": "debug"}`)
		} else {
			body = strings.NewReader(`{}`)
		}

		req := httptest.NewRequest(route.method, route.path, body)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)
		assert.True(t, resp.Code < 500, "expected no 5xx errors for %s", route.path)
	}

	mockOperator.AssertExpectations(t)
}
