package adapters_test

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/adapters"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOTELAdapter(t *testing.T) {
	wg := &sync.WaitGroup{}
	adapter := adapters.NewOTELAdapter(wg)
	assert.NotNil(t, adapter)
}

func TestInitializeTwiceShouldFail(t *testing.T)       { legacyScenario(t, "initialize_twice") }
func TestStartAgentTwiceShouldFail(t *testing.T)       { legacyScenario(t, "start_twice") }
func TestStopAgentWithoutStartShouldFail(t *testing.T) { legacyScenario(t, "stop_without_start") }
func TestStopAgentAfterStart(t *testing.T)             { legacyScenario(t, "stop_after_start") }
func TestUpdateConfig(t *testing.T)                    { legacyScenario(t, "update_config") }

// Legacy reloading signals the process. Run these cases in child test processes
// with a real config and bounded cleanup so signals cannot escape into the suite.
func legacyScenario(t *testing.T, scenario string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestLegacyAdapterLifecycleHelper$", "-test.timeout=10s")
	cmd.Env = append(os.Environ(), "CTRL_LEGACY_TEST_SCENARIO="+scenario)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "%s", output)
}

func TestLegacyAdapterLifecycleHelper(t *testing.T) {
	scenario := os.Getenv("CTRL_LEGACY_TEST_SCENARIO")
	if scenario == "" {
		t.Skip("subprocess helper")
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	endpoint := listener.Addr().String()
	require.NoError(t, listener.Close())
	path := filepath.Join(t.TempDir(), "collector.json")
	config := fmt.Sprintf(`{"receivers":{"otlp":{"protocols":{"grpc":{"endpoint":%q}}}},"exporters":{"debug":{}},"service":{"telemetry":{"metrics":{"level":"none"}},"pipelines":{"logs":{"receivers":["otlp"],"exporters":["debug"]}}}}`, endpoint)
	require.NoError(t, os.WriteFile(path, []byte(config), 0600))
	constants.AGENT_CONFIG_PATH = path
	wg := &sync.WaitGroup{}
	adapter := adapters.NewOTELAdapter(wg)
	if scenario == "stop_without_start" {
		require.ErrorContains(t, adapter.StopAgent(), "not currently running")
		return
	}
	if scenario == "initialize_twice" {
		require.NoError(t, adapter.Initialize())
		require.ErrorContains(t, adapter.Initialize(), "already initialized")
	} else {
		require.NoError(t, adapter.StartAgent())
		if scenario == "start_twice" {
			require.ErrorContains(t, adapter.StartAgent(), "already running")
		}
	}
	defer func() {
		_ = adapter.StopAgent()
		done := make(chan struct{})
		go func() { wg.Wait(); close(done) }()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("legacy Collector did not stop")
		}
	}()
	require.Eventually(t, func() bool {
		connection, err := net.DialTimeout("tcp", endpoint, 20*time.Millisecond)
		if err != nil {
			return false
		}
		connection.Close()
		return true
	}, 3*time.Second, 10*time.Millisecond)
	if scenario == "update_config" {
		require.NoError(t, adapter.UpdateConfig())
	} else if scenario == "stop_after_start" {
		require.NoError(t, adapter.StopAgent())
	}
}

func TestGetVersion(t *testing.T) {
	wg := &sync.WaitGroup{}
	adapter := adapters.NewOTELAdapter(wg)

	version, err := adapter.GetVersion()
	if err != nil {
		t.Log("Could not get OTEL version, skipping: ", err)
	} else {
		assert.NotEmpty(t, version)
	}
}

func TestGracefulShutdown(t *testing.T) {
	t.Skip("GracefulShutdown calls os.Exit(0), which terminates the test process")
}
