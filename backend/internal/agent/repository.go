package agent

import (
	"database/sql"
	"errors"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/constants"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
)

// AgentRepository interacts with the agent database.
type AgentRepository struct {
	db *sql.DB // Database connection
}

// NewAgentRepository creates a new AgentRepository.
func NewAgentRepository(db *sql.DB) *AgentRepository {
	return &AgentRepository{db: db}
}

// RegisterAgent registers a new agent in the database.
func (ar *AgentRepository) RegisterAgent(req *models.AgentRegisterRequest) (*AgentRegisterResponse, error) {
	// var existingAgent int64

	// Check if the agent is already registered
	// err := ar.db.QueryRow("SELECT ID FROM agents WHERE hostname = ?", req.Hostname).Scan(&existingAgent)
	// if err == nil {
	// 	return nil, errors.New("agent for host " + req.Hostname + " already exists")
	// } else if err != sql.ErrNoRows {
	// 	return nil, errors.New("error checking database: " + err.Error())
	// }

	response := &AgentRegisterResponse{}
	// Insert the new agent into the database
	result, err := ar.db.Exec("INSERT INTO agents (name, type, version, hostname, platform, registered_at, ip) VALUES (?, ?, ?, ?, ?, ?, ?)", req.Name, req.Type, req.Version, req.Hostname, req.Platform, req.RegisteredAt, req.IP)

	if err != nil {
		return nil, errors.New("error inserting agent: " + err.Error())
	}

	// Get the ID of the newly inserted agent
	id, err := result.LastInsertId()
	if err != nil {
		return nil, errors.New("error getting last insert ID: " + err.Error())
	}
	response.ID = id

	// Get the default configuration for the agent
	response.Config = map[string]any{
		"receivers": map[string]any{
			"otlp": map[string]any{
				"protocols": map[string]any{
					"http": map[string]any{}, // Default: 0.0.0.0:4318
					"grpc": map[string]any{}, // Default: 0.0.0.0:4317
				},
			},
		},
		"processors": map[string]any{},
		"exporters": map[string]any{
			"debug": map[string]any{}, // Debug exporter for logs
		},
		"service": map[string]any{
			"telemetry": constants.TelemetryService,
			"pipelines": map[string]any{
				"logs/default": map[string]any{
					"receivers":  []any{"otlp"},
					"processors": []any{},
					"exporters":  []any{"debug"},
				},
			},
		},
	}

	return response, nil
}
