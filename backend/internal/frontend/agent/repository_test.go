package frontendagent

import (
	"database/sql"
	"strconv"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *FrontendAgentRepository) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	repo := NewFrontendAgentRepository(db)
	return db, mock, repo
}

func TestAgentExists(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery("SELECT id FROM agents WHERE id = ?").
		WithArgs("agent-1").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("agent-1"))

	exists := repo.AgentExists("agent-1")
	if !exists {
		t.Error("expected agent to exist")
	}
}

func TestGetAllAgents(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	agentRows := sqlmock.NewRows([]string{"id", "name", "version", "pipeline_name"}).
		AddRow(1, "agent1", "v1.0", sql.NullString{String: "pipeline1", Valid: true})

	mock.ExpectQuery("SELECT id, name, version, pipeline_name FROM agents").
		WillReturnRows(agentRows)

	mock.ExpectQuery("SELECT logs_rate_sent, traces_rate_sent, metrics_rate_sent, status FROM aggregated_agent_metrics WHERE agent_id = ?").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"logs_rate_sent", "traces_rate_sent", "metrics_rate_sent", "status"}).
			AddRow(1, 2, 3, "healthy"))

	agents, err := repo.GetAllAgents()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(agents) != 1 {
		t.Fatalf("expected 1 agent, got %d", len(agents))
	}
}

func TestGetAgent(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	agentRow := sqlmock.NewRows([]string{"id", "name", "version", "pipeline_id", "pipeline_name", "hostname", "ip", "platform"}).
		AddRow("1", "agent1", "v1.0", sql.NullInt64{Int64: 123, Valid: true}, sql.NullString{String: "pipeline1", Valid: true}, "host", "1.2.3.4", "linux")

	mock.ExpectQuery("SELECT id, name, version, pipeline_id, pipeline_name, hostname, ip, platform FROM agents WHERE id = ?").
		WithArgs("1").
		WillReturnRows(agentRow)

	mock.ExpectQuery("SELECT status FROM aggregated_agent_metrics WHERE agent_id = ?").
		WithArgs("1").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("healthy"))

	mock.ExpectQuery("SELECT key, value FROM agents_labels WHERE agent_id = ?").
		WithArgs("1").
		WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).
			AddRow("env", "prod"))

	agent, err := repo.GetAgent("1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if agent.Name != "agent1" || agent.Labels["env"] != "prod" {
		t.Errorf("unexpected agent data: %+v", agent)
	}
}

func TestDeleteAgent(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	mock.ExpectExec("DELETE FROM agents WHERE id = ?").
		WithArgs("1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.DeleteAgent("1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAgentStatus(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery("SELECT status FROM aggregated_agent_metrics WHERE agent_id = ?").
		WithArgs("1").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("healthy"))

	status := repo.AgentStatus("1")
	if status != "healthy" {
		t.Errorf("expected healthy, got %s", status)
	}
}

func TestGetHealthMetricsForGraph(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"cpu_utilization", "memory_utilization", "timestamp"}).
		AddRow(50.5, 30.2, time.Now().Unix())

	mock.ExpectQuery("SELECT cpu_utilization, memory_utilization, timestamp FROM realtime_agent_metrics WHERE agent_id = ?").
		WithArgs("1").
		WillReturnRows(rows)

	metrics, err := repo.GetHealthMetricsForGraph("1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*metrics) != 2 {
		t.Errorf("expected 2 metrics, got %d", len(*metrics))
	}
}

func TestAddLabels(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO agents_labels").
		WithArgs("1", "key1", "val1").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.AddLabels("1", map[string]string{"key1": "val1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetLatestAgentSince(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	unixTimestamp := time.Now().Unix()
	unixTimestampStr := strconv.FormatInt(unixTimestamp, 10)

	mock.ExpectQuery("SELECT id, name, registered_at FROM agents WHERE registered_at > ?").
		WithArgs(unixTimestampStr).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "registered_at"}).
			AddRow("id1", "agent1", unixTimestamp))

	agent, err := repo.GetLatestAgentSince(unixTimestampStr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if agent.Name != "agent1" {
		t.Errorf("expected agent1, got %s", agent.Name)
	}
}
