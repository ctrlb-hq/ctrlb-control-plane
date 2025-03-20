package operators

import (
	"encoding/json"
	"fmt"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/adapters"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/constants"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/models"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/pkg"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/utils"
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
		pkg.Logger.Info("Started procecss of initializing otel agent context")
		otc.Adapter.Initialize()
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

func (otc *OtelOperator) StartAgent() (map[string]string, error) {
	jsonStr := `{"message": "Otel Agent starting up"}`
	err := otc.Adapter.StartAgent()
	if err != nil {
		jsonStr = fmt.Sprintf(`{"message": "%s"}`, err.Error())
	}

	// Create a map to hold the result
	var result map[string]string

	// Unmarshal the JSON into the map
	err = json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (otc *OtelOperator) StopAgent() (map[string]string, error) {
	jsonStr := `{"message": "Otel Agent stopping"}`
	err := otc.Adapter.StopAgent()
	if err != nil {
		jsonStr = fmt.Sprintf(`{"message": "%s"}`, err.Error())
	}

	// Create a map to hold the result
	var result map[string]string

	// Unmarshal the JSON into the map
	err = json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (otc *OtelOperator) GracefulShutdown() (map[string]string, error) {
	jsonStr := `{"message": "Otel agent shutting down"}`
	go func() {
		err := otc.Adapter.GracefulShutdown()
		if err != nil {
			pkg.Logger.Error(fmt.Sprintf("Error occured while shutting down agent: %s", err))
		}
	}()

	// Create a map to hold the result
	var result map[string]string

	// Unmarshal the JSON into the map
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (otc *OtelOperator) UpdateCurrentConfig(updateConfigRequest map[string]any) error {

	if err := utils.SaveToYAML(updateConfigRequest, constants.AGENT_CONFIG_PATH); err != nil {
		return err
	}

	return nil
}

func (otc *OtelOperator) CurrentStatus() (*models.AgentMetrics, error) {
	return otc.Adapter.CurrentStatus()
}
