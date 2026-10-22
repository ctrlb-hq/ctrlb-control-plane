package kubernetes_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/api"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/constants"
	database "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/db"
	frontendagent "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontend/agent"
	frontendpipeline "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontend/pipeline"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/kubernetes"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/pkg/queue"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
	"github.com/stretchr/testify/require"
)

const initialConfig = `{"receivers":{"otlp":{"protocols":{"http":{"endpoint":"0.0.0.0:4318"}}}},"exporters":{"debug":{}},"service":{"pipelines":{"logs":{"receivers":["otlp"],"exporters":["debug"]}}}}`

func fixture(t *testing.T) (*sql.DB, *kubernetes.Service) {
	t.Helper()
	utils.InitLogger()
	db, err := database.DBInit(filepath.Join(t.TempDir(), "management.db"))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	_, err = db.Exec("INSERT INTO pipelines (pipeline_id, name, created_by, config_json) VALUES (1, 'test-pipeline', 'tester', ?)", initialConfig)
	require.NoError(t, err)
	return db, kubernetes.NewService(kubernetes.NewRepository(db))
}

func installation(t *testing.T, service *kubernetes.Service) *kubernetes.CreateInstallationResponse {
	t.Helper()
	i, err := service.Create(context.Background(), kubernetes.CreateInstallationRequest{Name: "test-installation", ClusterName: "test-cluster", PipelineID: 1}, "tester")
	require.NoError(t, err)
	return i
}

func checkIn(t *testing.T, service *kubernetes.Service, i *kubernetes.CreateInstallationResponse, req kubernetes.CheckInRequest) *kubernetes.CheckInResponse {
	t.Helper()
	response, err := service.CheckIn(context.Background(), i.Installation.ID, i.Token, req)
	require.NoError(t, err)
	return response
}

func TestInstallationCreationBeforeEnrollmentAndTokenStorage(t *testing.T) {
	db, service := fixture(t)
	i := installation(t, service)
	require.Empty(t, i.Installation.Collectors)
	require.NotNil(t, i.Installation.Collectors)
	require.Zero(t, i.Installation.Rollout.Total)
	var hash string
	require.NoError(t, db.QueryRow("SELECT token_hash FROM kubernetes_installations WHERE id = ?", i.Installation.ID).Scan(&hash))
	expected := sha256.Sum256([]byte(i.Token))
	require.Equal(t, hex.EncodeToString(expected[:]), hash)
	require.NotEqual(t, i.Token, hash)
	encoded, err := json.Marshal(i.Installation)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), i.Token)
	require.NotContains(t, string(encoded), hash)
	require.NotContains(t, string(encoded), "config_json")

	_, err = service.Create(context.Background(), kubernetes.CreateInstallationRequest{Name: "missing", ClusterName: "test-cluster", PipelineID: 999}, "tester")
	require.ErrorIs(t, err, kubernetes.ErrNotFound)
	_, err = db.Exec("INSERT INTO pipelines (pipeline_id, name, created_by) VALUES (2, 'empty', 'tester')")
	require.NoError(t, err)
	_, err = service.Create(context.Background(), kubernetes.CreateInstallationRequest{Name: "empty", ClusterName: "test-cluster", PipelineID: 2}, "tester")
	require.ErrorIs(t, err, kubernetes.ErrInvalid)
	_, err = db.Exec("DELETE FROM pipelines WHERE pipeline_id = 1")
	require.Error(t, err, "an installation must retain its assigned pipeline")
}

func TestCanonicalRevisionIgnoresObjectOrderAndFormatting(t *testing.T) {
	_, first, err := kubernetes.CanonicalConfig([]byte(`{"b":[2,1],"a":{"z":true,"x":1}}`))
	require.NoError(t, err)
	_, same, err := kubernetes.CanonicalConfig([]byte("{\n\"a\":{\"x\":1,\"z\":true}, \"b\":[2,1]}"))
	require.NoError(t, err)
	require.Equal(t, first, same)
	_, changed, err := kubernetes.CanonicalConfig([]byte(`{"a":{"x":1,"z":true},"b":[1,2]}`))
	require.NoError(t, err)
	require.NotEqual(t, first, changed)
}

