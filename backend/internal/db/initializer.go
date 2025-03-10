package database

import (
	"database/sql"
)

func InitializeDB() (*sql.DB, error) {
	db, err := DBInit()
	if err != nil {
		return nil, err
	}
	return db, nil
}
