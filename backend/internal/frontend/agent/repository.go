package frontendagent

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
)

type FrontendAgentRepository struct {
	db *sql.DB
}

// NewFrontendAgentRepository creates a new FrontendAgentRepository
func NewFrontendAgentRepository(db *sql.DB) *FrontendAgentRepository {
	return &FrontendAgentRepository{db: db}
}

func (f *FrontendAgentRepository) AgentExists(id string) bool {
	var existingId string
	err := f.db.QueryRow("SELECT id FROM agents WHERE id = ? LIMIT 1", id).Scan(&existingId)
	return err == nil
}

func (f *FrontendAgentRepository) GetAllAgents() ([]models.AgentInfoHome, error) {
	var agents []models.AgentInfoHome
	row, err := f.db.Query("SELECT id, name, version, pipeline_name FROM agents")
	if err != nil {
		return nil, err
	}
	defer row.Close()

	for row.Next() {
		agent := models.AgentInfoHome{}
		var pipelineName sql.NullString
		err := row.Scan(&agent.ID, &agent.Name, &agent.Version, &pipelineName)
		if err != nil {
			return nil, err
		}

		if pipelineName.Valid {
			agent.PipelineName = pipelineName.String
		} else {
			agent.PipelineName = ""
		}
		agents = append(agents, agent)
	}

	for i := range agents {
		// Get the status of the agent
		agentStatus := f.db.QueryRow("SELECT logs_rate_sent, traces_rate_sent, metrics_rate_sent, status FROM aggregated_agent_metrics WHERE agent_id = ?", agents[i].ID)

		err := agentStatus.Scan(&agents[i].LogRate, &agents[i].TraceRate, &agents[i].MetricsRate, &agents[i].Status)
		if err != nil {
			if err == sql.ErrNoRows {
				// Optionally set default values if no metrics found
				agents[i].LogRate = 0
				agents[i].TraceRate = 0
				agents[i].MetricsRate = 0
				agents[i].Status = "unknown" // or any fallback
				continue
			}
			return nil, err
		}
	}
	return agents, nil
}

func (f *FrontendAgentRepository) GetAllUnmanagedAgents() ([]UnmanagedAgents, error) {
	var agents []UnmanagedAgents
	rows, err := f.db.Query("SELECT a.id, a.name, a.type, a.version, a.hostname, a.platform FROM agents AS a LEFT JOIN aggregated_agent_metrics AS aam ON aam.agent_id = a.id WHERE a.pipeline_id IS NULL AND (aam.status IS NULL OR aam.status != 'disconnected');")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var agent UnmanagedAgents
		if err := rows.Scan(&agent.ID, &agent.Name, &agent.Type, &agent.Version, &agent.Hostname, &agent.Platform); err != nil {
			return nil, err
		}
		agents = append(agents, agent)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return agents, nil
}

// GetAgent retrieves a specific agent by ID
func (f *FrontendAgentRepository) GetAgent(id string) (*AgentInfoWithLabels, error) {
	agent := &AgentInfoWithLabels{}
	var pipelineName sql.NullString
	var pipelineId sql.NullInt64

	err := f.db.QueryRow("SELECT id, name, version, pipeline_id, pipeline_name, hostname, ip, platform FROM agents WHERE id = ?", id).Scan(&agent.ID, &agent.Name, &agent.Version, &pipelineId, &pipelineName, &agent.Hostname, &agent.IP, &agent.Platform)
	if err != nil {
		return nil, err
	}

	if pipelineName.Valid {
		agent.PipelineID = strconv.FormatInt(pipelineId.Int64, 10)
		agent.PipelineName = pipelineName.String
	} else {
		agent.PipelineName = ""
		agent.PipelineID = ""
	}

	err = f.db.QueryRow("SELECT status FROM aggregated_agent_metrics WHERE agent_id = ?", id).Scan(&agent.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			agent.Status = "unknown"
		} else {
			return nil, err
		}
	}

	agent.Labels = make(map[string]string)
	rows, err := f.db.Query("SELECT key, value FROM agents_labels WHERE agent_id = ?", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return agent, nil
		}
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

func (f *FrontendAgentRepository) GetAgentNetworkInfoByID(agentID string) (hostname, ip string, err error) {
	query := `SELECT hostname, ip FROM agents WHERE id = ?`

	err = f.db.QueryRow(query, agentID).Scan(&hostname, &ip)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch network info for agent ID %s: %w", agentID, err)
	}

	return hostname, ip, nil
}

// DeleteAgent removes an agent by ID
func (f *FrontendAgentRepository) DeleteAgent(id string) error {
	// This will delete all related labels, metrics and extenstions as well
	if _, err := f.db.Exec("DELETE FROM agents WHERE id = ?", id); err != nil {
		return err
	}
	return nil
}

func (f *FrontendAgentRepository) AgentStatus(id string) string {
	var status string
	err := f.db.QueryRow("SELECT status FROM aggregated_agent_metrics WHERE agent_id = ?", id).Scan(&status)
	if err != nil {
		status = "unknown"
	}
	return status
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
		if key == "" || value == "" {
			return fmt.Errorf("key and value cannot be empty")
		}
		if _, err := tx.Exec("INSERT INTO agents_labels (agent_id, key, value) VALUES (?, ?, ?) ON CONFLICT(agent_id, key) DO UPDATE SET value = excluded.value", agentId, key, value); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (f *FrontendAgentRepository) GetLatestAgentSince(since string) (*LatestAgentResponse, error) {
	query := `
	SELECT id, name, registered_at
	FROM agents
	WHERE registered_at > ? AND pipeline_id IS NULL
	ORDER BY registered_at DESC
	LIMIT 1;
`

	row := f.db.QueryRow(query, since)

	var agent LatestAgentResponse
	err := row.Scan(&agent.ID, &agent.Name, &agent.RegisteredAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // no new agent found
		}
		return nil, fmt.Errorf("failed to query latest agent: %w", err)
	}

	return &agent, nil
}
