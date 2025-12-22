package adapters

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFluentBitAdapter(t *testing.T) {
	var wg sync.WaitGroup
	adapter := NewFluentBitAdapter(&wg)

	assert.NotNil(t, adapter)
	assert.NotNil(t, adapter.httpClient)
	assert.NotNil(t, adapter.ctx)
	assert.NotNil(t, adapter.cancel)
	assert.False(t, adapter.isRunning)
	assert.Equal(t, "http://127.0.0.1:2020", adapter.baseURL)
}

func TestFluentBitAdapter_ValidateConfigInMemory_Success(t *testing.T) {
	var wg sync.WaitGroup
	adapter := NewFluentBitAdapter(&wg)

	validConfig := map[string]any{
		"service": map[string]any{
			"http_server": true,
			"http_listen": "0.0.0.0",
			"http_port":   2020,
			"hot_reload":  true,
		},
		"pipeline": map[string]any{
			"inputs": []any{
				map[string]any{
					"name": "cpu",
				},
			},
			"outputs": []any{
				map[string]any{
					"name":  "stdout",
					"match": "*",
				},
			},
		},
	}

	err := adapter.ValidateConfigInMemory(&validConfig)
	assert.NoError(t, err)
}

func TestFluentBitAdapter_ValidateConfigInMemory_NilConfig(t *testing.T) {
	var wg sync.WaitGroup
	adapter := NewFluentBitAdapter(&wg)

	err := adapter.ValidateConfigInMemory(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configuration data is nil")
}

func TestFluentBitAdapter_ValidateConfigInMemory_MissingService(t *testing.T) {
	var wg sync.WaitGroup
	adapter := NewFluentBitAdapter(&wg)

	invalidConfig := map[string]any{
		"pipeline": map[string]any{},
	}

	err := adapter.ValidateConfigInMemory(&invalidConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required 'service' section")
}

func TestFluentBitAdapter_ValidateConfigInMemory_MissingInputs(t *testing.T) {
	var wg sync.WaitGroup
	adapter := NewFluentBitAdapter(&wg)

	invalidConfig := map[string]any{
		"service": map[string]any{
			"http_server": true,
		},
		"pipeline": map[string]any{
			"outputs": []any{
				map[string]any{
					"name": "stdout",
				},
			},
		},
	}

	err := adapter.ValidateConfigInMemory(&invalidConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing 'inputs' section")
}

func TestFluentBitAdapter_ValidateConfigInMemory_MissingOutputs(t *testing.T) {
	var wg sync.WaitGroup
	adapter := NewFluentBitAdapter(&wg)

	invalidConfig := map[string]any{
		"service": map[string]any{
			"http_server": true,
		},
		"pipeline": map[string]any{
			"inputs": []any{
				map[string]any{
					"name": "cpu",
				},
			},
		},
	}

	err := adapter.ValidateConfigInMemory(&invalidConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing 'outputs' section")
}

func TestFluentBitAdapter_ValidateConfigInMemory_EmptyInputs(t *testing.T) {
	var wg sync.WaitGroup
	adapter := NewFluentBitAdapter(&wg)

	invalidConfig := map[string]any{
		"service": map[string]any{
			"http_server": true,
		},
		"pipeline": map[string]any{
			"inputs": []any{},
			"outputs": []any{
				map[string]any{
					"name": "stdout",
				},
			},
		},
	}

	err := adapter.ValidateConfigInMemory(&invalidConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a non-empty array")
}

func TestFluentBitAdapter_checkHealth_Success(t *testing.T) {
	// Create mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/health" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		}
	}))
	defer server.Close()

	var wg sync.WaitGroup
	adapter := NewFluentBitAdapter(&wg)
	adapter.baseURL = server.URL
	adapter.isRunning = true

	err := adapter.checkHealth()
	assert.NoError(t, err)
}

func TestFluentBitAdapter_checkHealth_Failure(t *testing.T) {
	// Create mock HTTP server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/health" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error"))
		}
	}))
	defer server.Close()

	var wg sync.WaitGroup
	adapter := NewFluentBitAdapter(&wg)
	adapter.baseURL = server.URL
	adapter.isRunning = true

	err := adapter.checkHealth()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "health check failed")
}

