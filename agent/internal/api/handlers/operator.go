package api

import (
	"fmt"
	"net/http"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/core/operators"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/pkg"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/utils"
)

func NewOperatorHandler(operatorService *operators.OperatorService) *OperatorHandler {
	operatorHandler := &OperatorHandler{
		OperatorService: operatorService,
	}
	return operatorHandler
}

func (o *OperatorHandler) StartAgent(w http.ResponseWriter, r *http.Request) {
	pkg.Logger.Info("Request received to start agent")

	response, err := o.OperatorService.StartAgent()
	if err != nil {
		pkg.Logger.Error(fmt.Sprintf("Error starting agent: %v", err.Error()))
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	pkg.Logger.Info("Successfully started agent")
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (o *OperatorHandler) StopAgent(w http.ResponseWriter, r *http.Request) {
	pkg.Logger.Info("Request received to stop agent")

	response, err := o.OperatorService.StopAgent()
	if err != nil {
		pkg.Logger.Error(fmt.Sprintf("Error stoping agent: %v", err.Error()))
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	pkg.Logger.Info("Successfully stopped agent")
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (o *OperatorHandler) GracefulShutdown(w http.ResponseWriter, r *http.Request) {
	pkg.Logger.Info("Request received for graceful shutdown")

	response, err := o.OperatorService.GracefulShutdown()
	if err != nil {
		pkg.Logger.Error(fmt.Sprintf("Error shutting down agent: %v", err.Error()))
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// ---
// FIXME: work on these methods
// ---
func (o *OperatorHandler) GetCurrentConfig(w http.ResponseWriter, r *http.Request) {
	pkg.Logger.Info("Request received to get current config")

	response, err := o.OperatorService.GetCurrentConfig()
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	pkg.Logger.Info("Successfully retrieved current config")
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (o *OperatorHandler) UpdateCurrentConfig(w http.ResponseWriter, r *http.Request) {
	pkg.Logger.Info("Request received to update current config")

	var updateConfigRequest map[string]any
	if err := utils.UnmarshalJSONRequest(r, &updateConfigRequest); err != nil {
		pkg.Logger.Error(fmt.Sprintf("Invalid request body: %v", err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := o.OperatorService.UpdateCurrentConfig(updateConfigRequest)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	pkg.Logger.Info("Successfully updated current config")
	response := map[string]string{"message": "Successfully updated current config"}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (o *OperatorHandler) CurrentStatus(w http.ResponseWriter, r *http.Request) {
	pkg.Logger.Info("Request received to get current status")

	response, err := o.OperatorService.CurrentStatus()
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	pkg.Logger.Info("Successfully retrieved current status")
	utils.WriteJSONResponse(w, http.StatusOK, response)
}
