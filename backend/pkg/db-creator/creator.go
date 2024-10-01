package dbcreator

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func DBCreator() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./all-father.db")
	if err != nil {
		return nil, err
	}

	// Create necessary tables
	err = createAgentTable(db)
	if err != nil {
		return nil, err
	}

	err = createAgentInfoTable(db)
	if err != nil {
		return nil, err
	}

	log.Println("Database and tables created/verified successfully")

	// Return the open *sql.DB connection
	return db, nil
}

func createAgentTable(db *sql.DB) error {
	createAgentTableSQL := `CREATE TABLE IF NOT EXISTS agents (
		"ID" TEXT PRIMARY KEY,
		"Name" TEXT,
		"Type" TEXT,
		"Version" TEXT,
		"Hostname" TEXT,
		"Platform" TEXT,
		"Config" TEXT
	);`
	_, err := db.Exec(createAgentTableSQL)
	if err != nil {
		log.Printf("Error creating Agent table: %s", err)
		return err
	}
	log.Println("Agent table created or already exists")
	return nil
}

func createAgentInfoTable(db *sql.DB) error {
	createAgentInfoTableSQL := `CREATE TABLE IF NOT EXISTS agent_info (
    	"AgentID" TEXT PRIMARY KEY,
    	"Status" TEXT,
    	"ExportedDataVolume" INTEGER,
    	"Uptime" INTEGER,
    	"DroppedRecords" INTEGER
	);`
	_, err := db.Exec(createAgentInfoTableSQL)
	if err != nil {
		log.Printf("Error creating AgentInfo table: %s", err)
		return err
	}
	log.Println("AgentInfo table created or already exists")
	return nil
}
