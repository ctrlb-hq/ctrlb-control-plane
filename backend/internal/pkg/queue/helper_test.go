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
