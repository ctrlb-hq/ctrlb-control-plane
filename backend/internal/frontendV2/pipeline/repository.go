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
