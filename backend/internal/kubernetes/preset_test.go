package kubernetes_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/api"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/constants"
	frontendpipeline "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontend/pipeline"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/kubernetes"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/pkg/configcompiler"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
	"github.com/stretchr/testify/require"
)

func presetRequest() kubernetes.CreatePipelineRequest {
	return kubernetes.CreatePipelineRequest{
		Name: "Kubernetes demo", Endpoint: "https://ctrlb.explore.ctrlb.ai/engine/api/default",
		LogsEndpoint:    "https://ctrlb.explore.ctrlb.ai/engine/api/default/control_plane/_otel/v1/logs",
		MetricsEndpoint: "https://ctrlb.explore.ctrlb.ai/engine/api/default/v1/metrics",
		Headers:         map[string]string{"Authorization": "Basic CHANGE_ME"},
	}
}

func TestKubernetesPresetCollectsBothSignalsAndPersistsOffsets(t *testing.T) {
	req := presetRequest()
	graph, err := kubernetes.PipelineGraph(req)
	require.NoError(t, err)
	config, err := configcompiler.CompileGraphToJSON(graph)
	require.NoError(t, err)
	pipelines := (*config)["service"].(map[string]any)["pipelines"].(configcompiler.Pipelines)
	require.Len(t, pipelines, 2)
	for name, pipeline := range pipelines {
		require.Len(t, pipeline.Receivers, 1)
		require.Len(t, pipeline.Processors, 4)
		for index, typ := range []string{"memory_limiter/", "k8s_attributes/", "resource/", "batch/"} {
			require.True(t, strings.HasPrefix(pipeline.Processors[index], typ))
		}
		require.Len(t, pipeline.Exporters, 1)
		require.True(t, strings.HasPrefix(pipeline.Exporters[0], "otlp_http/"))
		if strings.HasPrefix(name, "logs/") {
			require.True(t, strings.HasPrefix(pipeline.Receivers[0], "file_log/"))
		} else {
			require.True(t, strings.HasPrefix(name, "metrics/"))
			require.True(t, strings.HasPrefix(pipeline.Receivers[0], "kubelet_stats/"))
		}
	}
	for _, node := range graph.Nodes {
		switch node.ComponentName {
		case "filelog_receiver":
			require.Equal(t, []string{"/var/log/pods/*/*/*.log"}, node.Config["include"])
			require.Equal(t, []string{"/var/log/pods/${env:POD_NAMESPACE}_${env:POD_NAME}_${env:POD_UID}/*/*.log"}, node.Config["exclude"])
			require.Equal(t, "end", node.Config["start_at"])
			require.Equal(t, "file_storage", node.Config["storage"])
			require.Equal(t, true, node.Config["include_file_path"])
			operator := node.Config["operators"].([]any)[0].(map[string]any)
			require.Equal(t, "container", operator["type"])
			require.Equal(t, true, operator["add_metadata_from_filepath"])
		case "kubeletstats_receiver":
			require.Equal(t, "30s", node.Config["collection_interval"])
			require.Equal(t, []string{"node", "pod", "container"}, node.Config["metric_groups"])
			require.Equal(t, "serviceAccount", node.Config["auth_type"])
			require.Equal(t, false, node.Config["insecure_skip_verify"])
			require.NotContains(t, node.Config, "extra_metadata_labels")
		case "otlphttp_exporter":
			require.Equal(t, req.Headers, node.Config["headers"])
			require.Equal(t, req.LogsEndpoint, node.Config["logs_endpoint"])
			require.Equal(t, req.MetricsEndpoint, node.Config["metrics_endpoint"])
			require.Equal(t, "proto", node.Config["encoding"])
			require.Equal(t, "none", node.Config["compression"])
		}
	}
	extensions := (*config)["extensions"].(map[string]any)
	require.Equal(t, map[string]any{"directory": "${env:STATE_DIR}/offsets", "create_directory": true}, extensions["file_storage"])
	// An explicit opt-in supports development clusters with self-signed kubelets.
	req.KubeletInsecureSkipVerify = true
	insecure, err := kubernetes.PipelineGraph(req)
	require.NoError(t, err)
	require.Equal(t, true, insecure.Nodes[1].Config["insecure_skip_verify"])

	// Export the actual compiled config for the agent's upstream compatibility test.
	// The fixture contains only a documented credential placeholder.
	if path := os.Getenv("KUBERNETES_PRESET_FIXTURE"); path != "" {
		data, err := json.MarshalIndent(config, "", "  ")
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(path, data, 0600))
	}
}

