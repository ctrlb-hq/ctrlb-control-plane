package queue_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/pkg/queue"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

func TestDefaultMetricsHelper_Fetch_Success(t *testing.T) {
	// Simulate Prometheus endpoint
	metricsText := `
# TYPE otelcol_exporter_sent_log_records counter
otelcol_exporter_sent_log_records 42
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(metricsText))
	}))
	defer server.Close()

	helper := queue.DefaultMetricsHelper{}
	metrics, err := helper.Fetch(server.URL)
	assert.NoError(t, err)
	assert.Contains(t, metrics, "otelcol_exporter_sent_log_records")
}

func TestDefaultMetricsHelper_Fetch_HTTPFailure(t *testing.T) {
	helper := queue.DefaultMetricsHelper{}
	_, err := helper.Fetch("http://localhost:9999/invalid")
	assert.Error(t, err)
}

func TestDefaultMetricsHelper_ExtractValue_Counter(t *testing.T) {
	counterVal := 100.5
	metricFamily := map[string]*io_prometheus_client.MetricFamily{
		"custom_counter": {
			Metric: []*io_prometheus_client.Metric{
				{
					Counter: &io_prometheus_client.Counter{
						Value: &counterVal,
					},
				},
			},
		},
	}

	helper := queue.DefaultMetricsHelper{}
	value := helper.ExtractValue(metricFamily, "custom_counter")
	assert.Equal(t, 100.5, value)
}

func TestDefaultMetricsHelper_ExtractValue_Gauge(t *testing.T) {
	gaugeVal := 55.75
	metricFamily := map[string]*io_prometheus_client.MetricFamily{
		"custom_gauge": {
			Metric: []*io_prometheus_client.Metric{
				{
					Gauge: &io_prometheus_client.Gauge{
						Value: &gaugeVal,
					},
				},
			},
		},
	}

	helper := queue.DefaultMetricsHelper{}
	value := helper.ExtractValue(metricFamily, "custom_gauge")
	assert.Equal(t, 55.75, value)
}

func TestDefaultMetricsHelper_ExtractValue_MissingMetric(t *testing.T) {
	metricFamily := map[string]*io_prometheus_client.MetricFamily{}

	helper := queue.DefaultMetricsHelper{}
	value := helper.ExtractValue(metricFamily, "non_existent")
	assert.Equal(t, 0.0, value)
}

func TestDefaultMetricsHelper_ExtractValue_NoGaugeOrCounter(t *testing.T) {
	metricFamily := map[string]*io_prometheus_client.MetricFamily{
		"weird_metric": {
			Metric: []*io_prometheus_client.Metric{
				{}, // no gauge or counter set
			},
		},
	}

	helper := queue.DefaultMetricsHelper{}
	value := helper.ExtractValue(metricFamily, "weird_metric")
	assert.Equal(t, 0.0, value)
}

func TestDefaultMetricsHelper_ExtractValueWithLabels_MatchingLabels(t *testing.T) {
	cpuVal := 42.5
	nameLabel := "fluent-bit"
	modeLabel := "user"

	metricFamily := map[string]*io_prometheus_client.MetricFamily{
		"process_cpu_seconds_total": {
			Metric: []*io_prometheus_client.Metric{
				{
					Label: []*io_prometheus_client.LabelPair{
						{Name: &[]string{"name"}[0], Value: &nameLabel},
						{Name: &[]string{"mode"}[0], Value: &modeLabel},
					},
					Counter: &io_prometheus_client.Counter{
						Value: &cpuVal,
					},
				},
			},
		},
	}

	helper := queue.DefaultMetricsHelper{}
	value := helper.ExtractValueWithLabels(metricFamily, "process_cpu_seconds_total", map[string]string{
		"name": "fluent-bit",
		"mode": "user",
	})
	assert.Equal(t, 42.5, value)
}

func TestDefaultMetricsHelper_ExtractValueWithLabels_NoMatch(t *testing.T) {
	cpuVal := 42.5
	nameLabel := "fluent-bit"
	modeLabel := "user"

	metricFamily := map[string]*io_prometheus_client.MetricFamily{
		"process_cpu_seconds_total": {
			Metric: []*io_prometheus_client.Metric{
				{
					Label: []*io_prometheus_client.LabelPair{
						{Name: &[]string{"name"}[0], Value: &nameLabel},
						{Name: &[]string{"mode"}[0], Value: &modeLabel},
					},
					Counter: &io_prometheus_client.Counter{
						Value: &cpuVal,
					},
				},
			},
		},
	}

	helper := queue.DefaultMetricsHelper{}
	value := helper.ExtractValueWithLabels(metricFamily, "process_cpu_seconds_total", map[string]string{
		"name": "otel-collector",
		"mode": "user",
	})
	assert.Equal(t, 0.0, value)
}

func TestDefaultMetricsHelper_ExtractValueWithLabels_MultipleMetrics(t *testing.T) {
	cpuVal1 := 10.0
	cpuVal2 := 20.0
	cpuVal3 := 30.0
	nameLabel1 := "fluent-bit"
	nameLabel2 := "fluent-bit"
	nameLabel3 := "other-process"
	modeLabel1 := "user"
	modeLabel2 := "system"
	modeLabel3 := "user"

	metricFamily := map[string]*io_prometheus_client.MetricFamily{
		"process_cpu_seconds_total": {
			Metric: []*io_prometheus_client.Metric{
				{
					Label: []*io_prometheus_client.LabelPair{
						{Name: &[]string{"name"}[0], Value: &nameLabel1},
						{Name: &[]string{"mode"}[0], Value: &modeLabel1},
					},
					Counter: &io_prometheus_client.Counter{
						Value: &cpuVal1,
					},
				},
				{
					Label: []*io_prometheus_client.LabelPair{
						{Name: &[]string{"name"}[0], Value: &nameLabel2},
						{Name: &[]string{"mode"}[0], Value: &modeLabel2},
					},
					Counter: &io_prometheus_client.Counter{
						Value: &cpuVal2,
					},
				},
				{
					Label: []*io_prometheus_client.LabelPair{
						{Name: &[]string{"name"}[0], Value: &nameLabel3},
						{Name: &[]string{"mode"}[0], Value: &modeLabel3},
					},
					Counter: &io_prometheus_client.Counter{
						Value: &cpuVal3,
					},
				},
			},
		},
	}

	helper := queue.DefaultMetricsHelper{}

	// Should get the first matching metric (user mode)
	value := helper.ExtractValueWithLabels(metricFamily, "process_cpu_seconds_total", map[string]string{
		"name": "fluent-bit",
		"mode": "user",
	})
	assert.Equal(t, 10.0, value)

	// Should get the system mode metric
	value = helper.ExtractValueWithLabels(metricFamily, "process_cpu_seconds_total", map[string]string{
		"name": "fluent-bit",
		"mode": "system",
	})
	assert.Equal(t, 20.0, value)
}

func TestDefaultMetricsHelper_ExtractValueWithLabels_MissingMetric(t *testing.T) {
	metricFamily := map[string]*io_prometheus_client.MetricFamily{}

	helper := queue.DefaultMetricsHelper{}
	value := helper.ExtractValueWithLabels(metricFamily, "non_existent", map[string]string{
		"name": "fluent-bit",
	})
	assert.Equal(t, 0.0, value)
}
