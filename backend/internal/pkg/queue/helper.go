package queue

import (
	"bufio"
	"net/http"

	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

// MetricsHelper defines the interface for Prometheus metrics interactions.
type MetricsHelper interface {
	Fetch(url string) (map[string]*io_prometheus_client.MetricFamily, error)
	ExtractValue(metrics map[string]*io_prometheus_client.MetricFamily, name string) float64
}

// DefaultMetricsHelper is the production implementation of MetricsHelper.
type DefaultMetricsHelper struct{}

func (DefaultMetricsHelper) Fetch(url string) (map[string]*io_prometheus_client.MetricFamily, error) {
	resp, err := http.Get(url)
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
