package managed

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
)

type counterSample struct {
	value float64
	time  time.Time
}

type MetricsSampler struct {
	mu       sync.Mutex
	previous map[string]counterSample
	uptime   float64
	client   *http.Client
	url      string
}

func (s *MetricsSampler) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.previous, s.uptime = nil, 0
}

func (s *MetricsSampler) Sample(ctx context.Context) (map[string]float64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Collector metrics returned HTTP %d", resp.StatusCode)
	}
	const limit = 4 << 20
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if len(data) > limit {
		return nil, fmt.Errorf("Collector metrics exceed size limit")
	}
	parser := expfmt.NewTextParser(model.LegacyValidation)
	families, err := parser.TextToMetricFamilies(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return s.normalize(families, time.Now()), nil
}

func (s *MetricsSampler) normalize(families map[string]*dto.MetricFamily, now time.Time) map[string]float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := map[string]float64{}
	next := map[string]counterSample{}
	if family := families["otelcol_process_uptime"]; family != nil {
		if uptime, ok := sumValues(family); ok {
			if uptime < s.uptime {
				s.previous = nil
			}
			s.uptime = uptime
		}
	}
	for _, metric := range []struct {
		name, output string
		scale        float64
	}{
		{"otelcol_exporter_sent_log_records", "logs_sent_per_second", 1},
		{"otelcol_exporter_sent_metric_points", "metrics_sent_per_second", 1},
		{"otelcol_exporter_sent_spans", "spans_sent_per_second", 1},
		{"otelcol_process_cpu_seconds", "cpu_percent", 100},
	} {
		family := families[metric.name]
		if family == nil {
			family = families[metric.name+"_total"]
		}
		if family == nil {
			continue
		}
		var rate float64
		available := false
		for _, series := range family.Metric {
			value, ok := sampleValue(series)
			if !ok {
				continue
			}
			key := metric.name + labelKey(series)
			next[key] = counterSample{value, now}
			previous, exists := s.previous[key]
			if !exists || !now.After(previous.time) {
				continue
			}
			delta := value - previous.value
			if delta < 0 {
				delta = value
			}
			rate += delta / now.Sub(previous.time).Seconds()
			available = true
		}
		if available && finite(rate*metric.scale) {
			result[metric.output] = rate * metric.scale
		}
	}
	for _, name := range []string{"otelcol_process_memory_rss", "otelcol_process_memory_rss_bytes"} {
		if value, ok := sumValues(families[name]); ok {
			result["memory_bytes"] = value
			break
		}
	}
	s.previous = next
	return result
}

func sampleValue(metric *dto.Metric) (float64, bool) {
	var value float64
	if metric.Counter != nil {
		value = metric.Counter.GetValue()
	} else if metric.Gauge != nil {
		value = metric.Gauge.GetValue()
	} else {
		return 0, false
	}
	return value, finite(value)
}

func finite(value float64) bool { return value >= 0 && !math.IsNaN(value) && !math.IsInf(value, 0) }

func sumValues(family *dto.MetricFamily) (float64, bool) {
	if family == nil {
		return 0, false
	}
	var total float64
	ok := false
	for _, metric := range family.Metric {
		if value, valid := sampleValue(metric); valid {
			total += value
			ok = true
		}
	}
	return total, ok && finite(total)
}

func labelKey(metric *dto.Metric) string {
	labels := make([]string, 0, len(metric.Label))
	for _, label := range metric.Label {
		labels = append(labels, fmt.Sprintf("%q=%q", label.GetName(), label.GetValue()))
	}
	sort.Strings(labels)
	return "{" + strings.Join(labels, ",") + "}"
}
