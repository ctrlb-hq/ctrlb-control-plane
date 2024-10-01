package services

import (
	"github.com/ctrlb-hq/all-father/internal/repositories"
)

type AgentService struct {
	AgentRepository *repositories.AgentRepository
	AgentQueue      *AgentQueue
}

type Services struct {
	AgentService *AgentService
}
