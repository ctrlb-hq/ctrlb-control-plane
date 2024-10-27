package api

import (
	"log"
	"net/http"

	"github.com/ctrlb-hq/ctrlb-collector/internal/operators"
	"github.com/ctrlb-hq/ctrlb-collector/internal/utils"
)

var operatorHandler *OperatorHandler

func NewOperatorHandler(operatorService *operators.OperatorService) *OperatorHandler {
	operatorHandler = &OperatorHandler{
		OperatorService: operatorService,
	}
	return operatorHandler
}

func (o *OperatorHandler) GetCurrentConfig(w http.ResponseWriter, r *http.Request) {
	//TODO: Add Auth

	response, err := o.OperatorService.GetCurrentConfig()
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)

}

type UpdateConfigRequest struct {
	Name        string `json:"name"`        // Configuration name
	Description string `json:"description"` // Brief description of the configuration
	Config      string `json:"config"`      // Configuration content (e.g., JSON or YAML)
	TargetAgent string `json:"targetAgent"` // Agent type the configuration targets
}

func (o *OperatorHandler) UpdateCurrentConfig(w http.ResponseWriter, r *http.Request) {
	// TODO: Add Auth

	var updateConfigRequest interface{}

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

func (o *OperatorHandler) StartAgent(w http.ResponseWriter, r *http.Request) {
	//TODO: Add Auth

	response, err := o.OperatorService.StartAgent()
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)

}

func (o *OperatorHandler) StopAgent(w http.ResponseWriter, r *http.Request) {
	//TODO: Add Auth

	response, err := o.OperatorService.StopAgent()
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)

}

func (o *OperatorHandler) GracefulShutdown(w http.ResponseWriter, r *http.Request) {
	//TODO: Add Auth

	response, err := o.OperatorService.GracefulShutdown()
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)

}

func (o *OperatorHandler) CurrentStatus(w http.ResponseWriter, r *http.Request) {
	//TODO: Add Auth

	response, err := o.OperatorService.CurrentStatus()
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)

}
