package frontendconfig

import (
	"log"
	"net/http"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
	"github.com/gorilla/mux"
)

// FrontendConfigHandler handles configuration-related requests
type FrontendConfigHandler struct {
	FrontendConfigService *FrontendConfigService
}

// NewFrontendAgentHandler initializes FrontendConfigHandler
func NewFrontendAgentHandler(frontendConfigServices *FrontendConfigService) *FrontendConfigHandler {
	return &FrontendConfigHandler{
		FrontendConfigService: frontendConfigServices,
	}
}

// GetAllConfig retrieves all configurations
func (f *FrontendConfigHandler) GetAllConfig(w http.ResponseWriter, r *http.Request) {

	response, err := f.FrontendConfigService.GetAllConfigs()
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// CreateConfig creates a new configuration
func (f *FrontendConfigHandler) CreateConfig(w http.ResponseWriter, r *http.Request) {

	var createConfigRequest ConfigUpsertRequest
	if err := utils.UnmarshalJSONRequest(r, &createConfigRequest); err != nil {
		log.Println("Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := f.FrontendConfigService.CreateConfig(createConfigRequest)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// GetConfig retrieves a specific configuration by ID
func (f *FrontendConfigHandler) GetConfig(w http.ResponseWriter, r *http.Request) {

	id := mux.Vars(r)["id"]
	response, err := f.FrontendConfigService.GetConfig(id)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// DeleteConfig deletes a configuration by ID
func (f *FrontendConfigHandler) DeleteConfig(w http.ResponseWriter, r *http.Request) {

	id := mux.Vars(r)["id"]
	if err := f.FrontendConfigService.DeleteConfig(id); err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "Config deleted successfully"})
}

// UpdateConfig updates an existing configuration by ID
func (f *FrontendConfigHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {

	id := mux.Vars(r)["id"]
	var configUpdateRequest ConfigUpsertRequest
	if err := utils.UnmarshalJSONRequest(r, &configUpdateRequest); err != nil {
		log.Println("Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := f.FrontendConfigService.UpdateConfig(id, configUpdateRequest); err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "Config updated successfully"})
}
