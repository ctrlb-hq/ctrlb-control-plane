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

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/core/shutdown"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/models"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/utils"
	"github.com/prometheus/common/expfmt"
	"go.opentelemetry.io/collector/otelcol"
)

type OTELAdapter struct {
	svc      *otelcol.Collector
	isActive bool
	mu       sync.Mutex
	wg       *sync.WaitGroup
	errChan  chan error
	ctx      context.Context
	cancel   context.CancelFunc
	baseUrl  string
}

func NewOTELAdapter(wg *sync.WaitGroup) *OTELAdapter {
	ctx, cancel := context.WithCancel(context.Background())
	return &OTELAdapter{
		wg:      wg,
		errChan: make(chan error, 1),
		ctx:     ctx,
		cancel:  cancel,
		baseUrl: "http://0.0.0.0:8888",
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
			log.Printf("OTEL collector stopped with error: %v", err)
		}
	}()
	a.isActive = true
	return nil
}

func (a *OTELAdapter) StartAgent() error {
	if a.isActive {
		return fmt.Errorf("fluent-bit instance already running")
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
			log.Printf("OTEL collector stopped with error: %v", err)
		}
	}()
	a.isActive = true
	return nil
}

func (a *OTELAdapter) StopAgent() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.isActive {
		return fmt.Errorf("OTEL collector instance not currently running")
	}

	if a.svc != nil {
		a.svc.Shutdown()
	} else {
		log.Println("otel collector is not running")
		return fmt.Errorf("otel collector is not running")
	}
	a.isActive = false
	log.Println("OTEL collector stopped")
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

func (a *OTELAdapter) GracefulShutdown() error {
	log.Println("Initiating Server shutdown...")

	shutdown.ShutdownServer(a.wg)

	log.Printf("Initiating graceful shutdown of Otel agent...")

	a.StopAgent()

	log.Printf("Waiting for all goroutines to finish...")
	done := make(chan struct{})
	a.wg.Wait()
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

func (a *OTELAdapter) CurrentStatus() (*models.AgentMetrics, error) {
	url := a.baseUrl + "/metrics"
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

func (a *OTELAdapter) GetVersion() (string, error) {
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