func TestFluentBitAdapter_GetVersion_ViaHTTP(t *testing.T) {
	buildInfo := FluentBitBuildInfo{}
	buildInfo.FluentBit.Version = "3.2.0"
	buildInfo.FluentBit.Edition = "Community"

	// Create mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(buildInfo)
		}
	}))
	defer server.Close()

	var wg sync.WaitGroup
	adapter := NewFluentBitAdapter(&wg)
	adapter.baseURL = server.URL
	adapter.isRunning = true

	version, err := adapter.GetVersion()
	assert.NoError(t, err)
	assert.Equal(t, "3.2.0", version)
}

func TestFluentBitAdapter_GetMetrics_Success(t *testing.T) {
	mockMetrics := map[string]any{
		"input": map[string]any{
			"cpu.0": map[string]any{
				"records": 8,
				"bytes":   2536,
			},
		},
		"output": map[string]any{
			"stdout.0": map[string]any{
				"proc_records":   5,
				"proc_bytes":     1585,
				"errors":         0,
				"retries":        0,
				"retries_failed": 0,
			},
		},
	}

	// Create mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/metrics" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockMetrics)
		}
	}))
	defer server.Close()

	var wg sync.WaitGroup
	adapter := NewFluentBitAdapter(&wg)
	adapter.baseURL = server.URL
	adapter.isRunning = true

	metrics, err := adapter.GetMetrics()
	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.Contains(t, metrics, "input")
	assert.Contains(t, metrics, "output")
}

func TestFluentBitAdapter_GetMetrics_NotRunning(t *testing.T) {
	var wg sync.WaitGroup
	adapter := NewFluentBitAdapter(&wg)
	adapter.isRunning = false

	metrics, err := adapter.GetMetrics()
	assert.Error(t, err)
	assert.Nil(t, metrics)
	assert.Contains(t, err.Error(), "not running")
}

func TestFluentBitAdapter_GetUptime_Success(t *testing.T) {
	mockUptime := FluentBitUptimeInfo{
		UptimeSec: 8950000,
		UptimeHr:  "Fluent Bit has been running: 103 days, 14 hours, 6 minutes and 40 seconds",
	}

	// Create mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/uptime" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockUptime)
		}
	}))
	defer server.Close()

	var wg sync.WaitGroup
	adapter := NewFluentBitAdapter(&wg)
	adapter.baseURL = server.URL
	adapter.isRunning = true

	uptime, err := adapter.GetUptime()
	assert.NoError(t, err)
	assert.NotNil(t, uptime)
	assert.Equal(t, int64(8950000), uptime.UptimeSec)
	assert.Contains(t, uptime.UptimeHr, "103 days")
}

func TestFluentBitAdapter_GetUptime_NotRunning(t *testing.T) {
	var wg sync.WaitGroup
	adapter := NewFluentBitAdapter(&wg)
	adapter.isRunning = false

	uptime, err := adapter.GetUptime()
	assert.Error(t, err)
	assert.Nil(t, uptime)
	assert.Contains(t, err.Error(), "not running")
}

