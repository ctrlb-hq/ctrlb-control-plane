package fluentbit

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
	"unsafe"

	"github.com/ctrlb-hq/ctrlb-collector/internal/constants"
	"github.com/ctrlb-hq/ctrlb-collector/internal/helper"
)

type FluentBitAdapter struct {
	fluentbitCtx *FlbLibCtx
	isActive     bool
	mu           *sync.Mutex
	wg           *sync.WaitGroup
}

func NewFluentBitAdapter(wg *sync.WaitGroup) *FluentBitAdapter {
	return &FluentBitAdapter{wg: wg}
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
