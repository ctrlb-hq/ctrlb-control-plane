package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func dbCreator() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./backend.db")
	if err != nil {
		return nil, err
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		log.Printf("Error enabling foreign keys: %s", err)
		return nil, err
	}

	// Create necessary tables
	if err := createUserTable(db); err != nil {
		return nil, err
	}
	if err := createConfigTable(db); err != nil {
		return nil, err
	}
	if err := createAgentTable(db); err != nil {
		return nil, err
	}
	if err := createAgentMetricsTable(db); err != nil {
		return nil, err
	}
	if err := createAgentStatusTable(db); err != nil {
		return nil, err
	}
	if err := createNewConfigTables(db); err != nil {
		return nil, err
	}

	log.Println("Database and tables created/verified successfully")
	return db, nil
}

func createUserTable(db *sql.DB) error {
	// Create user table
	createUserTableSQL := `
    CREATE TABLE IF NOT EXISTS user (
        "Email" TEXT PRIMARY KEY,  -- Unique email address
        "Name" TEXT,
        "Password" TEXT
    );
    `
	_, err := db.Exec(createUserTableSQL)
	if err != nil {
		log.Printf("Error creating User table: %s", err)
		return err
	}
	return nil
}

func createConfigTable(db *sql.DB) error {
	// Create config table
	createConfigTableSQL := `
    CREATE TABLE IF NOT EXISTS config (
        "ID" TEXT PRIMARY KEY,
        "Name" TEXT,
        "Description" TEXT,
        "Config" TEXT,
        "TargetAgent" TEXT,
        "CreatedAt" DATETIME,
        "UpdatedAt" DATETIME
    );
    `
	_, err := db.Exec(createConfigTableSQL)
	if err != nil {
		log.Printf("Error creating Config table: %s", err)
		return err
	}
	return nil
}

func createAgentTable(db *sql.DB) error {
	// Create agents table with foreign key to config.ID
	createAgentTableSQL := `
    CREATE TABLE IF NOT EXISTS agents (
        "ID" TEXT PRIMARY KEY,
        "Name" TEXT,
        "Type" TEXT,
        "Version" TEXT,
        "Hostname" TEXT,
        "Platform" TEXT,
        "ConfigID" TEXT,
        "IsPipeline" BOOL,
        "RegisteredAt" DATETIME,
        FOREIGN KEY ("ConfigID") REFERENCES config("ID") ON DELETE SET NULL
    );
    `
	_, err := db.Exec(createAgentTableSQL)
	if err != nil {
		log.Printf("Error creating Agent table: %s", err)
		return err
	}
	return nil
}

func createAgentMetricsTable(db *sql.DB) error {
	// Create agent_metrics table with a foreign key referencing agents(ID)
	createAgentMetricsTableSQL := `
    CREATE TABLE IF NOT EXISTS agent_metrics (
        "AgentID" TEXT PRIMARY KEY,
        "Status" TEXT,
        "ExportedDataVolume" INTEGER,
        "UptimeSeconds" INTEGER,
        "DroppedRecords" INTEGER,
        "UpdatedAt" DATETIME,
        FOREIGN KEY ("AgentID") REFERENCES agents("ID") ON DELETE CASCADE
    );
    `
	_, err := db.Exec(createAgentMetricsTableSQL)
	if err != nil {
		log.Printf("Error creating AgentMetrics table: %s", err)
		return err
	}
	return nil
}

func createAgentStatusTable(db *sql.DB) error {
	// Create agent_status table with a foreign key referencing agents(ID)
	createAgentStatusTableSQL := `
    CREATE TABLE IF NOT EXISTS agent_status (
        "AgentID" TEXT PRIMARY KEY,
        "Hostname" TEXT,
        "CurrentStatus" TEXT,
        "RetryRemaining" INTEGER,
        "UpdatedAt" DATETIME,
        FOREIGN KEY ("AgentID") REFERENCES agents("ID") ON DELETE CASCADE
    );
    `
	_, err := db.Exec(createAgentStatusTableSQL)
	if err != nil {
		log.Printf("Error creating AgentStatus table: %s", err)
		return err
	}
	return nil
}

// If you want to create the new config tables with foreign keys as well:
func createNewConfigTables(db *sql.DB) error {
	// Enable foreign key enforcement again (harmless if already on, safe to re-run)
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		log.Printf("Error enabling foreign keys: %s", err)
		return err
	}

	statements := []string{
		`CREATE TABLE IF NOT EXISTS config_sets (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            version TEXT,
            log_level TEXT,
            credentials TEXT,
            created_at TEXT DEFAULT (datetime('now')),
            updated_at TEXT DEFAULT (datetime('now'))
        )`,
		`CREATE TABLE IF NOT EXISTS extensions (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            config_set_id INTEGER NOT NULL,
            extension_name TEXT,
            enabled BOOLEAN,
            endpoint TEXT,
            extra TEXT,
            created_at TEXT DEFAULT (datetime('now')),
            updated_at TEXT DEFAULT (datetime('now')),
            FOREIGN KEY (config_set_id) REFERENCES config_sets(id) ON DELETE CASCADE
        )`,
		`CREATE TABLE IF NOT EXISTS pipelines (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            config_set_id INTEGER NOT NULL,
            name TEXT NOT NULL,
            created_at TEXT DEFAULT (datetime('now')),
            updated_at TEXT DEFAULT (datetime('now')),
            FOREIGN KEY (config_set_id) REFERENCES config_sets(id) ON DELETE CASCADE
        )`,
		`CREATE TABLE IF NOT EXISTS pipeline_components (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            pipeline_id INTEGER NOT NULL,
            section TEXT CHECK (section IN ('source','processor','destination')),
            type TEXT,
            name TEXT,
            config TEXT,
            created_at TEXT DEFAULT (datetime('now')),
            updated_at TEXT DEFAULT (datetime('now')),
            FOREIGN KEY (pipeline_id) REFERENCES pipelines(id) ON DELETE CASCADE
        )`,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			log.Printf("Error executing statement: %s\n%s", err, stmt)
			return err
		}
	}

	fmt.Println("New schema created successfully.")
	return nil
}
