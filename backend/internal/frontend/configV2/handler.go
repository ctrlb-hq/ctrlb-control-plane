package frontendconfigV2

import (
	"log"
	"net/http"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
	"github.com/gorilla/mux"
)

// FrontendConfigHandlerV2 handles configuration-related requests
type FrontendConfigHandlerV2 struct {
	FrontendConfigService *FrontendConfigServiceV2
}

// NewFrontendAgentHandlerV2 initializes FrontendConfigHandler
func NewFrontendAgentHandlerV2(frontendConfigServices *FrontendConfigServiceV2) *FrontendConfigHandlerV2 {
	return &FrontendConfigHandlerV2{
		FrontendConfigService: frontendConfigServices,
	}
}

// GetAllConfig retrieves all configurations
func (f *FrontendConfigHandlerV2) GetAllConfig(w http.ResponseWriter, r *http.Request) {

	response, err := f.FrontendConfigService.GetAllConfigs()
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// GetAllConfig retrieves all configurations V2
func (f *FrontendConfigHandlerV2) GetAllConfigV2(w http.ResponseWriter, r *http.Request) {

	response, err := f.FrontendConfigService.GetAllConfigsV2()
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// CreateConfig creates a new configuration
func (f *FrontendConfigHandlerV2) CreateConfig(w http.ResponseWriter, r *http.Request) {

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
func (f *FrontendConfigHandlerV2) GetConfig(w http.ResponseWriter, r *http.Request) {

	id := mux.Vars(r)["id"]
	response, err := f.FrontendConfigService.GetConfig(id)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// DeleteConfig deletes a configuration by ID
func (f *FrontendConfigHandlerV2) DeleteConfig(w http.ResponseWriter, r *http.Request) {

	id := mux.Vars(r)["id"]
	if err := f.FrontendConfigService.DeleteConfig(id); err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "Config deleted successfully"})
}

// UpdateConfig updates an existing configuration by ID
func (f *FrontendConfigHandlerV2) UpdateConfig(w http.ResponseWriter, r *http.Request) {

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
