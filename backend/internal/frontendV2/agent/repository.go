package frontendagent

import (
	"database/sql"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
)

type FrontendAgentRepository struct {
	db *sql.DB
}

// NewFrontendAgentRepository creates a new FrontendAgentRepository
func NewFrontendAgentRepository(db *sql.DB) *FrontendAgentRepository {
	return &FrontendAgentRepository{db: db}
}

func (f *FrontendAgentRepository) GetAllAgents() ([]AgentInfoHome, error) {
	var agents []AgentInfoHome
	row, err := f.db.Query("SELECT id, name, version, pipeline_name FROM agents")
	if err != nil {
		return nil, err
	}
	defer row.Close()

	for row.Next() {
		agent := AgentInfoHome{}
		err := row.Scan(&agent.ID, &agent.Name, &agent.Version, &agent.PipelineName)
		if err != nil {
			return nil, err
		}
		agents = append(agents, agent)
	}

	for i := range agents {
		// Get the status of the agent
		agentStatus := f.db.QueryRow("SELECT log_rate_sent, traces_rate_sent, metrics_rate_sent, status FROM aggregated_agent_metrics WHERE agent_id = ?", agents[i].ID)

		err := agentStatus.Scan(&agents[i].LogRate, &agents[i].TraceRate, &agents[i].MetricsRate, &agents[i].Status)
		if err != nil {
			return nil, err
		}
	}
	return agents, nil
}

// GetAgent retrieves a specific agent by ID
func (f *FrontendAgentRepository) GetAgent(id string) (*AgentInfoWithLabels, error) {
	agent := &AgentInfoWithLabels{}

	err := f.db.QueryRow("SELECT id, name, version, pipeline_id, pipeline_name, hostname, platform FROM agents WHERE id = ?", id).Scan(&agent.ID, &agent.Name, &agent.Version, &agent.PipelineID, &agent.PipelineName, &agent.Hostname, &agent.Platform)
	if err != nil {
		return nil, err
	}

	err = f.db.QueryRow("SELECT status FROM aggregated_agent_metrics WHERE agent_id = ?", id).Scan(&agent.Status)
	if err != nil {
		return nil, err
	}

	agent.Labels = make(map[string]string)
	rows, err := f.db.Query("SELECT key, value FROM agents_labels WHERE agent_id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var key, value string
		err := rows.Scan(&key, &value)
		if err != nil {
			return nil, err
		}
		agent.Labels[key] = value
	}

	return agent, nil
}

func (f *FrontendAgentRepository) GetAgentHostname(id string) (string, error) {
	var hostname string

	err := f.db.QueryRow("SELECT hostname FROM agents WHERE id = ?", id).Scan(&hostname)
	if err != nil {
		return "", err
	}

	return hostname, nil
}

// DeleteAgent removes an agent by ID
func (f *FrontendAgentRepository) DeleteAgent(id string) error {
	// This will delete all related labels, metrics and extenstions as well
	if _, err := f.db.Exec("DELETE FROM agents WHERE id = ?", id); err != nil {
		return err
	}
	return nil
}

// GetMetricsForGraph retrieves metrics for a specific agent
func (f *FrontendAgentRepository) GetMetricsForGraph(id string) (*models.AgentMetrics, error) {
	//TODO: Implement this
	return nil, nil
}
