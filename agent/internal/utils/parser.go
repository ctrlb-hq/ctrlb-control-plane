package utils

import (
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

func LoadYAML(yamlFilePath string) (interface{}, error) {
	data, err := os.ReadFile(yamlFilePath)
	if err != nil {
		return "", err
	}

	var content interface{}
	if err := yaml.Unmarshal(data, &content); err != nil {
		return "", fmt.Errorf("invalid YAML format: %w", err)
	}

	return string(data), nil
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
