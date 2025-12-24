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
		// OTEL metrics
		"otelcol_exporter_sent_log_records":   100.0,
		"otelcol_exporter_sent_spans":         200.0,
		"otelcol_exporter_sent_metric_points": 300.0,
		"otelcol_exporter_sent_bytes":         1024.0,
		"otelcol_receiver_accepted_bytes":     2048.0,
		"system_cpu_utilization":              50.5,
		"system_memory_utilization":           75.3,
		"otelcol_process_cpu_seconds_total":   10.0,
		"otelcol_process_memory_rss":          512.0,
		// Fluent Bit metrics
		"fluentbit_input_records_total":     150.0,
		"fluentbit_output_proc_bytes_total": 5120.0,
		"fluentbit_input_bytes_total":       3072.0,
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

func TestFluentBitMetricsStrategy_GetMetricsSources(t *testing.T) {
	strategy := queue.FluentBitMetricsStrategy{}
	sources := strategy.GetMetricsSources()

	assert.Len(t, sources, 2, "FluentBit should have 2 metrics sources")

	// Verify first source (process exporter)
	assert.Equal(t, 2021, sources[0].Port)
	assert.Equal(t, "/metrics", sources[0].Path)

	// Verify second source (fluent bit internal metrics)
	assert.Equal(t, 2020, sources[1].Port)
	assert.Equal(t, "/api/v2/metrics/prometheus", sources[1].Path)
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

	// Verify Fluent Bit specific metrics (now extracted from port 2020)
	assert.Equal(t, 150.0, data.LogsRateSent, "Logs rate from fluentbit_input_records_total")
	assert.Equal(t, 0.0, data.TracesRateSent, "Fluent Bit doesn't handle traces")
	assert.Equal(t, 0.0, data.MetricsRateSent, "Fluent Bit doesn't expose metrics rate separately")
	assert.Equal(t, 5120.0, data.DataSentBytes, "Data sent from fluentbit_output_proc_bytes_total")
	assert.Equal(t, 3072.0, data.DataReceivedBytes, "Data received from fluentbit_input_bytes_total")
}

func TestOtelMetricsStrategy_GetMetricsSources(t *testing.T) {
	strategy := queue.OtelMetricsStrategy{}
	sources := strategy.GetMetricsSources()

	assert.Len(t, sources, 1, "OTEL should have 1 metrics source")
	assert.Equal(t, 8888, sources[0].Port)
	assert.Equal(t, "/metrics", sources[0].Path)
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

func TestMetricsSourceStructure(t *testing.T) {
	// Test that MetricsSource can be created and accessed
	source := queue.MetricsSource{
		Port: 9090,
		Path: "/custom/metrics",
	}

	assert.Equal(t, 9090, source.Port)
	assert.Equal(t, "/custom/metrics", source.Path)
}

func TestFluentBitMetricsStrategy_MultipleSourcesIntegration(t *testing.T) {
	strategy := queue.FluentBitMetricsStrategy{}
	helper := mockMetricsHelperForStrategy{}

	// Simulate merged metrics from both sources
	metrics := map[string]*io_prometheus_client.MetricFamily{}

	data := strategy.ExtractMetrics(metrics, helper)

	// Verify we get both CPU/memory (from port 2021) and fluent bit metrics (from port 2020)
	assert.NotZero(t, data.CPUUtilization, "Should have CPU metrics from process exporter")
	assert.NotZero(t, data.MemoryUtilization, "Should have memory metrics from process exporter")
	assert.NotZero(t, data.LogsRateSent, "Should have logs rate from Fluent Bit metrics")
	assert.NotZero(t, data.DataSentBytes, "Should have data sent from Fluent Bit metrics")
	assert.NotZero(t, data.DataReceivedBytes, "Should have data received from Fluent Bit metrics")
}

func TestFluentBitMetricsStrategy_ExtractMetrics_EmptyMetrics(t *testing.T) {
	strategy := queue.FluentBitMetricsStrategy{}

	// Use a helper that returns zero for all metrics
	helper := &mockMetricsHelperEmpty{}
	metrics := map[string]*io_prometheus_client.MetricFamily{}

	data := strategy.ExtractMetrics(metrics, helper)

	// All values should be zero when metrics are not available
	assert.Equal(t, 0.0, data.CPUUtilization)
	assert.Equal(t, 0.0, data.MemoryUtilization)
	assert.Equal(t, 0.0, data.LogsRateSent)
	assert.Equal(t, 0.0, data.TracesRateSent)
	assert.Equal(t, 0.0, data.MetricsRateSent)
	assert.Equal(t, 0.0, data.DataSentBytes)
	assert.Equal(t, 0.0, data.DataReceivedBytes)
}

func TestOtelMetricsStrategy_ExtractMetrics_EmptyMetrics(t *testing.T) {
	strategy := queue.OtelMetricsStrategy{}
	helper := &mockMetricsHelperEmpty{}
	metrics := map[string]*io_prometheus_client.MetricFamily{}

	data := strategy.ExtractMetrics(metrics, helper)

	// All values should be zero when metrics are not available
	assert.Equal(t, 0.0, data.CPUUtilization)
	assert.Equal(t, 0.0, data.MemoryUtilization)
	assert.Equal(t, 0.0, data.LogsRateSent)
	assert.Equal(t, 0.0, data.TracesRateSent)
	assert.Equal(t, 0.0, data.MetricsRateSent)
	assert.Equal(t, 0.0, data.DataSentBytes)
	assert.Equal(t, 0.0, data.DataReceivedBytes)
}

// mockMetricsHelperEmpty returns 0 for all metrics
type mockMetricsHelperEmpty struct{}

func (m *mockMetricsHelperEmpty) Fetch(url string) (map[string]*io_prometheus_client.MetricFamily, error) {
	return nil, nil
}

func (m *mockMetricsHelperEmpty) ExtractValue(metrics map[string]*io_prometheus_client.MetricFamily, name string) float64 {
	return 0
}

func (m *mockMetricsHelperEmpty) ExtractValueWithLabels(metrics map[string]*io_prometheus_client.MetricFamily, name string, labels map[string]string) float64 {
	return 0
}
