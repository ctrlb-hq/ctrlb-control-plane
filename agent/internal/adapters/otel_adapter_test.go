package adapters_test

import (
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/adapters"
	"github.com/stretchr/testify/assert"
)

func TestNewOTELAdapter(t *testing.T) {
	wg := &sync.WaitGroup{}
	adapter := adapters.NewOTELAdapter(wg)
	assert.NotNil(t, adapter)
}

func TestInitializeTwiceShouldFail(t *testing.T) {
	wg := &sync.WaitGroup{}
	adapter := adapters.NewOTELAdapter(wg)

	err := adapter.Initialize()
	assert.NoError(t, err)

	err = adapter.Initialize()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already initialized")
}

func TestStartAgentTwiceShouldFail(t *testing.T) {
	wg := &sync.WaitGroup{}
	adapter := adapters.NewOTELAdapter(wg)

	err := adapter.StartAgent()
	assert.NoError(t, err)

	err = adapter.StartAgent()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")
}

func TestStopAgentWithoutStartShouldFail(t *testing.T) {
	wg := &sync.WaitGroup{}
	adapter := adapters.NewOTELAdapter(wg)

	err := adapter.StopAgent()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not currently running")
}

func TestStopAgentAfterStart(t *testing.T) {
	wg := &sync.WaitGroup{}
	adapter := adapters.NewOTELAdapter(wg)

	err := adapter.StartAgent()
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond) // give goroutine time to run
	err = adapter.StopAgent()
	assert.NoError(t, err)
}

func TestUpdateConfig(t *testing.T) {
	wg := &sync.WaitGroup{}
	adapter := adapters.NewOTELAdapter(wg)

	go func() {
		time.Sleep(1 * time.Second)
		// simulate config reload signal to avoid blocking
		_ = syscall.Kill(os.Getpid(), syscall.SIGHUP)
	}()

	err := adapter.UpdateConfig()
	assert.NoError(t, err)
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
