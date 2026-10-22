package managed

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/adapters"
	"github.com/stretchr/testify/require"
)

func runtimeConfig(t *testing.T, backend string) Config {
	t.Helper()
	return Config{BackendURL: backend, InstallationID: "installation", Token: "installation-secret", NodeName: "worker-1",
		PodName: "collector-abc", PodUID: "pod-uid", Namespace: "observability", StateDir: t.TempDir(), MetricsURL: "http://127.0.0.1:1/metrics"}
}

func TestRuntimeCheckInPreservesBackendAPIPrefix(t *testing.T) {
	for _, test := range []struct {
		name, basePath, expectedPath string
	}{
		{"legacy root", "", "/api/agent/v2/installations/installation/check-in"},
		{"legacy root slash", "/", "/api/agent/v2/installations/installation/check-in"},
		{"explicit API", "/api", "/api/agent/v2/installations/installation/check-in"},
		{"hosted root", "/control-plane", "/control-plane/api/agent/v2/installations/installation/check-in"},
		{"hosted root slash", "/control-plane/", "/control-plane/api/agent/v2/installations/installation/check-in"},
		{"hosted API", "/control-plane/api", "/control-plane/api/agent/v2/installations/installation/check-in"},
		{"hosted API slash", "/control-plane/api/", "/control-plane/api/agent/v2/installations/installation/check-in"},
	} {
		t.Run(test.name, func(t *testing.T) {
			type requestDetails struct{ path, authorization string }
			requests := make(chan requestDetails, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				requests <- requestDetails{req.URL.Path, req.Header.Get("Authorization")}
				w.WriteHeader(http.StatusUnauthorized)
			}))
			t.Cleanup(server.Close)
			r, err := New(runtimeConfig(t, server.URL+test.basePath), &fakeEngine{})
			require.NoError(t, err)
			_, err = r.CheckIn(context.Background())
			require.ErrorContains(t, err, "401")
			request := <-requests
			require.Equal(t, test.expectedPath, request.path)
			require.Equal(t, "Bearer installation-secret", request.authorization)
		})
	}
}

func TestRuntimeAPIHostnameStillGetsAPIPath(t *testing.T) {
	r, err := New(runtimeConfig(t, "https://api"), &fakeEngine{})
	require.NoError(t, err)
	require.Equal(t, "https://api/api/agent/v2/installations/installation/check-in", r.url)
}

func TestRuntimeBackendSequenceAndOfflineRestart(t *testing.T) {
	ctx := context.Background()
	good := minimalConfig(t)
	goodRevision := revision(t, good)
	invalid := minimalConfig(t)
	invalid["receivers"] = map[string]any{"unknown_receiver": map[string]any{}}
	failedRevision := revision(t, invalid)
	var mu sync.Mutex
	var reports []checkInRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/api/agent/v2/installations/installation/check-in" || req.Header.Get("Authorization") != "Bearer installation-secret" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		var report checkInRequest
		if err := json.NewDecoder(req.Body).Decode(&report); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		mu.Lock()
		reports = append(reports, report)
		count := len(reports)
		mu.Unlock()
		response := checkInResponse{DesiredRevision: failedRevision, PollIntervalSeconds: 10}
		if count == 1 {
			response.DesiredRevision, response.Config = goodRevision, good
		} else if count == 2 {
			response.Config = invalid
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	cfg := runtimeConfig(t, server.URL)
	r, err := New(cfg, &adapters.ManagedEngine{})
	require.NoError(t, err)
	defer r.manager.Close(ctx)
	require.NoError(t, r.manager.Restore(ctx))
	for index := range 3 {
		interval, err := r.CheckIn(ctx)
		require.NoError(t, err)
		if index < 2 {
			require.Equal(t, 100*time.Millisecond, interval)
		} else {
			require.Equal(t, 10*time.Second, interval)
		}
	}
	mu.Lock()
	require.Len(t, reports, 3)
	require.Equal(t, "waiting", reports[0].EngineStatus)
	require.Empty(t, reports[0].AppliedRevision)
	require.Equal(t, "running", reports[1].EngineStatus)
	require.Equal(t, goodRevision, reports[1].AppliedRevision)
	require.Equal(t, "running", reports[2].EngineStatus)
	require.Equal(t, goodRevision, reports[2].AppliedRevision)
	require.Equal(t, failedRevision, reports[2].AttemptedRevision)
	require.Contains(t, reports[2].ConfigError, "validate configuration")
	require.Equal(t, "worker-1", reports[2].NodeName)
	require.Equal(t, "pod-uid", reports[2].PodUID)
	mu.Unlock()
	server.Close()
	_, err = r.CheckIn(ctx)
	require.Error(t, err)
	require.Equal(t, goodRevision, r.manager.Status().AppliedRevision)
	require.NoError(t, r.manager.Close(ctx))
	resumed, err := New(cfg, &adapters.ManagedEngine{})
	require.NoError(t, err)
	defer resumed.manager.Close(ctx)
	require.NoError(t, resumed.manager.Restore(ctx))
	_, err = resumed.CheckIn(ctx)
	require.Error(t, err)
	require.Equal(t, "running", resumed.manager.Status().EngineStatus)
	require.Equal(t, goodRevision, resumed.manager.Status().AppliedRevision)
	require.Equal(t, failedRevision, resumed.manager.Status().AttemptedRevision)
}

func TestRuntimeNeverForwardsTokenOnRedirect(t *testing.T) {
	var forwarded atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { forwarded.Add(1) }))
	defer target.Close()
	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
	}))
	defer redirect.Close()
	r, err := New(runtimeConfig(t, redirect.URL), &fakeEngine{})
	require.NoError(t, err)
	_, err = r.CheckIn(context.Background())
	require.ErrorContains(t, err, "307")
	require.Equal(t, int32(0), forwarded.Load())
	require.Equal(t, "waiting", r.manager.Status().EngineStatus)
}

