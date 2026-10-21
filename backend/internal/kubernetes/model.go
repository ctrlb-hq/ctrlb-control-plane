package kubernetes

type CreateInstallationRequest struct {
	Name        string `json:"name"`
	ClusterName string `json:"cluster_name"`
	PipelineID  int    `json:"pipeline_id"`
}

type Installation struct {
	ID              string      `json:"id"`
	Name            string      `json:"name"`
	ClusterName     string      `json:"cluster_name"`
	PipelineID      int         `json:"pipeline_id"`
	CreatedBy       string      `json:"created_by"`
	CreatedAt       int64       `json:"created_at"`
	DesiredRevision string      `json:"desired_revision"`
	Collectors      []Collector `json:"collectors"`
	Rollout         Rollout     `json:"rollout"`
	tokenHash       string
	configJSON      []byte
}

type CreateInstallationResponse struct {
	Installation *Installation `json:"installation"`
	Token        string        `json:"token"`
}

type Rollout struct {
	Total        int `json:"total"`
	Pending      int `json:"pending"`
	Applied      int `json:"applied"`
	Failed       int `json:"failed"`
	Disconnected int `json:"disconnected"`
}

// Metrics have explicit units; nil distinguishes unavailable from a measured zero.
type Metrics struct {
	CPUPercent           *float64 `json:"cpu_percent,omitempty"`
	MemoryBytes          *float64 `json:"memory_bytes,omitempty"`
	LogsSentPerSecond    *float64 `json:"logs_sent_per_second,omitempty"`
	MetricsSentPerSecond *float64 `json:"metrics_sent_per_second,omitempty"`
	SpansSentPerSecond   *float64 `json:"spans_sent_per_second,omitempty"`
}

type CheckInRequest struct {
	NodeName          string   `json:"node_name"`
	NodeUID           string   `json:"node_uid"`
	Namespace         string   `json:"namespace"`
	PodName           string   `json:"pod_name"`
	PodUID            string   `json:"pod_uid"`
	PodIP             string   `json:"pod_ip"`
	Version           string   `json:"version"`
	Platform          string   `json:"platform"`
	EngineStatus      string   `json:"engine_status"`
	AppliedRevision   string   `json:"applied_revision"`
	AttemptedRevision string   `json:"attempted_revision"`
	ConfigError       string   `json:"config_error"`
	Metrics           *Metrics `json:"metrics,omitempty"`
}

type CheckInResponse struct {
	CollectorID         int64          `json:"collector_id"`
	DesiredRevision     string         `json:"desired_revision"`
	Config              map[string]any `json:"config,omitempty"`
	PollIntervalSeconds int            `json:"poll_interval_seconds"`
	ServerTime          int64          `json:"server_time"`
}

type Collector struct {
	ID                int64    `json:"id"`
	InstallationID    string   `json:"installation_id"`
	NodeName          string   `json:"node_name"`
	NodeUID           string   `json:"node_uid"`
	Namespace         string   `json:"namespace"`
	PodName           string   `json:"pod_name"`
	PodUID            string   `json:"pod_uid"`
	PodIP             string   `json:"pod_ip"`
	Version           string   `json:"version"`
	Platform          string   `json:"platform"`
	LastSeen          int64    `json:"last_seen"`
	ConnectionStatus  string   `json:"connection_status"`
	EngineStatus      string   `json:"engine_status"`
	AppliedRevision   string   `json:"applied_revision"`
	AttemptedRevision string   `json:"attempted_revision"`
	ConfigError       string   `json:"config_error"`
	RolloutStatus     string   `json:"rollout_status"`
	Metrics           *Metrics `json:"metrics,omitempty"`
	MetricsUpdatedAt  *int64   `json:"metrics_updated_at,omitempty"`
}
