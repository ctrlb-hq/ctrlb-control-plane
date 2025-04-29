package database_test

import (
	"database/sql"
	"testing"

	database "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/db"
	_ "github.com/mattn/go-sqlite3"
)

func TestDBInitWithPath_CreatesAllTables(t *testing.T) {
	db, err := database.DBInit(":memory:") // <-- use in-memory DB
	if err != nil {
		t.Fatalf("DBInitWithPath failed: %v", err)
	}
	defer db.Close()

	expectedTables := []string{
		"user",
		"agents",
		"agents_labels",
		"aggregated_agent_metrics",
		"realtime_agent_metrics",
		"extensions",
		"pipelines",
		"pipeline_components",
		"pipeline_component_edges",
		"component_schemas",
	}

	for _, table := range expectedTables {
		if !tableExists(t, db, table) {
			t.Errorf("expected table %s to exist", table)
		}
	}
}

// Helper to check if a table exists in SQLite
func tableExists(t *testing.T, db *sql.DB, tableName string) bool {
	query := `SELECT name FROM sqlite_master WHERE type='table' AND name=?`
	var name string
	err := db.QueryRow(query, tableName).Scan(&name)
	if err == sql.ErrNoRows {
		return false
	}
	if err != nil {
		t.Fatalf("error querying sqlite_master for table %s: %v", tableName, err)
	}
	return true
}
