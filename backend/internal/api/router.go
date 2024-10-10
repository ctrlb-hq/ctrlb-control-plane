package api

import (
	handler "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/api/handlers"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/services"
	"github.com/gorilla/mux"
)

func NewRouter(services *services.Services) *mux.Router {
	router := mux.NewRouter()

	agentHandler := handler.NewAgentHandler(services.AgentService)
	authHandler := handler.NewAuthHandler(services.AuthService)

	authAPIsV1 := router.NewRoute().Subrouter()

	authAPIsV1.HandleFunc("/register", authHandler.Register).Methods("POST")
	authAPIsV1.HandleFunc("/login", authHandler.Login).Methods("POST")

	agentAPIsV1 := router.PathPrefix("/api/v1/agent").Subrouter()

	agentAPIsV1.HandleFunc("/register", agentHandler.RegisterAgent).Methods("PUT")
	agentAPIsV1.HandleFunc("/config", agentHandler.UpdateConfig).Methods("PUT")
	agentAPIsV1.HandleFunc("/remove", agentHandler.RemoveAgent).Methods("PUT")
	agentAPIsV1.HandleFunc("/start", agentHandler.StartAgent).Methods("POST")
	agentAPIsV1.HandleFunc("/stop", agentHandler.StopAgent).Methods("POST")
	agentAPIsV1.HandleFunc("/config", agentHandler.GetAgentConfig).Methods("GET")
	agentAPIsV1.HandleFunc("/uptime", agentHandler.GetAgentUptime).Methods("GET")
	agentAPIsV1.HandleFunc("/status", agentHandler.GetAgentStatus).Methods("GET")

	return router
}
