package services

import (
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/repositories"
)

type AgentService struct {
	AgentRepository *repositories.AgentRepository
	AgentQueue      *AgentQueue
}

type AuthService struct {
	AuthRepository *repositories.AuthRepository
}

type Services struct {
	AgentService *AgentService
	AuthService  *AuthService
}
