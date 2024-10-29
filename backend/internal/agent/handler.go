package agent

import (
	"log"
	"net/http"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
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
	var registerRequest AgentRegisterRequest // Define a variable to hold the registration request

	// Unmarshal the JSON request body into the registerRequest struct
	if err := utils.UnmarshalJSONRequest(r, &registerRequest); err != nil {
		log.Println("Invalid request body")                          // Log the error for debugging
		http.Error(w, "Invalid request body", http.StatusBadRequest) // Respond with a bad request status
		return
	}

	// Call the RegisterAgent method of the AgentService to process the registration
	reponse, err := a.AgentService.RegisterAgent(registerRequest)
	if err != nil {
		// If an error occurs, send a JSON error response with a server error status
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// If successful, write the JSON response with a status OK
	utils.WriteJSONResponse(w, http.StatusOK, reponse)
}