func TestFluentBitAdapter_UpdateConfig_NotRunning(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `service:
  http_server: on
  http_port: 2020
  hot_reload: on

pipeline:
  inputs:
    - name: cpu
  outputs:
    - name: stdout
      match: '*'
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Save original and restore after test
	originalPath := constants.AGENT_CONFIG_PATH
	constants.AGENT_CONFIG_PATH = configPath
	defer func() { constants.AGENT_CONFIG_PATH = originalPath }()

	var wg sync.WaitGroup
	adapter := NewFluentBitAdapter(&wg)
	adapter.isRunning = false

	// Since fluent-bit binary may not be available in test environment,
	// we expect an error but we're testing the flow
	err = adapter.UpdateConfig()
	// We expect error because fluent-bit binary is not available in test
	assert.Error(t, err)
}

func TestFluentBitAdapter_UpdateConfig_WithReload(t *testing.T) {
	mockReloadResp := FluentBitReloadResponse{
		Reload: "done",
		Status: 0,
	}

	// Create mock HTTP server for reload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/reload":
			if r.Method == "POST" {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(mockReloadResp)
			}
		case "/api/v1/health":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		}
	}))
	defer server.Close()

	var wg sync.WaitGroup
	adapter := NewFluentBitAdapter(&wg)
	adapter.baseURL = server.URL
	adapter.isRunning = true

	err := adapter.UpdateConfig()
	assert.NoError(t, err)
}

func TestFluentBitAdapter_StopAgent_NotRunning(t *testing.T) {
	var wg sync.WaitGroup
	adapter := NewFluentBitAdapter(&wg)
	adapter.isRunning = false

	err := adapter.StopAgent()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not currently running")
}

func TestFluentBitAdapter_StartAgent_AlreadyRunning(t *testing.T) {
	var wg sync.WaitGroup
	adapter := NewFluentBitAdapter(&wg)
	adapter.isRunning = true

	err := adapter.StartAgent()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")
}

func TestFluentBitAdapter_GracefulShutdown(t *testing.T) {
	var wg sync.WaitGroup
	adapter := NewFluentBitAdapter(&wg)
	adapter.isRunning = false // Not actually running

	err := adapter.GracefulShutdown()
	// Should not error even if not running (error from StopAgent is only logged, not returned)
	assert.NoError(t, err)
}

func TestFluentBitAdapter_ContextCancellation(t *testing.T) {
	var wg sync.WaitGroup
	adapter := NewFluentBitAdapter(&wg)

	// Verify context is not cancelled initially
	select {
	case <-adapter.ctx.Done():
		t.Fatal("context should not be cancelled initially")
	default:
		// Expected
	}

	// Cancel the context
	adapter.cancel()

	// Verify context is cancelled
	select {
	case <-adapter.ctx.Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Fatal("context should be cancelled")
	}

	assert.Equal(t, context.Canceled, adapter.ctx.Err())
}

func TestFluentBitLogWriter(t *testing.T) {
	writer := &fluentBitLogWriter{prefix: "[test] "}

	n, err := writer.Write([]byte("test message\n"))
	assert.NoError(t, err)
	assert.Equal(t, 13, n) // Length of "test message\n"
}

func TestFluentBitAdapter_StartProcess_ConfigNotFound(t *testing.T) {
	// Save original and restore after test
	originalPath := constants.AGENT_CONFIG_PATH
	constants.AGENT_CONFIG_PATH = "/nonexistent/config.yaml"
	defer func() { constants.AGENT_CONFIG_PATH = originalPath }()

	var wg sync.WaitGroup
	adapter := NewFluentBitAdapter(&wg)

	err := adapter.Initialize()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config file not found")
}

func TestFluentBitAdapter_MultipleInitialize(t *testing.T) {
	var wg sync.WaitGroup
	adapter := NewFluentBitAdapter(&wg)
	adapter.isRunning = true

	err := adapter.Initialize()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already initialized")
}

// Integration-style test (requires fluent-bit to be installed)
func TestFluentBitAdapter_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Check if fluent-bit is available
	if _, err := os.Stat("/usr/local/bin/fluent-bit"); os.IsNotExist(err) {
		if _, err := os.Stat("/opt/fluent-bit/bin/fluent-bit"); os.IsNotExist(err) {
			t.Skip("fluent-bit not installed, skipping integration test")
		}
	}

	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `service:
  http_server: on
  http_listen: 127.0.0.1
  http_port: 2020
  hot_reload: on
  log_level: info

pipeline:
  inputs:
    - name: dummy
      dummy: '{"message":"test"}'
      rate: 1
  
  outputs:
    - name: stdout
      match: '*'
      format: json_lines
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Save original and restore after test
	originalPath := constants.AGENT_CONFIG_PATH
	constants.AGENT_CONFIG_PATH = configPath
	defer func() { constants.AGENT_CONFIG_PATH = originalPath }()

	var wg sync.WaitGroup
	adapter := NewFluentBitAdapter(&wg)

	// Test Initialize
	err = adapter.Initialize()
	if err != nil {
		t.Logf("Initialize failed (this is expected if fluent-bit is not properly installed): %v", err)
		return
	}
	defer adapter.GracefulShutdown()

	// Wait a bit for startup
	time.Sleep(3 * time.Second)

	// Test GetVersion
	version, err := adapter.GetVersion()
	assert.NoError(t, err)
	assert.NotEmpty(t, version)
	t.Logf("Fluent Bit version: %s", version)

	// Test GetMetrics
	metrics, err := adapter.GetMetrics()
	assert.NoError(t, err)
	assert.NotNil(t, metrics)

	// Test GetUptime
	uptime, err := adapter.GetUptime()
	assert.NoError(t, err)
	assert.NotNil(t, uptime)
	assert.Greater(t, uptime.UptimeSec, int64(0))

	// Test health check
	err = adapter.checkHealth()
	assert.NoError(t, err)

	// Test UpdateConfig (hot reload)
	err = adapter.UpdateConfig()
	assert.NoError(t, err)

	// Test StopAgent
	err = adapter.StopAgent()
	assert.NoError(t, err)

	// Wait for goroutines
	wg.Wait()
}
