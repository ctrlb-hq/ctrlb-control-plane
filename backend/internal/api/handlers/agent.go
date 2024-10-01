package handler

import (
	"log"
	"net/http"

	"github.com/ctrlb-hq/all-father/internal/models"
	"github.com/ctrlb-hq/all-father/internal/services"
	"github.com/ctrlb-hq/all-father/internal/utils"
)

var agentHandler *AgentHandler

func NewAgentHandler(services *services.Services) *AgentHandler {
	agentHandler = &AgentHandler{
		Services: services,
	}
	return agentHandler
}

func (a *AgentHandler) RegisterAgent(w http.ResponseWriter, r *http.Request) {
	var registerRequest models.AgentRegisterRequest

	if err := utils.UnmarshalJSONRequest(r, &registerRequest); err != nil {
		log.Println("Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	reponse, err := a.Services.AgentService.RegisterAgent(registerRequest)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, reponse)

}

func (a *AgentHandler) RemoveAgent(w http.ResponseWriter, r *http.Request) {
	var unregisterRequest models.AgentRequest

	if err := utils.UnmarshalJSONRequest(r, &unregisterRequest); err != nil {
		log.Println("Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := a.Services.AgentService.RemoveAgent(unregisterRequest)
	if err != nil {
		if err.Error() == "no agent found to delete" {
			http.Error(w, err.Error(), http.StatusNotFound) // Return 404 if agent not found
		} else {
			utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (a *AgentHandler) StartAgent(w http.ResponseWriter, r *http.Request) {
	var startRequest models.AgentRequest

	if err := utils.UnmarshalJSONRequest(r, &startRequest); err != nil {
		log.Println("Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := a.Services.AgentService.StartAgent(startRequest)
	if err != nil {
		if err.Error() == "no agent found to start" {
			http.Error(w, err.Error(), http.StatusNotFound) // Return 404 if agent not found
		} else {
			utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (a *AgentHandler) StopAgent(w http.ResponseWriter, r *http.Request) {
	var stopRequest models.AgentRequest

	if err := utils.UnmarshalJSONRequest(r, &stopRequest); err != nil {
		log.Println("Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := a.Services.AgentService.StopAgent(stopRequest)
	if err != nil {
		if err.Error() == "no agent found to stop" {
			http.Error(w, err.Error(), http.StatusNotFound) // Return 404 if agent not found
		} else {
			utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (a *AgentHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var updateRequest models.ConfigUpdateRequest

	if err := utils.UnmarshalJSONRequest(r, &updateRequest); err != nil {
		log.Println("Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	reponse, err := a.Services.AgentService.UpdateConfig(updateRequest)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, reponse)

}

func (a *AgentHandler) GetAgentConfig(w http.ResponseWriter, r *http.Request) {
	var agentConfigRequest models.AgentRequest

	if err := utils.UnmarshalJSONRequest(r, &agentConfigRequest); err != nil {
		log.Println("Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := a.Services.AgentService.GetAgentConfig(agentConfigRequest)
	if err != nil {
		if err.Error() == "no agent found to fetch config" {
			http.Error(w, err.Error(), http.StatusNotFound) // Return 404 if agent not found
		} else {
			utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (a *AgentHandler) GetAgentUptime(w http.ResponseWriter, r *http.Request) {
	var agentUptimeRequest models.AgentRequest

	if err := utils.UnmarshalJSONRequest(r, &agentUptimeRequest); err != nil {
		log.Println("Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := a.Services.AgentService.GetAgentUptime(agentUptimeRequest)
	if err != nil {
		if err.Error() == "no agent found to fetch uptime" {
			http.Error(w, err.Error(), http.StatusNotFound) // Return 404 if agent not found
		} else {
			utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (a *AgentHandler) GetAgentStatus(w http.ResponseWriter, r *http.Request) {
	var agentStatusRequest models.AgentRequest

	if err := utils.UnmarshalJSONRequest(r, &agentStatusRequest); err != nil {
		log.Println("Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := a.Services.AgentService.GetAgentStatus(agentStatusRequest)
	if err != nil {
		if err.Error() == "no agent found to fetch status" {
			http.Error(w, err.Error(), http.StatusNotFound) // Return 404 if agent not found
		} else {
			utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}
