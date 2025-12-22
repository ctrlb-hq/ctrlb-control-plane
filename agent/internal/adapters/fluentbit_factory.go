package adapters

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/constants"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/pkg/logger"
)

// startProcess is the internal method to start the Fluent Bit process
// Must be called with mutex locked
func (a *FluentBitAdapter) startProcess() error {
	// Verify config file exists
	if _, err := os.Stat(constants.AGENT_CONFIG_PATH); err != nil {
		return fmt.Errorf("config file not found at %s: %w", constants.AGENT_CONFIG_PATH, err)
	}

	// Find fluent-bit executable
	fluentBitPath, err := exec.LookPath("fluent-bit")
	if err != nil {
		// Try common installation paths
		possiblePaths := []string{
			"/opt/fluent-bit/bin/fluent-bit",
			"/usr/local/bin/fluent-bit",
			"/usr/bin/fluent-bit",
			"./fluent-bit",
		}
		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				fluentBitPath = path
				break
			}
		}
		if fluentBitPath == "" {
			return fmt.Errorf("fluent-bit executable not found in PATH or common locations")
		}
	}

	// Create command with config file
	cmd := exec.CommandContext(a.ctx, fluentBitPath, "-c", constants.AGENT_CONFIG_PATH)

	// Set up stdout and stderr logging
	cmd.Stdout = &fluentBitLogWriter{prefix: "[fluent-bit] ", isStderr: false}
	cmd.Stderr = &fluentBitLogWriter{prefix: "[fluent-bit] ", isStderr: true}

	// Start the process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start fluent-bit process: %w", err)
	}

	a.process = cmd
	a.isRunning = true
	a.processExited = make(chan struct{})

	// Monitor process in goroutine
	a.wg.Add(1)
	go a.monitorProcess()

	// Wait for Fluent Bit to be ready
	if err := a.waitForReady(); err != nil {
		// If startup fails, kill the process
		if a.process != nil && a.process.Process != nil {
			_ = a.process.Process.Kill()
		}
		a.isRunning = false
		return fmt.Errorf("fluent-bit failed to become ready: %w", err)
	}

	logger.Logger.Sugar().Infof("Fluent Bit process started with PID: %d", cmd.Process.Pid)
	return nil
}

// monitorProcess watches the Fluent Bit process and handles unexpected exits
func (a *FluentBitAdapter) monitorProcess() {
	defer a.wg.Done()

	// Wait for process to exit
	err := a.process.Wait()

	a.mu.Lock()
	a.isRunning = false
	close(a.processExited)
	a.mu.Unlock()

	if err != nil {
		logger.Logger.Sugar().Errorf("Fluent Bit process exited with error: %v", err)
	} else {
		logger.Logger.Info("Fluent Bit process exited cleanly")
	}
}

// waitForReady polls the health endpoint until Fluent Bit is ready or timeout
func (a *FluentBitAdapter) waitForReady() error {
	logger.Logger.Info("Waiting for Fluent Bit to become ready...")

	timeout := time.After(startupTimeout)
	ticker := time.NewTicker(healthCheckInterval)
	defer ticker.Stop()

	for attempt := 1; attempt <= healthCheckRetry; attempt++ {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for fluent-bit to start after %v", startupTimeout)
		case <-a.processExited:
			return fmt.Errorf("fluent-bit process exited during startup")
		case <-ticker.C:
			if err := a.checkHealth(); err == nil {
				logger.Logger.Sugar().Infof("Fluent Bit is ready (attempt %d/%d)", attempt, healthCheckRetry)
				return nil
			}
			logger.Logger.Sugar().Debugf("Health check failed (attempt %d/%d), retrying...", attempt, healthCheckRetry)
		}
	}

	return fmt.Errorf("fluent-bit failed health checks after %d attempts", healthCheckRetry)
}

// checkHealth performs a health check on the running Fluent Bit instance
func (a *FluentBitAdapter) checkHealth() error {
	req, err := http.NewRequestWithContext(a.ctx, "GET", a.baseURL+apiV1Health, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("health check failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// checkHealthWithRetry performs health check with retries and exponential backoff
func (a *FluentBitAdapter) checkHealthWithRetry(maxRetries int, initialDelay time.Duration) error {
	var lastErr error
	delay := initialDelay

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			logger.Logger.Sugar().Infof("Retrying health check (attempt %d/%d) after %v...", i+1, maxRetries, delay)
			time.Sleep(delay)
			// Exponential backoff with cap at 10 seconds
			delay *= 2
			if delay > 10*time.Second {
				delay = 10 * time.Second
			}
		}

		if err := a.checkHealth(); err != nil {
			lastErr = err
			logger.Logger.Sugar().Warnf("Health check failed (attempt %d/%d): %v", i+1, maxRetries, err)
			continue
		}

		// Health check succeeded
		if i > 0 {
			logger.Logger.Sugar().Infof("Health check succeeded after %d attempts", i+1)
		}
		return nil
	}

	return fmt.Errorf("health check failed after %d attempts: %w", maxRetries, lastErr)
}

// getVersionFromBinary gets version by executing fluent-bit --version
func (a *FluentBitAdapter) getVersionFromBinary() (string, error) {
	fluentBitPath, err := exec.LookPath("fluent-bit")
	if err != nil {
		// Try common paths
		possiblePaths := []string{
			"/opt/fluent-bit/bin/fluent-bit",
			"/usr/local/bin/fluent-bit",
			"/usr/bin/fluent-bit",
		}
		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				fluentBitPath = path
				break
			}
		}
		if fluentBitPath == "" {
			return "", fmt.Errorf("fluent-bit executable not found")
		}
	}

	cmd := exec.Command(fluentBitPath, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get fluent-bit version: %w", err)
	}

	// Parse version from output (format: "Fluent Bit v3.2.0")
	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")
	if len(lines) > 0 {
		parts := strings.Fields(lines[0])
		for i, part := range parts {
			if part == "Bit" && i+1 < len(parts) {
				version := strings.TrimPrefix(parts[i+1], "v")
				return version, nil
			}
		}
	}

	return "", fmt.Errorf("failed to parse version from output: %s", outputStr)
}

// fluentBitLogWriter is a custom writer for Fluent Bit process output
type fluentBitLogWriter struct {
	prefix   string
	isStderr bool
}

func (w *fluentBitLogWriter) Write(p []byte) (n int, err error) {
	msg := strings.TrimSpace(string(p))
	if msg != "" {
		// For stderr, check if it's actually an error or just startup/info messages
		if w.isStderr {
			msgLower := strings.ToLower(msg)
			// Check if it contains actual error indicators
			if strings.Contains(msgLower, "[error]") || 
			   strings.Contains(msgLower, "failed") || 
			   strings.Contains(msgLower, "cannot") ||
			   strings.Contains(msgLower, "error:") {
				logger.Logger.Error(w.prefix + msg)
			} else {
				// It's just informational output on stderr (banner, copyright, etc.)
				logger.Logger.Debug(w.prefix + msg)
			}
		} else {
			logger.Logger.Info(w.prefix + msg)
		}
	}
	return len(p), nil
}
