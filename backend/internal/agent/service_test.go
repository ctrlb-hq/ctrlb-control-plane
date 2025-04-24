package agent

import (
	"errors"
	"testing"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/stretchr/testify/assert"
)

// --- Mock AgentRepository ---
type mockAgentRepo struct {
	registerCalled bool
	returnError    bool
}

func (m *mockAgentRepo) RegisterAgent(req *models.AgentRegisterRequest) (*AgentRegisterResponse, error) {
	m.registerCalled = true
	if m.returnError {
		return nil, errors.New("db insert failed")
	}
	return &AgentRegisterResponse{ID: 42, Config: map[string]any{}}, nil
}

// --- Mock AgentQueue ---
type mockAgentQueue struct {
	addedAgents []string
}

func (mq *mockAgentQueue) AddAgent(id, hostname, ip string) error {
	mq.addedAgents = append(mq.addedAgents, id)
	return nil
}

// --- Mock FrontendPipelineService ---
type mockFrontendPipelineService struct {
	calledWith string
	err        error
}

func (m *mockFrontendPipelineService) SyncConfig(agentID string) error {
	m.calledWith = agentID
	return m.err
}

// --- Tests ---
func TestRegisterAgentSuccess(t *testing.T) {
	repo := &mockAgentRepo{}
	queue := &mockAgentQueue{}
	service := NewAgentService(repo, queue)

	req := &models.AgentRegisterRequest{
		Version:  "v1.2.3",
		Hostname: "test-host",
		Platform: "linux",
		IP:       "127.0.0.1",
	}

	resp, err := service.RegisterAgent(req)
	assert.NoError(t, err)
	assert.Equal(t, int64(42), resp.ID)
	assert.True(t, repo.registerCalled)
	assert.Equal(t, 1, len(queue.addedAgents))
}

func TestRegisterAgentFailure(t *testing.T) {
	repo := &mockAgentRepo{returnError: true}
	queue := &mockAgentQueue{}
	service := NewAgentService(repo, queue)

	req := &models.AgentRegisterRequest{
		Version:  "v1.0",
		Hostname: "fail-host",
		Platform: "linux",
		IP:       "192.168.0.1",
	}

	resp, err := service.RegisterAgent(req)
	assert.Nil(t, resp)
	assert.Error(t, err)
}

func TestConfigChangedPing(t *testing.T) {
	mockSync := &mockFrontendPipelineService{}
	svc := &AgentService{FrontendAgentService: mockSync}

	err := svc.ConfigChangedPing("agent-123")
	assert.NoError(t, err)
	assert.Equal(t, "agent-123", mockSync.calledWith)
}

func TestConfigChangedPingError(t *testing.T) {
	mockSync := &mockFrontendPipelineService{err: errors.New("sync failed")}
	svc := &AgentService{FrontendAgentService: mockSync}

	err := svc.ConfigChangedPing("agent-fail")
	assert.EqualError(t, err, "sync failed")
}
