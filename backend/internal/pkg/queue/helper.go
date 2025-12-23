package queue

import (
	"bufio"
	"net/http"
	"time"

	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

// MetricsHelper defines the interface for Prometheus metrics interactions.
type MetricsHelper interface {
	Fetch(url string) (map[string]*io_prometheus_client.MetricFamily, error)
	ExtractValue(metrics map[string]*io_prometheus_client.MetricFamily, name string) float64
	ExtractValueWithLabels(metrics map[string]*io_prometheus_client.MetricFamily, name string, labels map[string]string) float64
}

// DefaultMetricsHelper is the production implementation of MetricsHelper.
type DefaultMetricsHelper struct{}

func (DefaultMetricsHelper) Fetch(url string) (map[string]*io_prometheus_client.MetricFamily, error) {
	// Create HTTP client with timeout to prevent indefinite hangs
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	parser := expfmt.TextParser{}
	return parser.TextToMetricFamilies(bufio.NewReader(resp.Body))
}

func (DefaultMetricsHelper) ExtractValue(metrics map[string]*io_prometheus_client.MetricFamily, name string) float64 {
	if mf, ok := metrics[name]; ok {
		for _, m := range mf.Metric {
			if m.GetGauge() != nil {
				return m.GetGauge().GetValue()
			} else if m.GetCounter() != nil {
				return m.GetCounter().GetValue()
			}
		}
	}
	return 0
}

// ExtractValueWithLabels extracts a metric value that matches the given labels.
// It searches through all metrics with the given name and returns the value of the first
// metric that has all the specified label key-value pairs.
func (DefaultMetricsHelper) ExtractValueWithLabels(metrics map[string]*io_prometheus_client.MetricFamily, name string, labels map[string]string) float64 {
	if mf, ok := metrics[name]; ok {
		for _, m := range mf.Metric {
			// Check if all required labels match
			if labelsMatch(m.Label, labels) {
				if m.GetGauge() != nil {
					return m.GetGauge().GetValue()
				} else if m.GetCounter() != nil {
					return m.GetCounter().GetValue()
				}
			}
		}
	}
	return 0
}

// labelsMatch checks if a metric's labels contain all the required key-value pairs
func labelsMatch(metricLabels []*io_prometheus_client.LabelPair, required map[string]string) bool {
	for key, value := range required {
		found := false
		for _, label := range metricLabels {
			if label.GetName() == key && label.GetValue() == value {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
