package frontendpipeline

import (
	"database/sql"
	"fmt"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
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

func (f *FrontendPipelineRepository) VerifyPipelineExists(pipelineId int) error {
	var verifyId int
	err := f.db.QueryRow(`SELECT pipeline_id FROM pipelines WHERE pipeline_id = ?`, pipelineId).Scan(&verifyId)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("no pipeline found with id %d", pipelineId)
		}
		return err
	}
	return nil
}

func (f *FrontendPipelineRepository) DeletePipeline(pipelineId int) error {
	_, err := f.db.Exec("DELETE FROM pipelines WHERE id = ?", pipelineId)
	if err != nil {
		return err
	}
	return nil
}

func (f *FrontendPipelineRepository) GetAllAgentsAttachedToPipeline(PipelineId int) ([]models.AgentInfoHome, error) {
	var agents []models.AgentInfoHome

	// Optimized query for SQLite
	query := `
		SELECT a.id, a.name, a.version, a.pipeline_name, 
		       IFNULL(m.logs_rate_sent, 0), IFNULL(m.traces_rate_sent, 0), 
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
		agent := models.AgentInfoHome{}
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
	setQuery := `UPDATE agents SET pipeline_id = NULL, pipeline_name = NULL WHERE id = ?`

	_, err := f.db.Exec(setQuery, agentId)
	if err != nil {
		return fmt.Errorf("failed to detach agent: %w", err)
	}

	return nil
}

func (f *FrontendPipelineRepository) AttachAgentToPipeline(pipelineId int, agentId int) error {
	setQuery := `UPDATE agents SET pipeline_id = ?, pipeline_name = (SELECT name FROM pipelines WHERE pipeline_id = ?) WHERE id = ?`

	_, err := f.db.Exec(setQuery, pipelineId, pipelineId, agentId)
	if err != nil {
		return fmt.Errorf("failed to attach agent: %w", err)
	}

	return nil
}

func (f *FrontendPipelineRepository) GetPipelineGraph(pipelineId int) (*PipelineGraph, error) {
	// Get pipeline components (nodes)
	nodes, err := f.getPipelineComponents(pipelineId)
	if err != nil {
		return nil, fmt.Errorf("failed to get pipeline components: %w", err)
	}

	// Get component dependencies (edges)
	edges, err := f.getPipelineEdges(pipelineId)
	if err != nil {
		return nil, fmt.Errorf("failed to get pipeline dependencies: %w", err)
	}

	return &PipelineGraph{
		Nodes: nodes,
		Edges: edges,
	}, nil
}

func (f *FrontendPipelineRepository) getPipelineComponents(pipelineId int) ([]PipelineComponent, error) {
	rows, err := f.db.Query(`
		SELECT component_id, name, component_role, plugin_name 
		FROM pipeline_components 
		WHERE pipeline_id = ?`, pipelineId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []PipelineComponent
	for rows.Next() {
		var node PipelineComponent
		if err := rows.Scan(&node.ComponentID, &node.Name, &node.ComponentRole, &node.PluginName); err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return nodes, rows.Err()
}

func (f *FrontendPipelineRepository) getPipelineEdges(pipelineId int) ([]PipelineEdge, error) {
	rows, err := f.db.Query(`
		SELECT parent_component_id, child_component_id 
		FROM pipeline_component_edges 
		WHERE pipeline_id = ?`, pipelineId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []PipelineEdge
	for rows.Next() {
		var edge PipelineEdge
		if err := rows.Scan(&edge.FromComponentID, &edge.ToComponentID); err != nil {
			return nil, err
		}
		edges = append(edges, edge)
	}
	return edges, rows.Err()
}

func (f *FrontendPipelineRepository) SyncPipelineGraph(pipelineID int, components []PipelineComponent, edges []PipelineEdge) error {

	tx, err := f.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	_, err = tx.Exec(`DELETE FROM pipeline_components WHERE pipeline_id = ?`, pipelineID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete existing components: %w", err)
	}

	componentIDMap := make(map[string]int) // If you need to map names to IDs for edge linking
	insertComponentStmt, err := tx.Prepare(`
		INSERT INTO pipeline_components (pipeline_id, component_role, plugin_name, name)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to prepare component insert: %w", err)
	}
	defer insertComponentStmt.Close()

	for _, comp := range components {
		res, err := insertComponentStmt.Exec(pipelineID, comp.ComponentRole, comp.PluginName, comp.Name)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert component %s: %w", comp.Name, err)
		}
		id, _ := res.LastInsertId()
		componentIDMap[comp.Name] = int(id) // Optional
	}

	insertEdgeStmt, err := tx.Prepare(`
		INSERT INTO pipeline_component_edges (pipeline_id, parent_component_id, child_component_id)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to prepare edge insert: %w", err)
	}
	defer insertEdgeStmt.Close()

	for _, edge := range edges {
		_, err := insertEdgeStmt.Exec(pipelineID, edge.FromComponentID, edge.ToComponentID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert edge (%d → %d): %w", edge.FromComponentID, edge.ToComponentID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit graph sync: %w", err)
	}

	return nil
}
