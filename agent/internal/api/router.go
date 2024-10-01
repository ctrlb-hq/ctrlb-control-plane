package api

import (
	api "github.com/ctrlb-hq/ctrlb-collector/internal/api/handlers"
	"github.com/ctrlb-hq/ctrlb-collector/internal/services"
	"github.com/gorilla/mux"
)

func NewRouter(operatorService *services.OperatorService) *mux.Router {
	router := mux.NewRouter()
	operatorHandler := api.NewOperatorHandler(operatorService)

	operatorApiV1 := router.PathPrefix("/api/v1").Subrouter()

	operatorApiV1.HandleFunc("/uptime", operatorHandler.GetUptime).Methods("GET")

	operatorApiV1.HandleFunc("/config", operatorHandler.UpdateCurrentConfig).Methods("PUT")

	operatorApiV1.HandleFunc("/config", operatorHandler.GetCurrentConfig).Methods("GET")

	operatorApiV1.HandleFunc("/start", operatorHandler.StartAgent).Methods("POST")

	operatorApiV1.HandleFunc("/stop", operatorHandler.StopAgent).Methods("POST")

	operatorApiV1.HandleFunc("/shutdown", operatorHandler.GracefulShutdown).Methods("POST")

	operatorApiV1.HandleFunc("/status", operatorHandler.CurrentStatus).Methods("GET")

	// V2 works
	// operatorApiV1.HandleFunc("/update", operatorHandler.UpdateAgent).Methods("POST")

	return router
}