func TestKubernetesPipelinePersistsEditableGraphAndStorageExtension(t *testing.T) {
	db, service := fixture(t)
	created, err := service.CreatePipeline(context.Background(), presetRequest(), "actor@example.com")
	require.NoError(t, err)
	var actor string
	var persisted []byte
	require.NoError(t, db.QueryRow("SELECT created_by, config_json FROM pipelines WHERE pipeline_id = ?", created.PipelineID).Scan(&actor, &persisted))
	require.Equal(t, "actor@example.com", actor)
	_, originalRevision, err := kubernetes.CanonicalConfig(persisted)
	require.NoError(t, err)
	repository := frontendpipeline.NewFrontendPipelineRepository(db)
	graph, err := repository.GetPipelineGraph(created.PipelineID)
	require.NoError(t, err)
	compiled, err := configcompiler.CompileGraphToJSON(*graph)
	require.NoError(t, err)
	data, err := json.Marshal(compiled)
	require.NoError(t, err)
	_, revision, err := kubernetes.CanonicalConfig(data)
	require.NoError(t, err)
	require.Equal(t, originalRevision, revision)
	// Saving reassigns component database IDs; unchanged config keeps its revision.
	require.NoError(t, repository.SyncPipelineGraph(nil, created.PipelineID, *graph))
	require.NoError(t, db.QueryRow("SELECT config_json FROM pipelines WHERE pipeline_id = ?", created.PipelineID).Scan(&persisted))
	_, revision, err = kubernetes.CanonicalConfig(persisted)
	require.NoError(t, err)
	require.Equal(t, originalRevision, revision)
	graph, err = repository.GetPipelineGraph(created.PipelineID)
	require.NoError(t, err)
	for index := range graph.Nodes {
		if graph.Nodes[index].ComponentName == "batch_processor" {
			graph.Nodes[index].Config["timeout"] = "10s"
		}
	}
	require.NoError(t, frontendpipeline.NewFrontendPipelineService(repository).SyncPipelineGraph(created.PipelineID, *graph))
	i, err := service.Create(context.Background(), kubernetes.CreateInstallationRequest{Name: "demo", ClusterName: "demo", PipelineID: created.PipelineID}, "tester")
	require.NoError(t, err)
	require.NotEqual(t, originalRevision, i.Installation.DesiredRevision)
	response := checkIn(t, service, i, kubernetes.CheckInRequest{NodeName: "worker", EngineStatus: "waiting"})
	require.Contains(t, response.Config["extensions"], "file_storage")
	encoded, err := json.Marshal(response.Config)
	require.NoError(t, err)
	require.Contains(t, string(encoded), `"timeout":"10s"`)
	require.Contains(t, string(encoded), `"enabled":true`)
}

func TestKubernetesPipelineRouteAndInputValidation(t *testing.T) {
	_, service := fixture(t)
	previous := constants.JWT_SECRET
	constants.JWT_SECRET = "kubernetes-preset-test-secret"
	t.Cleanup(func() { constants.JWT_SECRET = previous })
	token, err := utils.GenerateAccessToken("actor@example.com")
	require.NoError(t, err)
	router := api.NewRouter(&api.Handler{KubernetesHandler: kubernetes.NewHandler(service)})
	request := func(token string, payload any) *httptest.ResponseRecorder {
		data, err := json.Marshal(payload)
		require.NoError(t, err)
		req := httptest.NewRequest("POST", "/api/frontend/v2/kubernetes-pipelines", bytes.NewReader(data))
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		return recorder
	}
	require.Equal(t, http.StatusUnauthorized, request("", presetRequest()).Code)
	valid := request(token, presetRequest())
	require.Equal(t, http.StatusCreated, valid.Code, valid.Body.String())
	var response kubernetes.CreatePipelineResponse
	require.NoError(t, json.Unmarshal(valid.Body.Bytes(), &response))
	require.Positive(t, response.PipelineID)
	for _, endpoint := range []string{"", "/relative", "ftp://example.com", "https://user:secret@example.com", "https://example.com#secret"} {
		req := presetRequest()
		req.Endpoint = endpoint
		invalid := request(token, req)
		require.Equal(t, http.StatusBadRequest, invalid.Code)
		require.NotContains(t, invalid.Body.String(), "secret")
	}
	req := presetRequest()
	req.LogsEndpoint = "ftp://example.com/secret"
	require.Equal(t, http.StatusBadRequest, request(token, req).Code)
	req = presetRequest()
	req.MetricsEndpoint = "/relative"
	require.Equal(t, http.StatusBadRequest, request(token, req).Code)
	req = presetRequest()
	req.Headers["Authorization"] = "secret\nvalue"
	invalid := request(token, req)
	require.Equal(t, http.StatusBadRequest, invalid.Code)
	require.NotContains(t, invalid.Body.String(), "secret")
	req = presetRequest()
	req.Headers["Authorization"] = "literal-${env:NOT_A_CREDENTIAL_VARIABLE}-$"
	graph, err := kubernetes.PipelineGraph(req)
	require.NoError(t, err)
	exporterHeaders := graph.Nodes[len(graph.Nodes)-1].Config["headers"].(map[string]string)
	require.Equal(t, "literal-$${env:NOT_A_CREDENTIAL_VARIABLE}-$$", exporterHeaders["Authorization"])
	require.Equal(t, "literal-${env:NOT_A_CREDENTIAL_VARIABLE}-$", req.Headers["Authorization"])
	// A plain HTTP destination is supported when intentionally used in a demo.
	req = presetRequest()
	req.Endpoint = "http://localhost:4318"
	require.Equal(t, http.StatusCreated, request(token, req).Code)
}
