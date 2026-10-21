package managed

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/adapters"
	"github.com/stretchr/testify/require"
)

func minimalConfig(t *testing.T) map[string]any {
	t.Helper()
	var config map[string]any
	require.NoError(t, json.Unmarshal([]byte(`{
  "receivers":{"otlp":{"protocols":{"grpc":{"endpoint":"127.0.0.1:0"}}}},
  "exporters":{"debug":{},"otlphttp":{"endpoint":"http://127.0.0.1:4318","sending_queue":{"enabled":false},"retry_on_failure":{"enabled":false}}},
  "service":{"telemetry":{"metrics":{"level":"none"}},"pipelines":{"logs":{"receivers":["otlp"],"exporters":["debug"]}}}
}`), &config))
	return config
}

func revision(t *testing.T, config map[string]any) string {
	t.Helper()
	value, err := configRevision(config)
	require.NoError(t, err)
	return value
}

func realManager(t *testing.T, dir string) *Manager {
	t.Helper()
	m, err := NewManager(dir, "installation", "worker-1", &adapters.ManagedEngine{})
	require.NoError(t, err)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		require.NoError(t, m.Close(ctx))
	})
	return m
}

func TestManagerValidationStartupRollbackAndCachedRestart(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	m := realManager(t, dir)
	require.NoError(t, m.Restore(ctx))
	require.Equal(t, "waiting", m.Status().EngineStatus)
	good := minimalConfig(t)
	goodRevision := revision(t, good)
	require.NoError(t, m.Apply(ctx, goodRevision, good))
	require.Equal(t, Status{EngineStatus: "running", AppliedRevision: goodRevision, AttemptedRevision: goodRevision}, m.Status())
	info, err := os.Stat(filepath.Join(dir, "state.json"))
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0600), info.Mode().Perm())
	saved, err := m.store.load()
	require.NoError(t, err)
	exporter := saved.WorkingConfig["exporters"].(map[string]any)["otlphttp"].(map[string]any)
	require.Equal(t, false, exporter["sending_queue"].(map[string]any)["enabled"])

	invalid := minimalConfig(t)
	invalid["receivers"] = map[string]any{"does_not_exist": map[string]any{}}
	invalidRevision := revision(t, invalid)
	require.ErrorContains(t, m.Apply(ctx, invalidRevision, invalid), "validate configuration")
	require.Equal(t, "running", m.Status().EngineStatus)
	require.Equal(t, goodRevision, m.Status().AppliedRevision)
	require.Equal(t, invalidRevision, m.Status().AttemptedRevision)

	// Extension endpoints bind during startup rather than DryRun, exercising a
	// real failed startup and restoration of the prior embedded Collector.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()
	startupFailure := minimalConfig(t)
	startupFailure["extensions"] = map[string]any{"health_check": map[string]any{"endpoint": listener.Addr().String()}}
	startupFailure["service"].(map[string]any)["extensions"] = []any{"health_check"}
	failedRevision := revision(t, startupFailure)
	require.ErrorContains(t, m.Apply(ctx, failedRevision, startupFailure), "start configuration")
	require.Equal(t, "running", m.Status().EngineStatus)
	require.Equal(t, goodRevision, m.Status().AppliedRevision)
	require.Equal(t, failedRevision, m.Status().AttemptedRevision)
	require.NotEmpty(t, m.Status().ConfigError)

	require.NoError(t, m.Close(ctx))
	restarted := realManager(t, dir)
	require.NoError(t, restarted.Restore(ctx))
	require.Equal(t, "running", restarted.Status().EngineStatus)
	require.Equal(t, goodRevision, restarted.Status().AppliedRevision)
	require.Equal(t, failedRevision, restarted.Status().AttemptedRevision)
	require.NotEmpty(t, restarted.Status().ConfigError)
}

