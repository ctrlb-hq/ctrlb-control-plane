package frontendagent

import (
	"database/sql"
)

type FrontendAgentRepository struct {
	db *sql.DB
}

// NewFrontendAgentRepository creates a new FrontendAgentRepository
func NewFrontendAgentRepository(db *sql.DB) *FrontendAgentRepository {
	return &FrontendAgentRepository{db: db}
}
