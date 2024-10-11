package otel

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/loggingexporter"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/batchprocessor"
	"go.opentelemetry.io/collector/processor/memorylimiterprocessor"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"

	"github.com/ctrlb-hq/ctrlb-collector/internal/constants"
	"github.com/ctrlb-hq/ctrlb-collector/internal/helper"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckextension"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver"
)

type OTELCollectorAdapter struct {
	svc       *otelcol.Collector
	isActive  bool
	startTime time.Time
	mu        sync.Mutex
	wg        *sync.WaitGroup
	errChan   chan error
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewOTELCollectorAdapter(wg *sync.WaitGroup) *OTELCollectorAdapter {
	ctx, cancel := context.WithCancel(context.Background())
	return &OTELCollectorAdapter{
		wg:      wg,
		errChan: make(chan error, 1),
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (o *OTELCollectorAdapter) Initialize() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.svc != nil {
		return fmt.Errorf("OTEL collector already initialized")
	}

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

	svc, err := otelcol.NewCollector(settings)
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
	o.startTime = time.Now()
	return nil
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
	); err != nil {
		return factories, fmt.Errorf("failed to create processor factory map: %w", err)
	}

	if factories.Exporters, err = exporter.MakeFactoryMap(
		otlpexporter.NewFactory(),
		loggingexporter.NewFactory(),
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

func (o *OTELCollectorAdapter) StartAgent() error {
	if o.isActive {
		return fmt.Errorf("fluent-bit instance already running")
	}

	o.mu.Lock()
	defer o.mu.Unlock()

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

	svc, err := otelcol.NewCollector(settings)
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

func (o *OTELCollectorAdapter) StopAgent() error {
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

func (o *OTELCollectorAdapter) UpdateConfig() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.svc == nil {
		return fmt.Errorf("OTEL collector service not initialized")
	}

	err := o.StopAgent()
	if err != nil {
		return fmt.Errorf("failed to stop OTEL collector for config update: %w", err)
	}

	o.svc = nil
	o.ctx, o.cancel = context.WithCancel(context.Background())

	err = o.StartAgent()
	if err != nil {
		return fmt.Errorf("failed to restart OTEL collector with new config: %w", err)
	}

	log.Println("Config updated. OTEL collector restarted")
	return nil
}

func (o *OTELCollectorAdapter) GracefulShutdown() error {
	log.Println("Initiating Server shutdown...")

	helper.ShutdownServer(o.wg)

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
