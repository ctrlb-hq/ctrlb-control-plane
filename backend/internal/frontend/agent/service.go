package frontendagent

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/pkg/queue"
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

func (f *FrontendAgentService) GetAllAgents() ([]AgentInfoHome, error) {
	return f.FrontendAgentRepository.GetAllAgents()
}

func (f *FrontendAgentService) GetAllUnmanagedAgents() ([]UnmanagedAgents, error) {
	return f.FrontendAgentRepository.GetAllUnmanagedAgents()
}

// GetAgent retrieves an agent along with its configuration
func (f *FrontendAgentService) GetAgent(id string) (*AgentInfoWithLabels, error) {
	agent, err := f.FrontendAgentRepository.GetAgent(id)
	if err != nil {
		return nil, err
	}
	return agent, nil
}

// DeleteAgent removes an agent by ID and shuts it down
func (f *FrontendAgentService) DeleteAgent(id string) error {
	hostname, err := f.FrontendAgentRepository.GetAgentHostname(id)
	if err != nil {
		return err
	}

	f.AgentQueue.RemoveAgent(id)

	if err := f.sendAgentCommand(hostname, "shutdown"); err != nil {
		f.AgentQueue.AddAgent(id, hostname)
		return fmt.Errorf("unable to shut down the agent — it is still running and currently under active monitoring")
	}

	if err := f.FrontendAgentRepository.DeleteAgent(id); err != nil {
		return err
	}

	return nil
}

// StartAgent sends a start request to the agent
func (f *FrontendAgentService) StartAgent(id string) error {
	hostname, err := f.FrontendAgentRepository.GetAgentHostname(id)
	if err != nil {
		return err
	}

	if f.sendAgentCommand(hostname, "start") != nil {
		return fmt.Errorf("error encountered while starting agent")
	}

	if err = f.AgentQueue.AddAgent(id, hostname); err != nil {
		return fmt.Errorf("error while starting agent monitoring")
	}

	return nil
}

// StopAgent sends a stop request to the agent
func (f *FrontendAgentService) StopAgent(id string) error {
	hostname, err := f.FrontendAgentRepository.GetAgentHostname(id)
	if err != nil {
		return err
	}

	if err = f.AgentQueue.RemoveAgent(id); err != nil {
		return err
	}

	if err := f.sendAgentCommand(hostname, "stop"); err != nil {
		f.AgentQueue.AddAgent(id, hostname)
		return fmt.Errorf("error encountered while stopping agent")
	}
	return nil

}

// RestartMonitoring restarts monitoring for the agent
func (f *FrontendAgentService) RestartMonitoring(id string) error {
	hostname, err := f.FrontendAgentRepository.GetAgentHostname(id)
	if err != nil {
		return err
	}

	if err = f.AgentQueue.AddAgent(id, hostname); err != nil {
		return err
	}

	return nil
}

func (f *FrontendAgentService) GetHealthMetricsForGraph(id string) (*[]AgentMetrics, error) {
	exists, err := f.FrontendAgentRepository.AgentExists(id)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("agent not found")
	}
	return f.FrontendAgentRepository.GetHealthMetricsForGraph(id)
}

func (f *FrontendAgentService) GetRateMetricsForGraph(id string) (*[]AgentMetrics, error) {
	exists, err := f.FrontendAgentRepository.AgentExists(id)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("agent not found")
	}
	return f.FrontendAgentRepository.GetRateMetricsForGraph(id)
}

func (f *FrontendAgentService) AddLabels(agentId string, labels map[string]string) error {
	exists, err := f.FrontendAgentRepository.AgentExists(agentId)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("agent not found")
	}
	return f.FrontendAgentRepository.AddLabels(agentId, labels)
}

func (f *FrontendAgentService) sendAgentCommand(hostname, command string) error {
	url := fmt.Sprintf("http://%s:443/agent/v1/%s", hostname, command)

	client := &http.Client{
		Timeout: 20 * time.Second, // Adjust timeout as needed
	}

	resp, err := client.Post(url, "application/json", nil)
	if err != nil {
		return fmt.Errorf("error encountered while %s agent: %w", command, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error encountered while %s agent: %s", command, resp.Status)
	}

	return nil
}
