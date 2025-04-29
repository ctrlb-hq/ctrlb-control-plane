package utils_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
	"github.com/mattn/go-sqlite3"
)

func TestValidatePipelineRequest(t *testing.T) {
	tests := []struct {
		name   string
		input  models.CreatePipelineRequest
		hasErr bool
	}{
		{
			name: "valid pipeline",
			input: models.CreatePipelineRequest{
				Name:      "TestPipeline",
				CreatedBy: "admin",
				PipelineGraph: models.PipelineGraph{
					Nodes: []models.PipelineNodes{{ComponentID: 1, Name: "source", ComponentRole: "input"}},
					Edges: []models.PipelineEdges{{Source: "1", Target: "2"}},
				},
			},
			hasErr: false,
		},
		{
			name:   "empty name",
			input:  models.CreatePipelineRequest{CreatedBy: "admin"},
			hasErr: true,
		},
		{
			name:   "empty createdBy",
			input:  models.CreatePipelineRequest{Name: "Pipeline"},
			hasErr: true,
		},
		{
			name:   "nil nodes",
			input:  models.CreatePipelineRequest{Name: "Pipeline", CreatedBy: "admin", PipelineGraph: models.PipelineGraph{}},
			hasErr: true,
		},
		{
			name: "nil edges",
			input: models.CreatePipelineRequest{
				Name:      "Pipeline",
				CreatedBy: "admin",
				PipelineGraph: models.PipelineGraph{
					Nodes: []models.PipelineNodes{{ComponentID: 1, Name: "source", ComponentRole: "input"}},
				},
			},
			hasErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := utils.ValidatePipelineRequest(&tt.input)
			if tt.hasErr && err == nil {
				t.Error("expected error but got nil")
			} else if !tt.hasErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateUserRegistrationRequest(t *testing.T) {
	tests := []struct {
		name   string
		input  models.UserRegisterRequest
		hasErr bool
	}{
		{
			name: "valid",
			input: models.UserRegisterRequest{
				Name:     "John",
				Email:    "john@example.com",
				Password: "pass",
				Role:     "admin",
			},
			hasErr: false,
		},
		{
			name: "empty name",
			input: models.UserRegisterRequest{
				Name:     "",
				Email:    "john@example.com",
				Password: "pass",
				Role:     "admin",
			},
			hasErr: true,
		},
		{
			name: "empty email",
			input: models.UserRegisterRequest{
				Name:     "John",
				Email:    "",
				Password: "pass",
				Role:     "admin",
			},
			hasErr: true,
		},
		{
			name: "empty password",
			input: models.UserRegisterRequest{
				Name:     "John",
				Email:    "john@example.com",
				Password: "",
				Role:     "admin",
			},
			hasErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := utils.ValidateUserRegistrationRequest(&tt.input)
			if tt.hasErr && err == nil {
				t.Errorf("expected error but got nil")
			} else if !tt.hasErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateUserLoginRequest(t *testing.T) {
	err := utils.ValidateUserLoginRequest(&models.LoginRequest{Email: "", Password: "p"})
	if err == nil {
		t.Error("expected error for empty email")
	}

	err = utils.ValidateUserLoginRequest(&models.LoginRequest{Email: "e@e.com", Password: ""})
	if err == nil {
		t.Error("expected error for empty password")
	}

	err = utils.ValidateUserLoginRequest(&models.LoginRequest{Email: "e@e.com", Password: "p"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateAgentRegisterRequest(t *testing.T) {
	err := utils.ValidateAgentRegisterRequest(&models.AgentRegisterRequest{Hostname: "", Platform: "Linux", Version: "1.0"})
	if err == nil {
		t.Error("expected error for empty hostname")
	}
}

func TestIsUniqueViolation(t *testing.T) {
	sqliteErr := sqlite3.Error{Code: sqlite3.ErrConstraint, ExtendedCode: sqlite3.ErrConstraintUnique}
	wrappedErr := fmt.Errorf("wrapped: %w", sqliteErr)
	if !utils.IsUniqueViolation(wrappedErr) {
		t.Error("expected true for unique violation")
	}

	notUniqueErr := errors.New("some other error")
	if utils.IsUniqueViolation(notUniqueErr) {
		t.Error("expected false for non-unique violation")
	}
}
