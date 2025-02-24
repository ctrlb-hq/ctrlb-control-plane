package frontendconfigV2

import (
	"context"
	"time"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
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

// CreateConfig creates a new configuration based on the provided request
func (f *FrontendConfigService) CreateConfig(ctx context.Context, createConfigRequest ConfigUpsertRequest) (*models.Config, error) {
	config := &models.Config{
		ID:          utils.CreateNewUUID(),
		Name:        createConfigRequest.Name,
		Description: createConfigRequest.Description,
		Config:      createConfigRequest.Config,
		TargetAgent: createConfigRequest.TargetAgent,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := f.FrontendConfigRepository.CreateConfig(ctx, config); err != nil {
		return nil, err
	}
	return config, nil
}

// GetConfig retrieves a specific configuration by ID
func (f *FrontendConfigService) GetConfig(ctx context.Context, id string) (*models.Config, error) {
	return f.FrontendConfigRepository.GetConfig(ctx, id)
}

// DeleteConfig removes a configuration by ID
func (f *FrontendConfigService) DeleteConfig(ctx context.Context, id string) error {
	return f.FrontendConfigRepository.DeleteConfig(ctx, id)
}

// UpdateConfig modifies an existing configuration by ID
func (f *FrontendConfigService) UpdateConfig(ctx context.Context, id string, configUpdateRequest ConfigUpsertRequest) error {
	return f.FrontendConfigRepository.UpdateConfig(ctx, id, configUpdateRequest)
}
