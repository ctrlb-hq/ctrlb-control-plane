package frontendpipeline_test

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	frontendpipeline "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontend/pipeline"
	"github.com/stretchr/testify/assert"
)

func setupTestRepo(t *testing.T) (*frontendpipeline.FrontendPipelineRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	return frontendpipeline.NewFrontendPipelineRepository(db), mock, func() { db.Close() }
}

func TestPipelineExists(t *testing.T) {
	repo, mock, cleanup := setupTestRepo(t)
	defer cleanup()

	mock.ExpectQuery("SELECT pipeline_id FROM pipelines WHERE pipeline_id = \\? LIMIT 1").
		WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"pipeline_id"}).AddRow(1))
	assert.True(t, repo.PipelineExists(1))

	mock.ExpectQuery("SELECT pipeline_id FROM pipelines WHERE pipeline_id = \\? LIMIT 1").
		WithArgs(999).WillReturnError(sql.ErrNoRows)
	assert.False(t, repo.PipelineExists(999))
}

func TestGetAllPipelines(t *testing.T) {
	repo, mock, cleanup := setupTestRepo(t)
	defer cleanup()

	// Use Unix timestamp (e.g., 1704067200 = 2024-01-01T00:00:00Z)
	mock.ExpectQuery("SELECT (.+) FROM pipelines p").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "agents", "incomingBytes", "outgoingBytes", "updated_at"}).
			AddRow(1, "Pipeline 1", 2, 100, 200, int64(1704067200)))

	pipelines, err := repo.GetAllPipelines()
	assert.NoError(t, err)
	assert.Len(t, pipelines, 1)
	assert.Equal(t, "Pipeline 1", pipelines[0].Name)
	assert.Equal(t, 1704067200, pipelines[0].UpdatedAt)
}

func TestGetPipelineInfo(t *testing.T) {
	repo, mock, cleanup := setupTestRepo(t)
	defer cleanup()

	createdAt := int64(1704067200) // 2024-01-01T00:00:00Z
	updatedAt := int64(1704153600) // 2024-01-02T00:00:00Z

	mock.ExpectQuery("SELECT pipeline_id, name, created_by, created_at, updated_at FROM pipelines WHERE pipeline_id = \\?").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"pipeline_id", "name", "created_by", "created_at", "updated_at"}).
			AddRow(1, "PipelineX", "admin", createdAt, updatedAt))

	info, err := repo.GetPipelineInfo(1)
	assert.NoError(t, err)
	assert.Equal(t, "PipelineX", info.Name)
	assert.Equal(t, 1, info.ID)
	assert.Equal(t, int(createdAt), info.CreatedAt)
	assert.Equal(t, int(updatedAt), info.UpdatedAt)
}

func TestGetAgentPipelineId(t *testing.T) {
	repo, mock, cleanup := setupTestRepo(t)
	defer cleanup()

	mock.ExpectQuery("SELECT pipeline_id FROM agents WHERE id = \\?").
		WithArgs("123").WillReturnRows(sqlmock.NewRows([]string{"pipeline_id"}).AddRow(42))

	id, err := repo.GetAgentPipelineId("123")
	assert.NoError(t, err)
	assert.NotNil(t, id)
	assert.Equal(t, 42, *id)

	mock.ExpectQuery("SELECT pipeline_id FROM agents WHERE id = \\?").
		WithArgs("124").WillReturnError(sql.ErrNoRows)

	id, err = repo.GetAgentPipelineId("124")
	assert.NoError(t, err)
	assert.Nil(t, id)
}

func TestAttachAgentToPipeline(t *testing.T) {
	repo, mock, cleanup := setupTestRepo(t)
	defer cleanup()

	mock.ExpectExec("UPDATE agents SET pipeline_id = \\?, pipeline_name = \\(SELECT name FROM pipelines WHERE pipeline_id = \\?\\) WHERE id = \\?").
		WithArgs(1, 1, 123).WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.AttachAgentToPipeline(1, 123)
	assert.NoError(t, err)
}

func TestDetachAgentFromPipeline(t *testing.T) {
	repo, mock, cleanup := setupTestRepo(t)
	defer cleanup()

	mock.ExpectExec("UPDATE agents SET pipeline_id = NULL, pipeline_name = NULL WHERE id = \\?").
		WithArgs(123).WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.DetachAgentFromPipeline(1, 123)
	assert.NoError(t, err)
}

func TestDeletePipeline(t *testing.T) {
	repo, mock, cleanup := setupTestRepo(t)
	defer cleanup()

	mock.ExpectExec("DELETE FROM pipelines WHERE pipeline_id = \\?").
		WithArgs(101).WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.DeletePipeline(101)
	assert.NoError(t, err)
}

func TestGetAllAgentsAttachedToPipeline(t *testing.T) {
	repo, mock, cleanup := setupTestRepo(t)
	defer cleanup()

	mock.ExpectQuery("SELECT a.id, a.name, a.version, a.pipeline_name, a.hostname, a.IP, (.+) FROM agents a").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "version", "pipeline_name", "hostname", "IP",
			"logs_rate_sent", "traces_rate_sent", "metrics_rate_sent", "status",
		}).AddRow(1, "agent-1", "v1", "p1", "host1", "10.0.0.1", 1, 2, 3, "active"))

	agents, err := repo.GetAllAgentsAttachedToPipeline(1)
	assert.NoError(t, err)
	assert.Len(t, agents, 1)
	assert.Equal(t, "agent-1", agents[0].Name)
}

func TestGetAgentInfo(t *testing.T) {
	repo, mock, cleanup := setupTestRepo(t)
	defer cleanup()

	mock.ExpectQuery("SELECT id, name, version, pipeline_name, hostname, ip FROM agents WHERE id = \\?").
		WithArgs(123).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "version", "pipeline_name", "hostname", "ip"}).
			AddRow(123, "agentX", "v2", "PipeY", "hostY", "192.168.1.1"))

	mock.ExpectQuery("SELECT status FROM aggregated_agent_metrics WHERE agent_id = \\?").
		WithArgs(123).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("running"))

	info, err := repo.GetAgentInfo(123)
	assert.NoError(t, err)
	assert.Equal(t, "agentX", info.Name)
	assert.Equal(t, "running", info.Status)
}
