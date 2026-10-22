package frontendpipeline_test

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"

	database "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/db"
	frontendpipeline "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontend/pipeline"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/kubernetes"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
	"github.com/stretchr/testify/require"
)

func deletionDatabase(t *testing.T) (*sql.DB, *frontendpipeline.FrontendPipelineRepository, *kubernetes.Service) {
	t.Helper()
	db, err := database.DBInit(filepath.Join(t.TempDir(), "deletion.db"))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	for id := 1; id <= 2; id++ {
		_, err := db.Exec(`INSERT INTO pipelines (pipeline_id, name, created_by, config_json)
			VALUES (?, ?, 'tester', '{"receivers":{"otlp":{}},"exporters":{"debug":{}},"service":{"pipelines":{}}}')`, id, fmt.Sprintf("pipeline-%d", id))
		require.NoError(t, err)
		_, err = db.Exec(`INSERT INTO agents (id, name, pipeline_id, pipeline_name)
			VALUES (?, ?, ?, ?)`, 100+id, fmt.Sprintf("vm-%d", id), id, fmt.Sprintf("pipeline-%d", id))
		require.NoError(t, err)
		for _, role := range []string{"receiver", "exporter"} {
			_, err = db.Exec(`INSERT INTO pipeline_components (pipeline_id, component_role, component_name, supported_signals)
				VALUES (?, ?, ?, 'logs')`, id, role, role)
			require.NoError(t, err)
		}
		_, err = db.Exec(`INSERT INTO pipeline_component_edges (pipeline_id, child_component_id, parent_component_id)
			SELECT ?, child.component_id, parent.component_id FROM pipeline_components child, pipeline_components parent
			WHERE child.pipeline_id = ? AND parent.pipeline_id = ?
			AND child.component_role = 'exporter' AND parent.component_role = 'receiver'`, id, id, id)
		require.NoError(t, err)
	}
	return db, frontendpipeline.NewFrontendPipelineRepository(db), kubernetes.NewService(kubernetes.NewRepository(db))
}

func deletionInstallation(t *testing.T, service *kubernetes.Service, pipelineID int, name string) *kubernetes.CreateInstallationResponse {
	t.Helper()
	i, err := service.Create(context.Background(), kubernetes.CreateInstallationRequest{
		Name: name, ClusterName: name + "-cluster", PipelineID: pipelineID,
	}, "tester")
	require.NoError(t, err)
	return i
}

func deletionCollector(t *testing.T, db *sql.DB, service *kubernetes.Service, installation *kubernetes.CreateInstallationResponse, node string) int64 {
	t.Helper()
	response, err := service.CheckIn(context.Background(), installation.Installation.ID, installation.Token,
		kubernetes.CheckInRequest{NodeName: node, EngineStatus: "waiting"})
	require.NoError(t, err)
	for _, query := range []string{
		"INSERT INTO agents_labels (agent_id, key, value) VALUES (?, 'environment', 'test')",
		"INSERT INTO aggregated_agent_metrics (agent_id, status) VALUES (?, 'connected')",
		"INSERT INTO realtime_agent_metrics (agent_id) VALUES (?)",
		"INSERT INTO extensions (agent_id, extension_name) VALUES (?, 'health_check')",
	} {
		_, err = db.Exec(query, response.CollectorID)
		require.NoError(t, err)
	}
	return response.CollectorID
}

func deletionCount(t *testing.T, db *sql.DB, query string, args ...any) int {
	t.Helper()
	var count int
	require.NoError(t, db.QueryRow(query, args...).Scan(&count))
	return count
}

func requireVMAssignment(t *testing.T, db *sql.DB, agentID int, pipelineID sql.NullInt64, pipelineName sql.NullString) {
	t.Helper()
	var actualID sql.NullInt64
	var actualName sql.NullString
	require.NoError(t, db.QueryRow("SELECT pipeline_id, pipeline_name FROM agents WHERE id = ?", agentID).Scan(&actualID, &actualName))
	require.Equal(t, pipelineID, actualID)
	require.Equal(t, pipelineName, actualName)
}

func TestDeleteVMPipelineDetachesAgentsAndDeletesGraph(t *testing.T) {
	db, repo, _ := deletionDatabase(t)
	require.NoError(t, frontendpipeline.NewFrontendPipelineService(repo).DeletePipeline(1))

	requireVMAssignment(t, db, 101, sql.NullInt64{}, sql.NullString{})
	requireVMAssignment(t, db, 102, sql.NullInt64{Int64: 2, Valid: true}, sql.NullString{String: "pipeline-2", Valid: true})
	require.Equal(t, 1, deletionCount(t, db, "SELECT COUNT(*) FROM pipelines"))
	require.Equal(t, 0, deletionCount(t, db, "SELECT COUNT(*) FROM pipeline_components WHERE pipeline_id = 1"))
	require.Equal(t, 0, deletionCount(t, db, "SELECT COUNT(*) FROM pipeline_component_edges WHERE pipeline_id = 1"))
	require.Equal(t, 2, deletionCount(t, db, "SELECT COUNT(*) FROM pipeline_components WHERE pipeline_id = 2"))
	require.Equal(t, 1, deletionCount(t, db, "SELECT COUNT(*) FROM pipeline_component_edges WHERE pipeline_id = 2"))
	require.ErrorIs(t, repo.DeletePipeline(1), utils.ErrPipelineDoesNotExists)
}

