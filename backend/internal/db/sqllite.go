package db

import (
	"database/sql"
)

func InitializeDB(ip string, port int64, dbName string) (*sql.DB, error) {
	return sql.Open("sqlite3", "sqlite.db")
}
