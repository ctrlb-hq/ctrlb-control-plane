package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/core/shutdown"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/pkg/logger"
)

const (
	// Default Fluent Bit HTTP server settings
	defaultFluentBitHTTPPort = "2020"
	defaultFluentBitHTTPHost = "127.0.0.1"

	// API endpoints
	apiV2Reload  = "/api/v2/reload"
	apiV1Health  = "/api/v1/health"
	apiV1Metrics = "/api/v1/metrics"
	apiV1Uptime  = "/api/v1/uptime"
	apiRoot      = "/"

	// Process management timeouts
	startupTimeout      = 30 * time.Second
	shutdownTimeout     = 20 * time.Second
	reloadTimeout       = 10 * time.Second
	healthCheckRetry    = 5
	healthCheckInterval = 2 * time.Second
)

// FluentBitAdapter manages Fluent Bit as a child process
type FluentBitAdapter struct {
	process    *exec.Cmd
	mu         sync.RWMutex
	wg         *sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	httpClient *http.Client
	baseURL    string

	// Process management
	processExited chan struct{}
	isRunning     bool
}

// FluentBitBuildInfo represents the Fluent Bit version information
type FluentBitBuildInfo struct {
	FluentBit struct {
		Version string   `json:"version"`
		Edition string   `json:"edition"`
		Flags   []string `json:"flags"`
	} `json:"fluent-bit"`
}

// FluentBitUptimeInfo represents uptime information
type FluentBitUptimeInfo struct {
	UptimeSec int64  `json:"uptime_sec"`
	UptimeHr  string `json:"uptime_hr"`
}

// FluentBitHealthResponse represents health check response
type FluentBitHealthResponse struct {
	Status string `json:"status,omitempty"`
}

// FluentBitReloadResponse represents hot reload response
type FluentBitReloadResponse struct {
	Reload string `json:"reload,omitempty"`
	Status int    `json:"status,omitempty"`
}

// NewFluentBitAdapter creates a new Fluent Bit adapter instance
func NewFluentBitAdapter(wg *sync.WaitGroup) *FluentBitAdapter {
	ctx, cancel := context.WithCancel(context.Background())
	return &FluentBitAdapter{
		wg:     wg,
		ctx:    ctx,
		cancel: cancel,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL:       fmt.Sprintf("http://%s:%s", defaultFluentBitHTTPHost, defaultFluentBitHTTPPort),
		processExited: make(chan struct{}),
		isRunning:     false,
	}
}

// Initialize starts the Fluent Bit process for the first time
func (a *FluentBitAdapter) Initialize() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.isRunning {
		return fmt.Errorf("fluent-bit process already initialized")
	}

	if err := a.startProcess(); err != nil {
		return fmt.Errorf("failed to initialize fluent-bit: %w", err)
	}

	logger.Logger.Info("Fluent Bit initialized successfully")
	return nil
}

// StartAgent starts a new Fluent Bit process
func (a *FluentBitAdapter) StartAgent() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.isRunning {
		return fmt.Errorf("fluent-bit process already running")
	}

	if err := a.startProcess(); err != nil {
		return fmt.Errorf("failed to start fluent-bit: %w", err)
	}

	logger.Logger.Info("Fluent Bit started successfully")
	return nil
}

// StopAgent stops the running Fluent Bit process
func (a *FluentBitAdapter) StopAgent() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.isRunning {
		return fmt.Errorf("fluent-bit process is not currently running")
	}

	logger.Logger.Info("Stopping Fluent Bit process...")

	// First try graceful shutdown with SIGTERM
	if a.process != nil && a.process.Process != nil {
		if err := a.process.Process.Signal(os.Interrupt); err != nil {
			logger.Logger.Sugar().Warnf("Failed to send SIGTERM to fluent-bit: %v", err)
		}

		// Wait for graceful shutdown
		shutdownDone := make(chan struct{})
		go func() {
			select {
			case <-a.processExited:
				close(shutdownDone)
			case <-time.After(shutdownTimeout):
				// Force kill if graceful shutdown times out
				logger.Logger.Warn("Graceful shutdown timeout, forcing kill...")
				if a.process != nil && a.process.Process != nil {
					_ = a.process.Process.Kill()
				}
				close(shutdownDone)
			}
		}()

		<-shutdownDone
	}

	a.process = nil
	a.isRunning = false
	logger.Logger.Info("Fluent Bit stopped")
	return nil
}

