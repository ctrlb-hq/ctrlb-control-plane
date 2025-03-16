package frontendpipeline

import (
	"database/sql"
)

type FrontendPipelineRepository struct {
	db *sql.DB
}

// NewFrontendPipelineRepository creates a new FrontendPipelineRepository
func NewFrontendPipelineRepository(db *sql.DB) *FrontendPipelineRepository {
	return &FrontendPipelineRepository{db: db}
}

func (f *FrontendPipelineRepository) GetAllPipelines() ([]*Pipeline, error) {
	return nil, nil
}

func (f *FrontendPipelineRepository) GetPipelineInfo(pipelineId int) (*PipelineInfo, error) {
	pipelineInfo := &PipelineInfo{}

	// Query the database for the pipeline info
	query := `SELECT id, name, created_by, created_at, updated_at FROM pipelines WHERE id = ?`
	err := f.db.QueryRow(query, pipelineId).Scan(&pipelineInfo.ID, &pipelineInfo.Name, &pipelineInfo.CreatedBy, &pipelineInfo.CreatedAt, &pipelineInfo.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return pipelineInfo, nil
}

func (f *FrontendPipelineRepository) DeletePipeline(pipelineId int) error {
	_, err := f.db.Exec("DELETE FROM pipelines WHERE id = ?", pipelineId)
	if err != nil {
		return err
	}
	return nil
}

func (f *FrontendPipelineRepository) GetAllAgentsAttachedToPipeline(PipelineId int) ([]AgentInfoHome, error) {
	var agents []AgentInfoHome
	row, err := f.db.Query("SELECT id, name, version, pipeline_name FROM agents where pipeline_id = ?", PipelineId)
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
