package repositories

import (
	"database/sql"
	"errors"
	"log"

	"github.com/ctrlb-hq/all-father/internal/models"
)

func NewAgentRepository(db *sql.DB) *AgentRepository {
	return &AgentRepository{db: db}
}

func (ar *AgentRepository) AddAgent(agent *models.Agent) error {
	var existingAgent string
	err := ar.db.QueryRow("SELECT ID FROM agents WHERE Name = ?", agent.Name).Scan(&existingAgent)
	if err == nil {
		log.Printf("Agent already registered: %s", agent.Name)
		return errors.New("agent " + agent.Name + " already exists")
	} else if err != sql.ErrNoRows {
		log.Println(err)
		return errors.New("error encountered while checking database to add agent" + err.Error())
	}

	_, err = ar.db.Exec("INSERT INTO agents (ID, Name, Type, Version, Hostname, Platform, Config) VALUES (?, ?, ?, ?, ?, ?, ?)",
		agent.ID, agent.Name, agent.Type, agent.Version, agent.Hostname, agent.Platform, agent.Config)
	if err != nil {
		log.Println(err)
		return errors.New("error encountered while adding new agent " + err.Error())
	}

	log.Println("New agent added:", agent.Name)
	return nil
}

func (ar *AgentRepository) UpdateConfig(config models.ConfigUpdateRequest) error {
	err := ar.db.QueryRow("SELECT 1 FROM agents WHERE ID = ?", config.AgentID).Scan(new(int))
	if err == sql.ErrNoRows {
		log.Println("Agent not registered:", config.AgentID)
		return errors.New("agent not registered with ID:" + config.AgentID)
	} else if err != nil {
		log.Println(err)
		return errors.New("error encountered while updating config:" + err.Error())
	}

	query := `UPDATE agents SET Config = ? WHERE ID = ?`
	_, err = ar.db.Exec(query, config.Config, config.AgentID)
	if err != nil {
		log.Println(err)
		return errors.New("error encountered while updating config:" + err.Error())
	}

	log.Println("Config updated for agent:", config.AgentID)
	return nil
}

func (ar *AgentRepository) RemoveAgent(agentID string) error {
	stmt, err := ar.db.Prepare("DELETE FROM agents WHERE ID = ?")
	if err != nil {
		log.Println("Error preparing DELETE statement:", err)
		return errors.New("error preparing DELETE statement for removing agent:" + agentID + err.Error())
	}
	defer stmt.Close()

	result, err := stmt.Exec(agentID)
	if err != nil {
		log.Println("Error executing DELETE:", err)
		return errors.New("error in removing agent:" + agentID + err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println("Error getting rows affected:", err)
		return errors.New("error getting rows affected while removing agent: " + agentID + err.Error())
	}

	if rowsAffected == 0 {
		log.Printf("No agent found with ID: %s", agentID)
		return errors.New("no agent found to delete with ID: " + agentID)
	}

	log.Printf("Agent with ID %s successfully removed", agentID)
	return nil
}

func (ar *AgentRepository) GetAgentHost(agentID string) (string, error) {
	var hostname string
	err := ar.db.QueryRow("SELECT Hostname FROM agents WHERE ID = ?", agentID).Scan(&hostname)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("no agent found with the ID: " + agentID)
		}
		return "", err
	}
	return hostname, nil
}

func (ar *AgentRepository) GetAgentConfig(agentID string) (string, error) {
	var config string
	err := ar.db.QueryRow("SELECT Config FROM agents WHERE ID = ?", agentID).Scan(&config)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("no agent found while searching for Config with ID: " + agentID)
		}
		return "", err
	}
	return config, nil
}

func (ar *AgentRepository) GetAgentType(agentID string) (string, error) {
	var config string
	err := ar.db.QueryRow("SELECT Type FROM agents WHERE ID = ?", agentID).Scan(&config)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("no agent found while searching for Type with ID: " + agentID)
		}
		return "", err
	}
	return config, nil
}
