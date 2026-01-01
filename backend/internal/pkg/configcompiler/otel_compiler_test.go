package configcompiler

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// OTEL Compiler Tests
// =============================================================================

func TestCompileGraphToOTEL_Success(t *testing.T) {
	graph := createSampleGraph()

	result, err := CompileGraphToOTEL(graph)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, *result, "receivers")
	assert.Contains(t, *result, "processors")
	assert.Contains(t, *result, "exporters")
	assert.Contains(t, *result, "service")

	service := (*result)["service"].(map[string]any)
	assert.Contains(t, service, "pipelines")
	assert.Contains(t, service, "telemetry")

	// Verify pipelines were created for both signals (metrics and logs)
	pipelines := service["pipelines"].(map[string]any)
	assert.GreaterOrEqual(t, len(pipelines), 1, "Should have at least 1 pipeline")
}

func TestCompileGraphToOTEL_MultiPathGraph(t *testing.T) {
	graph := createSampleGraph2()

	result, err := CompileGraphToOTEL(graph)

	b, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(b))

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, *result, "receivers")
	assert.Contains(t, *result, "processors")
	assert.Contains(t, *result, "exporters")
	assert.Contains(t, *result, "service")

	service := (*result)["service"].(map[string]any)
	assert.Contains(t, service, "pipelines")
	assert.Contains(t, service, "telemetry")

	// The graph has two separate paths that share an exporter:
	// Path 1: R1 -> P1 -> E1 (traces)
	// Path 2: R2 -> P2 -> E1 (logs)
	// These should create separate pipelines
	pipelines := service["pipelines"].(map[string]any)
	assert.GreaterOrEqual(t, len(pipelines), 2, "Should have at least 2 pipelines for different paths")
}

func TestCompileGraphToOTEL_EmptyGraph(t *testing.T) {
	graph := models.PipelineGraph{}
	result, err := CompileGraphToOTEL(graph)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "empty pipeline graph")
}

func TestCompileGraphToOTEL_NoReceivers(t *testing.T) {
	graph := models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{
				ComponentID:      1,
				Name:             "exporter",
				ComponentName:    "otlp_exporter",
				ComponentRole:    "exporter",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{},
			},
		},
		Edges: []models.PipelineEdges{},
	}

	result, err := CompileGraphToOTEL(graph)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "at least one receiver")
}

func TestCompileGraphToOTEL_NoExporters(t *testing.T) {
	graph := models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{
				ComponentID:      1,
				Name:             "receiver",
				ComponentName:    "otlp_receiver",
				ComponentRole:    "receiver",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{},
			},
		},
		Edges: []models.PipelineEdges{},
	}

	result, err := CompileGraphToOTEL(graph)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "at least one exporter")
}

func TestCompileGraphToOTEL_WithCycle(t *testing.T) {
	graph := models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{
				ComponentID:      1,
				Name:             "receiver",
				ComponentName:    "otlp_receiver",
				ComponentRole:    "receiver",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{},
			},
			{
				ComponentID:      2,
				Name:             "processor",
				ComponentName:    "batch_processor",
				ComponentRole:    "processor",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{},
			},
			{
				ComponentID:      3,
				Name:             "exporter",
				ComponentName:    "otlp_exporter",
				ComponentRole:    "exporter",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{},
			},
		},
		Edges: []models.PipelineEdges{
			{Source: "1", Target: "2"},
			{Source: "2", Target: "3"},
			{Source: "3", Target: "1"}, // Cycle
		},
	}

	result, err := CompileGraphToOTEL(graph)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "cycle detected")
}

