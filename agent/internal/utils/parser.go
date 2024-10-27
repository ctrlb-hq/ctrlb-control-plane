package utils

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ctrlb-hq/ctrlb-collector/internal/models"
	"gopkg.in/yaml.v3"

	io_prometheus_client "github.com/prometheus/client_model/go"
)

type Status struct {
	Uptime             float64
	ExportedDataVolume float64
	DroppedRecords     float64
}

func SaveToYAML(inputString string, filePath string) error {
	var validation map[string]interface{}
	if err := yaml.Unmarshal([]byte(inputString), &validation); err != nil {
		return fmt.Errorf("invalid YAML format: %v", err)
	}

	if _, err := os.Stat(filePath); err == nil {
		if err := os.Remove(filePath); err != nil {
			return fmt.Errorf("could not remove existing file at %s: %v", filePath, err)
		}
	}

	if err := os.WriteFile(filePath, []byte(inputString), 0644); err != nil {
		return fmt.Errorf("could not write YAML file: %v", err)
	}

	return nil
}

func LoadYAMLToJSON(yamlFilePath string, agentType string) (interface{}, error) {
	yamlData, err := os.ReadFile(yamlFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %v", err)
	}

	var config interface{}
	switch agentType {
	case "fluent-bit":
		var fluentBitConfig models.FluentBitConfig
		err = yaml.Unmarshal(yamlData, &fluentBitConfig)
		config = fluentBitConfig
	case "otel":
		var otelConfig models.OTELConfig
		err = yaml.Unmarshal(yamlData, &otelConfig)
		config = otelConfig
	default:
		return nil, fmt.Errorf("unsupported agent type: %s", agentType)
	}

	if err != nil {
		return nil, fmt.Errorf("error parsing YAML: %v", err)
	}

	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("error converting to JSON: %v", err)
	}

	var jsonInterface interface{}
	err = json.Unmarshal(jsonData, &jsonInterface)
	if err != nil {
		return nil, fmt.Errorf("error converting JSON to interface{}: %v", err)
	}

	return jsonInterface, nil
}

func ExtractStatusFromPrometheus(metrics map[string]*io_prometheus_client.MetricFamily, collector string) (*models.AgentMetrics, error) {
	agentMetrics := &models.AgentMetrics{}

	if collector == "fluent-bit" {
		if mf, ok := metrics["fluentbit_uptime"]; ok {
			for _, metric := range mf.Metric {
				if metric.Counter != nil {
					agentMetrics.UptimeSeconds = *metric.Counter.Value
				}
			}
		}

		if mf, ok := metrics["fluentbit_output_proc_bytes_total"]; ok {
			for _, metric := range mf.Metric {
				if metric.Counter != nil {
					agentMetrics.ExportedDataVolume = *metric.Counter.Value
				}
			}
		}

		if mf, ok := metrics["fluentbit_output_dropped_records_total"]; ok {
			for _, metric := range mf.Metric {
				if metric.Counter != nil {
					agentMetrics.DroppedRecords = *metric.Counter.Value
				}
			}
		}
	} else if collector == "otel" {
		if mf, ok := metrics["otelcol_process_uptime"]; ok {
			for _, metric := range mf.Metric {
				if metric.Counter != nil {
					agentMetrics.UptimeSeconds = *metric.Counter.Value
				}
			}
		}

		if mf, ok := metrics["otelcol_exporter_sent_log_records"]; ok {
			for _, metric := range mf.Metric {
				if metric.Counter != nil {
					agentMetrics.ExportedDataVolume = *metric.Counter.Value
				}
			}
		}

		if mf, ok := metrics["otelcol_exporter_send_failed_log_records"]; ok {
			for _, metric := range mf.Metric {
				if metric.Counter != nil {
					agentMetrics.DroppedRecords = *metric.Counter.Value
				}
			}
		}
	} else {
		return nil, fmt.Errorf("agent supplied for status metrics is not supported: %v", collector)
	}

	agentMetrics.Status = "DOWN"
	if agentMetrics.UptimeSeconds > 0 {
		agentMetrics.Status = "UP"
	}

	return agentMetrics, nil
}
