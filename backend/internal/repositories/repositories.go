package repositories

import "database/sql"

type AgentRepository struct {
	db *sql.DB
}
type AuthRepository struct {
	db *sql.DB
}
