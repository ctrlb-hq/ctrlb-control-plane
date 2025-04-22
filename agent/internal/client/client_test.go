package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockSystemInfo struct {
	Hostname string
	IP       string
	HostErr  error
	IPFail   bool
}

func (m *MockSystemInfo) GetHostname() (string, error) {
	return m.Hostname, nil
}

func (m *MockSystemInfo) LookupHost(name string) ([]string, error) {
	if m.HostErr != nil {
		return nil, m.HostErr
	}
	return []string{"127.0.0.1"}, nil
}

func (m *MockSystemInfo) GetLocalIP() (string, error) {
	if m.IPFail {
		return "", fmt.Errorf("mock IP error")
	}
	return m.IP, nil
}

func TestInformBackendServerStart_Success(t *testing.T) {
	// Setup mock system info
	mockSys := &MockSystemInfo{
		Hostname: "mock-host",
		IP:       "127.0.0.1",
	}

	// Setup a test HTTP server to simulate backend
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/agent/v1/agents", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()

		var req AgentRequest
		json.Unmarshal(body, &req)

		assert.Equal(t, "127.0.0.1", req.IP)
		assert.Equal(t, "mock-host", req.Hostname)

		resp := AgentResponse{
			ID: 12345,
			Config: map[string]any{
				"some_config": "value",
			},
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer testServer.Close()

	// Override backend URL for the test
	originalURL := constants.BACKEND_URL
	constants.BACKEND_URL = strings.TrimPrefix(testServer.URL, "http://")
	defer func() { constants.BACKEND_URL = originalURL }()

	client := &http.Client{}
	cfg, err := InformBackendServerStart(mockSys, client)

	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "value", cfg["some_config"])
	assert.Equal(t, int64(12345), constants.AGENTID)
}

func TestInformBackendConfigFileChanged_Success(t *testing.T) {
	// Set agent ID for test
	constants.AGENTID = 123

	// Start mock HTTP server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/agent/v1/agents/123/config-changed", r.URL.Path)

		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()

		var payload map[string]string
		json.Unmarshal(body, &payload)

		assert.Contains(t, payload["message"], "Config file modified")
		assert.NotEmpty(t, payload["time"])

		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	// Override backend URL temporarily
	originalBackend := constants.BACKEND_URL
	constants.BACKEND_URL = testServer.URL
	defer func() { constants.BACKEND_URL = originalBackend }()

	err := InformBackendConfigFileChanged(testServer.Client())
	assert.NoError(t, err)
}
