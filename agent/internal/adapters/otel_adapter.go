package adapters

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/core/shutdown"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/pkg/logger"

	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/otelcol"
)

type OTELAdapter struct {
	svc    *otelcol.Collector
	mu     sync.RWMutex
	wg     *sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func NewOTELAdapter(wg *sync.WaitGroup) *OTELAdapter {
	ctx, cancel := context.WithCancel(context.Background())
	return &OTELAdapter{
		wg:     wg,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (a *OTELAdapter) Initialize() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.svc != nil {
		return fmt.Errorf("OTEL collector already initialized")
	}

	svc, err := getNewOTELCollector()
	if err != nil {
		return fmt.Errorf("failed to create OTEL collector: %w", err)
	}

	a.svc = svc
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		if err := a.svc.Run(a.ctx); err != nil {
			logger.Logger.Sugar().Errorf("OTEL collector stopped with error: %v", err)
		}
		a.mu.Lock()
		a.svc = nil
		a.mu.Unlock()
	}()
	return nil
}

func (a *OTELAdapter) StartAgent() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.svc != nil {
		return fmt.Errorf("OTEL collector instance already running")
	}

	svc, err := getNewOTELCollector()
	if err != nil {
		return fmt.Errorf("failed to start OTEL collector: %w", err)
	}

	a.svc = svc
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		if err := a.svc.Run(a.ctx); err != nil {
			logger.Logger.Sugar().Errorf("OTEL collector stopped with error: %v", err)
		}
		a.mu.Lock()
		a.svc = nil
		a.mu.Unlock()
	}()
	return nil
}

func (a *OTELAdapter) StopAgent() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.svc == nil {
		return fmt.Errorf("OTEL collector instance not currently running")
	}

	a.svc.Shutdown()
	a.svc = nil
	logger.Logger.Info("OTEL collector stopped")
	return nil
}

func (a *OTELAdapter) UpdateConfig() error {
	a.mu.RLock()
	running := a.svc != nil
	a.mu.RUnlock()

	if !running {
		if err := a.StartAgent(); err != nil {
			return fmt.Errorf("failed to start OTEL collector: %w", err)
		}
		logger.Logger.Info("OTEL collector started with updated config")
		return nil
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP)
	defer signal.Stop(sigChan)

	logger.Logger.Info("Sending SIGHUP signal to hot-reload otel collector for updating config...")
	if err := syscall.Kill(os.Getpid(), syscall.SIGHUP); err != nil {
		return fmt.Errorf("failed to send SIGHUP signal: %w", err)
	}

	select {
	case sig := <-sigChan:
		logger.Logger.Sugar().Infof("Received signal for updating config in otel collector: %s", sig)
	case <-time.After(5 * time.Second):
		logger.Logger.Warn("Timeout waiting for config update signal")
	}

	logger.Logger.Info("Config updated. OTEL collector restarted")
	return nil
}

func (a *OTELAdapter) GracefulShutdown() error {
	logger.Logger.Info("Starting graceful shutdown...")

	// Cancel context to stop running goroutines
	a.cancel()

	// Shutdown external services
	shutdown.ShutdownServer()

	// Stop the agent
	if err := a.StopAgent(); err != nil {
		logger.Logger.Error(fmt.Sprintf("Error stopping agent during shutdown: %v", err))
	}

	logger.Logger.Info("Waiting for all goroutines to finish...")
	done := make(chan struct{})
	go func() {
		a.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Logger.Info("All goroutines finished successfully")
	case <-time.After(20 * time.Second):
		return fmt.Errorf("timed out waiting for goroutines to finish")
	}

	logger.Logger.Info("Agent shutdown successfully")
	return nil
}

func (a *OTELAdapter) GetVersion() (string, error) {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, dep := range info.Deps {
			if strings.HasPrefix(dep.Path, "go.opentelemetry.io/collector") {
				return strings.TrimPrefix(dep.Version, "v"), nil
			}
		}
	}
	return "", fmt.Errorf("failed to determine OpenTelemetry Collector version")
}

func (a *OTELAdapter) ValidateConfigInMemory(data *map[string]any) error {
	if data == nil || *data == nil {
		return fmt.Errorf("configuration data is nil")
	}

	// Get factories
	factories, err := componentsFactory()
	if err != nil {
		return fmt.Errorf("failed to create component factories: %w", err)
	}

	// Create in-memory config map
	configMap := confmap.NewFromStringMap(*data)

	// Create collector config
	collectorConfig := &otelcol.Config{}
	if err := configMap.Unmarshal(collectorConfig); err != nil {
		return fmt.Errorf("failed to unmarshal collector config: %w", err)
	}

	// Validate the configuration by attempting to create components
	if err := validateComponents(collectorConfig, factories); err != nil {
		return fmt.Errorf("component validation failed: %w", err)
	}

	logger.Logger.Info("In-memory configuration validation successful")
	return nil
}
