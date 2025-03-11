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
	if err := createAgentsLabelsTable(db); err != nil {
		return nil, err
	}
	if err := createAggregatedAgentMetricsTable(db); err != nil {
		return nil, err
	}
	if err := createRealtimeAgentMetricsTable(db); err != nil {
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
        email TEXT PRIMARY KEY,
        name TEXT NOT NULL,
        password TEXT NOT NULL,
        role TEXT NOT NULL
    );`
	_, err := db.Exec(createUserTableSQL)
	if err != nil {
		log.Printf("Error creating User table: %s", err)
		return err
	}
	return nil
}

// Agents table with a Unix timestamp for registered_at
func createAgentsTable(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS agents (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT,
        type TEXT,
        version TEXT,
        hostname TEXT,
        platform TEXT,
        pipeline_id INTEGER,
        pipeline_name TEXT,
        registered_at INTEGER DEFAULT (strftime('%s', 'now')), -- Stores Unix timestamp
        FOREIGN KEY (pipeline_id) REFERENCES pipelines(pipeline_id) ON DELETE SET NULL
    );
    `
	_, err := db.Exec(query)
	if err != nil {
		log.Printf("Error creating agents table: %v", err)
	}
	return err
}

// Agents labels table with a Unix timestamp for created_at
func createAgentsLabelsTable(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS agents_labels (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        agent_id INTEGER NOT NULL,
        key TEXT NOT NULL,
        value TEXT NOT NULL,
        created_at INTEGER DEFAULT (strftime('%s', 'now')), -- Unix timestamp
        FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
    );
    `
	_, err := db.Exec(query)
	if err != nil {
		log.Printf("Error creating agents_labels table: %v", err)
	}
	return err
}

// Aggregated agent metrics table with a Unix timestamp for updated_at
func createAggregatedAgentMetricsTable(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS aggregated_agent_metrics (
        agent_id INTEGER PRIMARY KEY,
        log_rate_sent INTEGER DEFAULT 0,
        traces_rate_sent INTEGER DEFAULT 0,
        metrics_rate_sent INTEGER DEFAULT 0,
        status TEXT CHECK(status IN ('connected', 'disconnected', 'stopped')),
        updated_at INTEGER DEFAULT (strftime('%s', 'now')), -- Unix timestamp
        FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
    );
    `
	_, err := db.Exec(query)
	if err != nil {
		log.Printf("Error creating aggregated_agent_metrics table: %v", err)
	}
	return err
}

// Realtime agent metrics table with a Unix timestamp for updated_at
func createRealtimeAgentMetricsTable(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS realtime_agent_metrics (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        agent_id INTEGER NOT NULL,
        log_rate_sent INTEGER DEFAULT 0,
        traces_rate_sent INTEGER DEFAULT 0,
        metrics_rate_sent INTEGER DEFAULT 0,
        cpu_utilization REAL,
        memory_utilization REAL,
        updated_at INTEGER DEFAULT (strftime('%s', 'now')), -- Unix timestamp
        FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
    );
    `
	_, err := db.Exec(query)
	if err != nil {
		log.Printf("Error creating realtime_agent_metrics table: %v", err)
	}
	return err
}

// Extensions table with Unix timestamps for created_at and updated_at
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
        created_at INTEGER DEFAULT (strftime('%s', 'now')), -- Unix timestamp
        updated_at INTEGER DEFAULT (strftime('%s', 'now')), -- Unix timestamp
        FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
    );
    `
	_, err := db.Exec(query)
	if err != nil {
		log.Printf("Error creating extensions table: %v", err)
	}
	return err
}

// Pipelines table with a Unix timestamp for created_at
func createPipelinesTable(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS pipelines (
        pipeline_id INTEGER PRIMARY KEY AUTOINCREMENT,
        type TEXT CHECK (
            type IN ('metrics','traces','logs')
        ),
        name TEXT NOT NULL,
        description TEXT,
        created_at INTEGER DEFAULT (strftime('%s', 'now')) -- Unix timestamp
    );
    `
	_, err := db.Exec(query)
	if err != nil {
		log.Printf("Error creating pipelines table: %v", err)
	}
	return err
}

// Pipeline components table with a Unix timestamp for created_at
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
        created_at INTEGER DEFAULT (strftime('%s', 'now')), -- Unix timestamp
        FOREIGN KEY (pipeline_id) REFERENCES pipelines(pipeline_id) ON DELETE CASCADE
    );
    `
	_, err := db.Exec(query)
	if err != nil {
		log.Printf("Error creating pipeline_components table: %v", err)
	}
	return err
}
