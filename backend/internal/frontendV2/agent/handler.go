package frontendagent

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