func TestCompileGraphToOTEL_FanIn(t *testing.T) {
	// Two receivers feeding into one processor -> one exporter
	graph := models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{
				ComponentID:      1,
				Name:             "receiver_one",
				ComponentName:    "otlp_receiver",
				ComponentRole:    "receiver",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{"endpoint": "0.0.0.0:4317"},
			},
			{
				ComponentID:      2,
				Name:             "receiver_two",
				ComponentName:    "otlp_receiver",
				ComponentRole:    "receiver",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{"endpoint": "0.0.0.0:4318"},
			},
			{
				ComponentID:      3,
				Name:             "processor",
				ComponentName:    "batch_processor",
				ComponentRole:    "processor",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{},
			},
			{
				ComponentID:      4,
				Name:             "exporter",
				ComponentName:    "otlp_exporter",
				ComponentRole:    "exporter",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{},
			},
		},
		Edges: []models.PipelineEdges{
			{Source: "1", Target: "3"},
			{Source: "2", Target: "3"},
			{Source: "3", Target: "4"},
		},
	}

	result, err := CompileGraphToOTEL(graph)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	service := (*result)["service"].(map[string]any)
	pipelines := service["pipelines"].(map[string]any)

	// Should have one pipeline with multiple receivers
	assert.GreaterOrEqual(t, len(pipelines), 1)

	// Find a logs pipeline and verify it has 2 receivers
	for name, p := range pipelines {
		if name[:4] == "logs" {
			pipeline := p.(map[string]any)
			receivers := pipeline["receivers"].([]string)
			assert.Len(t, receivers, 2, "Fan-in pipeline should have 2 receivers")
		}
	}
}

func TestCompileGraphToOTEL_FanOut(t *testing.T) {
	// One receiver -> one processor -> two exporters
	graph := models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{
				ComponentID:      1,
				Name:             "receiver",
				ComponentName:    "otlp_receiver",
				ComponentRole:    "receiver",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{},
			},
			{
				ComponentID:      2,
				Name:             "processor",
				ComponentName:    "batch_processor",
				ComponentRole:    "processor",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{},
			},
			{
				ComponentID:      3,
				Name:             "exporter_one",
				ComponentName:    "otlp_exporter",
				ComponentRole:    "exporter",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{"endpoint": "backend1:4317"},
			},
			{
				ComponentID:      4,
				Name:             "exporter_two",
				ComponentName:    "otlp_exporter",
				ComponentRole:    "exporter",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{"endpoint": "backend2:4317"},
			},
		},
		Edges: []models.PipelineEdges{
			{Source: "1", Target: "2"},
			{Source: "2", Target: "3"},
			{Source: "2", Target: "4"},
		},
	}

	result, err := CompileGraphToOTEL(graph)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	service := (*result)["service"].(map[string]any)
	pipelines := service["pipelines"].(map[string]any)
	assert.GreaterOrEqual(t, len(pipelines), 1)

	// Find a logs pipeline and verify it has 2 exporters
	for name, p := range pipelines {
		if name[:4] == "logs" {
			pipeline := p.(map[string]any)
			exporters := pipeline["exporters"].([]string)
			assert.Len(t, exporters, 2, "Fan-out pipeline should have 2 exporters")
		}
	}
}

func TestCompileGraphToOTEL_DisconnectedComponents(t *testing.T) {
	// Two completely separate pipelines
	graph := models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			// Pipeline 1
			{
				ComponentID:      1,
				Name:             "receiver_one",
				ComponentName:    "otlp_receiver",
				ComponentRole:    "receiver",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{},
			},
			{
				ComponentID:      2,
				Name:             "exporter_one",
				ComponentName:    "otlp_exporter",
				ComponentRole:    "exporter",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{},
			},
			// Pipeline 2
			{
				ComponentID:      3,
				Name:             "receiver_two",
				ComponentName:    "otlp_receiver",
				ComponentRole:    "receiver",
				SupportedSignals: []string{"metrics"},
				Config:           map[string]any{},
			},
			{
				ComponentID:      4,
				Name:             "exporter_two",
				ComponentName:    "otlp_exporter",
				ComponentRole:    "exporter",
				SupportedSignals: []string{"metrics"},
				Config:           map[string]any{},
			},
		},
		Edges: []models.PipelineEdges{
			{Source: "1", Target: "2"},
			{Source: "3", Target: "4"},
		},
	}

	result, err := CompileGraphToOTEL(graph)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	service := (*result)["service"].(map[string]any)
	pipelines := service["pipelines"].(map[string]any)

	// Should have 2 separate pipelines
	assert.GreaterOrEqual(t, len(pipelines), 2, "Should have at least 2 pipelines for disconnected components")
}

// =============================================================================
// COMPREHENSIVE OTEL CONFIG VALIDATION TESTS
// =============================================================================

