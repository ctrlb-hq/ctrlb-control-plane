package frontendconfigV2

import (
	"time"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
)

// FrontendConfigServiceV2 manages frontend configuration operations
type FrontendConfigServiceV2 struct {
	FrontendConfigRepository *FrontendConfigRepositoryV2
}

// NewFrontendAgentServiceV2 initializes FrontendConfigService
func NewFrontendAgentServiceV2(frontendConfigRepository *FrontendConfigRepositoryV2) *FrontendConfigServiceV2 {
	return &FrontendConfigServiceV2{FrontendConfigRepository: frontendConfigRepository}
}

// GetAllConfigs retrieves all configurations
func (f *FrontendConfigServiceV2) GetAllConfigs() ([]models.Config, error) {
	return f.FrontendConfigRepository.GetAllConfigs()
}

func (f *FrontendConfigServiceV2) GetAllConfigsV2() ([]models.Config, error) {
	return f.FrontendConfigRepository.GetAllConfigs()
}

// CreateConfig creates a new configuration based on the provided request
func (f *FrontendConfigServiceV2) CreateConfig(createConfigRequest ConfigUpsertRequest) (*models.Config, error) {
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
func (f *FrontendConfigServiceV2) GetConfig(id string) (*models.Config, error) {
	return f.FrontendConfigRepository.GetConfig(id)
}

// DeleteConfig removes a configuration by ID
func (f *FrontendConfigServiceV2) DeleteConfig(id string) error {
	return f.FrontendConfigRepository.DeleteConfig(id)
}

// UpdateConfig modifies an existing configuration by ID
func (f *FrontendConfigServiceV2) UpdateConfig(id string, configUpdateRequest ConfigUpsertRequest) error {
	return f.FrontendConfigRepository.UpdateConfig(id, configUpdateRequest)
}
