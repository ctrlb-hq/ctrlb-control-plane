package adapters

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/collector/otelcol"
)

type EngineStatus struct {
	Running bool
	Error   string
}

type collectorRun struct {
	collector *otelcol.Collector
	cancel    context.CancelFunc
	done      chan struct{}
	err       error
	stop      sync.Once
}

// ManagedEngine serializes programmatic Collector starts and stops. It never
// signals the process or changes the legacy adapter's global configuration.
type ManagedEngine struct {
	opMu sync.Mutex
	mu   sync.RWMutex
	run  *collectorRun
}

func (e *ManagedEngine) Validate(ctx context.Context, path string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	collector, err := NewManagedCollector(path)
	if err != nil {
		return err
	}
	return collector.DryRun(ctx)
}

func (e *ManagedEngine) Start(ctx context.Context, path string) error {
	e.opMu.Lock()
	defer e.opMu.Unlock()
	e.mu.RLock()
	previous := e.run
	e.mu.RUnlock()
	if previous != nil {
		select {
		case <-previous.done:
		default:
			return errors.New("collector is already running or stopping")
		}
	}
	collector, err := NewManagedCollector(path)
	if err != nil {
		return err
	}
	runCtx, cancel := context.WithCancel(context.Background())
	run := &collectorRun{collector: collector, cancel: cancel, done: make(chan struct{})}
	e.mu.Lock()
	e.run = run
	e.mu.Unlock()
	go func() {
		err := collector.Run(runCtx)
		e.mu.Lock()
		run.err = err
		close(run.done)
		e.mu.Unlock()
	}()

	ctx, stop := context.WithTimeout(ctx, 30*time.Second)
	defer stop()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-run.done:
			return e.runError(run)
		case <-ctx.Done():
			stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer stopCancel()
			return errors.Join(fmt.Errorf("waiting for Collector startup: %w", ctx.Err()), e.stopRun(stopCtx, run))
		case <-ticker.C:
			if collector.GetState() == otelcol.StateRunning {
				select {
				case <-run.done:
					return e.runError(run)
				default:
					return nil
				}
			}
		}
	}
}

func (e *ManagedEngine) runError(run *collectorRun) error {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if run.err != nil {
		return run.err
	}
	return errors.New("Collector exited before startup completed")
}

func (e *ManagedEngine) Stop(ctx context.Context) error {
	e.opMu.Lock()
	defer e.opMu.Unlock()
	e.mu.RLock()
	run := e.run
	e.mu.RUnlock()
	if run == nil {
		return nil
	}
	return e.stopRun(ctx, run)
}

func (e *ManagedEngine) stopRun(ctx context.Context, run *collectorRun) error {
	// Shutdown waits for cleanup in current Collector versions. Keep its wait
	// bounded here, while refusing another start until Run has actually exited.
	run.stop.Do(func() {
		run.cancel()
		go run.collector.Shutdown()
	})
	select {
	case <-run.done:
		e.mu.Lock()
		if e.run == run {
			e.run = nil
		}
		e.mu.Unlock()
		return nil
	case <-ctx.Done():
		return fmt.Errorf("waiting for Collector shutdown: %w", ctx.Err())
	}
}

func (e *ManagedEngine) Status() EngineStatus {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.run == nil {
		return EngineStatus{}
	}
	select {
	case <-e.run.done:
		message := "Collector exited unexpectedly"
		if e.run.err != nil {
			message = e.run.err.Error()
		}
		return EngineStatus{Error: message}
	default:
		return EngineStatus{Running: e.run.collector.GetState() == otelcol.StateRunning}
	}
}