func TestCheckInAuthenticationAndInstallationScoping(t *testing.T) {
	db, service := fixture(t)
	first, second := installation(t, service), installation(t, service)
	req := kubernetes.CheckInRequest{NodeName: "worker-1", EngineStatus: "waiting"}
	for _, test := range []struct{ id, token string }{
		{first.Installation.ID, ""}, {first.Installation.ID, "wrong-token"},
		{second.Installation.ID, first.Token}, {"missing", first.Token},
	} {
		_, err := service.CheckIn(context.Background(), test.id, test.token, req)
		require.ErrorIs(t, err, kubernetes.ErrUnauthorized)
	}
	var count int
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM agents").Scan(&count))
	require.Zero(t, count)
	firstResponse := checkIn(t, service, first, req)
	secondResponse := checkIn(t, service, second, req)
	require.NotEqual(t, firstResponse.CollectorID, secondResponse.CollectorID)
}

func TestCheckInPreservesIdentityAndRollsOutIndependently(t *testing.T) {
	db, service := fixture(t)
	i := installation(t, service)
	value := 0.0
	req := kubernetes.CheckInRequest{NodeName: "worker-1", PodUID: "first-pod", EngineStatus: "waiting", Metrics: &kubernetes.Metrics{LogsSentPerSecond: &value}}
	first := checkIn(t, service, i, req)
	require.NotEmpty(t, first.Config)
	require.Equal(t, i.Installation.DesiredRevision, first.DesiredRevision)
	req.PodUID, req.PodIP = "replacement-pod", "10.244.2.5"
	req.EngineStatus, req.AppliedRevision, req.AttemptedRevision = "running", first.DesiredRevision, first.DesiredRevision
	req.Metrics = nil
	updated := checkIn(t, service, i, req)
	require.Equal(t, first.CollectorID, updated.CollectorID)
	require.Nil(t, updated.Config)
	secondNode := kubernetes.CheckInRequest{NodeName: "worker-2", EngineStatus: "waiting"}
	checkIn(t, service, i, secondNode)
	var agents, pipelines int
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM agents WHERE pipeline_id = 1 AND management_mode = 'pull'").Scan(&agents))
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM pipelines").Scan(&pipelines))
	require.Equal(t, 2, agents)
	require.Equal(t, 1, pipelines)

	_, err := db.Exec("UPDATE pipelines SET config_json = ? WHERE pipeline_id = 1", `{"receivers":{"otlp":{}},"service":{"pipelines":{}},"exporters":{"debug":{"verbosity":"detailed"}}}`)
	require.NoError(t, err)
	candidate := checkIn(t, service, i, req)
	require.NotEqual(t, first.DesiredRevision, candidate.DesiredRevision)
	require.NotEmpty(t, candidate.Config)
	// A failed candidate rolls back to the old running engine and is not resent.
	req.AttemptedRevision, req.ConfigError = candidate.DesiredRevision, "candidate failed to start; restored previous configuration"
	failed := checkIn(t, service, i, req)
	require.Nil(t, failed.Config)
	view, err := service.Get(context.Background(), i.Installation.ID)
	require.NoError(t, err)
	require.Equal(t, kubernetes.Rollout{Total: 2, Failed: 1, Pending: 1}, view.Rollout)
	require.Equal(t, "running", view.Collectors[0].EngineStatus)
	require.Equal(t, "failed", view.Collectors[0].RolloutStatus)
	require.Equal(t, "replacement-pod", view.Collectors[0].PodUID)
	require.NotNil(t, view.Collectors[0].Metrics)
	require.NotNil(t, view.Collectors[0].Metrics.LogsSentPerSecond)
	require.Zero(t, *view.Collectors[0].Metrics.LogsSentPerSecond)
	require.Nil(t, view.Collectors[0].Metrics.CPUPercent)

	secondNode.EngineStatus, secondNode.AppliedRevision, secondNode.AttemptedRevision = "running", candidate.DesiredRevision, candidate.DesiredRevision
	checkIn(t, service, i, secondNode)
	_, err = db.Exec("UPDATE kubernetes_collectors SET last_seen = ? WHERE node_name = 'worker-1'", time.Now().Add(-40*time.Second).Unix())
	require.NoError(t, err)
	view, err = service.Get(context.Background(), i.Installation.ID)
	require.NoError(t, err)
	require.Equal(t, kubernetes.Rollout{Total: 2, Failed: 1, Applied: 1, Disconnected: 1}, view.Rollout)
	require.Equal(t, "disconnected", view.Collectors[0].ConnectionStatus)
	// Losing the saved configuration clears revisions and requests enrollment again.
	req.AppliedRevision, req.AttemptedRevision, req.ConfigError, req.EngineStatus = "", "", "", "waiting"
	require.NotEmpty(t, checkIn(t, service, i, req).Config)
}

