package api

import (
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/agent"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/auth"
	frontendagent "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontend/agent"
	frontendnode "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontend/node"
	frontendpipeline "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontend/pipeline"
)

type Handler struct {
	AgentHandler            *agent.AgentHandler
	AuthHandler             *auth.AuthHandler
	FrontendAgentHandler    *frontendagent.FrontendAgentHandler
	FrontendPipelineHandler *frontendpipeline.FrontendPipelineHandler
	FrontendNodeHandler     *frontendnode.FrontendNodeHandler
}

func NewHandler(
	agentService *agent.AgentService,
	authService *auth.AuthService,
	frontendAgentServiceV2 *frontendagent.FrontendAgentService,
	frontendPipelineServiceV2 *frontendpipeline.FrontendPipelineService,
	frontendNodeServiceV2 *frontendnode.FrontendNodeService,
) *Handler {
	return &Handler{
		AgentHandler:            agent.NewAgentHandler(agentService),
		AuthHandler:             auth.NewAuthHandler(authService),
		FrontendAgentHandler:    frontendagent.NewFrontendAgentHandler(frontendAgentServiceV2),
		FrontendPipelineHandler: frontendpipeline.NewFrontendPipelineHandler(frontendPipelineServiceV2),
		FrontendNodeHandler:     frontendnode.NewFrontendNodeHandler(frontendNodeServiceV2),
	}
}
