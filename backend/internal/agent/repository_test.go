package agent

import (
	"database/sql"
	"testing"
	"time"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/constants"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	assert.NoError(t, err)

	_, err = db.Exec(`
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
			registered_at INTEGER DEFAULT (strftime('%s', 'now'))
		);
	`)
	assert.NoError(t, err)

	// Optional: mock pipelines table if you want to test FK behavior
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS pipelines (
		pipeline_id INTEGER PRIMARY KEY
	);`)
	assert.NoError(t, err)

	return db
}

func TestAgentExists(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAgentRepository(db)

	// Insert a test agent
	_, err := db.Exec(`INSERT INTO agents (name, type, version, hostname, platform, registered_at, ip) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"TestAgent", "otel", "v1.0", "agent.local", "linux", time.Now().Unix(), "127.0.0.1")
	assert.NoError(t, err)

	// Check existence
	exists, err := repo.AgentExists("agent.local")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Check non-existence
	exists, err = repo.AgentExists("not-found.local")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestRegisterAgent(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAgentRepository(db)

	req := &models.AgentRegisterRequest{
		Name:         "NewAgent",
		Type:         "otel",
		Version:      "v1.1",
		Hostname:     "new-agent.local",
		Platform:     "linux",
		RegisteredAt: time.Now().Unix(),
		IP:           "192.168.1.10",
	}

	resp, err := repo.RegisterAgent(req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Greater(t, resp.ID, int64(0))

	// Validate config
	cfg := resp.Config
	assert.Contains(t, cfg, "receivers")
	assert.Contains(t, cfg, "processors")
	assert.Contains(t, cfg, "exporters")
	assert.Contains(t, cfg, "service")

	service := cfg["service"].(map[string]any)
	assert.Equal(t, constants.TelemetryService, service["telemetry"])
	assert.Contains(t, service["pipelines"], "logs/default")
}
