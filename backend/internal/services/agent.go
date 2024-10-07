package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/ctrlb-hq/all-father/internal/models"
	"github.com/ctrlb-hq/all-father/internal/repositories"
	"github.com/ctrlb-hq/all-father/internal/utils"
)

func NewAgentService(agentRepository *repositories.AgentRepository, agentQueue *AgentQueue) *AgentService {
	return &AgentService{
		AgentRepository: agentRepository,
		AgentQueue:      agentQueue,
	}
}

func (a *AgentService) RegisterAgent(request models.AgentRegisterRequest) (interface{}, error) {
	var agent models.Agent

	agent.Hostname = request.Hostname
	agent.Platform = request.Platform
	agent.Version = request.Version
	agent.Type = request.Type
	agent.Name = utils.GenerateAgentName(agent.Type, agent.Version, agent.Hostname)
	agent.Config = request.Config
	agent.ID = utils.CreateNewUUID()

	log.Println("Received registration request from agent:", agent.Name)

	err := a.AgentRepository.AddAgent(&agent)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	log.Println("Agent registered with ID: ", agent.ID)
	//TODO: Add agent to AgentQueue

	return agent, nil
}

func (a *AgentService) UpdateConfig(request models.ConfigUpdateRequest) (map[string]string, error) {

	reqBodyStruct := map[string]string{
		"config": request.Config,
	}

	reqBody, err := json.Marshal(reqBodyStruct)
	if err != nil {
		return nil, errors.New("unable to marshal agent config: Bad config")
	}

	hostname, err := a.AgentRepository.GetAgentHost(request.AgentID)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("http://%s:443/api/v1/config", hostname)

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, errors.New("failed in requesting agent to update config: " + err.Error())
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("failed in requesting agent to update config" + err.Error())
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("encountered error while parsing agent response for update config: " + err.Error())
	}
	log.Println("Response while updating agent config:", string(respBody))

	err = a.AgentRepository.UpdateConfig(request)
	if err != nil {
		return nil, err
	}
	log.Println("config updated for agent id:", request.AgentID)

	//TODO: Add agent to AgentQueue

	jsonStr := `{"message": "Configuration has been updated"}`
	var result map[string]string
	_ = json.Unmarshal([]byte(jsonStr), &result)

	return result, nil
}

func (a *AgentService) RemoveAgent(unregisterRequest models.AgentRequest) (map[string]string, error) {

	// shutting down registered agent
	hostname, err := a.AgentRepository.GetAgentHost(unregisterRequest.AgentID)
	if err != nil {
		return nil, err
	}

	// prepare the URL for the shutdown request using the extracted hostname
	url := fmt.Sprintf("http://%s:443/api/v1/shutdown", hostname)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return nil, fmt.Errorf("error encountered while removing agent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error encountered while removing agent: %s", resp.Status)
	}

	// removing agent from registry db
	err = a.AgentRepository.RemoveAgent(unregisterRequest.AgentID)
	if err != nil {
		return nil, err
	}

	jsonStr := `{"message": "Agent removed successfully"}`
	var result map[string]string
	_ = json.Unmarshal([]byte(jsonStr), &result)

	return result, nil
}

func (a *AgentService) StartAgent(startRequest models.AgentRequest) (map[string]string, error) {

	// starting registered agent
	hostname, err := a.AgentRepository.GetAgentHost(startRequest.AgentID)
	if err != nil {
		return nil, err
	}

	// prepare the URL for the start agent request using the extracted hostname
	url := fmt.Sprintf("http://%s:443/api/v1/start", hostname)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return nil, fmt.Errorf("error encountered while starting agent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error encountered while starting agent: %s", resp.Status)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response body while starting agent: %w", err)
	}

	return result, nil
}

func (a *AgentService) StopAgent(startRequest models.AgentRequest) (map[string]string, error) {

	// stopping registered agent
	hostname, err := a.AgentRepository.GetAgentHost(startRequest.AgentID)
	if err != nil {
		return nil, err
	}

	// prepare the URL for the stop agent request using the extracted hostname
	url := fmt.Sprintf("http://%s:443/api/v1/stop", hostname)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return nil, fmt.Errorf("error encountered while stopping agent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error encountered while stopping agent: %s", resp.Status)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response body while stopping agent: %w", err)
	}

	return result, nil
}

func (a *AgentService) GetAgentConfig(agentConfigRequest models.AgentRequest) (map[string]interface{}, error) {

	// stopping registered agent
	config, err := a.AgentRepository.GetAgentConfig(agentConfigRequest.AgentID)
	if err != nil {
		return nil, err
	}

	log.Println("Agent config fetched successfully")
	var result map[string]interface{}
	_ = json.Unmarshal([]byte(config), &result)

	return result, nil
}

func (a *AgentService) GetAgentUptime(agentUptimeRequest models.AgentRequest) (map[string]interface{}, error) {

	hostname, err := a.AgentRepository.GetAgentHost(agentUptimeRequest.AgentID)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("http://%s:443/api/v1/uptime", hostname)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error encountered while fetching agent uptime: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error encountered while fetching agent uptime: %s", resp.Status)
	}

	var uptimeResult map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&uptimeResult); err != nil {
		return nil, fmt.Errorf("failed to decode response body while fetching agent uptime: %w", err)
	}

	log.Println("Agent uptime fetched successfully")

	result := make(map[string]interface{})

	if status, ok := uptimeResult["status"].(string); ok {
		result["status"] = status
	}
	if uptime, ok := uptimeResult["uptime"].(float64); ok {
		result["uptime"] = fmt.Sprintf("%d", int(uptime))
	}

	return result, nil
}

func (a *AgentService) GetAgentStatus(agentStatusRequest models.AgentRequest) (map[string]string, error) {

	hostname, err := a.AgentRepository.GetAgentHost(agentStatusRequest.AgentID)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("http://%s:443/api/v1/status", hostname)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error encountered while fetching agent status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error encountered while fetching agent status: %s", resp.Status)
	}

	var statusResult map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&statusResult); err != nil {
		return nil, fmt.Errorf("failed to decode response body while fetching agent status: %w", err)
	}

	log.Println("Agent status fetched successfully")

	result := make(map[string]string)
	if uptime, ok := statusResult["Uptime"].(string); ok {
		result["Uptime"] = uptime
	}
	if exportedDataVolume, ok := statusResult["ExportedDataVolume"].(string); ok {
		result["ExportedDataVolume"] = exportedDataVolume
	}
	if droppedRecords, ok := statusResult["DroppedRecords"].(string); ok {
		result["DroppedRecords"] = droppedRecords
	}
	if status, ok := statusResult["Status"].(string); ok {
		result["Status"] = status
	}

	return result, nil
}

func (a *AgentService) GetAgentGraph(agentGraphRequest models.AgentRequest) (map[string]interface{}, error) {

	agentConfig, err := a.AgentRepository.GetAgentConfig(agentGraphRequest.AgentID)
	if err != nil {
		return nil, err
	}

	log.Println(agentConfig)

	agentType, err := a.AgentRepository.GetAgentType(agentGraphRequest.AgentID)
	if err != nil {
		return nil, err
	}

	graphData, err := utils.DrawGraph(agentConfig, agentType)
	if err != nil {
		return nil, err
	}

	return graphData, nil
}
