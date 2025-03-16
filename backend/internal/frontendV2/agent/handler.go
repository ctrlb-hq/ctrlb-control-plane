package frontendagent

import (
	"fmt"
	"net/http"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
	"github.com/gorilla/mux"
)

// FrontendAgentHandler handles frontend agent operations
type FrontendAgentHandler struct {
	FrontendAgentService *FrontendAgentService
}

// NewFrontendAgentHandler initializes the handler
func NewFrontendAgentHandler(frontendAgentServices *FrontendAgentService) *FrontendAgentHandler {
	return &FrontendAgentHandler{
		FrontendAgentService: frontendAgentServices,
	}
}

// GetAllAgents retrieves all agents
func (f *FrontendAgentHandler) GetAllAgents(w http.ResponseWriter, r *http.Request) {
	utils.Logger.Info("Received request to get all agents")
	response, err := f.FrontendAgentService.GetAllAgents()
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("Error getting all agents: %s", err))
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (f *FrontendAgentHandler) GetUnmanagedAgents(w http.ResponseWriter, r *http.Request) {
	utils.Logger.Info("Received request to get all agents")
	response, err := f.FrontendAgentService.GetAllUnmanagedAgents()
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("Error getting all agents: %s", err))
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// GetAgent retrieves a specific agent by ID
func (f *FrontendAgentHandler) GetAgent(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	utils.Logger.Info(fmt.Sprintf("Getting agent with ID: %s", id))

	response, err := f.FrontendAgentService.GetAgent(id)
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("Error getting agent [ID: %s]: %v", id, err))
		utils.SendJSONError(w, http.StatusNotFound, "Agent not found")
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// DeleteAgent removes an agent by ID
func (f *FrontendAgentHandler) DeleteAgent(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	utils.Logger.Info(fmt.Sprintf("Deleting agent with ID: %s", id))

	if err := f.FrontendAgentService.DeleteAgent(id); err != nil {
		utils.Logger.Error(fmt.Sprintf("Error deleting agent [ID: %s]: %s", id, err))
		if err.Error() == "agent not found" {
			utils.SendJSONError(w, http.StatusNotFound, err.Error())
		} else {
			utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "Agent deleted [ID: " + id + "]."})
}

// StartAgent starts an agent by ID
func (f *FrontendAgentHandler) StartAgent(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	utils.Logger.Info(fmt.Sprintf("Starting agent with ID: %s", id))

	if err := f.FrontendAgentService.StartAgent(id); err != nil {
		utils.Logger.Error(fmt.Sprintf("Error starting agent [ID: %s]: %s", id, err))
		if err.Error() == "no agent found to start" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "Agent started [ID: " + id + "]."})
}

// StopAgent stops an agent by ID
func (f *FrontendAgentHandler) StopAgent(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	utils.Logger.Info(fmt.Sprintf("Stopping agent with ID: %s", id))

	if err := f.FrontendAgentService.StopAgent(id); err != nil {
		utils.Logger.Error(fmt.Sprintf("Error stopping agent [ID: %s]: %s", id, err))
		if err.Error() == "no agent found to stop" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "Agent stopped [ID: " + id + "]."})
}

// RestartMonitoring restarts monitoring for a specific agent
func (f *FrontendAgentHandler) RestartMonitoring(w http.ResponseWriter, r *http.Request) {

	id := mux.Vars(r)["id"]

	if err := f.FrontendAgentService.RestartMonitoring(id); err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "Monitoring started for agent [ID: " + id + "]."})
}

// GetHealthMetricsForGraph retrieves metrics for a specific agent
func (f *FrontendAgentHandler) GetHealthMetricsForGraph(w http.ResponseWriter, r *http.Request) {

	id := mux.Vars(r)["id"]
	utils.Logger.Info(fmt.Sprintf("Getting health metrics for agent with ID: %s", id))

	response, err := f.FrontendAgentService.GetHealthMetricsForGraph(id)
	if err != nil {
		if err.Error() == "agent disconnected" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (f *FrontendAgentHandler) GetRateMetricsForGraph(w http.ResponseWriter, r *http.Request) {

	id := mux.Vars(r)["id"]
	utils.Logger.Info(fmt.Sprintf("Getting rate metrics for agent with ID: %s", id))

	response, err := f.FrontendAgentService.GetRateMetricsForGraph(id)
	if err != nil {
		if err.Error() == "agent disconnected" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}
