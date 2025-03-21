package agent

import (
	"database/sql"
	"errors"

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
func (ar *AgentRepository) RegisterAgent(req *AgentRegisterRequest) (*AgentRegisterResponse, error) {
	var existingAgent string
	var response *AgentRegisterResponse

	// Check if the agent is already registered
	err := ar.db.QueryRow("SELECT ID FROM agents WHERE hostname = ?", req.Hostname).Scan(&existingAgent)
	if err == nil {
		return nil, errors.New("agent for host" + req.Hostname + " already exists")
	} else if err != sql.ErrNoRows {
		return nil, errors.New("error checking database: " + err.Error())
	}

	// Insert the new agent into the database
	result, err := ar.db.Exec("INSERT INTO agents (type, version, hostname, platform, registered_at) VALUES (?, ?, ?, ?, ?)", req.Type, req.Version, req.Hostname, req.Platform, req.RegisteredAt)

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
	// FIXME: This is a placeholder, the actual configuration should be fetched from a configuration service

	return response, nil
}

func (f *AgentRepository) GetConfig(id string) (*models.Config, error) {
	// Initialize the config struct
	config := &models.Config{}

	// Use parameterized query to prevent SQL injection
	query := "SELECT ID, Name, Description, Config, TargetAgent, CreatedAt, UpdatedAt FROM config WHERE ID = ?"
	row := f.db.QueryRow(query, id)

	// Scan the result into the config struct
	err := row.Scan(
		&config.ID,
		&config.Name,
		&config.Description,
		&config.Config,
		&config.TargetAgent,
		&config.CreatedAt,
		&config.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// If no rows were returned, return a specific er
			return nil, errors.New("no config found with ID")
		}
		// Log and return other errors
		return nil, err
	}

	return config, nil
}
