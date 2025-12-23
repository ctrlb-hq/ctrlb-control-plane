package queue_test

import (
	"testing"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/pkg/queue"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

// mockMetricsHelperForStrategy is a mock helper for strategy tests
type mockMetricsHelperForStrategy struct{}

func (m mockMetricsHelperForStrategy) Fetch(url string) (map[string]*io_prometheus_client.MetricFamily, error) {
	return nil, nil
}

func (m mockMetricsHelperForStrategy) ExtractValue(metrics map[string]*io_prometheus_client.MetricFamily, name string) float64 {
	// Return mock values based on metric name
	mockValues := map[string]float64{
		"otelcol_exporter_sent_log_records":   100.0,
		"otelcol_exporter_sent_spans":         200.0,
		"otelcol_exporter_sent_metric_points": 300.0,
		"otelcol_exporter_sent_bytes":         1024.0,
		"otelcol_receiver_accepted_bytes":     2048.0,
		"system_cpu_utilization":              50.5,
		"system_memory_utilization":           75.3,
		"otelcol_process_cpu_seconds_total":   10.0,
		"otelcol_process_memory_rss":          512.0,
	}
	if val, ok := mockValues[name]; ok {
		return val
	}
	return 0
}

func (m mockMetricsHelperForStrategy) ExtractValueWithLabels(metrics map[string]*io_prometheus_client.MetricFamily, name string, labels map[string]string) float64 {
	// Return mock values based on metric name and labels
	if name == "process_cpu_seconds_total" {
		if labels["name"] == "fluent-bit" && labels["mode"] == "user" {
			return 5.5
		}
		if labels["name"] == "fluent-bit" && labels["mode"] == "system" {
			return 3.2
		}
	}
	if name == "process_memory_bytes" {
		if labels["name"] == "fluent-bit" && labels["type"] == "rss" {
			return 4096.0
		}
	}
	return 0
}

func TestFluentBitMetricsStrategy_GetPort(t *testing.T) {
	strategy := queue.FluentBitMetricsStrategy{}
	assert.Equal(t, 2021, strategy.GetPort())
}

func TestFluentBitMetricsStrategy_ExtractMetrics(t *testing.T) {
	strategy := queue.FluentBitMetricsStrategy{}
	helper := mockMetricsHelperForStrategy{}

	// Empty metrics map - the helper will return mock values
	metrics := map[string]*io_prometheus_client.MetricFamily{}

	data := strategy.ExtractMetrics(metrics, helper)

	// Verify CPU is sum of user + system
	assert.Equal(t, 8.7, data.CPUUtilization, "CPU should be user(5.5) + system(3.2)")

	// Verify memory
	assert.Equal(t, 4096.0, data.MemoryUtilization)

	// Verify application metrics are zero (not available from node_exporter)
	assert.Equal(t, 0.0, data.LogsRateSent)
	assert.Equal(t, 0.0, data.TracesRateSent)
	assert.Equal(t, 0.0, data.MetricsRateSent)
	assert.Equal(t, 0.0, data.DataSentBytes)
	assert.Equal(t, 0.0, data.DataReceivedBytes)
}

func TestOtelMetricsStrategy_GetPort(t *testing.T) {
	strategy := queue.OtelMetricsStrategy{}
	assert.Equal(t, 8888, strategy.GetPort())
}

func TestOtelMetricsStrategy_ExtractMetrics(t *testing.T) {
	strategy := queue.OtelMetricsStrategy{}
	helper := mockMetricsHelperForStrategy{}

	// Empty metrics map - the helper will return mock values
	metrics := map[string]*io_prometheus_client.MetricFamily{}

	data := strategy.ExtractMetrics(metrics, helper)

	// Verify all metrics are extracted
	assert.Equal(t, 100.0, data.LogsRateSent)
	assert.Equal(t, 200.0, data.TracesRateSent)
	assert.Equal(t, 300.0, data.MetricsRateSent)
	assert.Equal(t, 1024.0, data.DataSentBytes)
	assert.Equal(t, 2048.0, data.DataReceivedBytes)
	assert.Equal(t, 50.5, data.CPUUtilization)
	assert.Equal(t, 75.3, data.MemoryUtilization)
}

func TestOtelMetricsStrategy_ExtractMetrics_FallbackToProcessMetrics(t *testing.T) {
	strategy := queue.OtelMetricsStrategy{}

	// Create a helper that returns 0 for system metrics, forcing fallback
	helper := &mockMetricsHelperFallback{}

	metrics := map[string]*io_prometheus_client.MetricFamily{}

	data := strategy.ExtractMetrics(metrics, helper)

	// Should use process metrics as fallback
	assert.Equal(t, 10.0, data.CPUUtilization)
	assert.Equal(t, 512.0, data.MemoryUtilization)
}

// mockMetricsHelperFallback returns 0 for system metrics
type mockMetricsHelperFallback struct{}

func (m *mockMetricsHelperFallback) Fetch(url string) (map[string]*io_prometheus_client.MetricFamily, error) {
	return nil, nil
}

func (m *mockMetricsHelperFallback) ExtractValue(metrics map[string]*io_prometheus_client.MetricFamily, name string) float64 {
	mockValues := map[string]float64{
		"system_cpu_utilization":            0, // Return 0 to trigger fallback
		"system_memory_utilization":         0, // Return 0 to trigger fallback
		"otelcol_process_cpu_seconds_total": 10.0,
		"otelcol_process_memory_rss":        512.0,
	}
	if val, ok := mockValues[name]; ok {
		return val
	}
	return 0
}

func (m *mockMetricsHelperFallback) ExtractValueWithLabels(metrics map[string]*io_prometheus_client.MetricFamily, name string, labels map[string]string) float64 {
	return 0
}

func TestGetMetricsStrategy_FluentBit(t *testing.T) {
	strategy := queue.GetMetricsStrategy("fluent-bit")
	assert.IsType(t, queue.FluentBitMetricsStrategy{}, strategy)
}

func TestGetMetricsStrategy_Otel(t *testing.T) {
	strategy := queue.GetMetricsStrategy("otel")
	assert.IsType(t, queue.OtelMetricsStrategy{}, strategy)
}

func TestGetMetricsStrategy_OtelCollector(t *testing.T) {
	strategy := queue.GetMetricsStrategy("otel-collector")
	assert.IsType(t, queue.OtelMetricsStrategy{}, strategy)
}

func TestGetMetricsStrategy_UnknownType(t *testing.T) {
	// Unknown types should default to Otel strategy
	strategy := queue.GetMetricsStrategy("unknown-agent")
	assert.IsType(t, queue.OtelMetricsStrategy{}, strategy)
}

func TestGetMetricsStrategy_EmptyType(t *testing.T) {
	// Empty type should default to Otel strategy
	strategy := queue.GetMetricsStrategy("")
	assert.IsType(t, queue.OtelMetricsStrategy{}, strategy)
}