func TestConcurrentFirstCheckInsEnrollOneCollector(t *testing.T) {
	db, service := fixture(t)
	i := installation(t, service)
	const workers = 12
	ids := make(chan int64, workers)
	errors := make(chan error, workers)
	var wg sync.WaitGroup
	for index := 0; index < workers; index++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			response, err := service.CheckIn(context.Background(), i.Installation.ID, i.Token, kubernetes.CheckInRequest{NodeName: "worker-1", EngineStatus: "waiting"})
			if err != nil {
				errors <- err
				return
			}
			ids <- response.CollectorID
		}()
	}
	wg.Wait()
	close(ids)
	close(errors)
	for err := range errors {
		require.NoError(t, err)
	}
	var first int64
	for id := range ids {
		if first == 0 {
			first = id
		}
		require.Equal(t, first, id)
	}
	var count int
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM agents").Scan(&count))
	require.Equal(t, 1, count)
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM kubernetes_collectors").Scan(&count))
	require.Equal(t, 1, count)
}

func TestCheckInRejectsInvalidMetricsAndFields(t *testing.T) {
	db, service := fixture(t)
	i := installation(t, service)
	for _, value := range []float64{-1, math.NaN(), math.Inf(1)} {
		_, err := service.CheckIn(context.Background(), i.Installation.ID, i.Token, kubernetes.CheckInRequest{
			NodeName: "worker", EngineStatus: "running", Metrics: &kubernetes.Metrics{CPUPercent: &value},
		})
		require.ErrorIs(t, err, kubernetes.ErrInvalid)
	}
	_, err := service.CheckIn(context.Background(), i.Installation.ID, i.Token, kubernetes.CheckInRequest{EngineStatus: "waiting"})
	require.ErrorIs(t, err, kubernetes.ErrInvalid)
	_, err = service.CheckIn(context.Background(), i.Installation.ID, i.Token, kubernetes.CheckInRequest{NodeName: "worker", EngineStatus: "unknown"})
	require.ErrorIs(t, err, kubernetes.ErrInvalid)
	var count int
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM agents").Scan(&count))
	require.Zero(t, count)
}

func TestPullCollectorsAreExcludedFromVMControlPaths(t *testing.T) {
	db, service := fixture(t)
	i := installation(t, service)
	collector := checkIn(t, service, i, kubernetes.CheckInRequest{NodeName: "worker-1", EngineStatus: "waiting"})
	_, err := db.Exec("INSERT INTO aggregated_agent_metrics (agent_id, status, updated_at) VALUES (?, 'connected', ?)", collector.CollectorID, time.Now().Unix())
	require.NoError(t, err)
	queued, err := queue.NewQueueRepository(db).RefreshMonitoring()
	require.NoError(t, err)
	require.Empty(t, queued)
	agents := frontendagent.NewFrontendAgentRepository(db)
	_, _, err = agents.GetAgentNetworkInfoByID("1")
	require.Error(t, err, "legacy lifecycle commands cannot resolve a pull collector target")
	pipelines := frontendpipeline.NewFrontendPipelineRepository(db)
	require.Error(t, pipelines.DetachAgentFromPipeline(1, int(collector.CollectorID)))
	require.Error(t, pipelines.AttachAgentToPipeline(999, int(collector.CollectorID)))
	attached, err := pipelines.GetAllAgentsAttachedToPipeline(1)
	require.NoError(t, err)
	require.Len(t, attached, 1)
	require.Equal(t, "pull", attached[0].ManagementMode)
	info, err := pipelines.GetAgentInfo(int(collector.CollectorID))
	require.NoError(t, err)
	require.Equal(t, "pull", info.ManagementMode)
	// Persisting a graph on a pull-only pipeline must not contact node addresses.
	_, err = db.Exec("INSERT INTO pipeline_components (pipeline_id, component_role, component_name, name, config, supported_signals) VALUES (1, 'receiver', 'otlp_receiver', 'source', '{}', 'logs')")
	require.NoError(t, err)
	require.NoError(t, frontendpipeline.NewFrontendPipelineService(pipelines).SyncPipelineGraph(1, constants.DefaultPipelineGraph))
	// A pipeline with no enrolled collectors is also valid.
	_, err = db.Exec("INSERT INTO pipelines (pipeline_id, name, created_by) VALUES (2, 'new', 'tester')")
	require.NoError(t, err)
	require.NoError(t, frontendpipeline.NewFrontendPipelineService(pipelines).SyncPipelineGraph(2, constants.DefaultPipelineGraph))
}

