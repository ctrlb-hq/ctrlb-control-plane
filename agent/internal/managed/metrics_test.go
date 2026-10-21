package managed

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
)

func TestMetricsRatesSumSeriesAndHandleIndependentReset(t *testing.T) {
	sampler := &MetricsSampler{}
	t0 := time.Unix(1000, 0)
	parse := func(logA, logB, cpu, memory, uptime float64, alias string, now time.Time) map[string]float64 {
		parser := expfmt.NewTextParser(model.LegacyValidation)
		families, err := parser.TextToMetricFamilies(strings.NewReader(fmt.Sprintf(`
# TYPE otelcol_exporter_sent_log_records%s counter
otelcol_exporter_sent_log_records%s{exporter="a"} %g
otelcol_exporter_sent_log_records%s{exporter="b"} %g
# TYPE otelcol_exporter_sent_metric_points counter
otelcol_exporter_sent_metric_points{exporter="a"} %g
# TYPE otelcol_exporter_sent_spans counter
otelcol_exporter_sent_spans{exporter="a"} 0
# TYPE otelcol_process_cpu_seconds_total counter
otelcol_process_cpu_seconds_total %g
# TYPE otelcol_process_memory_rss gauge
otelcol_process_memory_rss %g
# TYPE otelcol_process_uptime gauge
otelcol_process_uptime %g
`, alias, alias, logA, alias, logB, logA, cpu, memory, uptime)))
		require.NoError(t, err)
		return sampler.normalize(families, now)
	}
	first := parse(100, 50, 10, 1024, 100, "", t0)
	require.Equal(t, map[string]float64{"memory_bytes": 1024}, first, "rates need two samples")
	second := parse(120, 5, 12, 2048, 110, "_total", t0.Add(10*time.Second))
	require.Equal(t, 2.5, second["logs_sent_per_second"], "sum independent exporter deltas, including reset to 5")
	require.Equal(t, 2.0, second["metrics_sent_per_second"])
	require.Equal(t, 0.0, second["spans_sent_per_second"])
	require.Equal(t, 20.0, second["cpu_percent"])
	require.Equal(t, 2048.0, second["memory_bytes"])
	restarted := parse(200, 100, 1, 1000, 1, "_total", t0.Add(20*time.Second))
	require.Equal(t, map[string]float64{"memory_bytes": 1000}, restarted, "uptime reset invalidates previous rate baselines")
}

func TestMetricsMissingAndInvalidValuesAreOmitted(t *testing.T) {
	parser := expfmt.NewTextParser(model.LegacyValidation)
	families, err := parser.TextToMetricFamilies(strings.NewReader(`# TYPE otelcol_process_memory_rss gauge
otelcol_process_memory_rss NaN
# TYPE otelcol_process_cpu_seconds counter
otelcol_process_cpu_seconds -1
`))
	require.NoError(t, err)
	sampler := &MetricsSampler{}
	require.Empty(t, sampler.normalize(families, time.Now()))
}