func TestOTELLogsConfigValidation(t *testing.T) {
	graph := createRealisticOTELLogsPipeline()

	result, err := CompileGraphToOTEL(graph)

	b, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("Generated OTEL Logs Config:\n%s\n", string(b))

	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Validate top-level structure
	config := *result
	assert.Contains(t, config, "receivers", "Config must have receivers section")
	assert.Contains(t, config, "processors", "Config must have processors section")
	assert.Contains(t, config, "exporters", "Config must have exporters section")
	assert.Contains(t, config, "service", "Config must have service section")

	// Validate receivers section
	receivers := config["receivers"].(map[string]any)
	assert.Len(t, receivers, 1, "Should have exactly 1 receiver")

	// Find the filelog receiver and validate its config
	var filelogReceiverConfig map[string]any
	for alias, cfg := range receivers {
		assert.Contains(t, alias, "filelog", "Receiver alias should contain component type")
		filelogReceiverConfig = cfg.(map[string]any)
		break
	}
	assert.NotNil(t, filelogReceiverConfig, "Filelog receiver config should exist")
	assert.Contains(t, filelogReceiverConfig, "include", "Filelog receiver must have include config")
	assert.Contains(t, filelogReceiverConfig, "start_at", "Filelog receiver must have start_at config")
	assert.Contains(t, filelogReceiverConfig, "operators", "Filelog receiver must have operators config")

	// Validate processors section
	processors := config["processors"].(map[string]any)
	assert.Len(t, processors, 2, "Should have exactly 2 processors (batch + memory_limiter)")

	// Check for batch processor config
	hasBatchProcessor := false
	hasMemoryLimiter := false
	for alias, cfg := range processors {
		procCfg := cfg.(map[string]any)
		if _, ok := procCfg["send_batch_size"]; ok {
			hasBatchProcessor = true
			assert.Contains(t, alias, "batch", "Batch processor alias should contain 'batch'")
			assert.Equal(t, 8192, procCfg["send_batch_size"], "Batch size should match config")
			assert.Equal(t, "200ms", procCfg["timeout"], "Batch timeout should match config")
		}
		if _, ok := procCfg["limit_mib"]; ok {
			hasMemoryLimiter = true
			assert.Contains(t, alias, "memorylimiter", "Memory limiter alias should contain 'memorylimiter'")
			assert.Equal(t, 400, procCfg["limit_mib"], "Memory limit should match config")
		}
	}
	assert.True(t, hasBatchProcessor, "Should have batch processor")
	assert.True(t, hasMemoryLimiter, "Should have memory limiter processor")

	// Validate exporters section
	exporters := config["exporters"].(map[string]any)
	assert.Len(t, exporters, 1, "Should have exactly 1 exporter")

	// Check exporter config
	for alias, cfg := range exporters {
		assert.Contains(t, alias, "otlp", "Exporter alias should contain 'otlp'")
		exporterCfg := cfg.(map[string]any)
		assert.Contains(t, exporterCfg, "endpoint", "Exporter must have endpoint")
		assert.Contains(t, exporterCfg, "tls", "Exporter must have TLS config")
		assert.Contains(t, exporterCfg, "retry_on_failure", "Exporter must have retry config")
	}

	// Validate service section
	service := config["service"].(map[string]any)
	assert.Contains(t, service, "pipelines", "Service must have pipelines")
	assert.Contains(t, service, "telemetry", "Service must have telemetry")

	// Validate pipelines
	pipelines := service["pipelines"].(map[string]any)
	assert.GreaterOrEqual(t, len(pipelines), 1, "Should have at least 1 pipeline")

	// Find logs pipeline and validate structure
	foundLogsPipeline := false
	for pipelineName, pipelineDef := range pipelines {
		if len(pipelineName) >= 4 && pipelineName[:4] == "logs" {
			foundLogsPipeline = true
			pdef := pipelineDef.(map[string]any)

			// Validate pipeline has receivers
			pipelineReceivers := pdef["receivers"].([]string)
			assert.Len(t, pipelineReceivers, 1, "Logs pipeline should have 1 receiver")

			// Validate pipeline has processors
			pipelineProcessors := pdef["processors"].([]string)
			assert.Len(t, pipelineProcessors, 2, "Logs pipeline should have 2 processors")

			// Validate pipeline has exporters
			pipelineExporters := pdef["exporters"].([]string)
			assert.Len(t, pipelineExporters, 1, "Logs pipeline should have 1 exporter")

			// Verify references exist in component sections
			for _, rcv := range pipelineReceivers {
				assert.Contains(t, receivers, rcv, "Pipeline receiver must exist in receivers section: "+rcv)
			}
			for _, proc := range pipelineProcessors {
				assert.Contains(t, processors, proc, "Pipeline processor must exist in processors section: "+proc)
			}
			for _, exp := range pipelineExporters {
				assert.Contains(t, exporters, exp, "Pipeline exporter must exist in exporters section: "+exp)
			}
		}
	}
	assert.True(t, foundLogsPipeline, "Should have a logs pipeline")
}