func TestManagementRoutesRequireUserAndInstallationTokens(t *testing.T) {
	_, service := fixture(t)
	previousSecret := constants.JWT_SECRET
	constants.JWT_SECRET = "kubernetes-management-test-secret"
	t.Cleanup(func() { constants.JWT_SECRET = previousSecret })
	userToken, err := utils.GenerateAccessToken("user@example.com")
	require.NoError(t, err)
	router := api.NewRouter(&api.Handler{KubernetesHandler: kubernetes.NewHandler(service)})
	request := func(method, path, token string, payload any) *httptest.ResponseRecorder {
		data, err := json.Marshal(payload)
		require.NoError(t, err)
		req := httptest.NewRequest(method, path, bytes.NewReader(data))
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		return recorder
	}
	creation := kubernetes.CreateInstallationRequest{Name: "test-installation", ClusterName: "test-cluster", PipelineID: 1}
	require.Equal(t, http.StatusUnauthorized, request("POST", "/api/frontend/v2/installations", "", creation).Code)
	created := request("POST", "/api/frontend/v2/installations", userToken, creation)
	require.Equal(t, http.StatusCreated, created.Code, created.Body.String())
	var i kubernetes.CreateInstallationResponse
	require.NoError(t, json.Unmarshal(created.Body.Bytes(), &i))
	require.Equal(t, "user@example.com", i.Installation.CreatedBy)
	path := "/api/agent/v2/installations/" + i.Installation.ID + "/check-in"
	req := kubernetes.CheckInRequest{NodeName: "worker-1", EngineStatus: "waiting"}
	require.Equal(t, http.StatusUnauthorized, request("POST", path, "", req).Code)
	require.Equal(t, http.StatusUnauthorized, request("POST", path, userToken, req).Code)
	require.Equal(t, http.StatusOK, request("POST", path, i.Token, req).Code)
	// Unknown fields cannot override the installation's pipeline assignment.
	require.Equal(t, http.StatusOK, request("POST", path, i.Token, map[string]any{"node_name": "worker-1", "engine_status": "waiting", "pipeline_id": 999}).Code)
	view := request("GET", "/api/frontend/v2/installations/"+i.Installation.ID, userToken, nil)
	require.Equal(t, http.StatusOK, view.Code)
	require.NotContains(t, view.Body.String(), i.Token)
	require.Contains(t, view.Body.String(), `"pipeline_id":1`)
	require.Equal(t, http.StatusOK, request("GET", "/api/frontend/v2/pipelines/1/installations", userToken, nil).Code)
	require.Equal(t, http.StatusNotFound, request("GET", "/api/frontend/v2/pipelines/999/installations", userToken, nil).Code)
	require.Equal(t, http.StatusNotFound, request("GET", "/api/frontend/v2/installations/missing", userToken, nil).Code)
	require.Equal(t, http.StatusOK, request("GET", "/api/frontend/v2/installations", userToken, nil).Code)
	bad := httptest.NewRequest("POST", path, bytes.NewBufferString(`{} {}`))
	bad.Header.Set("Authorization", "Bearer "+i.Token)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, bad)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}
