package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/core/operators"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/models"
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
	pkg.Logger.Info("Starting agent")

	response, err := o.OperatorService.StartAgent()
	if err != nil {
		pkg.Logger.Error(fmt.Sprintf("Error starting agent: %v", err.Error()))
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)

}

func (o *OperatorHandler) StopAgent(w http.ResponseWriter, r *http.Request) {
	pkg.Logger.Info("Stopping agent")

	response, err := o.OperatorService.StopAgent()
	if err != nil {
		pkg.Logger.Error(fmt.Sprintf("Error stoping agent: %v", err.Error()))
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)

}

func (o *OperatorHandler) GracefulShutdown(w http.ResponseWriter, r *http.Request) {
	pkg.Logger.Info("Shutting down agent")

	response, err := o.OperatorService.GracefulShutdown()
	if err != nil {
		pkg.Logger.Error(fmt.Sprintf("Error shutting down agent: %v", err.Error()))
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)

}

func (o *OperatorHandler) GetCurrentConfig(w http.ResponseWriter, r *http.Request) {

	response, err := o.OperatorService.GetCurrentConfig()
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)

}

func (o *OperatorHandler) UpdateCurrentConfig(w http.ResponseWriter, r *http.Request) {

	var updateConfigRequest models.ConfigUpsertRequest

	if err := utils.UnmarshalJSONRequest(r, &updateConfigRequest); err != nil {
		log.Println(err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := o.OperatorService.UpdateCurrentConfig(updateConfigRequest)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (o *OperatorHandler) CurrentStatus(w http.ResponseWriter, r *http.Request) {
	response, err := o.OperatorService.CurrentStatus()
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)

}
