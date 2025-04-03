package frontendpipeline

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
)

type FrontendPipelineRepository struct {
	db *sql.DB
}

// NewFrontendPipelineRepository creates a new FrontendPipelineRepository
func NewFrontendPipelineRepository(db *sql.DB) *FrontendPipelineRepository {
	return &FrontendPipelineRepository{db: db}
}

func (f *FrontendPipelineRepository) PipelineExists(pipelineId int) bool {
	var verifyId int
	err := f.db.QueryRow("SELECT pipeline_id FROM pipelines WHERE pipeline_id = ? LIMIT 1", pipelineId).Scan(&verifyId)
	return err == nil
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

func (f *FrontendPipelineRepository) CreatePipeline(createPipelineRequest CreatePipelineRequest) (string, error) {
	tx, err := f.db.Begin()
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Insert the pipeline
	res, err := tx.Exec(
		"INSERT INTO pipelines (name, created_by) VALUES (?, ?)",
		createPipelineRequest.Name,
		createPipelineRequest.CreatedBy,
	)
	if err != nil {
		_ = tx.Rollback()
		return "", fmt.Errorf("failed to insert pipeline: %w", err)
	}

	// Fetch the last-insert ID
	id, err := res.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		return "", fmt.Errorf("failed to retrieve pipeline ID: %w", err)
	}

	// Use the same transaction for everything else
	// (SyncPipelineGraph would also need to be updated not to require a context)
	if err := f.SyncPipelineGraph(tx, int(id), createPipelineRequest.PipelineGraph.Nodes, createPipelineRequest.PipelineGraph.Edges); err != nil {
		_ = tx.Rollback()
		return "", fmt.Errorf("failed to sync pipeline graph: %w", err)
	}
	// Commit
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	for _, agentId := range createPipelineRequest.AgentIDs {
		err = f.AttachAgentToPipeline(int(id), agentId)
		if err != nil {
			utils.Logger.Error(fmt.Sprintf("Failed to attach agent [ID: %v] to pipeline [ID: %v]", agentId, id))
		}
	}

	return strconv.FormatInt(id, 10), nil
}

func (f *FrontendPipelineRepository) DeletePipeline(pipelineId int) error {
	_, err := f.db.Exec("DELETE FROM pipelines WHERE pipeline_id = ?", pipelineId)
	if err != nil {
		return err
	}
	return nil
}

