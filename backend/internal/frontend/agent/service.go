package frontendagent

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/pkg/queue"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
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

func (f *FrontendAgentService) GetAllAgents() ([]models.AgentInfoHome, error) {
	return f.FrontendAgentRepository.GetAllAgents()
}

func (f *FrontendAgentService) GetAllUnmanagedAgents() ([]UnmanagedAgents, error) {
	return f.FrontendAgentRepository.GetAllUnmanagedAgents()
}

// GetAgent retrieves an agent along with its configuration
func (f *FrontendAgentService) GetAgent(id string) (*AgentInfoWithLabels, error) {
	if !f.FrontendAgentRepository.AgentExists(id) {
		return nil, utils.ErrAgentDoesNotExists
	}

	agent, err := f.FrontendAgentRepository.GetAgent(id)
	if err != nil {
		return nil, err
	}

	return agent, nil
}

// DeleteAgent removes an agent by ID and shuts it down
func (f *FrontendAgentService) DeleteAgent(id string) error {
	if !f.FrontendAgentRepository.AgentExists(id) {
		return utils.ErrAgentDoesNotExists
	}

	hostname, ip, err := f.FrontendAgentRepository.GetAgentNetworkInfoByID(id)
	if err != nil {
		return err
	}

	f.AgentQueue.RemoveAgent(id)

	if err := f.sendAgentCommand(hostname, ip, "shutdown"); err != nil {
		f.AgentQueue.AddAgent(id, hostname, ip)
		return fmt.Errorf("unable to shut down the agent — it is still running and currently under active monitoring")
	}

	if err := f.FrontendAgentRepository.DeleteAgent(id); err != nil {
		return err
	}

	return nil
}

// StartAgent sends a start request to the agent
func (f *FrontendAgentService) StartAgent(id string) error {
	if !f.FrontendAgentRepository.AgentExists(id) {
		return utils.ErrAgentDoesNotExists
	}

	hostname, ip, err := f.FrontendAgentRepository.GetAgentNetworkInfoByID(id)
	if err != nil {
		return err
	}

	if f.sendAgentCommand(hostname, ip, "start") != nil {
		return fmt.Errorf("error encountered while starting agent")
	}

	if err = f.AgentQueue.AddAgent(id, hostname, ip); err != nil {
		return fmt.Errorf("error while starting agent monitoring")
	}

	return nil
}

// StopAgent sends a stop request to the agent
func (f *FrontendAgentService) StopAgent(id string) error {
	if !f.FrontendAgentRepository.AgentExists(id) {
		return utils.ErrAgentDoesNotExists
	}

	hostname, ip, err := f.FrontendAgentRepository.GetAgentNetworkInfoByID(id)
	if err != nil {
		return err
	}

	if err = f.AgentQueue.RemoveAgent(id); err != nil {
		return err
	}

	if err := f.sendAgentCommand(hostname, ip, "stop"); err != nil {
		f.AgentQueue.AddAgent(id, hostname, ip)
		return fmt.Errorf("error encountered while stopping agent")
	}
	return nil

}

// RestartMonitoring restarts monitoring for the agent
func (f *FrontendAgentService) RestartMonitoring(id string) error {
	if !f.FrontendAgentRepository.AgentExists(id) {
		return utils.ErrAgentDoesNotExists
	}

	hostname, ip, err := f.FrontendAgentRepository.GetAgentNetworkInfoByID(id)
	if err != nil {
		return err
	}

	if err = f.AgentQueue.AddAgent(id, hostname, ip); err != nil {
		return err
	}

	return nil
}

func (f *FrontendAgentService) GetHealthMetricsForGraph(id string) (*[]AgentMetrics, error) {
	if !f.FrontendAgentRepository.AgentExists(id) {
		return nil, utils.ErrAgentDoesNotExists
	}

	return f.FrontendAgentRepository.GetHealthMetricsForGraph(id)
}

func (f *FrontendAgentService) GetRateMetricsForGraph(id string) (*[]AgentMetrics, error) {
	if !f.FrontendAgentRepository.AgentExists(id) {
		return nil, utils.ErrAgentDoesNotExists
	}

	return f.FrontendAgentRepository.GetRateMetricsForGraph(id)
}

func (f *FrontendAgentService) AddLabels(id string, labels map[string]string) error {
	if !f.FrontendAgentRepository.AgentExists(id) {
		return utils.ErrAgentDoesNotExists
	}

	return f.FrontendAgentRepository.AddLabels(id, labels)
}

func (f *FrontendAgentService) sendAgentCommand(hostname, ip, command string) error {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// First try using hostname
	err := f.trySendingAgentCommand(client, hostname, command)
	if err != nil {
		// Fallback to IP if hostname fails
		ipErr := f.trySendingAgentCommand(client, ip, command)
		if ipErr != nil {
			return fmt.Errorf("hostname attempt failed: %v; IP attempt failed: %v", err, ipErr)
		}
	}

	return nil
}

func (f *FrontendAgentService) trySendingAgentCommand(client *http.Client, target, command string) error {
	url := fmt.Sprintf("http://%s:443/agent/v1/%s", target, command)

	resp, err := client.Post(url, "application/json", nil)
	if err != nil {
		return fmt.Errorf("error sending %s command to agent at %s: %w", command, target, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK response for %s command at %s: %s", command, target, resp.Status)
	}

	return nil
}
