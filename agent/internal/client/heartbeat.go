package client

import (
	"bufio"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/constants"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/pkg/logger"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

// HeartbeatManager manages periodic heartbeat sending to the backend
type HeartbeatManager struct {
	httpClient *http.Client
	interval   time.Duration
	stopChan   chan struct{}
	wg         sync.WaitGroup
}

func NewHeartbeatManager(intervalSec int) *HeartbeatManager {
	return &HeartbeatManager{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		interval:   time.Duration(intervalSec) * time.Second,
		stopChan:   make(chan struct{}),
	}
}

func (h *HeartbeatManager) Start() {
	h.wg.Add(1)
	go h.run()
	logger.Logger.Info("HeartbeatManager started")
}

func (h *HeartbeatManager) Stop() {
	close(h.stopChan)
	h.wg.Wait()
	logger.Logger.Info("HeartbeatManager stopped")
}

func (h *HeartbeatManager) run() {
	defer h.wg.Done()

	// Wait a bit before first heartbeat to allow registration to complete
	time.Sleep(5 * time.Second)

	// Send first heartbeat immediately
	h.sendHeartbeat()

	for {
		// Calculate jitter: ±10% of the interval for each tick
		jitterFactor := (rand.Float64() * 0.2) - 0.1 // range: -0.1 to +0.1
		jitter := time.Duration(float64(h.interval) * jitterFactor)
		nextInterval := h.interval + jitter

		select {
		case <-h.stopChan:
			return
		case <-time.After(nextInterval):
			h.sendHeartbeat()
		}
	}
}

func (h *HeartbeatManager) sendHeartbeat() {
	metrics, err := h.scrapeMetrics()
	if err != nil {
		logger.Logger.Sugar().Warnf("Failed to scrape metrics: %v", err)
		// Send heartbeat with zero metrics to maintain connection status
		metrics = &HeartbeatRequest{}
	}

	if err := SendHeartbeat(h.httpClient, metrics); err != nil {
		logger.Logger.Sugar().Warnf("Failed to send heartbeat: %v", err)
	} else {
		logger.Logger.Sugar().Debugf("Heartbeat sent successfully")
	}
}

func (h *HeartbeatManager) scrapeMetrics() (*HeartbeatRequest, error) {
	sources := getMetricsSources(constants.AGENT_TYPE)

	metrics := &HeartbeatRequest{}

	for _, source := range sources {
		url := fmt.Sprintf("http://127.0.0.1:%d%s", source.Port, source.Path)

		resp, err := h.httpClient.Get(url)
		if err != nil {
			logger.Logger.Sugar().Debugf("Failed to fetch metrics from %s: %v", url, err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			continue
		}

		// Parse Prometheus metrics
		parser := expfmt.TextParser{}
		metricFamilies, err := parser.TextToMetricFamilies(bufio.NewReader(resp.Body))
		resp.Body.Close()

		if err != nil {
			logger.Logger.Sugar().Debugf("Failed to parse metrics from %s: %v", url, err)
			continue
		}

		// Extract metrics based on agent type
		extractMetrics(constants.AGENT_TYPE, metricFamilies, metrics)
	}

	return metrics, nil
}

// MetricsSource defines an endpoint to scrape metrics from
type MetricsSource struct {
	Port int
	Path string
}

func getMetricsSources(agentType string) []MetricsSource {
	switch agentType {
	case "fluent-bit":
		// Prometheus exporter output with node_exporter and fluentbit metrics
		return []MetricsSource{
			{Port: 2021, Path: "/metrics"},
		}
	case "otel", "otel-collector":
		return []MetricsSource{
			{Port: 8888, Path: "/metrics"},
		}
	default:
		return []MetricsSource{
			{Port: 8888, Path: "/metrics"},
		}
	}
}

func extractMetrics(agentType string, families map[string]*io_prometheus_client.MetricFamily, metrics *HeartbeatRequest) {
	switch agentType {
	case "fluent-bit":
		extractFluentBitMetrics(families, metrics)
	default:
		extractOTELMetrics(families, metrics)
	}
}

func extractOTELMetrics(families map[string]*io_prometheus_client.MetricFamily, metrics *HeartbeatRequest) {
	metrics.LogsRateSent = extractValue(families, "otelcol_exporter_sent_log_records")
	metrics.TracesRateSent = extractValue(families, "otelcol_exporter_sent_spans")
	metrics.MetricsRateSent = extractValue(families, "otelcol_exporter_sent_metric_points")
	metrics.DataSentBytes = extractValue(families, "otelcol_exporter_sent_bytes")
	metrics.DataReceivedBytes = extractValue(families, "otelcol_receiver_accepted_bytes")

	cpuUtil := extractValue(families, "system_cpu_utilization")
	if cpuUtil == 0 {
		cpuUtil = extractValue(families, "otelcol_process_cpu_seconds_total")
	}
	metrics.CPUUtilization = cpuUtil

	memUtil := extractValue(families, "system_memory_utilization")
	if memUtil == 0 {
		memUtil = extractValue(families, "otelcol_process_memory_rss")
	}
	metrics.MemoryUtilization = memUtil
}

func extractFluentBitMetrics(families map[string]*io_prometheus_client.MetricFamily, metrics *HeartbeatRequest) {
	// Extract CPU utilization from node_exporter metrics
	// Calculate as percentage of non-idle CPU time
	metrics.CPUUtilization = extractNodeCPUUtilization(families)

	// Extract memory utilization from node_exporter metrics
	// Calculate as (MemTotal - MemAvailable) / MemTotal * 100
	metrics.MemoryUtilization = extractNodeMemoryUtilization(families)

	// Fluent Bit telemetry metrics from fluentbit_metrics input
	metrics.LogsRateSent = extractValue(families, "fluentbit_output_proc_records_total")
	metrics.TracesRateSent = 0 // Fluent Bit is primarily for logs
	metrics.MetricsRateSent = 0
	metrics.DataSentBytes = extractValue(families, "fluentbit_output_proc_bytes_total")
	metrics.DataReceivedBytes = extractValue(families, "fluentbit_input_bytes_total")
}

// extractNodeCPUUtilization calculates CPU utilization from node_cpu_seconds_total
// Uses load average as a simpler proxy for CPU utilization
func extractNodeCPUUtilization(families map[string]*io_prometheus_client.MetricFamily) float64 {
	// Use 1-minute load average as CPU utilization proxy
	load1 := extractValue(families, "node_load1")
	if load1 > 0 {
		return load1
	}

	// Fallback: sum all non-idle CPU seconds as raw counter
	// This provides the total CPU seconds used (can be used for rate calculation)
	family, ok := families["node_cpu_seconds_total"]
	if !ok || family == nil {
		return 0
	}

	var totalNonIdle float64
	for _, m := range family.Metric {
		mode := getLabelValue(m.Label, "mode")
		// Count user, system, nice, softirq, irq, steal modes (exclude idle and iowait)
		if mode != "idle" && mode != "iowait" {
			if m.GetCounter() != nil {
				totalNonIdle += m.GetCounter().GetValue()
			}
		}
	}
	return totalNonIdle
}

// extractNodeMemoryUtilization calculates memory utilization percentage
// Formula: (MemTotal - MemAvailable) / MemTotal * 100
func extractNodeMemoryUtilization(families map[string]*io_prometheus_client.MetricFamily) float64 {
	memTotal := extractValue(families, "node_memory_MemTotal_bytes")
	memAvailable := extractValue(families, "node_memory_MemAvailable_bytes")

	if memTotal > 0 {
		usedPercent := ((memTotal - memAvailable) / memTotal) * 100
		return usedPercent
	}

	// Fallback to raw memory used bytes
	memFree := extractValue(families, "node_memory_MemFree_bytes")
	if memTotal > 0 && memFree > 0 {
		return memTotal - memFree
	}

	return 0
}

// getLabelValue extracts a label value from metric labels
func getLabelValue(labels []*io_prometheus_client.LabelPair, name string) string {
	for _, lp := range labels {
		if lp.GetName() == name {
			return lp.GetValue()
		}
	}
	return ""
}

func extractValue(families map[string]*io_prometheus_client.MetricFamily, name string) float64 {
	family, ok := families[name]
	if !ok || family == nil {
		return 0
	}

	var total float64
	for _, m := range family.Metric {
		if m.GetCounter() != nil {
			total += m.GetCounter().GetValue()
		} else if m.GetGauge() != nil {
			total += m.GetGauge().GetValue()
		} else if m.GetUntyped() != nil {
			total += m.GetUntyped().GetValue()
		}
	}
	return total
}

func extractValueWithLabels(families map[string]*io_prometheus_client.MetricFamily, name string, labels map[string]string) float64 {
	family, ok := families[name]
	if !ok || family == nil {
		return 0
	}

	for _, m := range family.Metric {
		if matchLabels(m.Label, labels) {
			if m.GetCounter() != nil {
				return m.GetCounter().GetValue()
			} else if m.GetGauge() != nil {
				return m.GetGauge().GetValue()
			} else if m.GetUntyped() != nil {
				return m.GetUntyped().GetValue()
			}
		}
	}
	return 0
}

func matchLabels(metricLabels []*io_prometheus_client.LabelPair, targetLabels map[string]string) bool {
	labelMap := make(map[string]string)
	for _, lp := range metricLabels {
		labelMap[lp.GetName()] = lp.GetValue()
	}

	for k, v := range targetLabels {
		if labelMap[k] != v {
			return false
		}
	}
	return true
}