func TestOTELMetricsConfigValidation(t *testing.T) {
	graph := createRealisticOTELMetricsPipeline()

	result, err := CompileGraphToOTEL(graph)

	b, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("Generated OTEL Metrics Config:\n%s\n", string(b))

	assert.NoError(t, err)
	assert.NotNil(t, result)

	config := *result

	// Validate receivers
	receivers := config["receivers"].(map[string]any)
	assert.Len(t, receivers, 1, "Should have 1 prometheus receiver")

	// Validate prometheus receiver config
	for alias, cfg := range receivers {
		assert.Contains(t, alias, "prometheus", "Should have prometheus receiver")
		promCfg := cfg.(map[string]any)
		assert.Contains(t, promCfg, "config", "Prometheus receiver must have config section")
		configSection := promCfg["config"].(map[string]any)
		assert.Contains(t, configSection, "scrape_configs", "Prometheus config must have scrape_configs")
	}

	// Validate processors
	processors := config["processors"].(map[string]any)
	assert.Len(t, processors, 2, "Should have 2 processors (filter + batch)")

	hasFilterProcessor := false
	for alias, cfg := range processors {
		procCfg := cfg.(map[string]any)
		if _, ok := procCfg["metrics"]; ok {
			hasFilterProcessor = true
			assert.Contains(t, alias, "filter", "Filter processor alias should contain 'filter'")
			metricsCfg := procCfg["metrics"].(map[string]any)
			assert.Contains(t, metricsCfg, "include", "Filter must have include config")
		}
	}
	assert.True(t, hasFilterProcessor, "Should have filter processor")

	// Validate exporters
	exporters := config["exporters"].(map[string]any)
	assert.Len(t, exporters, 1, "Should have 1 exporter")

	for alias, cfg := range exporters {
		assert.Contains(t, alias, "prometheusremotewrite", "Should have prometheus remote write exporter")
		expCfg := cfg.(map[string]any)
		assert.Contains(t, expCfg, "endpoint", "Exporter must have endpoint")
		assert.Contains(t, expCfg, "tls", "Exporter must have TLS config")
		assert.Contains(t, expCfg, "headers", "Exporter must have headers")
	}

	// Validate service pipelines
	service := config["service"].(map[string]any)
	pipelines := service["pipelines"].(map[string]any)

	// Should only have metrics pipeline since all nodes support metrics
	foundMetricsPipeline := false
	for pipelineName := range pipelines {
		if len(pipelineName) >= 7 && pipelineName[:7] == "metrics" {
			foundMetricsPipeline = true
		}
	}
	assert.True(t, foundMetricsPipeline, "Should have a metrics pipeline")
}

