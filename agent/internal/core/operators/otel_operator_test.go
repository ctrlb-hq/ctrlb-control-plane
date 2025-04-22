package operators_test

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/core/operators"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAdapter struct {
	mock.Mock
}

func (m *MockAdapter) Initialize() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAdapter) StartAgent() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAdapter) StopAgent() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAdapter) UpdateConfig() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAdapter) GracefulShutdown() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAdapter) GetVersion() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func TestOtelOperator_Initialize(t *testing.T) {
	mockAdapter := new(MockAdapter)
	mockAdapter.On("Initialize").Return(nil)

	op := operators.NewOtelOperator(mockAdapter)
	result, err := op.Initialize()

	assert.NoError(t, err)
	assert.Equal(t, "Otel Agent initializing", result["message"])

	// Wait briefly to allow the goroutine to run
	time.Sleep(10 * time.Millisecond)
	mockAdapter.AssertCalled(t, "Initialize")
}

func TestStartAgent_Success(t *testing.T) {
	mockAdapter := new(MockAdapter)
	mockAdapter.On("StartAgent").Return(nil)

	op := operators.NewOtelOperator(mockAdapter)
	err := op.StartAgent()

	assert.NoError(t, err)
	mockAdapter.AssertExpectations(t)
}

func TestStopAgent_Success(t *testing.T) {
	mockAdapter := new(MockAdapter)
	mockAdapter.On("StopAgent").Return(nil)

	op := operators.NewOtelOperator(mockAdapter)
	err := op.StopAgent()

	assert.NoError(t, err)
	mockAdapter.AssertExpectations(t)
}

func TestGracefulShutdown_Success(t *testing.T) {
	mockAdapter := new(MockAdapter)
	mockAdapter.On("GracefulShutdown").Return(nil)

	op := operators.NewOtelOperator(mockAdapter)
	err := op.GracefulShutdown()

	assert.NoError(t, err)

	time.Sleep(50 * time.Millisecond) // give goroutine a chance to execute

	mockAdapter.AssertExpectations(t)
}

func TestUpdateCurrentConfig_Success(t *testing.T) {
	mockAdapter := new(MockAdapter)
	op := operators.NewOtelOperator(mockAdapter)

	cfg := map[string]any{
		"log_level": "debug",
	}
	err := op.UpdateCurrentConfig(cfg)
	assert.NoError(t, err)

	_ = os.Remove("config.yaml")
}

func TestStartAgent_Failure(t *testing.T) {
	mockAdapter := new(MockAdapter)
	mockAdapter.On("StartAgent").Return(errors.New("fail to start"))

	op := operators.NewOtelOperator(mockAdapter)
	err := op.StartAgent()

	assert.Error(t, err)
	assert.Equal(t, "fail to start", err.Error())
	mockAdapter.AssertExpectations(t)
}
