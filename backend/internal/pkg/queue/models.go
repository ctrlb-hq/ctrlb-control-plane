package queue

// AggregatedAgentMetrics represents aggregated metrics stored in DB
type AggregatedAgentMetrics struct {
	AgentID           string
	LogsRateSent      float64
	TracesRateSent    float64
	MetricsRateSent   float64
	DataSentBytes     float64
	DataReceivedBytes float64
	Status            string
	UpdatedAt         int64
}

// RealtimeAgentMetrics represents point-in-time metrics for graphing
type RealtimeAgentMetrics struct {
	AgentID           string
	LogsRateSent      float64
	TracesRateSent    float64
	MetricsRateSent   float64
	DataSentBytes     float64
	DataReceivedBytes float64
	CPUUtilization    float64
	MemoryUtilization float64
	Timestamp         int64
}
