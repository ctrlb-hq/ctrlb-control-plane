package queue

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
)

type QueueRepository struct {
	db *sql.DB
}

func NewQueueRepository(db *sql.DB) *QueueRepository {
	return &QueueRepository{
		db: db,
	}
}

func (q *QueueRepository) UpdateAgentMetricsInDB(agg AggregatedAgentMetrics, rt RealtimeAgentMetrics) error {
	// Upsert for aggregated_agent_metrics
	_, err := q.db.Exec(`
		INSERT INTO aggregated_agent_metrics 
		(agent_id, logs_rate_sent, traces_rate_sent, metrics_rate_sent, data_sent_bytes, data_received_bytes, status, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(agent_id) DO UPDATE SET 
			logs_rate_sent = EXCLUDED.logs_rate_sent,
			traces_rate_sent = EXCLUDED.traces_rate_sent,
			metrics_rate_sent = EXCLUDED.metrics_rate_sent,
			data_sent_bytes = EXCLUDED.data_sent_bytes,
			data_received_bytes = EXCLUDED.data_received_bytes,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`, agg.AgentID, agg.LogsRateSent, agg.TracesRateSent, agg.MetricsRateSent, agg.DataSentBytes, agg.DataReceivedBytes, agg.Status, agg.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to upsert aggregated metrics: %w", err)
	}

	// Insert for realtime_agent_metrics
	_, err = q.db.Exec(`
		INSERT INTO realtime_agent_metrics 
		(agent_id, logs_rate_sent, traces_rate_sent, metrics_rate_sent, data_sent_bytes, data_received_bytes, cpu_utilization, memory_utilization, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, rt.AgentID, rt.LogsRateSent, rt.TracesRateSent, rt.MetricsRateSent, rt.DataSentBytes, rt.DataReceivedBytes, rt.CPUUtilization, rt.MemoryUtilization, rt.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to insert realtime metrics: %w", err)
	}

	return nil
}

func (q *QueueRepository) UpdateAgentStatus(agentID string, status string) error {
	_, err := q.db.Exec(`
		UPDATE aggregated_agent_metrics
		SET status = ?, updated_at = ?
		WHERE agent_id = ?
	`, status, time.Now(), agentID)

	if err != nil {
		utils.Logger.Sugar().Errorf("Failed to update status for agent %s: %w", agentID, err)
		return err
	}
	return nil
}

// GetStaleAgents returns agent IDs that haven't been updated since cutoffTime
func (q *QueueRepository) GetStaleAgents(cutoffTime int64) ([]string, error) {
	rows, err := q.db.Query(`
		SELECT agent_id FROM aggregated_agent_metrics
		WHERE status = 'connected' AND updated_at < ?
	`, cutoffTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query stale agents: %w", err)
	}
	defer rows.Close()

	var staleAgents []string
	for rows.Next() {
		var agentID string
		if err := rows.Scan(&agentID); err != nil {
			utils.Logger.Sugar().Errorf("Error scanning stale agent row: %v", err)
			continue
		}
		staleAgents = append(staleAgents, agentID)
	}

	return staleAgents, nil
}
