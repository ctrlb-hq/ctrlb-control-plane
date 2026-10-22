package database_test

import (
	"database/sql"
	"path/filepath"
	"testing"

	database "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/db"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
	"github.com/stretchr/testify/require"
)

func TestKubernetesMigrationPreservesLegacyRowsAndSurvivesReopen(t *testing.T) {
	utils.InitLogger()
	path := filepath.Join(t.TempDir(), "legacy.db")
	legacy, err := sql.Open("sqlite3", path)
	require.NoError(t, err)
	_, err = legacy.Exec(`CREATE TABLE agents (
		id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, type TEXT, version TEXT,
		hostname TEXT, platform TEXT, ip TEXT, pipeline_id INTEGER, pipeline_name TEXT,
		registered_at INTEGER
	); INSERT INTO agents (name, hostname) VALUES ('existing-vm', 'vm-1');`)
	require.NoError(t, err)
	require.NoError(t, legacy.Close())
	for count := 0; count < 2; count++ {
		db, err := database.DBInit(path)
		require.NoError(t, err)
		var name, mode string
		require.NoError(t, db.QueryRow("SELECT name, management_mode FROM agents WHERE id = 1").Scan(&name, &mode))
		require.Equal(t, "existing-vm", name)
		require.Equal(t, "push", mode)
		require.True(t, tableExists(t, db, "kubernetes_installations"))
		require.True(t, tableExists(t, db, "kubernetes_collectors"))
		// Force connection replacement. DSN options must apply to its successor.
		db.SetMaxIdleConns(0)
		var enabled int
		require.NoError(t, db.QueryRow("PRAGMA foreign_keys").Scan(&enabled))
		require.Equal(t, 1, enabled)
		require.NoError(t, db.Close())
	}
}
