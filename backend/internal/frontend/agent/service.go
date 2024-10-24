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

func NewFrontendAgentService(frontendAgentRepository *FrontendAgentRepository, agentQueue *queue.AgentQueue) *FrontendAgentService {
	return &FrontendAgentService{
		FrontendAgentRepository: frontendAgentRepository,
		AgentQueue:              agentQueue,
	}
}

func (f *FrontendAgentService) GetAllAgents() ([]Agent, error) {
	agents, err := f.FrontendAgentRepository.GetAllAgents()
	if err != nil {
		return nil, err
	}
	return agents, nil
}

func (f *FrontendAgentService) GetAgent(id string) (*models.AgentWithConfig, error) {
	agent, err := f.FrontendAgentRepository.GetAgent(id)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	config, err := f.FrontendAgentRepository.GetConfig(agent.ConfigID)
	if err != nil {
		return nil, err
	}

	agentWithConfig := &models.AgentWithConfig{
		ID:           agent.ID,
		Name:         agent.Name,
		Type:         agent.Type,
		Version:      agent.Version,
		Hostname:     agent.Hostname,
		Platform:     agent.Platform,
		Config:       *config,
		IsPipeline:   agent.IsPipeline,
		RegisteredAt: agent.RegisteredAt,
	}

	return agentWithConfig, nil
}

func (f *FrontendAgentService) DeleteAgent(id string) error {
	agent, err := f.FrontendAgentRepository.GetAgent(id)
	if err != nil {
		return err
	}

	err = f.FrontendAgentRepository.DeleteAgent(agent.ID)
	if err != nil {
		return err
	}

	//Shutting down the agent
	url := fmt.Sprintf("http://%s:443/api/v1/shutdown", agent.Hostname)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return fmt.Errorf("error encountered while removing agent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error encountered while removing agent: %s", resp.Status)
	}

	f.AgentQueue.RemoveAgent(agent.ID)

	return err
}

func (f *FrontendAgentService) StartAgent(id string) error {

	// starting registered agent
	agent, err := f.FrontendAgentRepository.GetAgent(id)
	if err != nil {
		return err
	}

	// prepare the URL for the start agent request using the extracted hostname
	url := fmt.Sprintf("http://%s:443/api/v1/start", agent.Hostname)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return fmt.Errorf("error encountered while starting agent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error encountered while starting agent: %s", resp.Status)
	}

	return nil
}

func (f *FrontendAgentService) StopAgent(id string) error {
	// starting registered agent
	agent, err := f.FrontendAgentRepository.GetAgent(id)
	if err != nil {
		return err
	}

	// prepare the URL for the start agent request using the extracted hostname
	url := fmt.Sprintf("http://%s:443/api/v1/stop", agent.Hostname)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return fmt.Errorf("error encountered while stopping agent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error encountered while stopping agent: %s", resp.Status)
	}

	return nil
}

func (f *FrontendAgentService) GetMetrics(id string) (*models.AgentMetrics, error) {
	agentMetrics, err := f.FrontendAgentRepository.GetMetrics(id)
	if err != nil {
		return nil, err
	}

	return agentMetrics, nil
}

func (f *FrontendAgentService) RestartMonitoring(id string) error {
	agent, err := f.FrontendAgentRepository.GetAgent(id)
	if err != nil {
		return err
	}

	f.AgentQueue.AddAgent(agent.ID, agent.Hostname)

	return nil
}
