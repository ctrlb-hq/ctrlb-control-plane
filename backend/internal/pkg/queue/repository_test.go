package queue

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	assert.NoError(t, err)

	// Create required tables
	_, err = db.Exec(`
		CREATE TABLE aggregated_agent_metrics (
			agent_id TEXT PRIMARY KEY,
			logs_rate_sent REAL,
			traces_rate_sent REAL,
			metrics_rate_sent REAL,
			data_sent_bytes REAL,
			data_received_bytes REAL,
			status TEXT,
			updated_at INTEGER
		);
		CREATE TABLE realtime_agent_metrics (
			agent_id TEXT,
			logs_rate_sent REAL,
			traces_rate_sent REAL,
			metrics_rate_sent REAL,
			data_sent_bytes REAL,
			data_received_bytes REAL,
			cpu_utilization REAL,
			memory_utilization REAL,
			timestamp INTEGER
		);
		CREATE TABLE agents (
			id TEXT PRIMARY KEY,
			hostname TEXT,
			ip TEXT
		);
	`)
	assert.NoError(t, err)
	return db
}

func TestUpdateAgentMetricsInDB(t *testing.T) {
	db := setupTestDB(t)
	repo := NewQueueRepository(db)

	agg := AggregatedAgentMetrics{
		AgentID:           "agent-1",
		LogsRateSent:      10.5,
		TracesRateSent:    5.2,
		MetricsRateSent:   7.3,
		DataSentBytes:     2048,
		DataReceivedBytes: 1024,
		Status:            "connected",
		UpdatedAt:         time.Now().Unix(),
	}
	rt := RealtimeAgentMetrics{
		AgentID:           "agent-1",
		LogsRateSent:      10.5,
		TracesRateSent:    5.2,
		MetricsRateSent:   7.3,
		DataSentBytes:     2048,
		DataReceivedBytes: 1024,
		CPUUtilization:    55.5,
		MemoryUtilization: 66.6,
		Timestamp:         time.Now().Unix(),
	}

	err := repo.UpdateAgentMetricsInDB(agg, rt)
	assert.NoError(t, err)
}

func TestUpdateAgentStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := NewQueueRepository(db)

	// Insert initial aggregated row
	_, err := db.Exec(`
		INSERT INTO aggregated_agent_metrics (agent_id, logs_rate_sent, traces_rate_sent, metrics_rate_sent, data_sent_bytes, data_received_bytes, status, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"agent-2", 0, 0, 0, 0, 0, "unknown", time.Now().Unix())
	assert.NoError(t, err)

	repo.UpdateAgentStatus("agent-2", "connected")

	var status string
	err = db.QueryRow(`SELECT status FROM aggregated_agent_metrics WHERE agent_id = ?`, "agent-2").Scan(&status)
	assert.NoError(t, err)
	assert.Equal(t, "connected", status)
}

func TestRefreshMonitoring(t *testing.T) {
	db := setupTestDB(t)
	repo := NewQueueRepository(db)

	// Insert test data
	_, err := db.Exec(`INSERT INTO agents (id, hostname, ip) VALUES (?, ?, ?)`, "agent-3", "host-3", "127.0.0.1")
	assert.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO aggregated_agent_metrics (agent_id, logs_rate_sent, traces_rate_sent, metrics_rate_sent, data_sent_bytes, data_received_bytes, status, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, "agent-3", 0, 0, 0, 0, 0, "connected", time.Now().Unix())
	assert.NoError(t, err)

	agents, err := repo.RefreshMonitoring()
	assert.NoError(t, err)
	assert.Len(t, agents, 1)
	assert.Equal(t, "agent-3", agents[0].AgentID)
	assert.Equal(t, "connected", agents[0].CurrentStatus)
}
