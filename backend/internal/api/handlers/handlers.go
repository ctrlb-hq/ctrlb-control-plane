package handler

import (
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/services"
)

type AgentHandler struct {
	AgentService *services.AgentService
}

type AuthHandler struct {
	AuthService *services.AuthService
}