func TestOTELTracesConfigValidation(t *testing.T) {
	graph := createRealisticOTELTracesPipeline()

	result, err := CompileGraphToOTEL(graph)

	b, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("Generated OTEL Traces Config:\n%s\n", string(b))

	assert.NoError(t, err)
	assert.NotNil(t, result)

	config := *result

	// Validate OTLP receiver config
	receivers := config["receivers"].(map[string]any)
	assert.Len(t, receivers, 1, "Should have 1 receiver")

	for alias, cfg := range receivers {
		assert.Contains(t, alias, "otlp", "Should have OTLP receiver")
		rcvCfg := cfg.(map[string]any)
		assert.Contains(t, rcvCfg, "protocols", "OTLP receiver must have protocols config")
		protocols := rcvCfg["protocols"].(map[string]any)
		assert.Contains(t, protocols, "grpc", "OTLP receiver must have gRPC protocol")
		assert.Contains(t, protocols, "http", "OTLP receiver must have HTTP protocol")
	}

	// Validate sampler processor
	processors := config["processors"].(map[string]any)
	hasSampler := false
	for alias, cfg := range processors {
		procCfg := cfg.(map[string]any)
		if _, ok := procCfg["sampling_percentage"]; ok {
			hasSampler = true
			assert.Contains(t, alias, "probabilisticsampler", "Should have probabilistic sampler")
			assert.Equal(t, 10.0, procCfg["sampling_percentage"], "Sampling percentage should be 10%")
		}
	}
	assert.True(t, hasSampler, "Should have sampler processor")

	// Validate Jaeger exporter
	exporters := config["exporters"].(map[string]any)
	assert.Len(t, exporters, 1, "Should have 1 exporter")

	for alias, cfg := range exporters {
		assert.Contains(t, alias, "jaeger", "Should have Jaeger exporter")
		expCfg := cfg.(map[string]any)
		assert.Contains(t, expCfg, "endpoint", "Jaeger exporter must have endpoint")
		assert.Equal(t, "jaeger-collector.tracing:14250", expCfg["endpoint"], "Jaeger endpoint should match")
	}

	// Validate traces pipeline
	service := config["service"].(map[string]any)
	pipelines := service["pipelines"].(map[string]any)

	foundTracesPipeline := false
	for pipelineName, pipelineDef := range pipelines {
		if len(pipelineName) >= 6 && pipelineName[:6] == "traces" {
			foundTracesPipeline = true
			pdef := pipelineDef.(map[string]any)
			processors := pdef["processors"].([]string)
			assert.Len(t, processors, 2, "Traces pipeline should have 2 processors")
		}
	}
	assert.True(t, foundTracesPipeline, "Should have a traces pipeline")
}

func TestOTELMultiSignalConfigValidation(t *testing.T) {
	graph := createMultiSignalOTELPipeline()

	result, err := CompileGraphToOTEL(graph)

	b, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("Generated OTEL Multi-Signal Config:\n%s\n", string(b))

	assert.NoError(t, err)
	assert.NotNil(t, result)
	if result == nil {
		t.Fatal("Result is nil, cannot continue test")
	}

	config := *result

	// Should have 1 receiver
	receivers := config["receivers"].(map[string]any)
	assert.Len(t, receivers, 1, "Should have 1 OTLP receiver")

	// Should have 2 processors (resource + batch)
	processors := config["processors"].(map[string]any)
	assert.Len(t, processors, 2, "Should have 2 processors (resource + batch)")

	hasResourceProcessor := false
	hasBatchProcessor := false
	for alias, cfg := range processors {
		procCfg := cfg.(map[string]any)
		if _, ok := procCfg["attributes"]; ok {
			hasResourceProcessor = true
			assert.Contains(t, alias, "resource", "Should have resource processor")
			attrs := procCfg["attributes"].([]map[string]any)
			assert.Len(t, attrs, 2, "Should have 2 attribute operations")
		}
		if _, ok := procCfg["timeout"]; ok {
			hasBatchProcessor = true
		}
	}
	assert.True(t, hasResourceProcessor, "Should have resource processor")
	assert.True(t, hasBatchProcessor, "Should have batch processor")

	// Should have 1 exporter that supports all signals
	exporters := config["exporters"].(map[string]any)
	assert.Len(t, exporters, 1, "Should have 1 exporter")

	// Validate service pipelines - should have separate pipelines for each signal type
	service := config["service"].(map[string]any)
	pipelines := service["pipelines"].(map[string]any)

	// Each signal should have its own pipeline since all components support all signals
	signalTypes := map[string]bool{"traces": false, "metrics": false, "logs": false}
	for pipelineName := range pipelines {
		for signal := range signalTypes {
			if len(pipelineName) >= len(signal) && pipelineName[:len(signal)] == signal {
				signalTypes[signal] = true
			}
		}
	}

	assert.True(t, signalTypes["traces"], "Should have traces pipeline")
	assert.True(t, signalTypes["metrics"], "Should have metrics pipeline")
	assert.True(t, signalTypes["logs"], "Should have logs pipeline")
}

