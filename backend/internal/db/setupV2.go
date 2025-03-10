package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// DBInit initializes the DB and creates all tables.
func DBInit() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./backend.db")
	if err != nil {
		return nil, err
	}

	// Enforce foreign keys in SQLite
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		log.Printf("Error enabling foreign keys: %v", err)
		return nil, err
	}

	// Create tables
	if err := createUserTable(db); err != nil {
		return nil, err
	}
	if err := createAgentsTable(db); err != nil {
		return nil, err
	}
	if err := createAgentMetricsTable(db); err != nil {
		return nil, err
	}
	if err := createExtensionsTable(db); err != nil {
		return nil, err
	}
	if err := createPipelinesTable(db); err != nil {
		return nil, err
	}
	if err := createPipelineComponentsTable(db); err != nil {
		return nil, err
	}

	log.Println("All tables created (or verified) successfully.")
	return db, nil
}

func createUserTable(db *sql.DB) error {
	createUserTableSQL := `
CREATE TABLE IF NOT EXISTS user (
    "email" TEXT PRIMARY KEY,
    "name" TEXT NOT NULL,
    "password" TEXT NOT NULL,
    "role" TEXT NOT NULL
);`
	_, err := db.Exec(createUserTableSQL)
	if err != nil {
		log.Printf("Error creating User table: %s", err)
		return err
	}
	return nil
}

// Note the removed trailing comma after 'registered_at'.
func createAgentsTable(db *sql.DB) error {
	query := `
CREATE TABLE IF NOT EXISTS agents (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT,
    type TEXT,
    version TEXT,
    hostname TEXT,
    platform TEXT,
    registered_at INTEGER DEFAULT (strftime('%s', 'now')) -- Stores Unix timestamp
);`
	_, err := db.Exec(query)
	if err != nil {
		log.Printf("Error creating agents table: %v", err)
	}
	return err
}

func createAgentMetricsTable(db *sql.DB) error {
	query := `
CREATE TABLE IF NOT EXISTS agent_metrics (
    agent_id INTEGER PRIMARY KEY,
    incoming_bytes INTEGER DEFAULT 0,
    outgoing_bytes INTEGER DEFAULT 0,
    uptime_seconds INTEGER DEFAULT 0,
    dropped_records INTEGER DEFAULT 0,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
);`
	_, err := db.Exec(query)
	if err != nil {
		log.Printf("Error creating agent_metrics table: %v", err)
	}
	return err
}

func createExtensionsTable(db *sql.DB) error {
	query := `
CREATE TABLE IF NOT EXISTS extensions (
    extension_id INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_id INTEGER NOT NULL,
    extension_name TEXT CHECK (
        extension_name IN (
            'health_check',
            'pprof',
            'zpages'
        )
    ) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT 0,
    config TEXT, -- JSON or other config data specific to the extension
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
);`
	_, err := db.Exec(query)
	if err != nil {
		log.Printf("Error creating extensions table: %v", err)
	}
	return err
}

// Added a comma after the CHECK constraint for 'type'.
func createPipelinesTable(db *sql.DB) error {
	query := `
CREATE TABLE IF NOT EXISTS pipelines (
    pipeline_id INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_id INTEGER NOT NULL,
    type TEXT CHECK (
        type IN ('metrics','traces','logs')
    ),
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
);`
	_, err := db.Exec(query)
	if err != nil {
		log.Printf("Error creating pipelines table: %v", err)
	}
	return err
}

func createPipelineComponentsTable(db *sql.DB) error {
	query := `
CREATE TABLE IF NOT EXISTS pipeline_components (
    component_id INTEGER PRIMARY KEY AUTOINCREMENT,
    pipeline_id INTEGER NOT NULL,
    component_role TEXT CHECK (
        component_role IN ('receiver','processor','exporter')
    ) NOT NULL,
    plugin_name TEXT NOT NULL,
    name TEXT,
    config TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (pipeline_id) REFERENCES pipelines(pipeline_id) ON DELETE CASCADE
);`
	_, err := db.Exec(query)
	if err != nil {
		log.Printf("Error creating pipeline_components table: %v", err)
	}
	return err
}
