package frontendpipeline

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
)

// FrontendPipelineRepository handles database operations for pipelines.
type FrontendPipelineRepository struct {
	db *sql.DB // Database connection
}

// NewFrontendPipelineRepository creates a new instance of FrontendPipelineRepository.
func NewFrontendPipelineRepository(db *sql.DB) *FrontendPipelineRepository {
	return &FrontendPipelineRepository{db: db}
}

// GetAllPipelines retrieves all pipelines from the database.
func (f *FrontendPipelineRepository) GetAllPipelines() ([]Pipeline, error) {
	var pipelines []Pipeline

	// Query the database for pipelines
	rows, err := f.db.Query("SELECT id, name, type, version, hostname, platform, configId, registeredAt FROM agents WHERE isPipeline = true")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate through the result set
	for rows.Next() {
		var registeredAt sql.NullTime
		var pipeline Pipeline
		err := rows.Scan(&pipeline.ID, &pipeline.Name, &pipeline.Type, &pipeline.Version, &pipeline.Hostname, &pipeline.Platform, &pipeline.ConfigID, &registeredAt)
		if err != nil {
			return nil, err
		}

		if registeredAt.Valid {
			pipeline.RegisteredAt = registeredAt.Time
		}
		pipelines = append(pipelines, pipeline)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return pipelines, nil
}

// GetPipeline retrieves a specific pipeline by its ID.
func (f *FrontendPipelineRepository) GetPipeline(id string) (*Pipeline, error) {
	pipeline := &Pipeline{}
	row := f.db.QueryRow("SELECT id, name, type, version, hostname, platform, configId, isPipeline, registeredAt FROM agents WHERE id = ?", id)
	err := row.Scan(&pipeline.ID, &pipeline.Name, &pipeline.Type, &pipeline.Version, &pipeline.Hostname, &pipeline.Platform, &pipeline.ConfigID, &pipeline.IsPipeline, &pipeline.RegisteredAt)
	if err != nil {
		return nil, err
	}
	return pipeline, nil
}

// DeletePipeline removes a pipeline from the database.
func (f *FrontendPipelineRepository) DeletePipeline(id string) error {
	result, err := f.db.Exec("DELETE FROM agents WHERE id = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no pipeline found with id %s", id)
	}

	return nil
}

// GetConfig retrieves configuration details by ID.
func (f *FrontendPipelineRepository) GetConfig(id string) (*models.Config, error) {
	config := &models.Config{}
	query := "SELECT ID, Name, Description, Config, TargetAgent, CreatedAt, UpdatedAt FROM config WHERE ID = ?"
	row := f.db.QueryRow(query, id)

	err := row.Scan(&config.ID, &config.Name, &config.Description, &config.Config, &config.TargetAgent, &config.CreatedAt, &config.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No config found with ID: %s", id)
			return nil, errors.New("no config found with ID")
		}
		log.Printf("Error scanning config with ID %s: %v", id, err)
		return nil, err
	}

	return config, nil
}

// GetMetrics retrieves metrics for a specific pipeline.
func (f *FrontendPipelineRepository) GetMetrics(id string) (*models.AgentMetrics, error) {
	pipelineMetrics := &models.AgentMetrics{}
	row := f.db.QueryRow("SELECT AgentID, Status, ExportedDataVolume, UptimeSeconds, DroppedRecords, UpdatedAt FROM agent_metrics WHERE AgentID = ?", id)

	err := row.Scan(&pipelineMetrics.AgentID, &pipelineMetrics.Status, &pipelineMetrics.ExportedDataVolume, &pipelineMetrics.UptimeSeconds, &pipelineMetrics.DroppedRecords, &pipelineMetrics.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("no metrics collected yet")
		}
		return nil, err
	}

	return pipelineMetrics, nil
}
