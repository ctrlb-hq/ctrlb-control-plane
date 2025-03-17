package frontendpipeline

import (
	"database/sql"
	"fmt"
)

type FrontendPipelineRepository struct {
	db *sql.DB
}

// NewFrontendPipelineRepository creates a new FrontendPipelineRepository
func NewFrontendPipelineRepository(db *sql.DB) *FrontendPipelineRepository {
	return &FrontendPipelineRepository{db: db}
}

func (f *FrontendPipelineRepository) GetAllPipelines() ([]*Pipeline, error) {
	var pipelines []*Pipeline

	query := `
    SELECT 
        p.pipeline_id AS id,
        p.name,
        COUNT(a.id) AS agents,
        COALESCE(SUM(ag.data_received_bytes), 0) AS incomingBytes,
        COALESCE(SUM(ag.data_sent_bytes), 0) AS outgoingBytes,
        p.updated_at
    FROM pipelines p
    LEFT JOIN agents a ON p.pipeline_id = a.pipeline_id
    LEFT JOIN aggregated_agent_metrics ag ON a.id = ag.agent_id
    GROUP BY p.pipeline_id;
    `

	rows, err := f.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate through query results and store them in the slice
	for rows.Next() {
		pipeline := &Pipeline{}
		err := rows.Scan(&pipeline.ID, &pipeline.Name, &pipeline.Agents, &pipeline.IncomingBytes, &pipeline.OutgoingBytes, &pipeline.UpdatedAt)
		if err != nil {
			return nil, err
		}
		pipelines = append(pipelines, pipeline)
	}

	// Check for any errors from row iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return pipelines, nil
}

func (f *FrontendPipelineRepository) GetPipelineInfo(pipelineId int) (*PipelineInfo, error) {
	pipelineInfo := &PipelineInfo{}

	// Query the database for the pipeline info
	query := `SELECT pipeline_id, name, created_by, created_at, updated_at FROM pipelines WHERE pipeline_id = ?`
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

	// Optimized query for SQLite
	query := `
		SELECT a.id, a.name, a.version, a.pipeline_name, 
		       IFNULL(m.log_rate_sent, 0), IFNULL(m.traces_rate_sent, 0), 
		       IFNULL(m.metrics_rate_sent, 0), IFNULL(m.status, '')
		FROM agents a
		LEFT JOIN aggregated_agent_metrics m ON a.id = m.agent_id
		WHERE a.pipeline_id = ?
	`
	rows, err := f.db.Query(query, PipelineId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		agent := AgentInfoHome{}
		err := rows.Scan(&agent.ID, &agent.Name, &agent.Version, &agent.PipelineName,
			&agent.LogRate, &agent.TraceRate, &agent.MetricsRate, &agent.Status)
		if err != nil {
			return nil, err
		}
		agents = append(agents, agent)
	}

	return agents, nil
}

func (f *FrontendPipelineRepository) DetachAgentFromPipeline(pipelineId int, agentId int) error {
	query := `SELECT pipeline_id FROM agents WHERE id = ?`
	var verifyId int
	err := f.db.QueryRow(query, agentId).Scan(&verifyId)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("no agent found with id %d", agentId)
		}
		return err
	}

	if verifyId != pipelineId {
		return fmt.Errorf("agent is not attached to given pipeline")
	}

	setQuery := `UPDATE agents SET pipeline_id = NULL, pipeline_name = NULL WHERE id = ?`

	_, err = f.db.Exec(setQuery, agentId)
	if err != nil {
		return fmt.Errorf("failed to detach agent: %w", err)
	}

	return nil
}
