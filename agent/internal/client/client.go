package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/constants"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/pkg/systeminfo"
)

func InformBackendServerStart(sys systeminfo.SystemInfoProvider,
	httpClient *http.Client) (map[string]any, error) {
	// Step 1: Get hostname or fallback to IP
	hostname, err := sys.GetHostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %v", err)
	}

	// Check if the hostname resolves to a valid DNS entry
	if _, err := sys.LookupHost(hostname); err != nil {
		// If DNS resolution fails, fallback to IP address
		hostname, err = sys.GetLocalIP()
		if err != nil {
			return nil, fmt.Errorf("failed to get IP address: %v", err)
		}
	}

	// Step 2: Gather platform information
	platform := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	ip, err := sys.GetLocalIP()
	if err != nil {
		return nil, fmt.Errorf("failed to get IP address: %v", err)
	}
	// Step 3: Create the agent request
	agentRequest := AgentRequest{
		IP:           ip,
		Version:      constants.AGENT_VERSION,
		Platform:     platform,
		Hostname:     hostname,
		PipelineName: constants.PIPELINE_NAME,
	}

	// Step 4: Marshal the agent request into JSON
	requestBody, err := json.Marshal(agentRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal agent request: %v", err)
	}

	// Step 5: Create the HTTP request to inform the backend server
	url := fmt.Sprintf("http://%s/api/agent/v1/agents", constants.BACKEND_URL)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Step 6: Execute the HTTP request
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Step 7: Check the response status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	// Step 8: Unmarshal the response body into models.AgentWithConfig
	var agentResponse AgentResponse
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&agentResponse); err != nil {
		return nil, fmt.Errorf("error decoding response body: %v", err)
	}

	constants.AGENTID = agentResponse.ID

	return agentResponse.Config, nil
}

func InformBackendConfigFileChanged(client *http.Client) error {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	url := fmt.Sprintf("%s/api/agent/v1/agents/%v/config-changed", constants.BACKEND_URL, constants.AGENTID)

	info := map[string]string{
		"time":    time.Now().Format(time.RFC3339),
		"message": "Config file modified/deleted",
	}

	jsonPayload, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("non-2xx response: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}