func TestDeletePipelineRemovesAllInstallationsCollectorsAndCredentials(t *testing.T) {
	db, repo, service := deletionDatabase(t)
	first := deletionInstallation(t, service, 1, "first")
	second := deletionInstallation(t, service, 1, "second")
	unenrolled := deletionInstallation(t, service, 1, "unenrolled")
	unrelated := deletionInstallation(t, service, 2, "unrelated")
	deletedIDs := []int64{
		deletionCollector(t, db, service, first, "worker-1"),
		deletionCollector(t, db, service, first, "worker-2"),
		deletionCollector(t, db, service, second, "worker-1"),
	}
	unrelatedID := deletionCollector(t, db, service, unrelated, "worker-1")

	require.NoError(t, frontendpipeline.NewFrontendPipelineService(repo).DeletePipeline(1))

	requireVMAssignment(t, db, 101, sql.NullInt64{}, sql.NullString{})
	requireVMAssignment(t, db, 102, sql.NullInt64{Int64: 2, Valid: true}, sql.NullString{String: "pipeline-2", Valid: true})
	require.Equal(t, 0, deletionCount(t, db, "SELECT COUNT(*) FROM pipelines WHERE pipeline_id = 1"))
	require.Equal(t, 1, deletionCount(t, db, "SELECT COUNT(*) FROM pipelines WHERE pipeline_id = 2"))
	require.Equal(t, 0, deletionCount(t, db, "SELECT COUNT(*) FROM kubernetes_installations WHERE pipeline_id = 1"))
	require.Equal(t, 1, deletionCount(t, db, "SELECT COUNT(*) FROM kubernetes_installations WHERE pipeline_id = 2"))
	require.Equal(t, 1, deletionCount(t, db, "SELECT COUNT(*) FROM kubernetes_collectors"))
	for _, table := range []string{"agents", "kubernetes_collectors", "agents_labels", "aggregated_agent_metrics", "realtime_agent_metrics", "extensions"} {
		idColumn := "agent_id"
		if table == "agents" {
			idColumn = "id"
		}
		for _, id := range deletedIDs {
			require.Zero(t, deletionCount(t, db, "SELECT COUNT(*) FROM "+table+" WHERE "+idColumn+" = ?", id), table)
		}
		require.Equal(t, 1, deletionCount(t, db, "SELECT COUNT(*) FROM "+table+" WHERE "+idColumn+" = ?", unrelatedID), table)
	}
	for _, installation := range []*kubernetes.CreateInstallationResponse{first, second, unenrolled} {
		_, err := service.CheckIn(context.Background(), installation.Installation.ID, installation.Token,
			kubernetes.CheckInRequest{NodeName: "worker-1", EngineStatus: "waiting"})
		require.ErrorIs(t, err, kubernetes.ErrUnauthorized)
		_, _, err = service.Manifest(context.Background(), installation.Installation.ID,
			kubernetes.ManifestRequest{Token: installation.Token, BackendURL: "https://example.com/api"})
		require.ErrorIs(t, err, kubernetes.ErrUnauthorized)
	}
	_, err := service.CheckIn(context.Background(), unrelated.Installation.ID, unrelated.Token,
		kubernetes.CheckInRequest{NodeName: "worker-1", EngineStatus: "waiting"})
	require.NoError(t, err)
	require.Equal(t, 3, deletionCount(t, db, "SELECT COUNT(*) FROM agents"), "revoked tokens cannot enroll new agents")
}

func TestDeletePipelineRollsBackAllChangesOnFailure(t *testing.T) {
	db, repo, service := deletionDatabase(t)
	installation := deletionInstallation(t, service, 1, "first")
	collectorID := deletionCollector(t, db, service, installation, "worker-1")
	_, err := db.Exec(`CREATE TRIGGER fail_pipeline_deletion BEFORE DELETE ON pipelines
		WHEN OLD.pipeline_id = 1 BEGIN SELECT RAISE(ABORT, 'injected deletion failure'); END`)
	require.NoError(t, err)

	err = repo.DeletePipeline(1)
	require.ErrorContains(t, err, "injected deletion failure")

	requireVMAssignment(t, db, 101, sql.NullInt64{Int64: 1, Valid: true}, sql.NullString{String: "pipeline-1", Valid: true})
	require.Equal(t, 2, deletionCount(t, db, "SELECT COUNT(*) FROM pipelines"))
	require.Equal(t, 1, deletionCount(t, db, "SELECT COUNT(*) FROM kubernetes_installations WHERE id = ?", installation.Installation.ID))
	require.Equal(t, 1, deletionCount(t, db, "SELECT COUNT(*) FROM kubernetes_collectors WHERE agent_id = ?", collectorID))
	for _, table := range []string{"agents_labels", "aggregated_agent_metrics", "realtime_agent_metrics", "extensions"} {
		require.Equal(t, 1, deletionCount(t, db, "SELECT COUNT(*) FROM "+table+" WHERE agent_id = ?", collectorID), table)
	}
	require.Equal(t, 2, deletionCount(t, db, "SELECT COUNT(*) FROM pipeline_components WHERE pipeline_id = 1"))
	require.Equal(t, 1, deletionCount(t, db, "SELECT COUNT(*) FROM pipeline_component_edges WHERE pipeline_id = 1"))
	response, err := service.CheckIn(context.Background(), installation.Installation.ID, installation.Token,
		kubernetes.CheckInRequest{NodeName: "worker-1", EngineStatus: "waiting"})
	require.NoError(t, err)
	require.Equal(t, collectorID, response.CollectorID)
}
