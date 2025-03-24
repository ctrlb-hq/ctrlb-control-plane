package queue

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
)

type QueueRepository struct {
	db *sql.DB
}

func NewQueueRepository(db *sql.DB) *QueueRepository {
	return &QueueRepository{
		db: db,
	}
}

func (qr *QueueRepository) UpdateStatusOnly(agentID, status string) error {
	query := `UPDATE agent_metrics SET STATUS = ? UpdatedAt = ? WHERE AGENTID = ?`
	_, err := qr.db.Exec(query, status, time.Now(), agentID)

	if err != nil {
		return fmt.Errorf("error updating agent status: %v", err)
	}

	query = `UPDATE agent_status SET STATUS = ? UpdatedAt = ? WHERE AGENTID = ?`
	_, err = qr.db.Exec(query, status, time.Now(), agentID)

	if err != nil {
		return fmt.Errorf("error updating agent status: %v", err)
	}

	return nil
}

func (qr *QueueRepository) UpdateStatusRetries(agentID string, retryRemaining int, status string) error {
	query := `UPDATE agent_metrics SET STATUS = ? UpdatedAt = ? WHERE AGENTID = ?`
	_, err := qr.db.Exec(query, status, time.Now(), agentID)

	if err != nil {
		return fmt.Errorf("error updating agent status: %v", err)
	}

	query = `UPDATE agent_status SET STATUS = ? UpdatedAt = ? RetryRemaining = ? WHERE AGENTID = ?`
	_, err = qr.db.Exec(query, status, time.Now(), retryRemaining, agentID)

	if err != nil {
		return fmt.Errorf("error updating agent status: %v", err)
	}

	return nil
}

func (qr *QueueRepository) UpdateMetrics(metrics *models.AgentMetrics) error {
	query := `
		UPDATE agent_status
		SET
			CurrentStatus = ?,
			RetryRemaining = 3,
			UpdatedAt = ?
		WHERE AgentID = ?`

	_, err := qr.db.Exec(query, metrics.Status, time.Now(), metrics.AgentID)
	if err != nil {
		return fmt.Errorf("error updating agent status: %w", err)
	}

	query = `
		UPDATE agent_metrics
		SET
			Status = ?,
			ExportedDataVolume = ?,
			UptimeSeconds = ?,
			DroppedRecords = ?,
			UpdatedAt = ?
		WHERE AgentID = ?`

	_, err = qr.db.Exec(query, metrics.Status, metrics.ExportedDataVolume, metrics.UptimeSeconds,
		metrics.DroppedRecords, metrics.UpdatedAt, metrics.AgentID)

	if err != nil {
		return fmt.Errorf("error updating agent metrics: %w", err)
	}

	return nil
}
