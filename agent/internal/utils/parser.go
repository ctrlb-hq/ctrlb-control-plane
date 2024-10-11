package utils

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ctrlb-hq/ctrlb-collector/internal/models"
	"gopkg.in/yaml.v3"

	io_prometheus_client "github.com/prometheus/client_model/go"
)

type FluentBitStatus struct {
	Uptime             float64
	ExportedDataVolume float64
	DroppedRecords     float64
}

func SaveToYAML(input interface{}, yamlFilePath string) error {
	// Convert the input interface to YAML format
	yamlData, err := yaml.Marshal(input)
	if err != nil {
		return fmt.Errorf("error converting to YAML: %v", err)
	}

	// Check if a YAML file already exists at the given path
	if _, err := os.Stat(yamlFilePath); err == nil {
		// Remove the existing file if found
		err := os.Remove(yamlFilePath)
		if err != nil {
			return fmt.Errorf("could not remove existing YAML file: %v", err)
		}
	}

	// Write the new YAML data to the specified path
	err = os.WriteFile(yamlFilePath, yamlData, 0644)
	if err != nil {
		return fmt.Errorf("could not write YAML file: %v", err)
	}

	// Return nil if no error occurred
	return nil
}

func LoadYAMLToJSON(yamlFilePath string) (interface{}, error) {
	yamlData, err := os.ReadFile(yamlFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %v", err)
	}

	var config models.FluentBitConfig
	err = yaml.Unmarshal(yamlData, &config)
	if err != nil {
		return nil, fmt.Errorf("error parsing YAML: %v", err)
	}

	// Convert config to JSON format
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("error converting to JSON: %v", err)
	}

	// Parse the JSON data into an interface{}
	var jsonInterface interface{}
	err = json.Unmarshal(jsonData, &jsonInterface)
	if err != nil {
		return nil, fmt.Errorf("error converting JSON to interface{}: %v", err)
	}

	return jsonInterface, nil
}

func ExtractFluentBitStatusFromPrometheus(metrics map[string]*io_prometheus_client.MetricFamily) FluentBitStatus {
	parsedMetrics := FluentBitStatus{
		Uptime:             0.0,
		ExportedDataVolume: 0.0,
		DroppedRecords:     0.0,
	}

	if mf, ok := metrics["fluentbit_uptime"]; ok {
		for _, metric := range mf.Metric {
			if metric.Counter != nil {
				parsedMetrics.Uptime = *metric.Counter.Value
			}
		}
	}

	if mf, ok := metrics["fluentbit_output_proc_bytes_total"]; ok {
		for _, metric := range mf.Metric {
			if metric.Counter != nil {
				parsedMetrics.ExportedDataVolume = *metric.Counter.Value
			}
		}
	}

	if mf, ok := metrics["fluentbit_output_dropped_records_total"]; ok {
		for _, metric := range mf.Metric {
			if metric.Counter != nil {
				parsedMetrics.DroppedRecords = *metric.Counter.Value
			}
		}
	}

	return parsedMetrics
}
