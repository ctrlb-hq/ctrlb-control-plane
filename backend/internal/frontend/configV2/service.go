package frontendconfigV2

import (
	"context"
	"time"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
)

// FrontendConfigService manages frontend configuration operations
type FrontendConfigService struct {
	FrontendConfigRepository *FrontendConfigRepository
}

// NewFrontendAgentService initializes FrontendConfigService
func NewFrontendAgentService(frontendConfigRepository *FrontendConfigRepository) *FrontendConfigService {
	return &FrontendConfigService{FrontendConfigRepository: frontendConfigRepository}
}

// GetAllConfigs retrieves all configurations
func (f *FrontendConfigService) GetAllConfigs(ctx context.Context) ([]models.ConfigSet, error) {
	return f.FrontendConfigRepository.GetAllConfigs(ctx)
}

// GetConfig retrieves a specific configuration by ID
func (f *FrontendConfigService) GetConfig(ctx context.Context, id string) (map[string]any, error) {
	return f.FrontendConfigRepository.GetConfig(ctx, id)
}

// CreateConfig creates a new configuration based on the provided request
func (f *FrontendConfigService) CreateConfigSet(ctx context.Context, congigSetUpsertRequest *ConfigSetUpsertRequest) (*models.ConfigSet, error) {
	congigSet := &models.ConfigSet{
		Name:        congigSetUpsertRequest.Name,
		Credentials: congigSetUpsertRequest.Credentials,
		Version:     "v0.0.1",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := f.FrontendConfigRepository.CreateConfigSet(ctx, congigSet); err != nil {
		return nil, err
	}
	return congigSet, nil
}

func (f *FrontendConfigService) GetPipelines(ctx context.Context, id string) ([]models.Pipeline, error) {
	return f.FrontendConfigRepository.GetPipelines(ctx, id)
}

func (f *FrontendConfigService) CreatePipelines(ctx context.Context, id int, createPipelinesRequest *CreatePipelinesRequest) (*models.Pipeline, error) {
	pipelines := models.Pipeline{
		Name:        createPipelinesRequest.Name,
		Type:        createPipelinesRequest.Type,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		ConfigSetID: id,
	}

	err := f.FrontendConfigRepository.CreatePipelines(ctx, pipelines)
	if err != nil {
		return nil, err
	}
	return &pipelines, nil
}

func (f *FrontendConfigService) CreatePipelineComponents(ctx context.Context, id int, createPipelineComponentRequest *CreatePipelineComponentRequest) (*models.PipelineComponent, error) {
	pipelineComponent := models.PipelineComponent{
		Name:          createPipelineComponentRequest.Name,
		Type:          createPipelineComponentRequest.Type,
		Config:        createPipelineComponentRequest.Config,
		ComponentType: createPipelineComponentRequest.ComponentType,
		PipelineID:    createPipelineComponentRequest.PipelineID,
	}

	err := f.FrontendConfigRepository.CreatePipelineComponents(ctx, pipelineComponent)
	if err != nil {
		return nil, err
	}
	return &pipelineComponent, nil
}

// DeleteConfig removes a configuration by ID
func (f *FrontendConfigService) DeleteConfig(ctx context.Context, id string) error {
	return f.FrontendConfigRepository.DeleteConfig(ctx, id)
}

// UpdateConfig modifies an existing configuration by ID
func (f *FrontendConfigService) UpdateConfig(ctx context.Context, id string, configUpdateRequest ConfigUpsertRequest) error {
	return f.FrontendConfigRepository.UpdateConfig(ctx, id, configUpdateRequest)
}
