package operators

import (
	"encoding/json"
	"fmt"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/adapters"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/config"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/constants"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/pkg/logger"
)

type FluentBitOperator struct {
	BaseURL string
	Adapter adapters.Adapter
}

func NewFluentBitOperator(adapter adapters.Adapter) *FluentBitOperator {
	baseURL := "http://0.0.0.0:2020"
	return &FluentBitOperator{
		BaseURL: baseURL,
		Adapter: adapter,
	}
}

func (otc *FluentBitOperator) GetAgentInfo() (map[string]string, error) {
	version, err := otc.Adapter.GetVersion()
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"version":       version,
		"agent_type":    constants.AGENT_TYPE,
		"pipeline_name": constants.PIPELINE_NAME,
		"started_by":    constants.STARTED_BY,
	}, nil
}

func (otc *FluentBitOperator) Initialize() (map[string]string, error) {
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

func (otc *FluentBitOperator) StartAgent() error {
	err := otc.Adapter.StartAgent()
	if err != nil {
		return err
	}
	return nil
}

func (otc *FluentBitOperator) StopAgent() error {
	return otc.Adapter.StopAgent()
}

func (otc *FluentBitOperator) GracefulShutdown() error {
	go func() {
		err := otc.Adapter.GracefulShutdown()
		if err != nil {
			logger.Logger.Error(fmt.Sprintf("Error occurred while shutting down agent: %s", err))
		}
	}()

	return nil
}

func (otc *FluentBitOperator) UpdateCurrentConfig(updateConfigRequest map[string]any) error {
	if updateConfigRequest == nil {
		return fmt.Errorf("configuration data is nil")
	}

	// Validate configuration before saving
	if err := otc.Adapter.ValidateConfigInMemory(&updateConfigRequest); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Validate configuration against Fluent Bit binary using a temp file
	if !constants.SKIP_CONFIG_VALIDATION {
		logger.Logger.Info("Validating configuration against Fluent Bit binary using a temp file")
		if err := otc.Adapter.ValidateConfigOnDisk(&updateConfigRequest); err != nil {
			return fmt.Errorf("fluent-bit config validation with binary failed: %w", err)
		}
	}

	// If validation passes, save to the actual config path
	if err := config.SaveToYAML(updateConfigRequest, constants.AGENT_CONFIG_PATH); err != nil {
		return fmt.Errorf("failed to save config to final location: %w", err)
	}

	logger.Logger.Info("Configuration updated and validated successfully")
	return nil
}

func (otc *FluentBitOperator) GetMetrics() (map[string]any, error) {
	metrics, err := otc.Adapter.GetMetrics()
	if err != nil {
		return nil, err
	}
	return metrics, nil
}
