package kubernetes_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/api"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/constants"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/kubernetes"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"gopkg.in/yaml.v3"
)

func manifestDocuments(t *testing.T, data []byte) map[string]map[string]any {
	t.Helper()
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	documents := map[string]map[string]any{}
	for {
		var document map[string]any
		err := decoder.Decode(&document)
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		kind := document["kind"].(string)
		require.NotContains(t, documents, kind)
		documents[kind] = document
	}
	return documents
}

func TestManifestHasIsolatedStateSecretAndLeastPrivilege(t *testing.T) {
	_, service := fixture(t)
	i := installation(t, service)
	data, filename, err := service.Manifest(context.Background(), i.Installation.ID, kubernetes.ManifestRequest{
		Token: i.Token, BackendURL: "https://control.example.com/prefix/",
	})
	require.NoError(t, err)
	name := "ctrlb-collector-" + i.Installation.ID[:12]
	require.Equal(t, name+".yaml", filename)
	documents := manifestDocuments(t, data)
	require.Len(t, documents, 6)
	require.Equal(t, "observability", documents["Namespace"]["metadata"].(map[string]any)["name"])
	for kind, document := range documents {
		if kind == "Namespace" {
			continue
		}
		metadata := document["metadata"].(map[string]any)
		require.Equal(t, name, metadata["name"])
		if kind == "Secret" || kind == "ServiceAccount" || kind == "DaemonSet" {
			require.Equal(t, "observability", metadata["namespace"])
		}
	}
	require.Equal(t, map[string]any{"token": i.Token}, documents["Secret"]["stringData"])
	rules := documents["ClusterRole"]["rules"].([]any)
	require.Len(t, rules, 2)
	require.Equal(t, []any{"pods", "namespaces"}, rules[0].(map[string]any)["resources"])
	require.Equal(t, []any{"get", "list", "watch"}, rules[0].(map[string]any)["verbs"])
	require.Equal(t, []any{"nodes/stats"}, rules[1].(map[string]any)["resources"])
	require.Equal(t, []any{"get"}, rules[1].(map[string]any)["verbs"])
	role := documents["ClusterRoleBinding"]
	require.Equal(t, name, role["roleRef"].(map[string]any)["name"])
	require.Equal(t, map[string]any{"kind": "ServiceAccount", "name": name, "namespace": "observability"}, role["subjects"].([]any)[0])
	pod := documents["DaemonSet"]["spec"].(map[string]any)["template"].(map[string]any)["spec"].(map[string]any)
	require.NotContains(t, pod, "hostNetwork")
	require.NotContains(t, pod, "hostPID")
	require.Equal(t, map[string]any{"kubernetes.io/os": "linux"}, pod["nodeSelector"])
	require.Equal(t, name, pod["serviceAccountName"])
	container := pod["containers"].([]any)[0].(map[string]any)
	require.Equal(t, kubernetes.DefaultCollectorImage, container["image"])
	require.Equal(t, "IfNotPresent", container["imagePullPolicy"])
	security := container["securityContext"].(map[string]any)
	require.Equal(t, false, security["allowPrivilegeEscalation"])
	require.Equal(t, true, security["readOnlyRootFilesystem"])
	require.NotContains(t, security, "privileged")
	require.Equal(t, []any{"ALL"}, security["capabilities"].(map[string]any)["drop"])
	env := map[string]map[string]any{}
	for _, item := range container["env"].([]any) {
		e := item.(map[string]any)
		env[e["name"].(string)] = e
	}
	require.Equal(t, "https://control.example.com/prefix", env["BACKEND_URL"]["value"])
	require.Equal(t, i.Installation.ID, env["INSTALLATION_ID"]["value"])
	require.Equal(t, "/var/lib/ctrlb", env["STATE_DIR"]["value"])
	require.Equal(t, "/var/run/secrets/ctrlb/token", env["INSTALLATION_TOKEN_FILE"]["value"])
	require.NotContains(t, env, "INSTALLATION_TOKEN")
	require.NotContains(t, env, "ALLOW_INSECURE_HTTP")
	for key, field := range map[string]string{"K8S_NODE_NAME": "spec.nodeName", "K8S_NODE_IP": "status.hostIP", "POD_IP": "status.podIP", "POD_UID": "metadata.uid", "POD_NAME": "metadata.name", "POD_NAMESPACE": "metadata.namespace"} {
		require.Equal(t, field, env[key]["valueFrom"].(map[string]any)["fieldRef"].(map[string]any)["fieldPath"])
	}
	volumes := map[string]map[string]any{}
	for _, item := range pod["volumes"].([]any) {
		v := item.(map[string]any)
		volumes[v["name"].(string)] = v
	}
	require.Equal(t, "/var/lib/ctrlb/installations/"+i.Installation.ID, volumes["state"]["hostPath"].(map[string]any)["path"])
	require.Equal(t, name, volumes["token"]["secret"].(map[string]any)["secretName"])
	require.Equal(t, 0400, volumes["token"]["secret"].(map[string]any)["defaultMode"])
	for _, item := range container["volumeMounts"].([]any) {
		mount := item.(map[string]any)
		if mount["name"] == "pod-logs" {
			require.Equal(t, "/var/log/pods", mount["mountPath"])
			require.Equal(t, true, mount["readOnly"])
		}
	}
	for probe, path := range map[string]string{"startupProbe": "/healthz", "livenessProbe": "/healthz", "readinessProbe": "/readyz"} {
		p := container[probe].(map[string]any)
		require.Equal(t, path, p["httpGet"].(map[string]any)["path"])
		require.Equal(t, "management", p["httpGet"].(map[string]any)["port"])
	}
	startup := container["startupProbe"].(map[string]any)
	require.GreaterOrEqual(t, startup["failureThreshold"].(int)*startup["periodSeconds"].(int), 90)
	require.Contains(t, container, "resources")
	// A second installation must not reuse names or host state paths.
	second := installation(t, service)
	other, otherFile, err := service.Manifest(context.Background(), second.Installation.ID, kubernetes.ManifestRequest{Token: second.Token, BackendURL: "http://host.docker.internal:8096", Image: "example.com/collector:v1"})
	require.NoError(t, err)
	require.NotEqual(t, filename, otherFile)
	require.NotContains(t, string(other), "/var/lib/ctrlb/installations/"+i.Installation.ID)
	require.Contains(t, string(other), "ALLOW_INSECURE_HTTP")
}

