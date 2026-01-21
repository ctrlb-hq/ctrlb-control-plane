package queue

// AgentQueueRepositoryInterface defines the interface for agent metrics storage
type AgentQueueRepositoryInterface interface {
	UpdateAgentMetricsInDB(agg AggregatedAgentMetrics, rt RealtimeAgentMetrics) error
	UpdateAgentStatus(agentID string, status string) error
	GetStaleAgents(cutoffTime int64) ([]string, error)
}
