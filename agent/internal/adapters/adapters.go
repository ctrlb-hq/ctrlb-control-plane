package adapters

// Design decision: Adapter wont be responsible to write file to disk. It would read config from disk
type Adapter interface {
	Initialize() error
	UpdateConfig() error
	StartAgent() error
	StopAgent() error
	GracefulShutdown() error
	GetUptime() (map[string]interface{}, error)
	CurrentStatus() (map[string]string, error)
}