func TestRuntimeHealthAndMissingDesiredConfiguration(t *testing.T) {
	good := minimalConfig(t)
	desired := revision(t, good)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(checkInResponse{DesiredRevision: desired})
	}))
	defer server.Close()
	engine := &fakeEngine{}
	r, err := New(runtimeConfig(t, server.URL), engine)
	require.NoError(t, err)
	for _, probe := range []struct {
		method, path string
		code         int
	}{{"GET", "/healthz", 200}, {"GET", "/readyz", 503}, {"POST", "/agent/v1/stop", 404}} {
		response := httptest.NewRecorder()
		r.Handler().ServeHTTP(response, httptest.NewRequest(probe.method, probe.path, nil))
		require.Equal(t, probe.code, response.Code)
	}
	_, err = r.CheckIn(context.Background())
	require.ErrorContains(t, err, "without its configuration")
	require.Empty(t, r.manager.Status().AppliedRevision)
	require.NoError(t, r.manager.Apply(context.Background(), desired, good))
	response := httptest.NewRecorder()
	r.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/readyz", nil))
	require.Equal(t, 200, response.Code)
	engine.crash()
	response = httptest.NewRecorder()
	r.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/readyz", nil))
	require.Equal(t, 503, response.Code)
}

func TestManagedEnvironmentAndTokenRotation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "token")
	require.NoError(t, os.WriteFile(path, []byte("first-secret\n"), 0600))
	t.Setenv("INSTALLATION_ID", "installation")
	t.Setenv("K8S_NODE_NAME", "worker-1")
	t.Setenv("BACKEND_URL", "http://host.docker.internal:8096")
	t.Setenv("INSTALLATION_TOKEN", "")
	t.Setenv("INSTALLATION_TOKEN_FILE", path)
	t.Setenv("STATE_DIR", t.TempDir())
	t.Setenv("ALLOW_INSECURE_HTTP", "false")
	_, err := ConfigFromEnv()
	require.ErrorContains(t, err, "HTTPS")
	t.Setenv("ALLOW_INSECURE_HTTP", "true")
	cfg, err := ConfigFromEnv()
	require.NoError(t, err)
	token, err := cfg.token()
	require.NoError(t, err)
	require.Equal(t, "first-secret", token)
	require.NoError(t, os.WriteFile(path, []byte("rotated-secret\n"), 0600))
	token, err = cfg.token()
	require.NoError(t, err)
	require.Equal(t, "rotated-secret", token)
}

func TestRuntimeLoopAcknowledgesThenShutsDownProgrammatically(t *testing.T) {
	good := minimalConfig(t)
	desired := revision(t, good)
	acknowledged := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var report checkInRequest
		if json.NewDecoder(req.Body).Decode(&report) != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		response := checkInResponse{DesiredRevision: desired}
		if report.AppliedRevision != desired {
			response.Config = good
		} else {
			select {
			case acknowledged <- struct{}{}:
			default:
			}
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	cfg := runtimeConfig(t, server.URL)
	cfg.ListenAddress = "127.0.0.1:0"
	engine := &fakeEngine{}
	r, err := New(cfg, engine)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- r.Run(ctx) }()
	select {
	case <-acknowledged:
	case <-time.After(2 * time.Second):
		t.Fatal("applied configuration was not acknowledged promptly")
	}
	cancel()
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("management loop did not shut down")
	}
	require.False(t, engine.Status().Running)
	saved, err := r.manager.store.load()
	require.NoError(t, err)
	require.Equal(t, desired, saved.AppliedRevision)
}
