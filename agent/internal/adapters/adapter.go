package adapters

import (
	"fmt"
	"sync"
)

// Adapter defines the interface for different telemetry collectors
type Adapter interface {
	Initialize() error
	StartAgent() error
	StopAgent() error
	UpdateConfig() error
	GracefulShutdown() error
	GetVersion() (string, error)
	ValidateConfigInMemory(data *map[string]any) error
	ValidateConfigOnDisk(data *map[string]any) error
	GetMetrics() (map[string]any, error)
}

func NewAdapter(wg *sync.WaitGroup, agentType string) (Adapter, error) {
	if agentType == "otel" || agentType == "" {
		return NewOTELAdapter(wg), nil
	}
	if agentType == "fluent-bit" || agentType == "fluentbit" || agentType == "fb" {
		return NewFluentBitAdapter(wg), nil
	}
	return nil, fmt.Errorf("unsupported agent type: %s", agentType)
}
