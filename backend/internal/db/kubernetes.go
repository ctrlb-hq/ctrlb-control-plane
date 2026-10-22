package database

import (
	"database/sql"
	"fmt"
)

// migrateKubernetes adds collector management without changing existing VM rows.
func migrateKubernetes(db *sql.DB) error {
	rows, err := db.Query("PRAGMA table_info(agents)")
	if err != nil {
		return err
	}
	hasMode := false
	for rows.Next() {
		var cid int
		var name, typ string
		var notNull, primaryKey int
		var defaultValue any
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &primaryKey); err != nil {
			rows.Close()
			return err
		}
		hasMode = hasMode || name == "management_mode"
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	if !hasMode {
		if _, err := db.Exec("ALTER TABLE agents ADD COLUMN management_mode TEXT NOT NULL DEFAULT 'push'"); err != nil {
			return fmt.Errorf("add agent management mode: %w", err)
		}
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS kubernetes_installations (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			cluster_name TEXT NOT NULL,
			pipeline_id INTEGER NOT NULL REFERENCES pipelines(pipeline_id) ON DELETE RESTRICT,
			token_hash TEXT NOT NULL,
			created_by TEXT NOT NULL,
			created_at INTEGER NOT NULL
		);
		CREATE INDEX IF NOT EXISTS kubernetes_installations_pipeline ON kubernetes_installations(pipeline_id);
		CREATE TABLE IF NOT EXISTS kubernetes_collectors (
			agent_id INTEGER PRIMARY KEY REFERENCES agents(id) ON DELETE CASCADE,
			installation_id TEXT NOT NULL REFERENCES kubernetes_installations(id) ON DELETE CASCADE,
			node_name TEXT NOT NULL,
			node_uid TEXT NOT NULL DEFAULT '',
			namespace TEXT NOT NULL DEFAULT '',
			pod_name TEXT NOT NULL DEFAULT '',
			pod_uid TEXT NOT NULL DEFAULT '',
			pod_ip TEXT NOT NULL DEFAULT '',
			last_seen INTEGER NOT NULL,
			engine_status TEXT NOT NULL,
			applied_revision TEXT NOT NULL DEFAULT '',
			attempted_revision TEXT NOT NULL DEFAULT '',
			config_error TEXT NOT NULL DEFAULT '',
			metrics_json TEXT,
			metrics_updated_at INTEGER,
			UNIQUE(installation_id, node_name)
		);
	`)
	return err
}
