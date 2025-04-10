package utils

import (
	"errors"
	"fmt"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/mattn/go-sqlite3"
)

func ValidateUserRegistrationRequest(request *models.UserRegisterRequest) error {
	if request.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if request.Email == "" {
		return fmt.Errorf("email cannot be empty")
	}
	if request.Password == "" {
		return fmt.Errorf("password cannot be empty")
	}
	return nil
}

func ValidateUserLoginRequest(request *models.LoginRequest) error {
	if request.Email == "" {
		return fmt.Errorf("email cannot be empty")
	}
	if request.Password == "" {
		return fmt.Errorf("password cannot be empty")
	}
	return nil
}

func ValidateAgentRegisterRequest(request *models.AgentRegisterRequest) error {
	if request.Hostname == "" {
		return fmt.Errorf("hostname cannot be empty")
	}
	if request.Platform == "" {
		return fmt.Errorf("platform cannot be empty")
	}
	if request.Version == "" {
		return fmt.Errorf("version cannot be empty")
	}
	return nil
}

func IsUniqueViolation(err error) bool {
	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr) {
		// Check for constraint violation (UNIQUE, PRIMARY KEY, etc.)
		return sqliteErr.Code == sqlite3.ErrConstraint && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique
	}
	return false
}

func ValidatePipelineRequest(request *models.CreatePipelineRequest) error {
	if request.Name == "" {
		return fmt.Errorf("pipeline name cannot be empty")
	}
	if request.CreatedBy == "" {
		return fmt.Errorf("created by cannot be empty")
	}
	if request.PipelineGraph.Nodes == nil {
		return fmt.Errorf("pipeline nodes cannot be empty")
	}
	if request.PipelineGraph.Edges == nil {
		return fmt.Errorf("pipeline edges cannot be empty")
	}
	return nil
}