// UpdateConfig reloads Fluent Bit configuration using hot reload
func (a *FluentBitAdapter) UpdateConfig() error {
	a.mu.RLock()
	running := a.isRunning
	a.mu.RUnlock()

	if !running {
		// If not running, start with new config
		if err := a.StartAgent(); err != nil {
			return fmt.Errorf("failed to start fluent-bit with updated config: %w", err)
		}
		logger.Logger.Info("Fluent Bit started with updated config")
		return nil
	}

	// Use HTTP hot reload API
	logger.Logger.Info("Triggering hot reload for Fluent Bit...")

	reloadURL := a.baseURL + apiV2Reload
	reqBody := []byte("{}")

	req, err := http.NewRequestWithContext(a.ctx, "POST", reloadURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create reload request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send reload request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("reload request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var reloadResp FluentBitReloadResponse
	if err := json.Unmarshal(body, &reloadResp); err != nil {
		logger.Logger.Sugar().Warnf("Failed to parse reload response: %v", err)
	} else {
		logger.Logger.Sugar().Infof("Hot reload successful: (reload: %s, status: %d)", reloadResp.Reload, reloadResp.Status)
	}

	// Verify the reload was successful by checking health with retries
	// Hot reload may cause temporary unavailability
	if err := a.checkHealthWithRetry(5, 2*time.Second); err != nil {
		return fmt.Errorf("health check failed after reload: %w", err)
	}

	logger.Logger.Info("Config updated successfully")
	return nil
}

// GracefulShutdown performs a graceful shutdown of the adapter
func (a *FluentBitAdapter) GracefulShutdown() error {
	logger.Logger.Info("Starting graceful shutdown of Fluent Bit adapter...")

	// Cancel context to stop any ongoing operations
	a.cancel()

	// Shutdown external services
	shutdown.ShutdownServer()

	// Stop the agent
	if err := a.StopAgent(); err != nil {
		logger.Logger.Sugar().Errorf("Error stopping Fluent Bit during shutdown: %v", err)
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
	case <-time.After(shutdownTimeout):
		return fmt.Errorf("timed out waiting for goroutines to finish")
	}

	logger.Logger.Info("Fluent Bit adapter shutdown successfully")
	return nil
}

// GetVersion retrieves the Fluent Bit version
func (a *FluentBitAdapter) GetVersion() (string, error) {
	a.mu.RLock()
	running := a.isRunning
	a.mu.RUnlock()

	// If running, query via HTTP API
	if running {
		req, err := http.NewRequestWithContext(a.ctx, "GET", a.baseURL+apiRoot, nil)
		if err != nil {
			return "", fmt.Errorf("failed to create version request: %w", err)
		}

		resp, err := a.httpClient.Do(req)
		if err != nil {
			return a.getVersionFromBinary()
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return a.getVersionFromBinary()
		}

		var buildInfo FluentBitBuildInfo
		if err := json.NewDecoder(resp.Body).Decode(&buildInfo); err != nil {
			return a.getVersionFromBinary()
		}

		return buildInfo.FluentBit.Version, nil
	}

	// If not running, get version from binary
	return a.getVersionFromBinary()
}

// ValidateConfigInMemory validates a Fluent Bit configuration
func (a *FluentBitAdapter) ValidateConfigInMemory(data *map[string]any) error {
	if data == nil || *data == nil {
		return fmt.Errorf("configuration data is nil")
	}

	// Validate required sections
	if _, hasService := (*data)["service"]; !hasService {
		return fmt.Errorf("configuration missing required 'service' section")
	}

	// Check if HTTP server is enabled (required for monitoring)
	service, ok := (*data)["service"].(map[string]any)
	if !ok {
		return fmt.Errorf("'service' section is not a valid object")
	}

	// Ensure hot reload is enabled
	if hotReload, exists := service["hot_reload"]; exists {
		if hr, ok := hotReload.(bool); ok && !hr {
			logger.Logger.Warn("hot_reload is disabled in config - hot reload functionality will not work")
		}
	}

	// Ensure HTTP server is enabled
	if httpServer, exists := service["http_server"]; exists {
		if hs, ok := httpServer.(bool); ok && !hs {
			logger.Logger.Warn("http_server is disabled in config - monitoring will not work")
		}
	}

	// Validate pipeline structure
	if pipeline, hasPipeline := (*data)["pipeline"]; hasPipeline {
		pipelineMap, ok := pipeline.(map[string]any)
		if !ok {
			return fmt.Errorf("'pipeline' section is not a valid object")
		}

		// Check for inputs
		if inputs, hasInputs := pipelineMap["inputs"]; !hasInputs {
			return fmt.Errorf("pipeline missing 'inputs' section")
		} else {
			if inputsList, ok := inputs.([]any); !ok || len(inputsList) == 0 {
				return fmt.Errorf("pipeline 'inputs' must be a non-empty array")
			}
		}

		// Check for outputs
		if outputs, hasOutputs := pipelineMap["outputs"]; !hasOutputs {
			return fmt.Errorf("pipeline missing 'outputs' section")
		} else {
			if outputsList, ok := outputs.([]any); !ok || len(outputsList) == 0 {
				return fmt.Errorf("pipeline 'outputs' must be a non-empty array")
			}
		}
	}

	logger.Logger.Info("Fluent Bit configuration validation successful")
	return nil
}

// GetMetrics retrieves metrics from Fluent Bit
func (a *FluentBitAdapter) GetMetrics() (map[string]any, error) {
	a.mu.RLock()
	running := a.isRunning
	a.mu.RUnlock()

	if !running {
		return nil, fmt.Errorf("fluent-bit is not running")
	}

	req, err := http.NewRequestWithContext(a.ctx, "GET", a.baseURL+apiV1Metrics, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics request: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("metrics request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("metrics request failed with status %d", resp.StatusCode)
	}

	var metrics map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
		return nil, fmt.Errorf("failed to decode metrics: %w", err)
	}

	return metrics, nil
}

// GetUptime retrieves uptime information from Fluent Bit
func (a *FluentBitAdapter) GetUptime() (*FluentBitUptimeInfo, error) {
	a.mu.RLock()
	running := a.isRunning
	a.mu.RUnlock()

	if !running {
		return nil, fmt.Errorf("fluent-bit is not running")
	}

	req, err := http.NewRequestWithContext(a.ctx, "GET", a.baseURL+apiV1Uptime, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create uptime request: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("uptime request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("uptime request failed with status %d", resp.StatusCode)
	}

	var uptime FluentBitUptimeInfo
	if err := json.NewDecoder(resp.Body).Decode(&uptime); err != nil {
		return nil, fmt.Errorf("failed to decode uptime: %w", err)
	}

	return &uptime, nil
}
