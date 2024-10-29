package frontendpipeline

import (
	"net/http"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
	"github.com/gorilla/mux"
)

type FrontendPipelineHandler struct {
	FrontendPipelineService *FrontendPipelineService
}

// NewFrontendPipelineHandler creates a new FrontendPipelineHandler
func NewFrontendPipelineHandler(frontendPipelineService *FrontendPipelineService) *FrontendPipelineHandler {
	return &FrontendPipelineHandler{
		FrontendPipelineService: frontendPipelineService,
	}
}

// GetAllPipelines retrieves all pipelines
func (f *FrontendPipelineHandler) GetAllPipelines(w http.ResponseWriter, r *http.Request) {

	response, err := f.FrontendPipelineService.GetAllPipelines()
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// GetPipeline retrieves a specific pipeline by ID
func (f *FrontendPipelineHandler) GetPipeline(w http.ResponseWriter, r *http.Request) {

	id := mux.Vars(r)["id"]

	response, err := f.FrontendPipelineService.GetPipeline(id)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// DeletePipeline removes a pipeline by ID
func (f *FrontendPipelineHandler) DeletePipeline(w http.ResponseWriter, r *http.Request) {

	id := mux.Vars(r)["id"]

	if err := f.FrontendPipelineService.DeletePipeline(id); err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]string{"message": "Pipeline deleted successfully"}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// StartPipeline starts a specific pipeline by ID
func (f *FrontendPipelineHandler) StartPipeline(w http.ResponseWriter, r *http.Request) {

	id := mux.Vars(r)["id"]

	if err := f.FrontendPipelineService.StartPipeline(id); err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]string{"message": "Pipeline started successfully"}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// StopPipeline stops a specific pipeline by ID
func (f *FrontendPipelineHandler) StopPipeline(w http.ResponseWriter, r *http.Request) {

	id := mux.Vars(r)["id"]

	if err := f.FrontendPipelineService.StopPipeline(id); err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]string{"message": "Pipeline stopped successfully"}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// GetMetrics retrieves metrics for a specific pipeline by ID
func (f *FrontendPipelineHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {

	id := mux.Vars(r)["id"]

	response, err := f.FrontendPipelineService.GetMetrics(id)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// RestartMonitoring restarts monitoring for a specific pipeline by ID
func (f *FrontendPipelineHandler) RestartMonitoring(w http.ResponseWriter, r *http.Request) {

	id := mux.Vars(r)["id"]

	if err := f.FrontendPipelineService.RestartMonitoring(id); err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]string{"message": "Started monitoring the pipeline"}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}