func TestManifestRejectsUnsafeInputsAndKeepsUserValuesAsScalars(t *testing.T) {
	_, service := fixture(t)
	i, err := service.Create(context.Background(), kubernetes.CreateInstallationRequest{Name: "demo", ClusterName: "quoted\"\n---\nkind: Pod", PipelineID: 1}, "tester")
	require.NoError(t, err)
	for _, address := range []string{"", "control.example.com", "ftp://control.example.com", "https://user:secret@control.example.com", "https://control.example.com?token=secret", "https://control.example.com?", "https://control.example.com#fragment", "https://control.example.com#", " https://control.example.com", "https://control.example.com:99999", "https://control.example.com/\nkind: Pod"} {
		_, _, err := service.Manifest(context.Background(), i.Installation.ID, kubernetes.ManifestRequest{Token: i.Token, BackendURL: address})
		require.ErrorIs(t, err, kubernetes.ErrInvalid, address)
	}
	for _, image := range []string{"repo/image:v1\n---", "image with spaces", strings.Repeat("x", 513)} {
		_, _, err := service.Manifest(context.Background(), i.Installation.ID, kubernetes.ManifestRequest{Token: i.Token, BackendURL: "https://control.example.com", Image: image})
		require.ErrorIs(t, err, kubernetes.ErrInvalid)
	}
	for _, id := range []string{"../etc/passwd", "short", strings.Repeat("a", 64)} {
		_, _, err := service.Manifest(context.Background(), id, kubernetes.ManifestRequest{Token: i.Token, BackendURL: "https://control.example.com"})
		require.ErrorIs(t, err, kubernetes.ErrUnauthorized)
	}
	data, _, err := service.Manifest(context.Background(), i.Installation.ID, kubernetes.ManifestRequest{Token: i.Token, BackendURL: "https://control.example.com"})
	require.NoError(t, err)
	documents := manifestDocuments(t, data)
	require.Len(t, documents, 6, "cluster name must remain a scalar rather than a seventh document")
}

func TestManifestRouteRequiresBothTokensWithoutLoggingSecrets(t *testing.T) {
	_, service := fixture(t)
	first, second := installation(t, service), installation(t, service)
	previousSecret, previousLogger := constants.JWT_SECRET, utils.Logger
	constants.JWT_SECRET = "manifest-test-secret"
	core, logs := observer.New(zap.DebugLevel)
	utils.Logger = zap.New(core)
	t.Cleanup(func() { constants.JWT_SECRET, utils.Logger = previousSecret, previousLogger })
	access, err := utils.GenerateAccessToken("user@example.com")
	require.NoError(t, err)
	refresh, err := utils.GenerateRefreshToken("user@example.com")
	require.NoError(t, err)
	router := api.NewRouter(&api.Handler{KubernetesHandler: kubernetes.NewHandler(service)})
	request := func(id, user, token string) *httptest.ResponseRecorder {
		payload, err := json.Marshal(kubernetes.ManifestRequest{Token: token, BackendURL: "https://control.example.com"})
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPost, "/api/frontend/v2/installations/"+id+"/manifest", bytes.NewReader(payload))
		if user != "" {
			req.Header.Set("Authorization", "Bearer "+user)
		}
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		return response
	}
	for _, test := range []struct{ id, user, token string }{
		{first.Installation.ID, "", first.Token}, {first.Installation.ID, first.Token, first.Token},
		{first.Installation.ID, refresh, first.Token}, {first.Installation.ID, access, ""},
		{first.Installation.ID, access, "wrong-secret"}, {second.Installation.ID, access, first.Token},
		{strings.Repeat("a", 32), access, first.Token},
	} {
		response := request(test.id, test.user, test.token)
		require.Equal(t, http.StatusUnauthorized, response.Code)
		require.NotContains(t, response.Body.String(), first.Token)
		require.NotContains(t, response.Body.String(), "wrong-secret")
	}
	response := request(first.Installation.ID, access, first.Token)
	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, "application/yaml", response.Header().Get("Content-Type"))
	require.Equal(t, "no-store", response.Header().Get("Cache-Control"))
	require.Equal(t, `attachment; filename="ctrlb-collector-`+first.Installation.ID[:12]+`.yaml"`, response.Header().Get("Content-Disposition"))
	require.Contains(t, response.Body.String(), first.Token)
	for _, entry := range logs.All() {
		require.NotContains(t, entry.Message, first.Token)
		require.NotContains(t, entry.ContextMap(), first.Token)
		encoded, err := json.Marshal(entry.ContextMap())
		require.NoError(t, err)
		require.NotContains(t, string(encoded), first.Token)
		require.NotContains(t, string(encoded), access)
	}
}
