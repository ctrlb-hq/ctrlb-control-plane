package api

import (
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/agent"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/auth"

	frontendagent "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontend/agent"
	frontendnode "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontend/node"
	frontendpipeline "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontend/pipeline"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/middleware"
	"github.com/gorilla/mux"
)

func NewRouter(agentService *agent.AgentService, authService *auth.AuthService, frontendAgentServiceV2 *frontendagent.FrontendAgentService, frontendPipelineServiceV2 *frontendpipeline.FrontendPipelineService, frontendNodeServiceV2 *frontendnode.FrontendNodeService) *mux.Router {
	router := mux.NewRouter()

	agentHandler := agent.NewAgentHandler(agentService)

	authHandler := auth.NewAuthHandler(authService)

	frontendAgentHandler := frontendagent.NewFrontendAgentHandler(frontendAgentServiceV2)
	frontendPipelineHandler := frontendpipeline.NewFrontendPipelineHandler(frontendPipelineServiceV2)
	frontendNodeHandler := frontendnode.NewFrontendNodeHandler(frontendNodeServiceV2)

	authAPIsV1 := router.PathPrefix("/api/auth/v1").Subrouter()

	authAPIsV1.HandleFunc("/register", authHandler.Register).Methods("POST")
	authAPIsV1.HandleFunc("/login", authHandler.Login).Methods("POST")
	authAPIsV1.HandleFunc("/refresh", authHandler.RefreshToken).Methods("POST")

	agentAPIsV1 := router.PathPrefix("/api/agent/v1").Subrouter()

	agentAPIsV1.HandleFunc("/agents", agentHandler.RegisterAgent).Methods("POST")
	agentAPIsV1.HandleFunc("/agents/{id}/config-changed", agentHandler.ConfigChangedPing).Methods("POST")

	frontendAgentAPIsV2 := router.PathPrefix("/api/frontend/v2").Subrouter()
	frontendAgentAPIsV2.Use(middleware.AuthMiddleware())

	frontendAgentAPIsV2.HandleFunc("/agents", frontendAgentHandler.GetAllAgents).Methods("GET")
	frontendAgentAPIsV2.HandleFunc("/agents/{id}", frontendAgentHandler.GetAgent).Methods("GET")
	frontendAgentAPIsV2.HandleFunc("/agents/{id}", frontendAgentHandler.DeleteAgent).Methods("DELETE")
	frontendAgentAPIsV2.HandleFunc("/agents/{id}/start", frontendAgentHandler.StartAgent).Methods("POST")
	frontendAgentAPIsV2.HandleFunc("/agents/{id}/stop", frontendAgentHandler.StopAgent).Methods("POST")
	frontendAgentAPIsV2.HandleFunc("/agents/{id}/restart-monitoring", frontendAgentHandler.RestartMonitoring).Methods("POST")
	frontendAgentAPIsV2.HandleFunc("/agents/{id}/healthmetrics", frontendAgentHandler.GetHealthMetricsForGraph).Methods("GET")
	frontendAgentAPIsV2.HandleFunc("/agents/{id}/ratemetrics", frontendAgentHandler.GetRateMetricsForGraph).Methods("GET")
	frontendAgentAPIsV2.HandleFunc("/agents/{id}/labels", frontendAgentHandler.AddLabels).Methods("POST")

	frontendAgentAPIsV2.HandleFunc("/unassigned-agents", frontendAgentHandler.GetUnmanagedAgents).Methods("GET")
	frontendAgentAPIsV2.HandleFunc("/agents/latest", frontendAgentHandler.GetAllAgents).Methods("GET")

	frontendAgentAPIsV2.HandleFunc("/pipelines", frontendPipelineHandler.GetAllPipelines).Methods("GET")
	frontendAgentAPIsV2.HandleFunc("/pipelines", frontendPipelineHandler.CreatePipeline).Methods("POST")

	frontendAgentAPIsV2.HandleFunc("/pipelines/{id}", frontendPipelineHandler.GetPipelineInfo).Methods("GET")
	frontendAgentAPIsV2.HandleFunc("/pipelines/{id}", frontendPipelineHandler.DeletePipeline).Methods("DELETE")

	frontendAgentAPIsV2.HandleFunc("/pipelines/{id}/graph", frontendPipelineHandler.GetPipelineGraph).Methods("GET")
	frontendAgentAPIsV2.HandleFunc("/pipelines/{id}/graph", frontendPipelineHandler.SyncPipelineGraph).Methods("POST")

	frontendAgentAPIsV2.HandleFunc("/pipelines/{id}/agents", frontendPipelineHandler.GetAllAgentsAttachedToPipeline).Methods("GET")
	frontendAgentAPIsV2.HandleFunc("/pipelines/{id}/agents/{agent_id}", frontendPipelineHandler.DetachAgentFromPipeline).Methods("DELETE")
	frontendAgentAPIsV2.HandleFunc("/pipelines/{id}/agents/{agent_id}", frontendPipelineHandler.AttachAgentToPipeline).Methods("POST")

	frontendAgentAPIsV2.HandleFunc("/component", frontendNodeHandler.GetComponent).Methods("GET")
	frontendAgentAPIsV2.HandleFunc("/component/schema/{name}", frontendNodeHandler.GetComponentSchema).Methods("GET")

	return router
}
