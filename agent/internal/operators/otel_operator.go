package operators

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/adapters"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/constants"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/models"
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
		log.Printf("Started procecss of initializing otel agent context")
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
	jsonStr := `{"message": "Otel lAgent starting up"}`
	log.Printf("Otel collector startup process initiated")
	err := otc.Adapter.StartAgent()
	if err != nil {
		log.Printf("error: %s", err.Error())
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
	log.Printf("Started process of stopping otel agent")
	err := otc.Adapter.StopAgent()
	if err != nil {
		log.Printf("error: %s", err.Error())
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
		log.Printf("Started process of Shutting down otel agent")
		err := otc.Adapter.GracefulShutdown()
		if err != nil {
			log.Printf("error: %s", err)
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

func (otc *OtelOperator) UpdateCurrentConfig(updateConfigRequest models.ConfigUpsertRequest) (map[string]string, error) {

	configString := updateConfigRequest.Config

	if err := utils.SaveToYAML(configString, constants.AGENT_CONFIG_PATH); err != nil {
		return nil, err
	}

	jsonStr := `{"message": "Configuration for otel agent has been updated"}`
	var result map[string]string
	_ = json.Unmarshal([]byte(jsonStr), &result)

	return result, nil
}

func (otc *OtelOperator) CurrentStatus() (*models.AgentMetrics, error) {
	return otc.Adapter.CurrentStatus()
}
