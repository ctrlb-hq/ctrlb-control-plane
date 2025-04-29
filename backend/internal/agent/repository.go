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

// AgentExists checks if an agent with the given hostname exists.
func (ar *AgentRepository) AgentExists(hostname string) (bool, error) {
	var exists bool
	err := ar.db.QueryRow("SELECT EXISTS(SELECT 1 FROM agents WHERE hostname = ?)", hostname).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// RegisterAgent registers a new agent in the database.
func (ar *AgentRepository) RegisterAgent(req *models.AgentRegisterRequest) (*AgentRegisterResponse, error) {
	response := &AgentRegisterResponse{}

	result, err := ar.db.Exec(`
		INSERT INTO agents (name, type, version, hostname, platform, registered_at, ip) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		req.Name, req.Type, req.Version, req.Hostname, req.Platform, req.RegisteredAt, req.IP,
	)
	if err != nil {
		return nil, errors.New("error inserting agent: " + err.Error())
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, errors.New("error getting last insert ID: " + err.Error())
	}
	response.ID = id

	// Setting default config
	response.Config = constants.DefaultConfig

	return response, nil
}
