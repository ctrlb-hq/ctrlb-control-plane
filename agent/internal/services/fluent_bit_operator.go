package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ctrlb-hq/ctrlb-collector/internal/adapters"
	"github.com/ctrlb-hq/ctrlb-collector/internal/constants"
	"github.com/ctrlb-hq/ctrlb-collector/internal/models"
	"github.com/ctrlb-hq/ctrlb-collector/internal/utils"
	"github.com/prometheus/common/expfmt"
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

func (f *FluentBitOperator) GetUptime() (map[string]interface{}, error) {
	// Define the URL of the endpoint
	url := f.BaseURL + "/api/v1/uptime"

	client := &http.Client{Timeout: 10 * time.Second}

	// Make the GET request
	resp, err := client.Get(url)
	if err != nil {
		// Error in pinging the address, return DOWN and 0 uptime
		return map[string]interface{}{
			"status": "DOWN",
			"uptime": 0,
		}, err
	}
	defer resp.Body.Close()

	// Check if status code is not OK
	if resp.StatusCode != http.StatusOK {
		return map[string]interface{}{
			"status": "DOWN",
			"uptime": 0,
		}, errors.New("failed to get valid response from server")
	}

	// Decode the JSON response into a map
	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return map[string]interface{}{
			"status": "DOWN",
			"uptime": 0,
		}, err
	}

	// Extract uptime_sec from the map
	uptimeSec, ok := data["uptime_sec"].(float64)
	if !ok {
		// If we can't get uptime_sec, return DOWN and 0 uptime
		return map[string]interface{}{
			"status": "DOWN",
			"uptime": 0,
		}, errors.New("failed to parse uptime_sec")
	}

	// Determine status based on uptime_sec
	status := "DOWN"
	if uptimeSec > 0 {
		status = "UP"
	}

	// Return the result with status and uptime
	return map[string]interface{}{
		"status": status,
		"uptime": int(uptimeSec), // Convert float64 to int
	}, nil
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

	url := f.BaseURL + "/api/v1/metrics/prometheus"
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metrics: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read metrics: %v", err)
	}

	parser := expfmt.TextParser{}
	metrics, err := parser.TextToMetricFamilies(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse metrics: %v", err)
	}

	parsedMetrics := utils.ExtractFluentBitStatusFromPrometheus(metrics)

	status := make(map[string]string)

	status["Uptime"] = fmt.Sprintf("%.0f", parsedMetrics.Uptime)
	status["ExportedDataVolume"] = fmt.Sprintf("%.0f", parsedMetrics.ExportedDataVolume)
	status["DroppedRecords"] = fmt.Sprintf("%.0f", parsedMetrics.DroppedRecords)

	if parsedMetrics.Uptime > 0 {
		status["Status"] = "ON"
	} else {
		status["Status"] = "OFF"
	}

	return status, nil
}
