package frontendnode

import (
	"fmt"
	"net/http"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
)

type FrontendNodeHandler struct {
	FrontendNodeService *FrontendNodeService
}

// NewFrontendAgentHandler initializes the handler
func NewFrontendNodeHandler(frontendNodeServices *FrontendNodeService) *FrontendNodeHandler {
	return &FrontendNodeHandler{
		FrontendNodeService: frontendNodeServices,
	}
}

func (f *FrontendNodeHandler) GetAllReceivers(w http.ResponseWriter, r *http.Request) {
	utils.Logger.Info("Received request to get all receivers")

	response, err := f.FrontendNodeService.GetAllReceivers()
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("Error occured while getting all receivers: %v", err))
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (f *FrontendNodeHandler) GetAllProcessors(w http.ResponseWriter, r *http.Request) {
	utils.Logger.Info("Received request to get all processors")

	response, err := f.FrontendNodeService.GetAllProcessors()
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("Error occured while getting all processors: %v", err))
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (f *FrontendNodeHandler) GetAllExporters(w http.ResponseWriter, r *http.Request) {
	utils.Logger.Info("Received request to get all exporters")

	response, err := f.FrontendNodeService.GetAllExporters()
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("Error occured while getting all exporters: %v", err))
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)

}
