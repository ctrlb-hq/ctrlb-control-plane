package serverclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/ctrlb-hq/ctrlb-collector/internal/constants"
	"github.com/ctrlb-hq/ctrlb-collector/internal/utils"
)

func InformBackendServerStart() error {
	var hostname string
	var err error

	if constants.ENV == "prod" {
		hostname, err = os.Hostname()
		if err != nil {
			return fmt.Errorf("failed to get hostname: %v", err)
		}
	} else {
		hostname = "localhost"
	}

	platform := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)

	jsonData, err := utils.LoadYAMLToJSON(constants.AGENT_CONFIG_PATH)
	if err != nil {
		return fmt.Errorf("failed to load YAML config while informing server: %v", err)
	}

	configJSON, err := json.Marshal(jsonData)
	if err != nil {
		return fmt.Errorf("failed to marshal config to JSON while informing server: %v", err)
	}

	configJSONString := string(configJSON)

	agentRequest := map[string]string{
		"type":     constants.AGENT_TYPE,
		"version":  constants.AGENT_VERSION,
		"hostname": hostname,
		"platform": platform,
		"config":   configJSONString,
	}

	requestBody, err := json.Marshal(agentRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal agent request while informing server: %v", err)
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("http://%s:8096/api/v1/agent/register", constants.BACKEND_URL), bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("error encountered while informing server: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error encountered while informing server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status while informing server: %v", resp.Status)
	}

	return nil
}