func TestOTELMultiExporterConfigValidation(t *testing.T) {
	graph := createMultiExporterOTELPipeline()

	result, err := CompileGraphToOTEL(graph)

	b, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("Generated OTEL Multi-Exporter Config:\n%s\n", string(b))

	assert.NoError(t, err)
	assert.NotNil(t, result)
	if result == nil {
		t.Fatal("Result is nil, cannot continue test")
	}

	config := *result

	// Should have 1 receiver
	receivers := config["receivers"].(map[string]any)
	assert.Len(t, receivers, 1, "Should have 1 OTLP receiver")

	// Should have 1 processor (batch)
	processors := config["processors"].(map[string]any)
	assert.Len(t, processors, 1, "Should have 1 processor")

	// Should have 3 exporters all going to different backends
	exporters := config["exporters"].(map[string]any)
	assert.Len(t, exporters, 3, "Should have 3 exporters")

	// Verify distinct endpoints for each exporter
	endpoints := make(map[string]bool)
	for _, cfg := range exporters {
		expCfg := cfg.(map[string]any)
		endpoint := expCfg["endpoint"].(string)
		assert.NotContains(t, endpoints, endpoint, "Each exporter should have unique endpoint")
		endpoints[endpoint] = true
	}
	assert.Len(t, endpoints, 3, "Should have 3 distinct endpoints")

	// Validate service pipelines
	service := config["service"].(map[string]any)
	pipelines := service["pipelines"].(map[string]any)

	// Should have pipelines for each signal type
	assert.GreaterOrEqual(t, len(pipelines), 3, "Should have at least 3 pipelines (one per signal)")

	// Each pipeline should reference all 3 exporters
	for pipelineName, pipelineDef := range pipelines {
		pdef := pipelineDef.(map[string]any)
		pipelineExporters := pdef["exporters"].([]string)
		assert.Len(t, pipelineExporters, 3, "Pipeline %s should have 3 exporters", pipelineName)

		// Verify all exporter references exist
		for _, exp := range pipelineExporters {
			assert.Contains(t, exporters, exp, "Exporter %s must exist in exporters section", exp)
		}
	}
}

// =============================================================================
// EDGE CASE AND COMPLEX TOPOLOGY TESTS - OTEL
// =============================================================================

func TestOTELDiamondTopology(t *testing.T) {
	// Diamond pattern: R -> P1 -> E
	//                  |         ^
	//                  +-> P2 ---+
	graph := models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{
				ComponentID:      1,
				Name:             "receiver",
				ComponentName:    "otlp_receiver",
				ComponentRole:    "receiver",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{"endpoint": "0.0.0.0:4317"},
			},
			{
				ComponentID:      2,
				Name:             "batch_processor",
				ComponentName:    "batch_processor",
				ComponentRole:    "processor",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{"timeout": "5s"},
			},
			{
				ComponentID:      3,
				Name:             "resource_processor",
				ComponentName:    "resource_processor",
				ComponentRole:    "processor",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{"attributes": []string{"env=prod"}},
			},
			{
				ComponentID:      4,
				Name:             "exporter",
				ComponentName:    "otlp_exporter",
				ComponentRole:    "exporter",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{"endpoint": "backend:4317"},
			},
		},
		Edges: []models.PipelineEdges{
			{Source: "1", Target: "2"},
			{Source: "1", Target: "3"},
			{Source: "2", Target: "4"},
			{Source: "3", Target: "4"},
		},
	}

	result, err := CompileGraphToOTEL(graph)

	b, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("Generated OTEL Diamond Config:\n%s\n", string(b))

	assert.NoError(t, err)
	assert.NotNil(t, result)

	config := *result
	processors := config["processors"].(map[string]any)
	assert.Len(t, processors, 2, "Diamond should have 2 processors")

	service := config["service"].(map[string]any)
	pipelines := service["pipelines"].(map[string]any)
	assert.GreaterOrEqual(t, len(pipelines), 1, "Should have at least 1 pipeline")
}

