package frontendagent

import (
	"net/http"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/auth"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
	"github.com/gorilla/mux"
)

// FrontendAgentHandler handles frontend agent operations
type FrontendAgentHandler struct {
	FrontendAgentService *FrontendAgentService
	BasicAuthenticator   *auth.BasicAuthenticator
}

// NewFrontendAgentHandler initializes the handler
func NewFrontendAgentHandler(frontendAgentServices *FrontendAgentService, basicAuthenticator *auth.BasicAuthenticator) *FrontendAgentHandler {
	return &FrontendAgentHandler{
		FrontendAgentService: frontendAgentServices,
		BasicAuthenticator:   basicAuthenticator,
	}
}

// authenticate validates the token from request headers
func (f *FrontendAgentHandler) authenticate(w http.ResponseWriter, r *http.Request) bool {
	token, err := utils.ExtractTokenFromHeaders(&r.Header)
	if err != nil || f.BasicAuthenticator.ValidateToken(token) != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return false
	}
	return true
}

// GetAllAgents retrieves all agents
func (f *FrontendAgentHandler) GetAllAgents(w http.ResponseWriter, r *http.Request) {
	if !f.authenticate(w, r) {
		return
	}

	response, err := f.FrontendAgentService.GetAllAgents()
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// GetAgent retrieves a specific agent by ID
func (f *FrontendAgentHandler) GetAgent(w http.ResponseWriter, r *http.Request) {
	if !f.authenticate(w, r) {
		return
	}

	id := mux.Vars(r)["id"]

	response, err := f.FrontendAgentService.GetAgent(id)
	if err != nil {
		utils.SendJSONError(w, http.StatusNotFound, "Agent not found")
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// DeleteAgent removes an agent by ID
func (f *FrontendAgentHandler) DeleteAgent(w http.ResponseWriter, r *http.Request) {
	if !f.authenticate(w, r) {
		return
	}

	id := mux.Vars(r)["id"]

	if err := f.FrontendAgentService.DeleteAgent(id); err != nil {
		if err.Error() == "agent not found" {
			utils.SendJSONError(w, http.StatusNotFound, err.Error())
		} else {
			utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "Agent deleted successfully"})
}

// StartAgent starts an agent by ID
func (f *FrontendAgentHandler) StartAgent(w http.ResponseWriter, r *http.Request) {
	if !f.authenticate(w, r) {
		return
	}

	id := mux.Vars(r)["id"]

	if err := f.FrontendAgentService.StartAgent(id); err != nil {
		if err.Error() == "no agent found to start" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "Agent started successfully"})
}

// StopAgent stops an agent by ID
func (f *FrontendAgentHandler) StopAgent(w http.ResponseWriter, r *http.Request) {
	if !f.authenticate(w, r) {
		return
	}

	id := mux.Vars(r)["id"]

	if err := f.FrontendAgentService.StopAgent(id); err != nil {
		if err.Error() == "no agent found to stop" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "Agent stopped successfully"})
}

// GetMetrics retrieves metrics for a specific agent
func (f *FrontendAgentHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	if !f.authenticate(w, r) {
		return
	}

	id := mux.Vars(r)["id"]

	response, err := f.FrontendAgentService.GetMetrics(id)
	if err != nil {
		if err.Error() == "no agent found to fetch config" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// RestartMonitoring restarts monitoring for a specific agent
func (f *FrontendAgentHandler) RestartMonitoring(w http.ResponseWriter, r *http.Request) {
	if !f.authenticate(w, r) {
		return
	}

	id := mux.Vars(r)["id"]

	if err := f.FrontendAgentService.RestartMonitoring(id); err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "Started monitoring the agent"})
}
