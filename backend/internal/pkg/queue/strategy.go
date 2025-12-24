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

// MetricsSource defines a single endpoint to fetch metrics from
type MetricsSource struct {
	Port int
	Path string
}

// AgentMetricsStrategy defines the interface for extracting metrics from different agent types
type AgentMetricsStrategy interface {
	// GetMetricsSources returns all endpoints that need to be queried for this agent type
	GetMetricsSources() []MetricsSource

	// ExtractMetrics extracts metrics from the provided Prometheus metric families
	// The metrics map contains merged data from all sources
	ExtractMetrics(metrics map[string]*io_prometheus_client.MetricFamily, helper MetricsHelper) AgentMetricsData
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

func (s FluentBitMetricsStrategy) GetMetricsSources() []MetricsSource {
	return []MetricsSource{
		{Port: 2021, Path: "/metrics"},                   // Process exporter for CPU/Memory
		{Port: 2020, Path: "/api/v2/metrics/prometheus"}, // Fluent Bit internal metrics
	}
}

func (s FluentBitMetricsStrategy) ExtractMetrics(metrics map[string]*io_prometheus_client.MetricFamily, helper MetricsHelper) AgentMetricsData {
	// Extract CPU and memory from process exporter (port 2021)
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

	// Extract Fluent Bit specific metrics (from port 2020)
	// Sum up input records across all inputs for logs rate
	logsRate := helper.ExtractValue(metrics, "fluentbit_input_records_total")

	// For traces and metrics, Fluent Bit typically doesn't expose these separately
	// as it primarily handles logs. These would be 0 for Fluent Bit agents.
	tracesRate := float64(0)
	metricsRate := float64(0)

	// Sum up output bytes across all outputs
	dataSent := helper.ExtractValue(metrics, "fluentbit_output_proc_bytes_total")

	// Sum up input bytes across all inputs
	dataReceived := helper.ExtractValue(metrics, "fluentbit_input_bytes_total")

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

type OtelMetricsStrategy struct{}

func (s OtelMetricsStrategy) GetMetricsSources() []MetricsSource {
	return []MetricsSource{
		{Port: 8888, Path: "/metrics"}, // OTEL collector metrics endpoint
	}
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
