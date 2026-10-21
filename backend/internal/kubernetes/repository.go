package kubernetes

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Create(ctx context.Context, installation *Installation) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO kubernetes_installations
		(id, name, cluster_name, pipeline_id, token_hash, created_by, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, installation.ID, installation.Name, installation.ClusterName,
		installation.PipelineID, installation.tokenHash, installation.CreatedBy, installation.CreatedAt)
	return err
}

func (r *Repository) PipelineConfig(ctx context.Context, pipelineID int) ([]byte, error) {
	var config []byte
	err := r.db.QueryRowContext(ctx, "SELECT config_json FROM pipelines WHERE pipeline_id = ?", pipelineID).Scan(&config)
	return config, err
}

func (r *Repository) Get(ctx context.Context, id string) (*Installation, error) {
	i := &Installation{}
	err := r.db.QueryRowContext(ctx, `SELECT i.id, i.name, i.cluster_name, i.pipeline_id,
		i.token_hash, i.created_by, i.created_at, p.config_json
		FROM kubernetes_installations i JOIN pipelines p ON p.pipeline_id = i.pipeline_id
		WHERE i.id = ?`, id).Scan(&i.ID, &i.Name, &i.ClusterName, &i.PipelineID,
		&i.tokenHash, &i.CreatedBy, &i.CreatedAt, &i.configJSON)
	return i, err
}

func (r *Repository) ListIDs(ctx context.Context, pipelineID *int) ([]string, error) {
	query := "SELECT id FROM kubernetes_installations"
	var args []any
	if pipelineID != nil {
		query += " WHERE pipeline_id = ?"
		args = append(args, *pipelineID)
	}
	query += " ORDER BY created_at, id"
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *Repository) CheckIn(ctx context.Context, installation *Installation, req CheckInRequest, now int64) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	var id int64
	err = tx.QueryRowContext(ctx, "SELECT agent_id FROM kubernetes_collectors WHERE installation_id = ? AND node_name = ?",
		installation.ID, req.NodeName).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		result, insertErr := tx.ExecContext(ctx, `INSERT INTO agents
			(name, type, version, hostname, platform, ip, registered_at, pipeline_id, pipeline_name, management_mode)
			VALUES (?, 'OTEL', ?, ?, ?, ?, ?, ?, (SELECT name FROM pipelines WHERE pipeline_id = ?), 'pull')`,
			fmt.Sprintf("%s/%s", installation.Name, req.NodeName), req.Version, req.NodeName,
			req.Platform, req.PodIP, now, installation.PipelineID, installation.PipelineID)
		if insertErr != nil {
			return 0, insertErr
		}
		id, err = result.LastInsertId()
		if err != nil {
			return 0, err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO kubernetes_collectors
			(agent_id, installation_id, node_name, last_seen, engine_status) VALUES (?, ?, ?, ?, ?)`,
			id, installation.ID, req.NodeName, now, req.EngineStatus)
	}
	if err != nil {
		return 0, err
	}
	_, err = tx.ExecContext(ctx, `UPDATE agents SET version = ?, hostname = ?, platform = ?, ip = ?,
		pipeline_id = ?, pipeline_name = (SELECT name FROM pipelines WHERE pipeline_id = ?) WHERE id = ?`,
		req.Version, req.NodeName, req.Platform, req.PodIP, installation.PipelineID, installation.PipelineID, id)
	if err != nil {
		return 0, err
	}
	var metricsJSON any
	var metricsUpdatedAt any
	if req.Metrics != nil {
		data, err := json.Marshal(req.Metrics)
		if err != nil {
			return 0, err
		}
		metricsJSON, metricsUpdatedAt = string(data), now
	}
	_, err = tx.ExecContext(ctx, `UPDATE kubernetes_collectors SET
		node_uid = ?, namespace = ?, pod_name = ?, pod_uid = ?, pod_ip = ?, last_seen = ?, engine_status = ?,
		applied_revision = ?, attempted_revision = ?, config_error = ?,
		metrics_json = COALESCE(?, metrics_json), metrics_updated_at = COALESCE(?, metrics_updated_at)
		WHERE agent_id = ?`, req.NodeUID, req.Namespace, req.PodName, req.PodUID, req.PodIP,
		now, req.EngineStatus, req.AppliedRevision, req.AttemptedRevision, req.ConfigError,
		metricsJSON, metricsUpdatedAt, id)
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *Repository) Collectors(ctx context.Context, installationID string) ([]Collector, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT c.agent_id, c.installation_id, c.node_name, c.node_uid,
		c.namespace, c.pod_name, c.pod_uid, c.pod_ip, a.version, a.platform,
		c.last_seen, c.engine_status, c.applied_revision, c.attempted_revision, c.config_error,
		c.metrics_json, c.metrics_updated_at
		FROM kubernetes_collectors c JOIN agents a ON a.id = c.agent_id
		WHERE c.installation_id = ? ORDER BY c.node_name`, installationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	collectors := []Collector{}
	for rows.Next() {
		var c Collector
		var metrics sql.NullString
		var metricsAt sql.NullInt64
		if err := rows.Scan(&c.ID, &c.InstallationID, &c.NodeName, &c.NodeUID, &c.Namespace,
			&c.PodName, &c.PodUID, &c.PodIP, &c.Version, &c.Platform, &c.LastSeen, &c.EngineStatus,
			&c.AppliedRevision, &c.AttemptedRevision, &c.ConfigError, &metrics, &metricsAt); err != nil {
			return nil, err
		}
		if metrics.Valid {
			if err := json.Unmarshal([]byte(metrics.String), &c.Metrics); err != nil {
				return nil, fmt.Errorf("decode collector metrics: %w", err)
			}
		}
		if metricsAt.Valid {
			c.MetricsUpdatedAt = &metricsAt.Int64
		}
		collectors = append(collectors, c)
	}
	return collectors, rows.Err()
}