func TestManagerPersistenceFailureDoesNotAcknowledgeCandidate(t *testing.T) {
	ctx := context.Background()
	m := realManager(t, t.TempDir())
	good := minimalConfig(t)
	goodRevision := revision(t, good)
	require.NoError(t, m.Apply(ctx, goodRevision, good))
	m.save = func(savedState) error { return errors.New("disk full") }
	candidate := minimalConfig(t)
	candidate["exporters"].(map[string]any)["debug"] = map[string]any{"verbosity": "detailed"}
	candidateRevision := revision(t, candidate)
	require.ErrorContains(t, m.Apply(ctx, candidateRevision, candidate), "disk full")
	status := m.Status()
	require.Equal(t, "running", status.EngineStatus)
	require.Equal(t, goodRevision, status.AppliedRevision)
	require.Equal(t, candidateRevision, status.AttemptedRevision)
	require.Contains(t, status.ConfigError, "persist working configuration")
	saved, err := m.store.load()
	require.NoError(t, err)
	require.Equal(t, goodRevision, saved.AppliedRevision)
}

func TestManagerRejectsMismatchedRevisionsAndCorruptedCache(t *testing.T) {
	ctx := context.Background()
	engine := &fakeEngine{}
	m, err := NewManager(t.TempDir(), "installation", "node", engine)
	require.NoError(t, err)
	config := minimalConfig(t)
	wrongRevision := "sha256:" + strings.Repeat("0", 64)
	require.ErrorContains(t, m.Apply(ctx, wrongRevision, config), "does not match")
	require.Empty(t, m.Status().AppliedRevision)
	require.Equal(t, 0, engine.starts)
	require.NoError(t, m.store.save(savedState{InstallationID: "installation", NodeName: "node", WorkingConfig: config, AppliedRevision: wrongRevision}))
	restarted, err := NewManager(m.dir, "installation", "node", engine)
	require.NoError(t, err)
	require.ErrorContains(t, restarted.Restore(ctx), "does not match")
	require.Empty(t, restarted.Status().AppliedRevision)
	require.Empty(t, restarted.Status().AttemptedRevision)
	require.Equal(t, "waiting", restarted.Status().EngineStatus)
}

func TestManagerRecoversOnlyWorkingConfigurationWithBoundedRetry(t *testing.T) {
	ctx := context.Background()
	engine := &fakeEngine{}
	m, err := NewManager(t.TempDir(), "installation", "node", engine)
	require.NoError(t, err)
	good := minimalConfig(t)
	goodRevision := revision(t, good)
	require.NoError(t, m.Apply(ctx, goodRevision, good))
	engine.crash()
	require.Equal(t, "failed", m.Status().EngineStatus)
	require.Empty(t, m.Status().AppliedRevision)
	require.NoError(t, m.Recover(ctx))
	require.Equal(t, goodRevision, m.Status().AppliedRevision)
	require.Equal(t, 2, engine.starts)
	engine.crash()
	engine.startError = errors.New("runtime failure")
	require.Error(t, m.Recover(ctx))
	attempts := engine.starts
	require.NoError(t, m.Recover(ctx))
	require.Equal(t, attempts, engine.starts, "immediate recovery retry must be suppressed")
	require.Equal(t, "failed", m.Status().EngineStatus)
	require.Empty(t, m.Status().AppliedRevision)

	fresh, err := NewManager(t.TempDir(), "installation", "node", &fakeEngine{startError: errors.New("bad candidate")})
	require.NoError(t, err)
	require.Error(t, fresh.Apply(ctx, goodRevision, good))
	require.NoError(t, fresh.Recover(ctx), "no working configuration exists to retry")
}

type fakeEngine struct {
	mu         sync.Mutex
	running    bool
	message    string
	starts     int
	startError error
}

func (e *fakeEngine) Validate(context.Context, string) error { return nil }
func (e *fakeEngine) Start(context.Context, string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.starts++
	if e.startError != nil {
		return e.startError
	}
	e.running, e.message = true, ""
	return nil
}
func (e *fakeEngine) Stop(context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.running, e.message = false, ""
	return nil
}
func (e *fakeEngine) Status() adapters.EngineStatus {
	e.mu.Lock()
	defer e.mu.Unlock()
	return adapters.EngineStatus{Running: e.running, Error: e.message}
}
func (e *fakeEngine) crash() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.running, e.message = false, "engine crashed"
}
