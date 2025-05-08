package frontendagent_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	frontendagent "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontend/agent"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
)

type MockRepo struct {
	mock.Mock
}

type MockQueue struct {
	mock.Mock
}

func (m *MockRepo) GetAllAgents() ([]models.AgentInfoHome, error) {
	args := m.Called()
	return args.Get(0).([]models.AgentInfoHome), args.Error(1)
}
func (m *MockRepo) GetAllUnmanagedAgents() ([]frontendagent.UnmanagedAgents, error) {
	args := m.Called()
	return args.Get(0).([]frontendagent.UnmanagedAgents), args.Error(1)
}
func (m *MockRepo) GetAgent(id string) (*frontendagent.AgentInfoWithLabels, error) {
	args := m.Called(id)
	return args.Get(0).(*frontendagent.AgentInfoWithLabels), args.Error(1)
}
func (m *MockRepo) AgentExists(id string) bool {
	args := m.Called(id)
	return args.Bool(0)
}
func (m *MockRepo) AgentStatus(id string) string {
	args := m.Called(id)
	return args.String(0)
}
func (m *MockRepo) GetAgentNetworkInfoByID(id string) (string, string, error) {
	args := m.Called(id)
	return args.String(0), args.String(1), args.Error(2)
}
func (m *MockRepo) DeleteAgent(id string) error {
	args := m.Called(id)
	return args.Error(0)
}
func (m *MockRepo) GetHealthMetricsForGraph(id string) (*[]frontendagent.AgentMetrics, error) {
	args := m.Called(id)
	return args.Get(0).(*[]frontendagent.AgentMetrics), args.Error(1)
}
func (m *MockRepo) GetRateMetricsForGraph(id string) (*[]frontendagent.AgentMetrics, error) {
	args := m.Called(id)
	return args.Get(0).(*[]frontendagent.AgentMetrics), args.Error(1)
}
func (m *MockRepo) AddLabels(id string, labels map[string]string) error {
	args := m.Called(id, labels)
	return args.Error(0)
}
func (m *MockRepo) GetLatestAgentSince(since string) (*frontendagent.LatestAgentResponse, error) {
	args := m.Called(since)
	return args.Get(0).(*frontendagent.LatestAgentResponse), args.Error(1)
}

func (mq *MockQueue) AddAgent(id, hostname, ip string) error {
	args := mq.Called(id, hostname, ip)
	return args.Error(0)
}
func (mq *MockQueue) RemoveAgent(id string) error {
	args := mq.Called(id)
	return args.Error(0)
}

func (mq *MockQueue) RefreshMonitoring() error {
	args := mq.Called()
	return args.Error(0)
}

// --- Tests ---

func TestGetAllUnmanagedAgents(t *testing.T) {
	repo := new(MockRepo)
	q := new(MockQueue)
	svc := frontendagent.NewFrontendAgentService(repo, q)

	expected := []frontendagent.UnmanagedAgents{{ID: "1"}}
	repo.On("GetAllUnmanagedAgents").Return(expected, nil)

	result, err := svc.GetAllUnmanagedAgents()
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestGetAgent_Success(t *testing.T) {
	repo := new(MockRepo)
	q := new(MockQueue)
	svc := frontendagent.NewFrontendAgentService(repo, q)

	agent := &frontendagent.AgentInfoWithLabels{}
	repo.On("AgentExists", "1").Return(true)
	repo.On("GetAgent", "1").Return(agent, nil)

	result, err := svc.GetAgent("1")
	assert.NoError(t, err)
	assert.Equal(t, agent, result)
}

func TestGetAgent_NotFound(t *testing.T) {
	repo := new(MockRepo)
	q := new(MockQueue)
	svc := frontendagent.NewFrontendAgentService(repo, q)

	repo.On("AgentExists", "2").Return(false)

	result, err := svc.GetAgent("2")
	assert.Nil(t, result)
	assert.ErrorIs(t, err, utils.ErrAgentDoesNotExists)
}

func TestStopAgent_Success(t *testing.T) {
	repo := new(MockRepo)
	q := new(MockQueue)
	svc := frontendagent.NewFrontendAgentService(repo, q)

	repo.On("AgentExists", "agent-1").Return(true)
	repo.On("GetAgentNetworkInfoByID", "agent-1").Return("host", "ip", nil)
	q.On("RemoveAgent", "agent-1").Return(nil)
	q.On("AddAgent", "agent-1", "host", "ip").Return(nil)

	// Simulate unreachable HTTP (sendAgentCommand returns error)
	err := svc.StopAgent("agent-1")
	assert.Error(t, err) // fallback added back to queue should trigger
}

func TestRestartMonitoring(t *testing.T) {
	repo := new(MockRepo)
	q := new(MockQueue)
	svc := frontendagent.NewFrontendAgentService(repo, q)

	repo.On("AgentExists", "agent-1").Return(true)
	repo.On("GetAgentNetworkInfoByID", "agent-1").Return("host", "ip", nil)
	q.On("AddAgent", "agent-1", "host", "ip").Return(nil)

	err := svc.RestartMonitoring("agent-1")
	assert.NoError(t, err)
}

func TestGetHealthMetrics(t *testing.T) {
	repo := new(MockRepo)
	q := new(MockQueue)
	svc := frontendagent.NewFrontendAgentService(repo, q)

	mockMetrics := &[]frontendagent.AgentMetrics{{}}
	repo.On("AgentExists", "a1").Return(true)
	repo.On("GetHealthMetricsForGraph", "a1").Return(mockMetrics, nil)

	m, err := svc.GetHealthMetricsForGraph("a1")
	assert.NoError(t, err)
	assert.Equal(t, mockMetrics, m)
}

func TestGetRateMetrics(t *testing.T) {
	repo := new(MockRepo)
	q := new(MockQueue)
	svc := frontendagent.NewFrontendAgentService(repo, q)

	mockMetrics := &[]frontendagent.AgentMetrics{{}}
	repo.On("AgentExists", "a1").Return(true)
	repo.On("GetRateMetricsForGraph", "a1").Return(mockMetrics, nil)

	m, err := svc.GetRateMetricsForGraph("a1")
	assert.NoError(t, err)
	assert.Equal(t, mockMetrics, m)
}

func TestGetLatestAgentSince(t *testing.T) {
	repo := new(MockRepo)
	q := new(MockQueue)
	svc := frontendagent.NewFrontendAgentService(repo, q)

	mockResp := &frontendagent.LatestAgentResponse{ID: "latest"}
	repo.On("GetLatestAgentSince", "2024-01-01T00:00:00Z").Return(mockResp, nil)

	resp, err := svc.GetLatestAgentSince("2024-01-01T00:00:00Z")
	assert.NoError(t, err)
	assert.Equal(t, mockResp, resp)
}
