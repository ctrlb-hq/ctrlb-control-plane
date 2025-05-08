package api

import (
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/middleware"
	"github.com/gorilla/mux"
)

func NewRouter(handler *Handler) *mux.Router {
	router := mux.NewRouter()

	authAPIsV1 := router.PathPrefix("/api/auth/v1").Subrouter()

	authAPIsV1.HandleFunc("/register", handler.AuthHandler.Register).Methods("POST")
	authAPIsV1.HandleFunc("/login", handler.AuthHandler.Login).Methods("POST")
	authAPIsV1.HandleFunc("/refresh", handler.AuthHandler.RefreshToken).Methods("POST")

	agentAPIsV1 := router.PathPrefix("/api/agent/v1").Subrouter()

	agentAPIsV1.HandleFunc("/agents", handler.AgentHandler.RegisterAgent).Methods("POST")
	agentAPIsV1.HandleFunc("/agents/{id}/config-changed", handler.AgentHandler.ConfigChangedPing).Methods("POST")

	frontendAgentAPIsV2 := router.PathPrefix("/api/frontend/v2").Subrouter()
	frontendAgentAPIsV2.Use(middleware.AuthMiddleware())

	frontendAgentAPIsV2.HandleFunc("/agents", handler.FrontendAgentHandler.GetAllAgents).Methods("GET")
	frontendAgentAPIsV2.HandleFunc("/agents/{id}", handler.FrontendAgentHandler.GetAgent).Methods("GET")
	frontendAgentAPIsV2.HandleFunc("/agents/{id}", handler.FrontendAgentHandler.DeleteAgent).Methods("DELETE")
	frontendAgentAPIsV2.HandleFunc("/agents/{id}/start", handler.FrontendAgentHandler.StartAgent).Methods("POST")
	frontendAgentAPIsV2.HandleFunc("/agents/{id}/stop", handler.FrontendAgentHandler.StopAgent).Methods("POST")
	frontendAgentAPIsV2.HandleFunc("/agents/{id}/restart-monitoring", handler.FrontendAgentHandler.RestartMonitoring).Methods("POST")
	frontendAgentAPIsV2.HandleFunc("/agents/{id}/healthmetrics", handler.FrontendAgentHandler.GetHealthMetricsForGraph).Methods("GET")
	frontendAgentAPIsV2.HandleFunc("/agents/{id}/ratemetrics", handler.FrontendAgentHandler.GetRateMetricsForGraph).Methods("GET")
	frontendAgentAPIsV2.HandleFunc("/agents/{id}/labels", handler.FrontendAgentHandler.AddLabels).Methods("POST")

	frontendAgentAPIsV2.HandleFunc("/unassigned-agents", handler.FrontendAgentHandler.GetUnmanagedAgents).Methods("GET")
	frontendAgentAPIsV2.HandleFunc("/latest-agent", handler.FrontendAgentHandler.GetLatestAgentSince).Methods("GET")

	frontendAgentAPIsV2.HandleFunc("/pipelines", handler.FrontendPipelineHandler.GetAllPipelines).Methods("GET")
	frontendAgentAPIsV2.HandleFunc("/pipelines", handler.FrontendPipelineHandler.CreatePipeline).Methods("POST")
	frontendAgentAPIsV2.HandleFunc("/pipelines/{id}", handler.FrontendPipelineHandler.GetPipelineInfo).Methods("GET")
	frontendAgentAPIsV2.HandleFunc("/pipelines/{id}", handler.FrontendPipelineHandler.DeletePipeline).Methods("DELETE")
	frontendAgentAPIsV2.HandleFunc("/pipelines-overview/{id}", handler.FrontendPipelineHandler.GetPipelineOverview).Methods("GET")

	frontendAgentAPIsV2.HandleFunc("/pipelines/{id}/graph", handler.FrontendPipelineHandler.GetPipelineGraph).Methods("GET")
	frontendAgentAPIsV2.HandleFunc("/pipelines/{id}/graph", handler.FrontendPipelineHandler.SyncPipelineGraph).Methods("POST")

	frontendAgentAPIsV2.HandleFunc("/pipelines/{id}/agents", handler.FrontendPipelineHandler.GetAllAgentsAttachedToPipeline).Methods("GET")
	frontendAgentAPIsV2.HandleFunc("/pipelines/{id}/agents/{agent_id}", handler.FrontendPipelineHandler.DetachAgentFromPipeline).Methods("DELETE")
	frontendAgentAPIsV2.HandleFunc("/pipelines/{id}/agents/{agent_id}", handler.FrontendPipelineHandler.AttachAgentToPipeline).Methods("POST")

	frontendAgentAPIsV2.HandleFunc("/component", frontendNodeHandler.GetComponent).Methods("GET")
	frontendAgentAPIsV2.HandleFunc("/component/schema/{name}", frontendNodeHandler.GetComponentSchema).Methods("GET")
	frontendAgentAPIsV2.HandleFunc("/component", handler.FrontendNodeHandler.GetComponent).Methods("GET")
	frontendAgentAPIsV2.HandleFunc("/component/schema/{name}", handler.FrontendNodeHandler.GetComponentSchema).Methods("GET")

	return router
}
