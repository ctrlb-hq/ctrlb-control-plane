package api

import (
	"fmt"
	"net/http"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/core/operators"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/pkg/logger"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/utils"
)

func NewOperatorHandler(operatorService *operators.OperatorService) *OperatorHandler {
	operatorHandler := &OperatorHandler{
		OperatorService: operatorService,
	}
	return operatorHandler
}

func (o *OperatorHandler) StartAgent(w http.ResponseWriter, r *http.Request) {
	logger.Logger.Info("Request received to start agent")

	err := o.OperatorService.StartAgent()
	if err != nil {
		logger.Logger.Error(fmt.Sprintf("Error starting agent: %v", err.Error()))
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	logger.Logger.Info("Successfully started agent")
	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "Successfully started agent"})
}

func (o *OperatorHandler) StopAgent(w http.ResponseWriter, r *http.Request) {
	logger.Logger.Info("Request received to stop agent")

	err := o.OperatorService.StopAgent()
	if err != nil {
		logger.Logger.Error(fmt.Sprintf("Error stoping agent: %v", err.Error()))
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	logger.Logger.Info("Successfully stopped agent")
	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "Successfully stopped agent"})
}

func (o *OperatorHandler) GracefulShutdown(w http.ResponseWriter, r *http.Request) {
	logger.Logger.Info("Request received for graceful shutdown")

	err := o.OperatorService.GracefulShutdown()
	if err != nil {
		logger.Logger.Error(fmt.Sprintf("Error shutting down agent: %v", err.Error()))
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "Successfully shutdown agent"})
}

func (o *OperatorHandler) UpdateCurrentConfig(w http.ResponseWriter, r *http.Request) {
	logger.Logger.Info("Request received to update current config")

	var updateConfigRequest map[string]any
	if err := utils.UnmarshalJSONRequest(r, &updateConfigRequest); err != nil {
		logger.Logger.Error(fmt.Sprintf("Invalid request body: %v", err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := o.OperatorService.UpdateCurrentConfig(updateConfigRequest)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	logger.Logger.Info("Successfully updated current config")
	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "Successfully updated current config"})
}
