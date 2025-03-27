package database

import (
	"database/sql"
	"fmt"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
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
		utils.Logger.Error(fmt.Sprintf("Error enabling foreign keys: %v", err))
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
	if err := createPipelineComponentDependenciesTable(db); err != nil {
		return nil, err
	}
	if err := createComponentSchemasTable(db); err != nil {
		return nil, err
	}

	utils.Logger.Info("All tables created (or verified) successfully.")
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
		utils.Logger.Error(fmt.Sprintf("Error creating User table: %s", err))
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
		ip TEXT,
        pipeline_id INTEGER DEFAULT NULL,
        pipeline_name TEXT DEFAULT NULL,
        registered_at INTEGER DEFAULT (strftime('%s', 'now')), -- Stores Unix timestamp
        FOREIGN KEY (pipeline_id) REFERENCES pipelines(pipeline_id) ON DELETE SET NULL
    );
    `
	_, err := db.Exec(query)
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("Error creating agents table: %v", err))
	}
	return err
}

// Agents labels table with a Unix timestamp for created_at
func createAgentsLabelsTable(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS agents_labels (
		agent_id INTEGER NOT NULL,
		key TEXT NOT NULL,
		value TEXT NOT NULL,
		created_at INTEGER DEFAULT (strftime('%s', 'now')),
		PRIMARY KEY (agent_id, key),
		FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
	);
    `
	_, err := db.Exec(query)
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("Error creating agents_labels table: %v", err))
	}
	return err
}

// Aggregated agent metrics table with a Unix timestamp for updated_at
func createAggregatedAgentMetricsTable(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS aggregated_agent_metrics (
		agent_id INTEGER PRIMARY KEY,
		logs_rate_sent INTEGER DEFAULT 0,
		traces_rate_sent INTEGER DEFAULT 0,
		metrics_rate_sent INTEGER DEFAULT 0,
		data_sent_bytes INTEGER DEFAULT 0, -- Total bytes sent
		data_received_bytes INTEGER DEFAULT 0, -- Total bytes received
		status TEXT CHECK(status IN ('connected', 'disconnected', 'stopped')),
		updated_at INTEGER DEFAULT (strftime('%s', 'now')), -- Unix timestamp
		FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
	);
    `
	_, err := db.Exec(query)
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("Error creating aggregated_agent_metrics table: %v", err))
	}
	return err
}

// Realtime agent metrics table with a Unix timestamp for updated_at
func createRealtimeAgentMetricsTable(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS realtime_agent_metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		agent_id INTEGER NOT NULL,
		logs_rate_sent INTEGER DEFAULT 0,
		traces_rate_sent INTEGER DEFAULT 0,
		metrics_rate_sent INTEGER DEFAULT 0,
		data_sent_bytes INTEGER DEFAULT 0,
		data_received_bytes INTEGER DEFAULT 0,
		cpu_utilization REAL DEFAULT 0,
		memory_utilization REAL DEFAULT 0,
		timestamp INTEGER DEFAULT (strftime('%s', 'now')), -- Unix timestamp
		FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
	);    
	`
	_, err := db.Exec(query)
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("Error creating realtime_agent_metrics table: %v", err))
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
		utils.Logger.Error(fmt.Sprintf("Error creating extensions table: %v", err))
	}
	return err
}

// Pipelines table with a Unix timestamp for created_at
func createPipelinesTable(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS pipelines (
        pipeline_id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        created_by TEXT NOT NULL,
        created_at INTEGER DEFAULT (strftime('%s', 'now')),
        updated_at INTEGER DEFAULT (strftime('%s', 'now'))
    );
    `
	_, err := db.Exec(query)
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("Error creating pipelines table: %v", err))
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
		utils.Logger.Error(fmt.Sprintf("Error creating pipeline_components table: %v", err))
	}
	return err
}

func createPipelineComponentDependenciesTable(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS pipeline_component_edges (
        edge_id INTEGER PRIMARY KEY AUTOINCREMENT,
        pipeline_id INTEGER NOT NULL,
        child_component_id INTEGER NOT NULL,  -- Component that depends on another (runs later)
        parent_component_id INTEGER NOT NULL, -- Component that must execute first
        created_at INTEGER DEFAULT (strftime('%s', 'now')), -- Unix timestamp
        FOREIGN KEY (pipeline_id) REFERENCES pipelines(pipeline_id) ON DELETE CASCADE,
        FOREIGN KEY (child_component_id) REFERENCES pipeline_components(component_id) ON DELETE CASCADE,
        FOREIGN KEY (parent_component_id) REFERENCES pipeline_components(component_id) ON DELETE CASCADE,
        UNIQUE (pipeline_id, child_component_id, parent_component_id) -- Prevent duplicate dependencies
    );
    `
	_, err := db.Exec(query)
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("Error creating pipeline_component_edges table: %v", err))
	}
	return err
}

func createComponentSchemasTable(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS component_schemas (
        name TEXT PRIMARY KEY,               				-- Internal unique ID (e.g., otlp_grpc)
        type TEXT NOT NULL,                  				-- receiver, exporter, processor, etc.
        display_name TEXT NOT NULL,          				-- Friendly name for UI
        supported_signals TEXT NOT NULL,     				-- Comma-separated: traces,metrics,logs
        schema_json TEXT NOT NULL,           				-- Full JSON schema
        created_at INTEGER DEFAULT (strftime('%s', 'now')) 	-- Unix timestamp
    );
    `
	_, err := db.Exec(query)
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("Error creating component_schemas table: %v", err))
	}
	return err
}
