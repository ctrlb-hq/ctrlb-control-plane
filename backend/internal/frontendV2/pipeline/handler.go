package frontendpipeline

import (
	"net/http"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
)

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

func (f *FrontendPipelineHandler) GetAllPipelines(w http.ResponseWriter, r *http.Request) {

	response, err := f.FrontendPipelineService.GetAllPipelines()
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}
