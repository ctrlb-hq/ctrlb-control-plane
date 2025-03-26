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
	AgentService *AgentService // Service for managing agent operations
}

// NewAgentHandler creates a new instance of AgentHandler with the provided services.
func NewAgentHandler(agentServices *AgentService) *AgentHandler {
	return &AgentHandler{
		AgentService: agentServices, // Assign the agent service\
	}
}

// RegisterAgent handles the registration of a new agent.
// It expects a JSON payload in the request body.
func (a *AgentHandler) RegisterAgent(w http.ResponseWriter, r *http.Request) {
	req := &models.AgentRegisterRequest{} // Define a variable to hold the registration request

	// Unmarshal the JSON request body into the registerRequest struct
	if err := utils.UnmarshalJSONRequest(r, req); err != nil {
		utils.Logger.Error("Invalid request body")                   // Log the error for debugging
		http.Error(w, "Invalid request body", http.StatusBadRequest) // Respond with a bad request status
		return
	}

	if err := utils.ValidateAgentRegisterRequest(req); err != nil {
		utils.Logger.Error(fmt.Sprintf("Invalid request: %v", err)) // Log the error for debugging
		utils.SendJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.Logger.Info(fmt.Sprintf("Received registration request from agent: %s", req.Hostname))

	reponse, err := a.AgentService.RegisterAgent(req)
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("Error registering agent: %v", err)) // Log the error for debugging
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// If successful, write the JSON response with a status OK
	utils.WriteJSONResponse(w, http.StatusOK, reponse)
}

// ConfigChangedPing handles the ping from an agent indicating that its configuration has changed.
func (a *AgentHandler) ConfigChangedPing(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["id"] // Get the agent ID from the URL
	utils.Logger.Info(fmt.Sprintf("Received config changed ping from agent: %s", agentID))
	//TODO: Implement the logic to handle the config changed ping
	utils.WriteJSONResponse(w, http.StatusOK, nil)
}
