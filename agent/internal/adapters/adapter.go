package adapters

import (
	"fmt"
	"sync"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/models"
)

// Adapter defines the interface for different telemetry collectors
type Adapter interface {
	Initialize() error
	StartAgent() error
	StopAgent() error
	UpdateConfig() error
	GracefulShutdown() error
	CurrentStatus() (*models.AgentMetrics, error)
	GetVersion() (string, error)
}

func NewAdapter(wg *sync.WaitGroup, agentType string) (Adapter, error) {
	if agentType == "otel" || agentType == "" {
		return NewOTELAdapter(wg), nil
	}
	return nil, fmt.Errorf("unsupported agent type: %s", agentType)
}
