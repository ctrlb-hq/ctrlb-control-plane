package operators

import (
	"encoding/json"
	"fmt"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/adapters"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/config"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/constants"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/pkg/logger"
)

type OtelOperator struct {
	BaseURL string
	Adapter adapters.Adapter
}

func NewOtelOperator(adapter adapters.Adapter) *OtelOperator {
	baseURL := "http://0.0.0.0:2020"
	return &OtelOperator{
		BaseURL: baseURL,
		Adapter: adapter,
	}
}

func (otc *OtelOperator) Initialize() (map[string]string, error) {
	go func() {
		logger.Logger.Info("Started process of initializing otel agent context")
		if err := otc.Adapter.Initialize(); err != nil {
			logger.Logger.Error(fmt.Sprintf("Failed to initialize adapter: %s", err))
		}
	}()
	jsonStr := `{"message": "Otel Agent initializing"}`

	// Create a map to hold the result
	var result map[string]string

	// Unmarshal the JSON into the map
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (otc *OtelOperator) StartAgent() error {
	err := otc.Adapter.StartAgent()
	if err != nil {
		return err
	}
	return nil
}

func (otc *OtelOperator) StopAgent() error {
	return otc.Adapter.StopAgent()
}

func (otc *OtelOperator) GracefulShutdown() error {
	go func() {
		err := otc.Adapter.GracefulShutdown()
		if err != nil {
			logger.Logger.Error(fmt.Sprintf("Error occurred while shutting down agent: %s", err))
		}
	}()

	return nil
}

func (otc *OtelOperator) UpdateCurrentConfig(updateConfigRequest map[string]any) error {
	if updateConfigRequest == nil {
		return fmt.Errorf("configuration data is nil")
	}

	// Validate configuration before saving
	if err := otc.Adapter.ValidateConfigInMemory(&updateConfigRequest); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// If validation passes, save to the actual config path
	if err := config.SaveToYAML(updateConfigRequest, constants.AGENT_CONFIG_PATH); err != nil {
		return fmt.Errorf("failed to save config to final location: %w", err)
	}

	logger.Logger.Info("Configuration updated and validated successfully")
	return nil
}
