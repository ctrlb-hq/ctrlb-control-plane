package frontendpipeline

// FrontendPipelineHandler handles frontend Pipeline operations
type FrontendPipelineHandler struct {
	FrontendPipelineService *FrontendPipelineService
}

// NewFrontendPipelineHandler initializes the handler
func NewFrontendPipelineHandler(frontendPipelineServices *FrontendPipelineService) *FrontendPipelineHandler {
	return &FrontendPipelineHandler{
		FrontendPipelineService: frontendPipelineServices,
	}
}
