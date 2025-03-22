package frontendagent

import (
	"database/sql"
)

type FrontendAgentRepository struct {
	db *sql.DB
}

// NewFrontendAgentRepository creates a new FrontendAgentRepository
func NewFrontendAgentRepository(db *sql.DB) *FrontendAgentRepository {
	return &FrontendAgentRepository{db: db}
}

func (f *FrontendAgentRepository) GetAllAgents() ([]AgentInfoHome, error) {
	var agents []AgentInfoHome
	row, err := f.db.Query("SELECT id, name, version, pipeline_name FROM agents")
	if err != nil {
		return nil, err
	}
	defer row.Close()

	for row.Next() {
		agent := AgentInfoHome{}
		err := row.Scan(&agent.ID, &agent.Name, &agent.Version, &agent.PipelineName)
		if err != nil {
			return nil, err
		}
		agents = append(agents, agent)
	}

	for i := range agents {
		// Get the status of the agent
		agentStatus := f.db.QueryRow("SELECT logs_rate_sent, traces_rate_sent, metrics_rate_sent, status FROM aggregated_agent_metrics WHERE agent_id = ?", agents[i].ID)

		err := agentStatus.Scan(&agents[i].LogRate, &agents[i].TraceRate, &agents[i].MetricsRate, &agents[i].Status)
		if err != nil {
			return nil, err
		}
	}
	return agents, nil
}

func (f *FrontendAgentRepository) GetAllUnmanagedAgents() ([]UnmanagedAgents, error) {
	var agents []UnmanagedAgents
	row, err := f.db.Query("SELECT id, name, type, version, hostname, platform FROM agents WHERE pipeline_id IS NULL AND status = 'connected'")
	if err != nil {
		return nil, err
	}
	defer row.Close()
	return agents, nil
}

// GetAgent retrieves a specific agent by ID
func (f *FrontendAgentRepository) GetAgent(id string) (*AgentInfoWithLabels, error) {
	agent := &AgentInfoWithLabels{}

	err := f.db.QueryRow("SELECT id, name, version, pipeline_id, pipeline_name, hostname, platform FROM agents WHERE id = ?", id).Scan(&agent.ID, &agent.Name, &agent.Version, &agent.PipelineID, &agent.PipelineName, &agent.Hostname, &agent.Platform)
	if err != nil {
		return nil, err
	}

	err = f.db.QueryRow("SELECT status FROM aggregated_agent_metrics WHERE agent_id = ?", id).Scan(&agent.Status)
	if err != nil {
		return nil, err
	}

	agent.Labels = make(map[string]string)
	rows, err := f.db.Query("SELECT key, value FROM agents_labels WHERE agent_id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var key, value string
		err := rows.Scan(&key, &value)
		if err != nil {
			return nil, err
		}
		agent.Labels[key] = value
	}

	return agent, nil
}

func (f *FrontendAgentRepository) GetAgentHostname(id string) (string, error) {
	var hostname string

	err := f.db.QueryRow("SELECT hostname FROM agents WHERE id = ?", id).Scan(&hostname)
	if err != nil {
		return "", err
	}

	return hostname, nil
}

// DeleteAgent removes an agent by ID
func (f *FrontendAgentRepository) DeleteAgent(id string) error {
	// This will delete all related labels, metrics and extenstions as well
	if _, err := f.db.Exec("DELETE FROM agents WHERE id = ?", id); err != nil {
		return err
	}
	return nil
}

// GetHealthMetricsForGraph retrieves metrics for a specific agent
func (f *FrontendAgentRepository) GetHealthMetricsForGraph(id string) (*[]AgentMetrics, error) {
	rows, err := f.db.Query("SELECT cpu_utilization, memory_utilization, timestamp FROM realtime_agent_metrics WHERE agent_id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cpuDataPoints []DataPoint
	var memoryDataPoints []DataPoint

	for rows.Next() {
		var cpu, memory float64
		var timestamp int64
		err := rows.Scan(&cpu, &memory, &timestamp)
		if err != nil {
			return nil, err
		}
		cpuDataPoints = append(cpuDataPoints, DataPoint{Timestamp: timestamp, Value: cpu})
		memoryDataPoints = append(memoryDataPoints, DataPoint{Timestamp: timestamp, Value: memory})
	}

	metrics := []AgentMetrics{
		{MetricName: "cpu_utilization", DataPoints: cpuDataPoints},
		{MetricName: "memory_utilization", DataPoints: memoryDataPoints},
	}

	return &metrics, nil
}

func (f *FrontendAgentRepository) GetRateMetricsForGraph(id string) (*[]AgentMetrics, error) {
	rows, err := f.db.Query("SELECT traces_rate_sent, metrics_rate_sent, logs_rate_sent, timestamp FROM realtime_agent_metrics WHERE agent_id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracesRateDataPoints, metricsRateDataPoints, logRateDataPoints []DataPoint

	for rows.Next() {
		var traceRate, metricsRate, logsRate float64
		var timestamp int64
		err := rows.Scan(&logsRate, &traceRate, &metricsRate, &timestamp)
		if err != nil {
			return nil, err
		}
		tracesRateDataPoints = append(tracesRateDataPoints, DataPoint{Timestamp: timestamp, Value: traceRate})
		logRateDataPoints = append(logRateDataPoints, DataPoint{Timestamp: timestamp, Value: logsRate})
		metricsRateDataPoints = append(metricsRateDataPoints, DataPoint{Timestamp: timestamp, Value: metricsRate})
	}

	metrics := []AgentMetrics{
		{MetricName: "traces_rate_sent", DataPoints: tracesRateDataPoints},
		{MetricName: "metrics_rate_sent", DataPoints: metricsRateDataPoints},
		{MetricName: "log_rate_sent", DataPoints: logRateDataPoints},
	}

	return &metrics, nil
}

func (f *FrontendAgentRepository) AddLabels(agentId string, labels map[string]string) error {
	tx, err := f.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for key, value := range labels {
		if _, err := tx.Exec("INSERT INTO agents_labels (agent_id, key, value) VALUES (?, ?, ?)", agentId, key, value); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