func (f *FrontendPipelineRepository) GetAllAgentsAttachedToPipeline(PipelineId int) ([]models.AgentInfoHome, error) {
	var agents []models.AgentInfoHome

	// Optimized query for SQLite
	query := `
		SELECT a.id, a.name, a.version, a.pipeline_name, a.hostname, a.IP, 
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
		err := rows.Scan(&agent.ID, &agent.Name, &agent.Version, &agent.PipelineName, &agent.Hostname, &agent.IP,
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
	//TODO: Add logic to push config to agent
	return nil
}

func (f *FrontendPipelineRepository) GetPipelineGraph(pipelineId int) (*models.PipelineGraph, error) {
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

	return &models.PipelineGraph{
		Nodes: nodes,
		Edges: edges,
	}, nil
}

func (f *FrontendPipelineRepository) getPipelineComponents(pipelineId int) ([]models.PipelineComponent, error) {
	rows, err := f.db.Query(`
		SELECT component_id, name, component_role, component_name, config 
		FROM pipeline_components 
		WHERE pipeline_id = ?`, pipelineId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []models.PipelineComponent
	for rows.Next() {
		var node models.PipelineComponent
		var configStr string // Declare configStr inside the loop

		if err := rows.Scan(&node.ComponentID, &node.Name, &node.ComponentRole, &node.ComponentName, &configStr); err != nil {
			return nil, err
		}

		if err = json.Unmarshal([]byte(configStr), &node.Config); err != nil {
			utils.Logger.Sugar().Errorf("Failed to unmarshal config: %v", err)
			return nil, err
		}

		nodes = append(nodes, node)
	}
	return nodes, rows.Err()
}

func (f *FrontendPipelineRepository) getPipelineEdges(pipelineId int) ([]models.PipelineEdge, error) {
	rows, err := f.db.Query(`
		SELECT parent_component_id, child_component_id 
		FROM pipeline_component_edges 
		WHERE pipeline_id = ?`, pipelineId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []models.PipelineEdge
	for rows.Next() {
		var edge models.PipelineEdge
		if err := rows.Scan(&edge.Source, &edge.Target); err != nil {
			return nil, err
		}
		edges = append(edges, edge)
	}
	return edges, rows.Err()
}

func (f *FrontendPipelineRepository) SyncPipelineGraph(tx *sql.Tx, pipelineID int, components []models.PipelineComponent, edges []models.PipelineEdge) error {
	shouldCommit := false
	var err error

	// Start a transaction only if one wasn't passed
	if tx == nil {
		tx, err = f.db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		shouldCommit = true
	}

	// First, delete existing components (cascades to edges due to FK constraints)
	_, err = tx.Exec(`DELETE FROM pipeline_components WHERE pipeline_id = ?`, pipelineID)
	if err != nil {
		if shouldCommit {
			_ = tx.Rollback()
		}
		return fmt.Errorf("failed to delete existing components: %w", err)
	}

	// Map to store the relationship between incoming component IDs and database IDs
	componentIDMap := make(map[int]int)

	// Prepare statement for component insertion
	insertComponentStmt, err := tx.Prepare(`
        INSERT INTO pipeline_components (pipeline_id, component_role, component_name, name, config)
        VALUES (?, ?, ?, ?, ?)
    `)
	if err != nil {
		if shouldCommit {
			_ = tx.Rollback()
		}
		return fmt.Errorf("failed to prepare component insert: %w", err)
	}
	defer insertComponentStmt.Close()

	// Insert all components first
	for _, comp := range components {
		// Marshal component config to JSON string
		configBytes, err := json.Marshal(comp.Config)
		if err != nil {
			if shouldCommit {
				_ = tx.Rollback()
			}
			return fmt.Errorf("failed to marshal config for component %s: %w", comp.Name, err)
		}

		// Insert the component
		res, err := insertComponentStmt.Exec(pipelineID, comp.ComponentRole, comp.ComponentName, comp.Name, string(configBytes))
		if err != nil {
			if shouldCommit {
				_ = tx.Rollback()
			}
			return fmt.Errorf("failed to insert component %s: %w", comp.Name, err)
		}

		// Store the mapping between incoming ID and database ID
		id, err := res.LastInsertId()
		if err != nil {
			if shouldCommit {
				_ = tx.Rollback()
			}
			return fmt.Errorf("failed to get last insert ID for component %s: %w", comp.Name, err)
		}
		componentIDMap[comp.ComponentID] = int(id)
	}

	// Prepare statement for edge insertion
	insertEdgeStmt, err := tx.Prepare(`
        INSERT INTO pipeline_component_edges (pipeline_id, parent_component_id, child_component_id)
        VALUES (?, ?, ?)
    `)
	if err != nil {
		if shouldCommit {
			_ = tx.Rollback()
		}
		return fmt.Errorf("failed to prepare edge insert: %w", err)
	}
	defer insertEdgeStmt.Close()

	// Insert all edges
	for _, edge := range edges {
		// Map the component IDs to their database IDs
		sourceID, sourceExists := componentIDMap[edge.Source]
		targetID, targetExists := componentIDMap[edge.Target]

		if !sourceExists || !targetExists {
			if shouldCommit {
				_ = tx.Rollback()
			}
			return fmt.Errorf("invalid edge reference: source or target component not found (source: %d, target: %d)", edge.Source, edge.Target)
		}

		// Insert the edge
		_, err := insertEdgeStmt.Exec(pipelineID, sourceID, targetID)
		if err != nil {
			if shouldCommit {
				_ = tx.Rollback()
			}
			return fmt.Errorf("failed to insert edge (%d → %d): %w", sourceID, targetID, err)
		}
	}

	// Commit the transaction if we started it
	if shouldCommit {
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit graph sync: %w", err)
		}
	}

	return nil
}

func (f *FrontendPipelineRepository) GetAgentInfo(agentId int) (*models.AgentInfoHome, error) {
	agent := &models.AgentInfoHome{}
	var pipelineName sql.NullString

	err := f.db.QueryRow("SELECT id, name, version, pipeline_name, hostname, ip FROM agents WHERE id = ?", agentId).Scan(&agent.ID, &agent.Name, &agent.Version, &pipelineName, &agent.Hostname, &agent.IP)
	if err != nil {
		return nil, err
	}

	if pipelineName.Valid {
		agent.PipelineName = pipelineName.String
	} else {
		agent.PipelineName = ""
	}

	err = f.db.QueryRow("SELECT status FROM aggregated_agent_metrics WHERE agent_id = ?", agentId).Scan(&agent.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			agent.Status = "unknown"
		} else {
			return nil, err
		}
	}

	return agent, nil
}
