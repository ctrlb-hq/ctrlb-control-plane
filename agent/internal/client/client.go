package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/constants"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/utils"
)

func InformBackendServerStart() (map[string]any, error) {
	// Step 1: Get hostname or fallback to IP
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %v", err)
	}

	// Check if the hostname resolves to a valid DNS entry
	if _, err := net.LookupHost(hostname); err != nil {
		// If DNS resolution fails, fallback to IP address
		hostname, err = utils.GetLocalIP()
		if err != nil {
			return nil, fmt.Errorf("failed to get IP address: %v", err)
		}
	}

	// Step 2: Gather platform information
	platform := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)

	// Step 3: Create the agent request
	agentRequest := AgentRequest{
		Version:  constants.AGENT_VERSION,
		Platform: platform,
		Hostname: hostname,
	}

	// Step 4: Marshal the agent request into JSON
	requestBody, err := json.Marshal(agentRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal agent request: %v", err)
	}

	// Step 5: Create the HTTP request to inform the backend server
	url := fmt.Sprintf("http://%s/api/agent/v1/agents", constants.BACKEND_URL)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Step 6: Execute the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
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

func InformBackendConfigFileChanged() error {
	// 1. Construct the POST URL
	url := fmt.Sprintf("%s/api/agent/v1/agents/%s/config-changed",
		constants.BACKEND_URL,
		constants.AGENTID,
	)

	// 2. Prepare the payload
	info := map[string]string{
		"time":    time.Now().Format(time.RFC3339),
		"message": "Config file modified/deleted",
	}

	// 3. Marshal the payload to JSON
	jsonPayload, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %v", err)
	}

	// 4. Create the HTTP request
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 5. Send the request
	client := &http.Client{
		Timeout: 5 * time.Second, // choose a sensible timeout
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// 6. Check for a non-2xx status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("received non-2xx status code: %d", resp.StatusCode)
	}

	return nil
}
