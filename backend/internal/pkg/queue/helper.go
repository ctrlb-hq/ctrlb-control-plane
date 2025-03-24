package queue

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func sendAgentCommand(hostname string, command string, body *any) error {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("error marshaling body: %w", err)
	}
	url := fmt.Sprintf("http://%s:443/agent/v1/%s", hostname, command)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error encountered while %s agent: %w", command, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error encountered while %s agent: %s", command, resp.Status)
	}

	return nil
}
