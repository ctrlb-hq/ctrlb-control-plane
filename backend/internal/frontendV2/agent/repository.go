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
func (f *FrontendAgentRepository) GetAgent(id string) (*AgentInfoHome, error) {
	agent := &AgentInfoHome{}

	return agent, nil
}

// DeleteAgent removes an agent by ID
func (f *FrontendAgentRepository) DeleteAgent(id string) error {

	return nil
}

// GetConfig retrieves the configuration for a specific agent
func (f *FrontendAgentRepository) GetConfig(id string) (*models.Config, error) {
	config := &models.Config{}
	return config, nil
}

// GetMetrics retrieves metrics for a specific agent
func (f *FrontendAgentRepository) GetMetrics(id string) (*models.AgentMetrics, error) {

	return nil, nil
}