func TestOTELMultipleReceiversToSingleExporter(t *testing.T) {
	// Multiple receivers with different configs going to same exporter
	graph := models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{
				ComponentID:      1,
				Name:             "otlp_receiver_grpc",
				ComponentName:    "otlp_receiver",
				ComponentRole:    "receiver",
				SupportedSignals: []string{"traces", "metrics", "logs"},
				Config: map[string]any{
					"protocols": map[string]any{
						"grpc": map[string]any{"endpoint": "0.0.0.0:4317"},
					},
				},
			},
			{
				ComponentID:      2,
				Name:             "otlp_receiver_http",
				ComponentName:    "otlp_receiver",
				ComponentRole:    "receiver",
				SupportedSignals: []string{"traces", "metrics", "logs"},
				Config: map[string]any{
					"protocols": map[string]any{
						"http": map[string]any{"endpoint": "0.0.0.0:4318"},
					},
				},
			},
			{
				ComponentID:      3,
				Name:             "jaeger_receiver",
				ComponentName:    "jaeger_receiver",
				ComponentRole:    "receiver",
				SupportedSignals: []string{"traces"},
				Config: map[string]any{
					"protocols": map[string]any{
						"grpc": map[string]any{"endpoint": "0.0.0.0:14250"},
					},
				},
			},
			{
				ComponentID:      4,
				Name:             "batch_processor",
				ComponentName:    "batch_processor",
				ComponentRole:    "processor",
				SupportedSignals: []string{"traces", "metrics", "logs"},
				Config:           map[string]any{"timeout": "10s"},
			},
			{
				ComponentID:      5,
				Name:             "otlp_exporter",
				ComponentName:    "otlp_exporter",
				ComponentRole:    "exporter",
				SupportedSignals: []string{"traces", "metrics", "logs"},
				Config:           map[string]any{"endpoint": "collector:4317"},
			},
		},
		Edges: []models.PipelineEdges{
			{Source: "1", Target: "4"},
			{Source: "2", Target: "4"},
			{Source: "3", Target: "4"},
			{Source: "4", Target: "5"},
		},
	}

	result, err := CompileGraphToOTEL(graph)

	b, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("Generated OTEL Multiple Receivers Config:\n%s\n", string(b))

	assert.NoError(t, err)
	assert.NotNil(t, result)

	config := *result

	// Should have 3 distinct receivers
	receivers := config["receivers"].(map[string]any)
	assert.Len(t, receivers, 3, "Should have 3 receivers")

	// Each receiver should have a unique alias
	receiverAliases := make(map[string]bool)
	for alias := range receivers {
		assert.NotContains(t, receiverAliases, alias, "Receiver aliases should be unique")
		receiverAliases[alias] = true
	}

	// Should have 1 processor and 1 exporter
	processors := config["processors"].(map[string]any)
	assert.Len(t, processors, 1, "Should have 1 processor")

	exporters := config["exporters"].(map[string]any)
	assert.Len(t, exporters, 1, "Should have 1 exporter")
}

func TestOTELProcessorChainOrdering(t *testing.T) {
	// Verify processors are ordered correctly in the pipeline
	// memory_limiter should come first, then batch, then resource
	graph := models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{
				ComponentID:      1,
				Name:             "receiver",
				ComponentName:    "otlp_receiver",
				ComponentRole:    "receiver",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{"endpoint": "0.0.0.0:4317"},
			},
			{
				ComponentID:      2,
				Name:             "memory_limiter",
				ComponentName:    "memorylimiter_processor",
				ComponentRole:    "processor",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{"limit_mib": 400},
			},
			{
				ComponentID:      3,
				Name:             "batch_processor",
				ComponentName:    "batch_processor",
				ComponentRole:    "processor",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{"timeout": "10s"},
			},
			{
				ComponentID:      4,
				Name:             "resource_processor",
				ComponentName:    "resource_processor",
				ComponentRole:    "processor",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{"attributes": []string{"env=prod"}},
			},
			{
				ComponentID:      5,
				Name:             "exporter",
				ComponentName:    "otlp_exporter",
				ComponentRole:    "exporter",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{"endpoint": "backend:4317"},
			},
		},
		Edges: []models.PipelineEdges{
			{Source: "1", Target: "2"},
			{Source: "2", Target: "3"},
			{Source: "3", Target: "4"},
			{Source: "4", Target: "5"},
		},
	}

	result, err := CompileGraphToOTEL(graph)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	config := *result
	service := config["service"].(map[string]any)
	pipelines := service["pipelines"].(map[string]any)

	// Find logs pipeline and check processor ordering
	for pipelineName, pipelineDef := range pipelines {
		if len(pipelineName) >= 4 && pipelineName[:4] == "logs" {
			pdef := pipelineDef.(map[string]any)
			procs := pdef["processors"].([]string)

			assert.Len(t, procs, 3, "Should have 3 processors")

			// First processor should be memory limiter
			assert.Contains(t, procs[0], "memorylimiter", "First processor should be memory limiter")
			// Second should be batch
			assert.Contains(t, procs[1], "batch", "Second processor should be batch")
			// Third should be resource
			assert.Contains(t, procs[2], "resource", "Third processor should be resource")
		}
	}
}

