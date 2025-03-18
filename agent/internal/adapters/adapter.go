package adapters

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/debugexporter"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/batchprocessor"
	"go.opentelemetry.io/collector/processor/memorylimiterprocessor"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/constants"
	shutdownhelper "github.com/ctrlb-hq/ctrlb-collector/agent/internal/core/shutdown"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/models"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/utils"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckextension"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/probabilisticsamplerprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver"
	"github.com/prometheus/common/expfmt"
)

type Adapter struct {
	svc      *otelcol.Collector
	isActive bool
	mu       sync.Mutex
	wg       *sync.WaitGroup
	errChan  chan error
	ctx      context.Context
	cancel   context.CancelFunc
	baseUrl  string
}

func NewAdapter(wg *sync.WaitGroup) *Adapter {
	ctx, cancel := context.WithCancel(context.Background())
	return &Adapter{
		wg:      wg,
		errChan: make(chan error, 1),
		ctx:     ctx,
		cancel:  cancel,
		baseUrl: "http://0.0.0.0:8888",
	}
}

func componentsFactory() (otelcol.Factories, error) {
	factories := otelcol.Factories{}
	var err error

	if factories.Receivers, err = receiver.MakeFactoryMap(
		otlpreceiver.NewFactory(),
		hostmetricsreceiver.NewFactory(),
	); err != nil {
		return factories, fmt.Errorf("failed to create receiver factory map: %w", err)
	}

	if factories.Processors, err = processor.MakeFactoryMap(
		batchprocessor.NewFactory(),
		memorylimiterprocessor.NewFactory(),
		probabilisticsamplerprocessor.NewFactory(),
	); err != nil {
		return factories, fmt.Errorf("failed to create processor factory map: %w", err)
	}

	if factories.Exporters, err = exporter.MakeFactoryMap(
		otlpexporter.NewFactory(),
		debugexporter.NewFactory(),
	); err != nil {
		return factories, fmt.Errorf("failed to create exporter factory map: %w", err)
	}

	if factories.Connectors, err = connector.MakeFactoryMap(); err != nil {
		return factories, fmt.Errorf("failed to create connector factory map: %w", err)
	}

	if factories.Extensions, err = extension.MakeFactoryMap(
		healthcheckextension.NewFactory(),
	); err != nil {
		return factories, fmt.Errorf("failed to create extension factory map: %w", err)
	}

	return factories, nil
}

func getNewOTELCollector() (*otelcol.Collector, error) {
	fmp := fileprovider.NewFactory()
	configProviderSettings := otelcol.ConfigProviderSettings{
		ResolverSettings: confmap.ResolverSettings{
			URIs:              []string{constants.AGENT_CONFIG_PATH},
			ProviderFactories: []confmap.ProviderFactory{fmp},
		},
	}

	settings := otelcol.CollectorSettings{
		BuildInfo:              component.NewDefaultBuildInfo(),
		Factories:              componentsFactory,
		ConfigProviderSettings: configProviderSettings,
	}

	return otelcol.NewCollector(settings)
}

func (o *Adapter) Initialize() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.svc != nil {
		return fmt.Errorf("OTEL collector already initialized")
	}

	svc, err := getNewOTELCollector()
	if err != nil {
		return fmt.Errorf("failed to create OTEL collector: %w", err)
	}

	o.svc = svc
	o.wg.Add(1)
	go func() {
		defer o.wg.Done()
		if err := o.svc.Run(o.ctx); err != nil {
			log.Printf("OTEL collector stopped with error: %v", err)
		}
	}()
	o.isActive = true
	return nil
}

func (o *Adapter) StartAgent() error {
	if o.isActive {
		return fmt.Errorf("fluent-bit instance already running")
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	svc, err := getNewOTELCollector()
	if err != nil {
		return fmt.Errorf("failed to start OTEL collector: %w", err)
	}

	o.svc = svc
	o.wg.Add(1)
	go func() {
		defer o.wg.Done()
		if err := o.svc.Run(o.ctx); err != nil {
			log.Printf("OTEL collector stopped with error: %v", err)
		}
	}()
	o.isActive = true
	return nil
}

func (o *Adapter) StopAgent() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if !o.isActive {
		return fmt.Errorf("OTEL collector instance not currently running")
	}

	if o.svc != nil {
		o.svc.Shutdown()
	} else {
		log.Println("otel collector is not running")
		return fmt.Errorf("otel collector is not running")
	}
	o.isActive = false
	log.Println("OTEL collector stopped")
	return nil
}

func (o *Adapter) UpdateConfig() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP)

	go func() {
		for {
			sig := <-sigChan
			fmt.Printf("Received signal for updating config in otel collector: %s\n", sig)
		}
	}()

	fmt.Println("Sending SIGHUP signal to Hot-reload Otel collector for updating config...")
	syscall.Kill(os.Getpid(), syscall.SIGHUP)

	time.Sleep(2 * time.Second)
	o.isActive = true

	log.Println("Config updated. OTEL collector restarted")
	return nil
}

func (o *Adapter) GracefulShutdown() error {
	log.Println("Initiating Server shutdown...")

	shutdownhelper.ShutdownServer(o.wg)

	log.Printf("Initiating graceful shutdown of Otel agent...")

	o.StopAgent()

	log.Printf("Waiting for all goroutines to finish...")
	done := make(chan struct{})
	o.wg.Wait()
	close(done)

	select {
	case <-done:
		log.Printf("All goroutines finished successfully")

	case <-time.After(20 * time.Second):
		return fmt.Errorf("timed out waiting for goroutines to finish")
	}

	log.Printf("Otel collector has been gracefully shutdown")
	os.Exit(0)
	return nil
}

func (o *Adapter) CurrentStatus() (*models.AgentMetrics, error) {
	url := o.baseUrl + "/metrics"
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metrics: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read metrics: %v", err)
	}

	parser := expfmt.TextParser{}
	metrics, err := parser.TextToMetricFamilies(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse metrics: %v", err)
	}

	agentMetrics, err := utils.ExtractStatusFromPrometheus(metrics, "otel")
	if err != nil {
		return nil, fmt.Errorf("failed to parse metrics: %v", err)
	}

	return agentMetrics, nil
}

func (o *Adapter) GetVersion() (string, error) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "", fmt.Errorf("failed to read build info")
	}

	for _, dep := range info.Deps {
		if dep.Path == "go.opentelemetry.io/collector" {
			return strings.TrimPrefix(dep.Version, "v"), nil
		}
	}

	return "", fmt.Errorf("OpenTelemetry Collector dependency not found")
}
