package frontendagent

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
)

type FrontendAgentRepository struct {
	db *sql.DB
}

// NewFrontendAgentRepository creates a new FrontendAgentRepository
func NewFrontendAgentRepository(db *sql.DB) *FrontendAgentRepository {
	return &FrontendAgentRepository{db: db}
}

// GetAllAgents retrieves all agents from the database
func (f *FrontendAgentRepository) GetAllAgents() ([]Agent, error) {
	var agents []Agent

	rows, err := f.db.Query("SELECT id, name, type, version, hostname, platform, configId, registeredAt FROM agents WHERE isPipeline = false")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var registeredAt sql.NullTime
		var agent Agent

		if err := rows.Scan(&agent.ID, &agent.Name, &agent.Type, &agent.Version, &agent.Hostname, &agent.Platform, &agent.ConfigID, &registeredAt); err != nil {
			return nil, err
		}

		if registeredAt.Valid {
			agent.RegisteredAt = registeredAt.Time // Assign if valid
		}
		agents = append(agents, agent)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return agents, nil
}

// GetAgent retrieves a specific agent by ID
func (f *FrontendAgentRepository) GetAgent(id string) (*Agent, error) {
	agent := &Agent{}
	row := f.db.QueryRow("SELECT id, name, type, version, hostname, platform, configID, isPipeline, registeredAt FROM agents WHERE id = ?", id)

	if err := row.Scan(&agent.ID, &agent.Name, &agent.Type, &agent.Version, &agent.Hostname, &agent.Platform, &agent.ConfigID, &agent.IsPipeline, &agent.RegisteredAt); err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No agent found with ID: %s", id)
			return nil, errors.New("agent not found")
		}
		return nil, err
	}

	return agent, nil
}

// DeleteAgent removes an agent by ID
func (f *FrontendAgentRepository) DeleteAgent(id string) error {
	result, err := f.db.Exec("DELETE FROM agents WHERE id = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		log.Printf("No agent found with ID: %s", id)
		return fmt.Errorf("no agent found with id %s", id)
	}

	return nil
}

// GetConfig retrieves the configuration for a specific agent
func (f *FrontendAgentRepository) GetConfig(id string) (*models.Config, error) {
	config := &models.Config{}
	query := "SELECT ID, Name, Description, Config, TargetAgent, CreatedAt, UpdatedAt FROM config WHERE ID = ?"
	row := f.db.QueryRow(query, id)

	if err := row.Scan(&config.ID, &config.Name, &config.Description, &config.Config, &config.TargetAgent, &config.CreatedAt, &config.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No config found with ID: %s", id)
			return nil, errors.New("no config found with ID")
		}
		log.Printf("Error scanning config with ID %s: %v", id, err)
		return nil, err
	}

	return config, nil
}

// GetMetrics retrieves metrics for a specific agent
func (f *FrontendAgentRepository) GetMetrics(id string) (*models.AgentMetrics, error) {
	agentMetrics := &models.AgentMetrics{}
	row := f.db.QueryRow("SELECT AgentID, Status, ExportedDataVolume, UptimeSeconds, DroppedRecords, UpdatedAt FROM agent_metrics WHERE AgentID = ?", id)

	if err := row.Scan(&agentMetrics.AgentID, &agentMetrics.Status, &agentMetrics.ExportedDataVolume, &agentMetrics.UptimeSeconds, &agentMetrics.DroppedRecords, &agentMetrics.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("no metrics collected yet")
		}
		return nil, err
	}

	return agentMetrics, nil
}
