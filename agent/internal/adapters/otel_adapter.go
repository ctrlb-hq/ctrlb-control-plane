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
	"go.opentelemetry.io/collector/otelcol"
)

type OTELAdapter struct {
	svc    *otelcol.Collector
	mu     sync.Mutex
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
			logger.Logger.Error(fmt.Sprintf("OTEL collector stopped with error: %v", err))
		}
	}()
	return nil
}

func (a *OTELAdapter) StartAgent() error {
	if a.svc != nil {
		return fmt.Errorf("OTEL collector instance already running")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	svc, err := getNewOTELCollector()
	if err != nil {
		return fmt.Errorf("failed to start OTEL collector: %w", err)
	}

	a.svc = svc
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		if err := a.svc.Run(a.ctx); err != nil {
			logger.Logger.Error(fmt.Sprintf("OTEL collector stopped with error: %v", err))
		}
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

func (o *OTELAdapter) UpdateConfig() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP)

	go func() {
		for {
			sig := <-sigChan
			logger.Logger.Info(fmt.Sprintf("Received signal for updating config in otel collector: %s\n", sig))
		}
	}()

	logger.Logger.Info("Sending SIGHUP signal to hot-reload otel collector for updating config...")
	syscall.Kill(os.Getpid(), syscall.SIGHUP)

	time.Sleep(2 * time.Second)

	logger.Logger.Info("Config updated. OTEL collector restarted")
	return nil
}

func (a *OTELAdapter) GracefulShutdown() error {
	shutdown.ShutdownServer(a.wg)
	a.StopAgent()

	logger.Logger.Info("Waiting for all goroutines to finish...")
	done := make(chan struct{})
	a.wg.Wait()
	close(done)

	select {
	case <-done:
		logger.Logger.Info("All goroutines finished successfully")

	case <-time.After(20 * time.Second):
		return fmt.Errorf("timed out waiting for goroutines to finish")
	}

	logger.Logger.Info("Agent shutdown successfully")
	os.Exit(0)
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
