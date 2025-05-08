package frontendagent

import (
	"fmt"
	"net/http"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
	"github.com/gorilla/mux"
)

// FrontendAgentHandler handles frontend agent operations
type FrontendAgentHandler struct {
	FrontendAgentService FrontendAgentServiceInterface
}

// NewFrontendAgentHandler initializes the handler
func NewFrontendAgentHandler(frontendAgentServices FrontendAgentServiceInterface) *FrontendAgentHandler {
	return &FrontendAgentHandler{
		FrontendAgentService: frontendAgentServices,
	}
}

// GetAllAgents retrieves all agents
func (f *FrontendAgentHandler) GetAllAgents(w http.ResponseWriter, r *http.Request) {
	utils.Logger.Info("Received request to get all agents")
	response, err := f.FrontendAgentService.GetAllAgents()
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("Error getting all agents: %s", err.Error()))
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (f *FrontendAgentHandler) GetUnmanagedAgents(w http.ResponseWriter, r *http.Request) {
	utils.Logger.Info("Received request to get all unmanaged agents")
	response, err := f.FrontendAgentService.GetAllUnmanagedAgents()
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("Error getting all agents: %s", err.Error()))
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
		utils.Logger.Error(fmt.Sprintf("Error getting agent [ID: %s]: %v", id, err.Error()))
		if err == utils.ErrAgentDoesNotExists {
			utils.SendJSONError(w, http.StatusOK, "Agent not found")
		} else {
			utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// DeleteAgent removes an agent by ID
func (f *FrontendAgentHandler) DeleteAgent(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	utils.Logger.Info(fmt.Sprintf("Deleting agent with ID: %s", id))

	if err := f.FrontendAgentService.DeleteAgent(id); err != nil {
		utils.Logger.Error(fmt.Sprintf("Error deleting agent [ID: %s]: %s", id, err.Error()))
		if err == utils.ErrAgentDoesNotExists {
			utils.SendJSONError(w, http.StatusOK, "Agent not found")
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
		utils.Logger.Error(fmt.Sprintf("Error starting agent [ID: %s]: %s", id, err.Error()))
		if err == utils.ErrAgentDoesNotExists {
			utils.SendJSONError(w, http.StatusOK, "Agent not found")
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
		utils.Logger.Error(fmt.Sprintf("Error stopping agent [ID: %s]: %s", id, err.Error()))
		if err == utils.ErrAgentDoesNotExists {
			utils.SendJSONError(w, http.StatusOK, "Agent not found")
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

	utils.Logger.Info(fmt.Sprintf("Got request to restart monitoring for agent [ID: %s]", id))
	if err := f.FrontendAgentService.RestartMonitoring(id); err != nil {
		utils.Logger.Error(fmt.Sprintf("Error occured while restarting monitoring for agent [ID: %s]: %s", id, err.Error()))
		if err == utils.ErrAgentDoesNotExists {
			utils.SendJSONError(w, http.StatusOK, "Agent not found")
		} else {
			utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		}
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
		if err == utils.ErrAgentDoesNotExists {
			utils.SendJSONError(w, http.StatusOK, "Agent not found")
		} else if err.Error() == "agent disconnected" {
			utils.SendJSONError(w, http.StatusOK, "Agent is in disconnected state")
		} else {
			utils.Logger.Error(fmt.Sprintf("Failed to get health metrics for agent [ID: %s]: %s", id, err.Error()))
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
		utils.Logger.Error(fmt.Sprintf("Failed to get rate metrics for agent [ID: %s]: %s", id, err.Error()))
		if err == utils.ErrAgentDoesNotExists {
			utils.SendJSONError(w, http.StatusOK, "Agent not found")
		} else if err.Error() == "agent disconnected" {
			utils.SendJSONError(w, http.StatusOK, "Agent is in disconnected state")
		} else {
			utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (f *FrontendAgentHandler) AddLabels(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	utils.Logger.Info(fmt.Sprintf("Adding labels to agent with ID: %s", id))

	var labels map[string]string
	if err := utils.UnmarshalJSONRequest(r, &labels); err != nil {
		utils.Logger.Error(fmt.Sprintf("Failed to decode request body: %s", err.Error()))
		utils.SendJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(labels) == 0 {
		utils.SendJSONError(w, http.StatusBadRequest, "No labels provided")
		return
	}

	if err := f.FrontendAgentService.AddLabels(id, labels); err != nil {
		utils.Logger.Error(fmt.Sprintf("Failed to add labels to agent [ID: %s]: %s", id, err.Error()))
		if err == utils.ErrAgentDoesNotExists {
			utils.SendJSONError(w, http.StatusOK, "Agent not found")
		} else {
			utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "Labels added to agent [ID: " + id + "]."})
}

func (f *FrontendAgentHandler) GetLatestAgentSince(w http.ResponseWriter, r *http.Request) {
	since := r.URL.Query().Get("since")
	if since == "" {
		utils.SendJSONError(w, http.StatusBadRequest, "Missing 'since' query parameter")
		return
	}

	utils.Logger.Info("Received request to get letest agents since: " + since)
	response, err := f.FrontendAgentService.GetLatestAgentSince(since)
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("Error getting all agents: %s", err.Error()))
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}
