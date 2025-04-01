package utils

import (
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
	if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.Code == sqlite3.ErrConstraint {
		return true
	}
	return false
}
