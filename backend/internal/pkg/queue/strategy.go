package queue

import (
	io_prometheus_client "github.com/prometheus/client_model/go"
)

type AgentMetricsData struct {
	CPUUtilization    float64
	MemoryUtilization float64
	LogsRateSent      float64
	TracesRateSent    float64
	MetricsRateSent   float64
	DataSentBytes     float64
	DataReceivedBytes float64
}

// AgentMetricsStrategy defines the interface for extracting metrics from different agent types
type AgentMetricsStrategy interface {
	// ExtractMetrics extracts metrics from the provided Prometheus metric families
	ExtractMetrics(metrics map[string]*io_prometheus_client.MetricFamily, helper MetricsHelper) AgentMetricsData

	// GetPort returns the metrics endpoint port for this agent type
	GetPort() int
}

func GetMetricsStrategy(agentType string) AgentMetricsStrategy {
	switch agentType {
	case "fluent-bit":
		return FluentBitMetricsStrategy{}
	case "otel", "otel-collector":
		return OtelMetricsStrategy{}
	default:
		return OtelMetricsStrategy{}
	}
}

type FluentBitMetricsStrategy struct{}

func (s FluentBitMetricsStrategy) GetPort() int {
	return 2021
}

func (s FluentBitMetricsStrategy) ExtractMetrics(metrics map[string]*io_prometheus_client.MetricFamily, helper MetricsHelper) AgentMetricsData {
	cpuUser := helper.ExtractValueWithLabels(metrics, "process_cpu_seconds_total", map[string]string{
		"name": "fluent-bit",
		"mode": "user",
	})
	cpuSystem := helper.ExtractValueWithLabels(metrics, "process_cpu_seconds_total", map[string]string{
		"name": "fluent-bit",
		"mode": "system",
	})
	cpuUtil := cpuUser + cpuSystem

	memUtil := helper.ExtractValueWithLabels(metrics, "process_memory_bytes", map[string]string{
		"name": "fluent-bit",
		"type": "rss",
	})

	return AgentMetricsData{
		CPUUtilization:    cpuUtil,
		MemoryUtilization: memUtil,
		LogsRateSent:      0,
		TracesRateSent:    0,
		MetricsRateSent:   0,
		DataSentBytes:     0,
		DataReceivedBytes: 0,
	}
}

type OtelMetricsStrategy struct{}

func (s OtelMetricsStrategy) GetPort() int {
	return 8888
}

func (s OtelMetricsStrategy) ExtractMetrics(metrics map[string]*io_prometheus_client.MetricFamily, helper MetricsHelper) AgentMetricsData {
	logsRate := helper.ExtractValue(metrics, "otelcol_exporter_sent_log_records")
	tracesRate := helper.ExtractValue(metrics, "otelcol_exporter_sent_spans")
	metricsRate := helper.ExtractValue(metrics, "otelcol_exporter_sent_metric_points")
	dataSent := helper.ExtractValue(metrics, "otelcol_exporter_sent_bytes")
	dataReceived := helper.ExtractValue(metrics, "otelcol_receiver_accepted_bytes")

	cpuUtil := helper.ExtractValue(metrics, "system_cpu_utilization")
	if cpuUtil == 0 {
		cpuUtil = helper.ExtractValue(metrics, "otelcol_process_cpu_seconds_total")
	}

	memUtil := helper.ExtractValue(metrics, "system_memory_utilization")
	if memUtil == 0 {
		memUtil = helper.ExtractValue(metrics, "otelcol_process_memory_rss")
	}

	return AgentMetricsData{
		CPUUtilization:    cpuUtil,
		MemoryUtilization: memUtil,
		LogsRateSent:      logsRate,
		TracesRateSent:    tracesRate,
		MetricsRateSent:   metricsRate,
		DataSentBytes:     dataSent,
		DataReceivedBytes: dataReceived,
	}
}
