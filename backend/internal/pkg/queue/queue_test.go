package queue

import (
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

const mockMetrics = `
# HELP otelcol_exporter_sent_log_records Total logs sent
# TYPE otelcol_exporter_sent_log_records counter
otelcol_exporter_sent_log_records 50

# HELP otelcol_exporter_sent_spans Total spans sent
# TYPE otelcol_exporter_sent_spans counter
otelcol_exporter_sent_spans 100

# HELP otelcol_exporter_sent_bytes Total bytes sent
# TYPE otelcol_exporter_sent_bytes counter
otelcol_exporter_sent_bytes 2048

# HELP otelcol_receiver_accepted_bytes Total bytes received
# TYPE otelcol_receiver_accepted_bytes counter
otelcol_receiver_accepted_bytes 1024

# HELP otelcol_exporter_sent_metric_points Total metrics sent
# TYPE otelcol_exporter_sent_metric_points counter
otelcol_exporter_sent_metric_points 75

# HELP otelcol_process_cpu_seconds_total CPU usage
# TYPE otelcol_process_cpu_seconds_total counter
otelcol_process_cpu_seconds_total 1.5

# HELP otelcol_process_memory_rss Memory usage
# TYPE otelcol_process_memory_rss gauge
otelcol_process_memory_rss 512000
`

func startFixedPortServer(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:8888")
	assert.NoError(t, err)

	go func() {
		err := http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, mockMetrics)
		}))
		if err != nil {
			t.Logf("Server stopped: %v", err)
		}
	}()
}

func TestWorkerProcessFixedPort(t *testing.T) {
	startFixedPortServer(t)

	db, err := sql.Open("sqlite3", ":memory:")
	assert.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE aggregated_agent_metrics (
			agent_id TEXT PRIMARY KEY,
			logs_rate_sent REAL,
			traces_rate_sent REAL,
			metrics_rate_sent REAL,
			data_sent_bytes REAL,
			data_received_bytes REAL,
			status TEXT,
			updated_at INTEGER
		);
		CREATE TABLE realtime_agent_metrics (
			agent_id TEXT,
			logs_rate_sent REAL,
			traces_rate_sent REAL,
			metrics_rate_sent REAL,
			data_sent_bytes REAL,
			data_received_bytes REAL,
			cpu_utilization REAL,
			memory_utilization REAL,
			timestamp INTEGER
		);
	`)
	assert.NoError(t, err)

	q := NewQueue(1, 1, db)

	err = q.AddAgent("agent-2", "127.0.0.1", "127.0.0.1")
	assert.NoError(t, err)

	q.checkAllAgents()
	time.Sleep(1 * time.Second) // give the worker a chance to run

	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM aggregated_agent_metrics WHERE agent_id = ?`, "agent-2").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}
