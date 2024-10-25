package frontendconfig

import (
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
func (f *FrontendConfigService) GetAllConfigs() ([]models.Config, error) {
	return f.FrontendConfigRepository.GetAllConfigs()
}

// CreateConfig creates a new configuration based on the provided request
func (f *FrontendConfigService) CreateConfig(createConfigRequest ConfigUpsertRequest) (*models.Config, error) {
	config := &models.Config{
		ID:          utils.CreateNewUUID(),
		Name:        createConfigRequest.Name,
		Description: createConfigRequest.Description,
		Config:      createConfigRequest.Config,
		TargetAgent: createConfigRequest.TargetAgent,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := f.FrontendConfigRepository.CreateConfig(config); err != nil {
		return nil, err
	}
	return config, nil
}

// GetConfig retrieves a specific configuration by ID
func (f *FrontendConfigService) GetConfig(id string) (*models.Config, error) {
	return f.FrontendConfigRepository.GetConfig(id)
}

// DeleteConfig removes a configuration by ID
func (f *FrontendConfigService) DeleteConfig(id string) error {
	return f.FrontendConfigRepository.DeleteConfig(id)
}

// UpdateConfig modifies an existing configuration by ID
func (f *FrontendConfigService) UpdateConfig(id string, configUpdateRequest ConfigUpsertRequest) error {
	return f.FrontendConfigRepository.UpdateConfig(id, configUpdateRequest)
}
