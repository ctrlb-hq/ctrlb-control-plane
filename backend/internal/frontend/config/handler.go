package frontendconfig

import (
	"log"
	"net/http"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/auth"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
	"github.com/gorilla/mux"
)

// FrontendConfigHandler handles configuration-related requests
type FrontendConfigHandler struct {
	FrontendConfigService *FrontendConfigService
	BasicAuthenticator    *auth.BasicAuthenticator
}

// NewFrontendAgentHandler initializes FrontendConfigHandler
func NewFrontendAgentHandler(frontendConfigServices *FrontendConfigService, basicAuthenticator *auth.BasicAuthenticator) *FrontendConfigHandler {
	return &FrontendConfigHandler{
		FrontendConfigService: frontendConfigServices,
		BasicAuthenticator:    basicAuthenticator,
	}
}

// GetAllConfig retrieves all configurations
func (f *FrontendConfigHandler) GetAllConfig(w http.ResponseWriter, r *http.Request) {
	token, err := utils.ExtractTokenFromHeaders(&r.Header)
	if err != nil || f.BasicAuthenticator.ValidateToken(token) != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	response, err := f.FrontendConfigService.GetAllConfigs()
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// CreateConfig creates a new configuration
func (f *FrontendConfigHandler) CreateConfig(w http.ResponseWriter, r *http.Request) {
	token, err := utils.ExtractTokenFromHeaders(&r.Header)
	if err != nil || f.BasicAuthenticator.ValidateToken(token) != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

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
	token, err := utils.ExtractTokenFromHeaders(&r.Header)
	if err != nil || f.BasicAuthenticator.ValidateToken(token) != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

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
	token, err := utils.ExtractTokenFromHeaders(&r.Header)
	if err != nil || f.BasicAuthenticator.ValidateToken(token) != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := mux.Vars(r)["id"]
	if err := f.FrontendConfigService.DeleteConfig(id); err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "Config deleted successfully"})
}

// UpdateConfig updates an existing configuration by ID
func (f *FrontendConfigHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	token, err := utils.ExtractTokenFromHeaders(&r.Header)
	if err != nil || f.BasicAuthenticator.ValidateToken(token) != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

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
