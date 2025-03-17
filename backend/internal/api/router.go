package api

import (
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/agent"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/auth"
	frontendagent "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontend/agent"
	frontendconfig "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontend/config"
	frontendpipeline "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontend/pipeline"

	frontendagentV2 "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontendV2/agent"
	frontendnode "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontendV2/node"
	frontendnodeV2 "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontendV2/node"
	frontendpipelineV2 "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontendV2/pipeline"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/middleware"
	"github.com/gorilla/mux"
)

func NewRouter(agentService *agent.AgentService, authService *auth.AuthService, frontendAgentService *frontendagent.FrontendAgentService, frontendPipelineService *frontendpipeline.FrontendPipelineService, frontendConfigServices *frontendconfig.FrontendConfigService, frontendAgentServiceV2 *frontendagentV2.FrontendAgentService, frontendPipelineServiceV2 *frontendpipelineV2.FrontendPipelineService, frontendNodeServiceV2 *frontendnode.FrontendNodeService) *mux.Router {
	router := mux.NewRouter()

	agentHandler := agent.NewAgentHandler(agentService)

	authHandler := auth.NewAuthHandler(authService)
	frontendAgentHandler := frontendagent.NewFrontendAgentHandler(frontendAgentService)
	frontendPipelineHandler := frontendpipeline.NewFrontendPipelineHandler(frontendPipelineService)
	frontendConfigHandler := frontendconfig.NewFrontendAgentHandler(frontendConfigServices)

	frontendAgentHandlerV2 := frontendagentV2.NewFrontendAgentHandler(frontendAgentServiceV2)
	frontendPipelineHandlerV2 := frontendpipelineV2.NewFrontendPipelineHandler(frontendPipelineServiceV2)
	frontendNodeHandlerV2 := frontendnodeV2.NewFrontendNodeHandler(frontendNodeServiceV2)

	authAPIsV1 := router.PathPrefix("/api/auth/v1").Subrouter()

	authAPIsV1.HandleFunc("/register", authHandler.Register).Methods("POST")
	authAPIsV1.HandleFunc("/login", authHandler.Login).Methods("POST")
	authAPIsV1.HandleFunc("/refresh", authHandler.RefreshToken).Methods("POST")

	agentAPIsV1 := router.PathPrefix("/api/agent/v1").Subrouter()

	agentAPIsV1.HandleFunc("/agents", agentHandler.RegisterAgent).Methods("PUT")

	frontendAgentAPIsV1 := router.PathPrefix("/api/frontend/v1").Subrouter()
	frontendAgentAPIsV1.Use(middleware.AuthMiddleware())

	frontendAgentAPIsV1.HandleFunc("/agents", frontendAgentHandler.GetAllAgents).Methods("GET")
	frontendAgentAPIsV1.HandleFunc("/agents/{id}", frontendAgentHandler.GetAgent).Methods("GET")
	frontendAgentAPIsV1.HandleFunc("/agents/{id}", frontendAgentHandler.DeleteAgent).Methods("DELETE")
	frontendAgentAPIsV1.HandleFunc("/agents/{id}/start", frontendAgentHandler.StartAgent).Methods("POST")
	frontendAgentAPIsV1.HandleFunc("/agents/{id}/stop", frontendAgentHandler.StopAgent).Methods("POST")
	frontendAgentAPIsV1.HandleFunc("/agents/{id}/metrics", frontendAgentHandler.GetMetrics).Methods("GET")
	frontendAgentAPIsV1.HandleFunc("/agents/{id}/restart-monitoring", frontendAgentHandler.RestartMonitoring).Methods("POST")

	frontendAgentAPIsV1.HandleFunc("/pipeline", frontendPipelineHandler.GetAllPipelines).Methods("GET")
	frontendAgentAPIsV1.HandleFunc("/pipeline/{id}", frontendPipelineHandler.GetPipeline).Methods("GET")
	frontendAgentAPIsV1.HandleFunc("/pipeline/{id}", frontendPipelineHandler.DeletePipeline).Methods("DELETE")
	frontendAgentAPIsV1.HandleFunc("/pipeline/{id}/start", frontendPipelineHandler.StartPipeline).Methods("POST")
	frontendAgentAPIsV1.HandleFunc("/pipeline/{id}/stop", frontendPipelineHandler.StopPipeline).Methods("POST")
	frontendAgentAPIsV1.HandleFunc("/pipeline/{id}/metrics", frontendPipelineHandler.GetMetrics).Methods("GET")
	frontendAgentAPIsV1.HandleFunc("/pipeline/{id}/restart-monitoring", frontendPipelineHandler.RestartMonitoring).Methods("POST")

	frontendAgentAPIsV1.HandleFunc("/configs", frontendConfigHandler.GetAllConfig).Methods("GET")
	frontendAgentAPIsV1.HandleFunc("/configs", frontendConfigHandler.CreateConfig).Methods("POST")
	frontendAgentAPIsV1.HandleFunc("/configs/{id}", frontendConfigHandler.GetConfig).Methods("GET")
	frontendAgentAPIsV1.HandleFunc("/configs/{id}", frontendConfigHandler.DeleteConfig).Methods("DELETE")
	frontendAgentAPIsV1.HandleFunc("/configs/{id}", frontendConfigHandler.UpdateConfig).Methods("PATCH")

	frontendAgentAPIsV2 := router.PathPrefix("/api/frontend/v2").Subrouter()
	frontendAgentAPIsV2.Use(middleware.AuthMiddleware())

	frontendAgentAPIsV2.HandleFunc("/agents", frontendAgentHandlerV2.GetAllAgents).Methods("GET")
	frontendAgentAPIsV2.HandleFunc("/agents/{id}", frontendAgentHandlerV2.GetAgent).Methods("GET")
	frontendAgentAPIsV2.HandleFunc("/agents/{id}", frontendAgentHandlerV2.DeleteAgent).Methods("DELETE")
	frontendAgentAPIsV2.HandleFunc("/agents/{id}/start", frontendAgentHandlerV2.StartAgent).Methods("POST")
	frontendAgentAPIsV2.HandleFunc("/agents/{id}/stop", frontendAgentHandlerV2.StopAgent).Methods("POST")
	frontendAgentAPIsV2.HandleFunc("/agents/{id}/restart-monitoring", frontendAgentHandlerV2.RestartMonitoring).Methods("POST")
	frontendAgentAPIsV2.HandleFunc("/agents/{id}/healthmetrics", frontendAgentHandlerV2.GetHealthMetricsForGraph).Methods("GET")
	frontendAgentAPIsV2.HandleFunc("/agents/{id}/ratemetrics", frontendAgentHandlerV2.GetRateMetricsForGraph).Methods("GET")

	frontendAgentAPIsV2.HandleFunc("/unassigned-agents", frontendAgentHandlerV2.GetUnmanagedAgents).Methods("GET")

	frontendAgentAPIsV2.HandleFunc("/pipeline", frontendPipelineHandlerV2.GetAllPipelines).Methods("GET")
	frontendAgentAPIsV2.HandleFunc("/pipeline/{id}", frontendPipelineHandlerV2.GetPipelineInfo).Methods("GET")
	frontendAgentAPIsV2.HandleFunc("/pipeline/{id}", frontendPipelineHandlerV2.DeletePipeline).Methods("DELETE")

	frontendAgentAPIsV2.HandleFunc("/pipeline/{id}/agents", frontendPipelineHandlerV2.GetAllAgentsAttachedToPipeline).Methods("GET")
	frontendAgentAPIsV2.HandleFunc("/pipeline/{id}/agent/{agent_id}", frontendPipelineHandlerV2.DetachAgentFromPipeline).Methods("DELETE")

	frontendAgentAPIsV2.HandleFunc("/receivers", frontendNodeHandlerV2.GetAllReceivers).Methods("GET")
	frontendAgentAPIsV2.HandleFunc("/processors", frontendNodeHandlerV2.GetAllProcessors).Methods("GET")
	frontendAgentAPIsV2.HandleFunc("/exporters", frontendNodeHandlerV2.GetAllExporters).Methods("GET")

	return router
}
