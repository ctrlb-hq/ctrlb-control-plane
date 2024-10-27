package frontendpipeline

import (
	"fmt"
	"net/http"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/queue"
)

// FrontendPipelineService handles business logic for pipelines.
type FrontendPipelineService struct {
	FrontendPipelineRepository *FrontendPipelineRepository // Repository for pipeline data
	PipelineQueue              *queue.AgentQueue           // Queue for managing agents
}

// NewFrontendPipelineService creates a new FrontendPipelineService instance.
func NewFrontendPipelineService(frontendRepository *FrontendPipelineRepository, pipelineQueue *queue.AgentQueue) *FrontendPipelineService {
	return &FrontendPipelineService{
		FrontendPipelineRepository: frontendRepository,
		PipelineQueue:              pipelineQueue,
	}
}

// GetAllPipelines retrieves all pipelines.
func (f *FrontendPipelineService) GetAllPipelines() ([]Pipeline, error) {
	pipelines, err := f.FrontendPipelineRepository.GetAllPipelines()
	if err != nil {
		return nil, err
	}
	return pipelines, nil
}

// GetPipeline retrieves a specific pipeline with its configuration.
func (f *FrontendPipelineService) GetPipeline(id string) (*models.AgentWithConfig, error) {
	pipeline, err := f.FrontendPipelineRepository.GetPipeline(id)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	config, err := f.FrontendPipelineRepository.GetConfig(pipeline.ConfigID)
	if err != nil {
		return nil, err
	}

	agentWithConfig := &models.AgentWithConfig{
		ID:           pipeline.ID,
		Name:         pipeline.Name,
		Type:         pipeline.Type,
		Version:      pipeline.Version,
		Hostname:     pipeline.Hostname,
		Platform:     pipeline.Platform,
		Config:       *config,
		IsPipeline:   pipeline.IsPipeline,
		RegisteredAt: pipeline.RegisteredAt,
	}

	return agentWithConfig, nil
}

// DeletePipeline removes a pipeline and shuts it down.
func (f *FrontendPipelineService) DeletePipeline(id string) error {
	pipeline, err := f.FrontendPipelineRepository.GetPipeline(id)
	if err != nil {
		return err
	}

	err = f.FrontendPipelineRepository.DeletePipeline(pipeline.ID)
	if err != nil {
		return err
	}

	// Shutdown the pipeline
	url := fmt.Sprintf("http://%s:443/agent/v1/shutdown", pipeline.Hostname)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return fmt.Errorf("error encountered while removing pipeline: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error encountered while removing pipeline: %s", resp.Status)
	}

	f.PipelineQueue.RemoveAgent(pipeline.ID)

	return nil
}

// StartPipeline starts a registered pipeline.
func (f *FrontendPipelineService) StartPipeline(id string) error {
	pipeline, err := f.FrontendPipelineRepository.GetPipeline(id)
	if err != nil {
		return err
	}

	// Prepare the URL for starting the pipeline
	url := fmt.Sprintf("http://%s:443/agent/v1/start", pipeline.Hostname)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return fmt.Errorf("error encountered while starting pipeline: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error encountered while starting pipeline: %s", resp.Status)
	}

	return nil
}

// StopPipeline stops a registered pipeline.
func (f *FrontendPipelineService) StopPipeline(id string) error {
	pipeline, err := f.FrontendPipelineRepository.GetPipeline(id)
	if err != nil {
		return err
	}

	// Prepare the URL for stopping the pipeline
	url := fmt.Sprintf("http://%s:443/agent/v1/stop", pipeline.Hostname)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return fmt.Errorf("error encountered while stopping pipeline: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error encountered while stopping pipeline: %s", resp.Status)
	}

	return nil
}

// GetMetrics retrieves metrics for a specific pipeline.
func (f *FrontendPipelineService) GetMetrics(id string) (*models.AgentMetrics, error) {
	pipelineMetrics, err := f.FrontendPipelineRepository.GetMetrics(id)
	if err != nil {
		return nil, err
	}

	return pipelineMetrics, nil
}

// RestartMonitoring restarts monitoring for a specific pipeline.
func (f *FrontendPipelineService) RestartMonitoring(id string) error {
	agent, err := f.FrontendPipelineRepository.GetPipeline(id)
	if err != nil {
		return err
	}

	f.PipelineQueue.AddAgent(agent.ID, agent.Hostname)

	return nil
}
