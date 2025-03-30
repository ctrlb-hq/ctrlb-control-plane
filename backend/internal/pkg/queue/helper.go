package queue

import (
	"bufio"
	"net/http"

	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

func fetchMetrics(url string) (map[string]*io_prometheus_client.MetricFamily, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	parser := expfmt.TextParser{}
	return parser.TextToMetricFamilies(bufio.NewReader(resp.Body))
}

func extractMetricValue(metrics map[string]*io_prometheus_client.MetricFamily, metricName string) float64 {
	if mf, ok := metrics[metricName]; ok {
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
