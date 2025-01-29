package adapters

import "github.com/ctrlb-hq/ctrlb-collector/agent/internal/models"

// Design decision: Adapter wont be responsible to write file to disk. It would read config from disk
type Adapter interface {
	Initialize() error
	UpdateConfig() error
	StartAgent() error
	StopAgent() error
	GracefulShutdown() error
	CurrentStatus() (*models.AgentMetrics, error)
	GetVersion() (string, error)
}
