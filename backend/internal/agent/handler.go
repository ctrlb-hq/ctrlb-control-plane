package agent

import (
	"fmt"
	"net/http"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
	"github.com/gorilla/mux"
)

// AgentHandler is responsible for handling HTTP requests related to agents.
type AgentHandler struct {
	AgentService AgentServiceInterface // Service for managing agent operations
}

// NewAgentHandler creates a new instance of AgentHandler with the provided services.
func NewAgentHandler(agentService AgentServiceInterface) *AgentHandler {
	return &AgentHandler{
		AgentService: agentService,
	}
}

// RegisterAgent handles the registration of a new agent.
// It expects a JSON payload in the request body.
func (a *AgentHandler) RegisterAgent(w http.ResponseWriter, r *http.Request) {
	req := &models.AgentRegisterRequest{}

	// Unmarshal the JSON request body into the request struct
	if err := utils.UnmarshalJSONRequest(r, req); err != nil {
		utils.Logger.Error("Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := utils.ValidateAgentRegisterRequest(req); err != nil {
		utils.Logger.Error(fmt.Sprintf("Invalid request: %v", err))
		utils.SendJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.Logger.Info(fmt.Sprintf("Received registration request from agent: %s", req.Hostname))

	response, err := a.AgentService.RegisterAgent(req)
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("Error registering agent: %v", err))
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// ConfigChangedPing handles the ping from an agent indicating that its configuration has changed.
func (a *AgentHandler) ConfigChangedPing(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["id"]

	utils.Logger.Info(fmt.Sprintf("Received config changed ping from agent: %s", agentID))

	if err := a.AgentService.ConfigChangedPing(agentID); err != nil {
		utils.Logger.Error(fmt.Sprintf("Error syncing config for agent %s: %v", agentID, err))
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, nil)
}
