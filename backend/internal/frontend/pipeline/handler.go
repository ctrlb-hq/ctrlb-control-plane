package frontendpipeline

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
	"github.com/gorilla/mux"
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

	utils.Logger.Info("Request received to get all pipelines")

	response, err := f.FrontendPipelineService.GetAllPipelines()
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (f *FrontendPipelineHandler) GetPipelineInfo(w http.ResponseWriter, r *http.Request) {

	pipelineId := mux.Vars(r)["id"]
	pipelineIdInt, err := strconv.Atoi(pipelineId)
	if err != nil {
		utils.SendJSONError(w, http.StatusBadRequest, "Invalid pipeline ID format")
		return
	}

	utils.Logger.Info(fmt.Sprintf("Request received to get pipeline with ID: %s", pipelineId))

	response, err := f.FrontendPipelineService.GetPipelineInfo(pipelineIdInt)
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("Error getting pipeline info [ID: %s]: %v", pipelineId, err))
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (f *FrontendPipelineHandler) DeletePipeline(w http.ResponseWriter, r *http.Request) {
	pipelineId := mux.Vars(r)["id"]
	pipelineIdInt, err := strconv.Atoi(pipelineId)
	if err != nil {
		utils.SendJSONError(w, http.StatusBadRequest, "Invalid pipeline ID format")
		return
	}

	utils.Logger.Info(fmt.Sprintf("Request received to delete pipeline with ID: %s", pipelineId))

	err = f.FrontendPipelineService.DeletePipeline(pipelineIdInt)
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("Error deleting pipeline [ID: %s]: %v", pipelineId, err))
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, "Pipeline deleted successfully")
}

func (f *FrontendPipelineHandler) GetAllAgentsAttachedToPipeline(w http.ResponseWriter, r *http.Request) {
	pipelineId := mux.Vars(r)["id"]
	pipelineIdInt, err := strconv.Atoi(pipelineId)
	if err != nil {
		utils.SendJSONError(w, http.StatusBadRequest, "Invalid pipeline ID format")
		return
	}

	utils.Logger.Info(fmt.Sprintf("Request received to get all agents attached to pipeline with ID: %s", pipelineId))

	response, err := f.FrontendPipelineService.GetAllAgentsAttachedToPipeline(pipelineIdInt)
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("Error getting agents attached to pipeline [ID: %s]: %v", pipelineId, err))
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (f *FrontendPipelineHandler) DetachAgentFromPipeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	pipelineId := vars["id"]
	agentId := vars["agent_id"]
	pipelineIdInt, err := strconv.Atoi(pipelineId)
	if err != nil {
		utils.SendJSONError(w, http.StatusBadRequest, "Invalid pipeline ID format")
		return
	}

	agentIdInt, err := strconv.Atoi(agentId)
	if err != nil {
		utils.SendJSONError(w, http.StatusBadRequest, "Invalid agent ID format")
		return
	}

	utils.Logger.Info(fmt.Sprintf("Request received to detach agent [ID: %s] from pipeline with ID: %s", agentId, pipelineId))

	err = f.FrontendPipelineService.DetachAgentFromPipeline(pipelineIdInt, agentIdInt)
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("Error detach agent [ID: %s] from pipeline [ID: %s]: %v", agentId, pipelineId, err))
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "Agent [ID: " + agentId + "] detached successfully from pipeline [ID: " + pipelineId + "]"})
}

func (f *FrontendPipelineHandler) AttachAgentToPipeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	pipelineId := vars["id"]
	agentId := vars["agent_id"]
	pipelineIdInt, err := strconv.Atoi(pipelineId)
	if err != nil {
		utils.SendJSONError(w, http.StatusBadRequest, "Invalid pipeline ID format")
		return
	}

	agentIdInt, err := strconv.Atoi(agentId)
	if err != nil {
		utils.SendJSONError(w, http.StatusBadRequest, "Invalid agent ID format")
		return
	}

	utils.Logger.Info(fmt.Sprintf("Request received to attach agent [ID: %s] to pipeline with ID: %s", agentId, pipelineId))

	err = f.FrontendPipelineService.AttachAgentToPipeline(pipelineIdInt, agentIdInt)
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("Error attach agent [ID: %s] to pipeline [ID: %s]: %v", agentId, pipelineId, err))
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "Agent [ID: " + agentId + "] attached successfully to pipeline [ID: " + pipelineId + "]"})
}
