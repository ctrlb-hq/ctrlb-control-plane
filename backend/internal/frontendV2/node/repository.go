package frontendnode

import (
	"database/sql"
)

type FrontendNodeRepository struct {
	db *sql.DB
}

func NewFrontendNodeRepository(db *sql.DB) *FrontendNodeRepository {
	return &FrontendNodeRepository{
		db: db,
	}
}

func (f *FrontendNodeRepository) GetAllReceivers() (any, error) {
	return nil, nil
}

func (f *FrontendNodeRepository) GetAllProcessors() (any, error) {
	return nil, nil
}

func (f *FrontendNodeRepository) GetAllExporters() (any, error) {
	return nil, nil
}
