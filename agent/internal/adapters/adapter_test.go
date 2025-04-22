package adapters_test

import (
	"sync"
	"testing"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/adapters"
	"github.com/stretchr/testify/assert"
)

func TestNewAdapter_OTEL(t *testing.T) {
	wg := &sync.WaitGroup{}

	adapter, err := adapters.NewAdapter(wg, "otel")
	assert.NoError(t, err)
	assert.NotNil(t, adapter)

	// Optional: check actual type
	_, ok := adapter.(*adapters.OTELAdapter)
	assert.True(t, ok, "adapter should be of type OTELAdapter")
}

func TestNewAdapter_DefaultEmptyType(t *testing.T) {
	wg := &sync.WaitGroup{}

	adapter, err := adapters.NewAdapter(wg, "")
	assert.NoError(t, err)
	assert.NotNil(t, adapter)
}

func TestNewAdapter_UnsupportedType(t *testing.T) {
	wg := &sync.WaitGroup{}

	adapter, err := adapters.NewAdapter(wg, "fluentbit")
	assert.Error(t, err)
	assert.Nil(t, adapter)
	assert.Contains(t, err.Error(), "unsupported agent type")
}
