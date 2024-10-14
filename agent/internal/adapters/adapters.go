package adapters

// Design decision: Adapter wont be responsible to write file to disk. It would read config from disk
type Adapter interface {
	Initialize() error
	UpdateConfig() error
	StartAgent() error
	StopAgent() error
	GracefulShutdown() error
	GetUptime(string) (map[string]interface{}, error)
	CurrentStatus(string) (map[string]string, error)
}
