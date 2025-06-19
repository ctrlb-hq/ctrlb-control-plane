package operators_test

import (
	"testing"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/core/operators"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockAdapter is a no-op implementation for interface satisfaction
type mockAdapter struct{}

func (m *mockAdapter) Initialize() error                                 { return nil }
func (m *mockAdapter) StartAgent() error                                 { return nil }
func (m *mockAdapter) StopAgent() error                                  { return nil }
func (m *mockAdapter) UpdateConfig() error                               { return nil }
func (m *mockAdapter) GracefulShutdown() error                           { return nil }
func (m *mockAdapter) GetVersion() (string, error)                       { return "mock", nil }
func (m *mockAdapter) ValidateConfigInMemory(data *map[string]any) error { return nil }

func TestNewOperatorService_ReturnsOtelOperator(t *testing.T) {
	adapter := &mockAdapter{}
	service := operators.NewOperatorService(adapter)

	assert.NotNil(t, service)
	assert.NotNil(t, service.Operator)

	// Optional: check that the operator is of type OtelOperator
	_, ok := service.Operator.(*operators.OtelOperator)
	assert.True(t, ok, "expected Operator to be OtelOperator")
}

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

func TestOperatorService_StartAgent(t *testing.T) {
	mockOp := new(MockOperator)
	mockOp.On("StartAgent").Return(nil)

	service := &operators.OperatorService{Operator: mockOp}
	err := service.StartAgent()

	assert.NoError(t, err)
	mockOp.AssertExpectations(t)
}

func TestOperatorService_StopAgent(t *testing.T) {
	mockOp := new(MockOperator)
	mockOp.On("StopAgent").Return(nil)

	service := &operators.OperatorService{Operator: mockOp}
	err := service.StopAgent()

	assert.NoError(t, err)
	mockOp.AssertExpectations(t)
}

func TestOperatorService_GracefulShutdown(t *testing.T) {
	mockOp := new(MockOperator)
	mockOp.On("GracefulShutdown").Return(nil)

	service := &operators.OperatorService{Operator: mockOp}
	err := service.GracefulShutdown()

	assert.NoError(t, err)
	mockOp.AssertExpectations(t)
}

func TestOperatorService_UpdateCurrentConfig(t *testing.T) {
	mockOp := new(MockOperator)
	mockOp.On("UpdateCurrentConfig", mock.Anything).Return(nil)

	service := &operators.OperatorService{Operator: mockOp}
	err := service.UpdateCurrentConfig(map[string]any{"log_level": "debug"})

	assert.NoError(t, err)
	mockOp.AssertExpectations(t)
}
