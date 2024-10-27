package frontendagent

import (
	"fmt"
	"net/http"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/queue"
)

type FrontendAgentService struct {
	FrontendAgentRepository *FrontendAgentRepository
	AgentQueue              *queue.AgentQueue
}

// NewFrontendAgentService creates a new FrontendAgentService
func NewFrontendAgentService(frontendAgentRepository *FrontendAgentRepository, agentQueue *queue.AgentQueue) *FrontendAgentService {
	return &FrontendAgentService{
		FrontendAgentRepository: frontendAgentRepository,
		AgentQueue:              agentQueue,
	}
}

// GetAllAgents retrieves all agents
func (f *FrontendAgentService) GetAllAgents() ([]Agent, error) {
	return f.FrontendAgentRepository.GetAllAgents()
}

// GetAgent retrieves an agent along with its configuration
func (f *FrontendAgentService) GetAgent(id string) (*models.AgentWithConfig, error) {
	agent, err := f.FrontendAgentRepository.GetAgent(id)
	if err != nil {
		return nil, err
	}

	config, err := f.FrontendAgentRepository.GetConfig(agent.ConfigID)
	if err != nil {
		return nil, err
	}

	return &models.AgentWithConfig{
		ID:           agent.ID,
		Name:         agent.Name,
		Type:         agent.Type,
		Version:      agent.Version,
		Hostname:     agent.Hostname,
		Platform:     agent.Platform,
		Config:       *config,
		IsPipeline:   agent.IsPipeline,
		RegisteredAt: agent.RegisteredAt,
	}, nil
}

// DeleteAgent removes an agent by ID and shuts it down
func (f *FrontendAgentService) DeleteAgent(id string) error {
	agent, err := f.FrontendAgentRepository.GetAgent(id)
	if err != nil {
		return err
	}

	// Attempt to delete the agent from the repository
	if err := f.FrontendAgentRepository.DeleteAgent(agent.ID); err != nil {
		return err
	}

	// Shut down the agent via HTTP request
	if err := f.shutdownAgent(agent.Hostname); err != nil {
		return err
	}

	// Remove agent from the queue
	f.AgentQueue.RemoveAgent(agent.ID)

	return nil
}

// shutdownAgent sends a shutdown request to the agent
func (f *FrontendAgentService) shutdownAgent(hostname string) error {
	url := fmt.Sprintf("http://%s:443/agent/v1/shutdown", hostname)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return fmt.Errorf("error encountered while removing agent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error encountered while shutting down agent: %s", resp.Status)
	}

	return nil
}

// StartAgent sends a start request to the agent
func (f *FrontendAgentService) StartAgent(id string) error {
	agent, err := f.FrontendAgentRepository.GetAgent(id)
	if err != nil {
		return err
	}

	return f.sendAgentCommand(agent.Hostname, "start")
}

// StopAgent sends a stop request to the agent
func (f *FrontendAgentService) StopAgent(id string) error {
	agent, err := f.FrontendAgentRepository.GetAgent(id)
	if err != nil {
		return err
	}

	return f.sendAgentCommand(agent.Hostname, "stop")
}

// sendAgentCommand sends a command (start/stop) to the agent
func (f *FrontendAgentService) sendAgentCommand(hostname, command string) error {
	url := fmt.Sprintf("http://%s:443/agent/v1/%s", hostname, command)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return fmt.Errorf("error encountered while %s agent: %w", command, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error encountered while %s agent: %s", command, resp.Status)
	}

	return nil
}

// GetMetrics retrieves the metrics for a specific agent
func (f *FrontendAgentService) GetMetrics(id string) (*models.AgentMetrics, error) {
	return f.FrontendAgentRepository.GetMetrics(id)
}

// RestartMonitoring restarts monitoring for the agent
func (f *FrontendAgentService) RestartMonitoring(id string) error {
	agent, err := f.FrontendAgentRepository.GetAgent(id)
	if err != nil {
		return err
	}

	f.AgentQueue.AddAgent(agent.ID, agent.Hostname)

	return nil
}
