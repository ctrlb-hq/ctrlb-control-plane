package api

import (
	api "github.com/ctrlb-hq/ctrlb-collector/agent/internal/api/handlers"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/core/operators"
	"github.com/gorilla/mux"
)

func NewRouter(operatorService *operators.OperatorService) *mux.Router {
	router := mux.NewRouter()
	operatorHandler := api.NewOperatorHandler(operatorService)

	// API version 1 for agent
	agentApiV1 := router.PathPrefix("/agent/v1").Subrouter()

	// Agent configuration (GET and PUT) - Retrieves or updates the current config of the agent
	agentApiV1.HandleFunc("/config", operatorHandler.GetCurrentConfig).Methods("GET")
	agentApiV1.HandleFunc("/config", operatorHandler.UpdateCurrentConfig).Methods("PUT")

	// Agent lifecycle actions (Start, Stop, Shutdown) - Manage agent's running state
	agentApiV1.HandleFunc("/start", operatorHandler.StartAgent).Methods("POST")
	agentApiV1.HandleFunc("/stop", operatorHandler.StopAgent).Methods("POST")
	agentApiV1.HandleFunc("/shutdown", operatorHandler.GracefulShutdown).Methods("POST")

	// Agent status (GET) - Fetches the current status of the agent
	agentApiV1.HandleFunc("/status", operatorHandler.CurrentStatus).Methods("GET")

	return router
}
