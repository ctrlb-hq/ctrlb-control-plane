package services

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ctrlb-hq/ctrlb-collector/internal/adapters"
	"github.com/ctrlb-hq/ctrlb-collector/internal/constants"
	"github.com/ctrlb-hq/ctrlb-collector/internal/models"
	"github.com/ctrlb-hq/ctrlb-collector/internal/utils"
)

type FluentBitOperator struct {
	Adapter adapters.Adapter
}

func NewFluentBitOperator(adapter adapters.Adapter) *FluentBitOperator {
	return &FluentBitOperator{
		Adapter: adapter,
	}
}

func (f *FluentBitOperator) GetUptime() (map[string]interface{}, error) {
	return f.Adapter.GetUptime()
}

func (f *FluentBitOperator) Initialize() (map[string]string, error) {
	go func() {
		log.Printf("Started procecss of initializing agent context")
		f.Adapter.Initialize()
	}()
	jsonStr := `{"message": "Agent initializing"}`

	// Create a map to hold the result
	var result map[string]string

	// Unmarshal the JSON into the map
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (f *FluentBitOperator) StartAgent() (map[string]string, error) {
	jsonStr := `{"message": "Agent starting up"}`
	log.Printf("Startup process initiated")
	err := f.Adapter.StartAgent()
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

func (f *FluentBitOperator) StopAgent() (map[string]string, error) {
	jsonStr := `{"message": "Agent stopping"}`
	log.Printf("Started process of stopping agent")
	err := f.Adapter.StopAgent()
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

func (f *FluentBitOperator) GracefulShutdown() (map[string]string, error) {
	jsonStr := `{"message": "Agent shutting down"}`
	go func() {
		log.Printf("Started process of Shutting down")
		err := f.Adapter.GracefulShutdown()
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

func (f *FluentBitOperator) UpdateCurrentConfig(updateConfigRequest interface{}) (map[string]string, error) {

	request, ok := updateConfigRequest.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid request body while updating config, expected a map[string]interface{}")
	}

	configString, ok := request["config"].(string)
	if !ok {
		return nil, fmt.Errorf("config field is missing or not a string")
	}

	var config models.FluentBitConfig

	err := json.Unmarshal([]byte(configString), &config)
	if err != nil {
		return nil, err
	}

	if err = utils.SaveToYAML(config, constants.AGENT_CONFIG_PATH); err != nil {
		return nil, err
	}

	f.Adapter.UpdateConfig()

	jsonStr := `{"message": "Configuration has been updated"}`
	var result map[string]string
	_ = json.Unmarshal([]byte(jsonStr), &result)

	return result, nil
}

func (f *FluentBitOperator) CurrentStatus() (map[string]string, error) {
	return f.Adapter.CurrentStatus()
}
