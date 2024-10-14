package fluentbit

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/ctrlb-hq/ctrlb-collector/internal/constants"
	"github.com/ctrlb-hq/ctrlb-collector/internal/helper"
	"github.com/ctrlb-hq/ctrlb-collector/internal/utils"
	"github.com/prometheus/common/expfmt"
)

type FluentBitAdapter struct {
	fluentbitCtx *FlbLibCtx
	isActive     bool
	mu           *sync.Mutex
	wg           *sync.WaitGroup
	baseUrl      string
}

func NewFluentBitAdapter(wg *sync.WaitGroup) *FluentBitAdapter {
	return &FluentBitAdapter{wg: wg, baseUrl: "http://0.0.0.0:2020"}
}

func (f *FluentBitAdapter) Initialize() error {
	f.fluentbitCtx = f.flbCreate()
	f.mu = &sync.Mutex{}
	if f.fluentbitCtx == nil {
		return fmt.Errorf("failed to create Fluent Bit context")
	}

	configFile := f.flbCString(constants.AGENT_CONFIG_PATH)
	configFile = f.flbStrdup(configFile)
	defer f.flbFreePointer(unsafe.Pointer(configFile))

	f.flbReadFromFile(configFile)

	ret := f.flbSetHTTPDefaultService()
	if ret != 0 {
		return fmt.Errorf("failed to set http service in Fluent Bit, required for agent metrics")
	}

	ret = f.flbStart()
	if ret != 0 {
		return fmt.Errorf("failed to start Fluent Bit")
	}

	f.isActive = true

	return nil
}

func (f *FluentBitAdapter) StartAgent() error {
	if f.isActive {
		return fmt.Errorf("fluent-bit instance already running")
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	f.fluentbitCtx = f.flbCreate()
	if f.fluentbitCtx == nil {
		return fmt.Errorf("failed to create Fluent Bit context")
	}

	configFile := f.flbCString(constants.AGENT_CONFIG_PATH)
	configFile = f.flbStrdup(configFile)
	defer f.flbFreePointer(unsafe.Pointer(configFile))

	f.flbReadFromFile(configFile)
	ret := f.flbStart()
	if ret != 0 {
		return fmt.Errorf("failed to start Fluent Bit")
	}

	f.isActive = true
	return nil
}

func (f *FluentBitAdapter) StopAgent() error {
	if !f.isActive {
		return fmt.Errorf("fluent-bit instance not currently running")
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	if f.fluentbitCtx == nil {
		return fmt.Errorf("fluent-bit context not initialized")
	}

	ret := f.flbStop()
	if ret != 0 {
		return fmt.Errorf("failed to stop Fluent Bit")
	}

	f.isActive = false
	f.flbDestroy()
	return nil
}

func (f *FluentBitAdapter) UpdateConfig() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	newContext := f.flbCreate()
	if f.fluentbitCtx == nil {
		return fmt.Errorf("failed to create Fluent Bit context")
	}

	configFile := f.flbCString(constants.AGENT_CONFIG_PATH)
	configFile = f.flbStrdup(configFile)
	defer f.flbFreePointer(unsafe.Pointer(configFile))

	f.flbReadFromFile(configFile)

	oldContext := f.fluentbitCtx
	f.fluentbitCtx = newContext

	f.flbDestroyContext(oldContext)

	ret := f.flbStop()
	if ret != 0 {
		return fmt.Errorf("failed to stop Fluent Bit")
	}
	cFilePath := f.flbCString(constants.AGENT_CONFIG_PATH)
	cFilePath = f.flbStrdup(cFilePath)
	defer f.flbFreePointer(unsafe.Pointer(cFilePath))

	f.flbReadFromFile(cFilePath)
	log.Printf("Config updated. Restarting fluent-bit")

	ret = f.flbStart()
	if ret != 0 {
		return fmt.Errorf("failed to restart Fluent Bit")
	}
	return nil
}

func (f *FluentBitAdapter) GracefulShutdown() error {
	log.Println("Initiating Server shutdown...")

	helper.ShutdownServer(f.wg)

	log.Printf("Initiating graceful shutdown of Fluent Bit...")

	f.StopAgent()

	log.Printf("Waiting for all goroutines to finish...")
	done := make(chan struct{})
	f.wg.Wait()
	close(done)

	select {
	case <-done:
		log.Printf("All goroutines finished successfully")

	case <-time.After(20 * time.Second):
		return fmt.Errorf("Timed out waiting for goroutines to finish")
	}

	log.Printf("FluentBit has been gracefully shutdown")
	os.Exit(0)
	return nil
}

func (f *FluentBitAdapter) GetUptime() (map[string]interface{}, error) {
	// Define the URL of the endpoint
	url := f.baseUrl + "/api/v1/uptime"

	client := &http.Client{Timeout: 10 * time.Second}

	// Make the GET request
	resp, err := client.Get(url)
	if err != nil {
		// Error in pinging the address, return DOWN and 0 uptime
		return map[string]interface{}{
			"status": "DOWN",
			"uptime": 0,
		}, err
	}
	defer resp.Body.Close()

	// Check if status code is not OK
	if resp.StatusCode != http.StatusOK {
		return map[string]interface{}{
			"status": "DOWN",
			"uptime": 0,
		}, errors.New("failed to get valid response from server")
	}

	// Decode the JSON response into a map
	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return map[string]interface{}{
			"status": "DOWN",
			"uptime": 0,
		}, err
	}

	// Extract uptime_sec from the map
	uptimeSec, ok := data["uptime_sec"].(float64)
	if !ok {
		// If we can't get uptime_sec, return DOWN and 0 uptime
		return map[string]interface{}{
			"status": "DOWN",
			"uptime": 0,
		}, errors.New("failed to parse uptime_sec")
	}

	// Determine status based on uptime_sec
	status := "DOWN"
	if uptimeSec > 0 {
		status = "UP"
	}

	// Return the result with status and uptime
	return map[string]interface{}{
		"status": status,
		"uptime": int(uptimeSec), // Convert float64 to int
	}, nil
}

func (f *FluentBitAdapter) CurrentStatus() (map[string]string, error) {
	url := f.baseUrl + "/api/v1/metrics/prometheus"
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

	parsedMetrics := utils.ExtractFluentBitStatusFromPrometheus(metrics)

	status := make(map[string]string)

	status["Uptime"] = fmt.Sprintf("%.0f", parsedMetrics.Uptime)
	status["ExportedDataVolume"] = fmt.Sprintf("%.0f", parsedMetrics.ExportedDataVolume)
	status["DroppedRecords"] = fmt.Sprintf("%.0f", parsedMetrics.DroppedRecords)

	if parsedMetrics.Uptime > 0 {
		status["Status"] = "ON"
	} else {
		status["Status"] = "OFF"
	}

	return status, nil
}