func TestOTELNoCommonSignalsError(t *testing.T) {
	// Create a pipeline where nodes don't share common signals
	graph := models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{
				ComponentID:      1,
				Name:             "receiver",
				ComponentName:    "otlp_receiver",
				ComponentRole:    "receiver",
				SupportedSignals: []string{"traces"},
				Config:           map[string]any{"endpoint": "0.0.0.0:4317"},
			},
			{
				ComponentID:      2,
				Name:             "processor",
				ComponentName:    "filter_processor",
				ComponentRole:    "processor",
				SupportedSignals: []string{"metrics"}, // Only supports metrics, not traces
				Config:           map[string]any{},
			},
			{
				ComponentID:      3,
				Name:             "exporter",
				ComponentName:    "otlp_exporter",
				ComponentRole:    "exporter",
				SupportedSignals: []string{"logs"}, // Only supports logs
				Config:           map[string]any{"endpoint": "backend:4317"},
			},
		},
		Edges: []models.PipelineEdges{
			{Source: "1", Target: "2"},
			{Source: "2", Target: "3"},
		},
	}

	result, err := CompileGraphToOTEL(graph)

	assert.Error(t, err, "Should error when no common signals")
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no common signals", "Error should mention no common signals")
}

func TestOTELEmptyProcessorsSection(t *testing.T) {
	// Test direct receiver to exporter connection (no processors)
	graph := models.PipelineGraph{
		Nodes: []models.PipelineNodes{
			{
				ComponentID:      1,
				Name:             "receiver",
				ComponentName:    "otlp_receiver",
				ComponentRole:    "receiver",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{"endpoint": "0.0.0.0:4317"},
			},
			{
				ComponentID:      2,
				Name:             "exporter",
				ComponentName:    "otlp_exporter",
				ComponentRole:    "exporter",
				SupportedSignals: []string{"logs"},
				Config:           map[string]any{"endpoint": "backend:4317"},
			},
		},
		Edges: []models.PipelineEdges{
			{Source: "1", Target: "2"},
		},
	}

	result, err := CompileGraphToOTEL(graph)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	config := *result

	// Processors section should exist but be empty
	processors := config["processors"].(map[string]any)
	assert.Empty(t, processors, "Processors section should be empty")

	// Pipeline should not have processors key or have empty list
	service := config["service"].(map[string]any)
	pipelines := service["pipelines"].(map[string]any)

	for _, pipelineDef := range pipelines {
		pdef := pipelineDef.(map[string]any)
		if procs, ok := pdef["processors"]; ok {
			procList := procs.([]string)
			assert.Empty(t, procList, "Pipeline processors should be empty")
		}
	}
}

func TestOTELConfigDeterminism(t *testing.T) {
	// Run compilation multiple times and verify output is deterministic
	graph := createRealisticOTELLogsPipeline()

	var results []string
	for i := 0; i < 5; i++ {
		result, err := CompileGraphToOTEL(graph)
		assert.NoError(t, err)

		b, _ := json.MarshalIndent(result, "", "  ")
		results = append(results, string(b))
	}

	// All results should be identical
	for i := 1; i < len(results); i++ {
		assert.Equal(t, results[0], results[i],
			"Compilation should be deterministic (run %d differs from run 0)", i)
	}
}
