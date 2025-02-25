package frontendconfigV2

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
)

// FrontendConfigRepository provides methods to interact with the config database table
type FrontendConfigRepository struct {
	db *sql.DB
}

// NewFrontendConfigRepository initializes FrontendConfigRepository
func NewFrontendConfigRepository(db *sql.DB) *FrontendConfigRepository {
	return &FrontendConfigRepository{db: db}
}

// GetAllConfigs retrieves all configurations from the database
func (f *FrontendConfigRepository) GetAllConfigs(ctx context.Context) ([]models.ConfigSet, error) {
	query := `
		SELECT id, version, credentials, created_at, updated_at 
		FROM config_sets
	`

	// Execute the query with context
	rows, err := f.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch config sets: %w", err)
	}
	defer rows.Close()

	var configSets []models.ConfigSet

	for rows.Next() {
		var configSet models.ConfigSet
		var credentialsJSON string
		var createdAt, updatedAt sql.NullTime

		// Scan values from the row
		if err := rows.Scan(
			&configSet.ID, &configSet.Version, &credentialsJSON,
			&createdAt, &updatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Parse credentials JSON if it's not empty
		if credentialsJSON != "" {
			if err := json.Unmarshal([]byte(credentialsJSON), &configSet.Credentials); err != nil {
				return nil, fmt.Errorf("failed to parse credentials JSON: %w", err)
			}
		}

		// Assign timestamps only if they are valid
		configSet.CreatedAt = utils.ParseNullTime(createdAt)
		configSet.UpdatedAt = utils.ParseNullTime(updatedAt)

		configSets = append(configSets, configSet)
	}

	// Check for errors during iteration
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return configSets, nil
}

// CreateConfig inserts a new configuration into the database
func (f *FrontendConfigRepository) CreateConfigSet(ctx context.Context, configSet *models.ConfigSet) error {
	query := `
		INSERT INTO config_sets (name, version, credentials, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := f.db.Exec(query, configSet.Name, configSet.Version, configSet.Version, configSet.CreatedAt.Unix(), configSet.UpdatedAt.Unix())
	if err != nil {
		return fmt.Errorf("failed to insert config: %w", err)
	}

	return nil
}

// GetConfig retrieves a specific configuration by ID
func (f *FrontendConfigRepository) GetConfig(ctx context.Context, id string) (map[string]any, error) {
	query := `
	SELECT json_object(
		'config_set', json_object(
			'id', cs.id,
			'version', cs.version,
			'credentials', cs.credentials,
			'created_at', cs.created_at,
			'updated_at', cs.updated_at
		),
		'telemetry_settings', json_object(
			'metrics_enabled', ts.metrics_enabled,
			'metrics_endpoint', ts.metrics_endpoint,
			'logs_level', ts.logs_level,
			'traces_enabled', ts.traces_enabled,
			'traces_endpoint', ts.traces_endpoint
		),
		'extensions', json_group_array(
			json_object(
				'id', e.id,
				'extension_name', e.extension_name,
				'enabled', e.enabled,
				'endpoint', e.endpoint,
				'extra', e.extra
			)
		),
		'pipelines', json_group_array(
			json_object(
				'id', p.id,
				'name', p.name,
				'type', p.type,
				'created_at', p.created_at,
				'updated_at', p.updated_at,
				'components', (
					SELECT json_group_array(
						json_object(
							'id', pc.id,
							'component_type', pc.component_type,
							'type', pc.type,
							'name', pc.name,
							'config', pc.config
						)
					)
					FROM pipeline_components pc WHERE pc.pipeline_id = p.id
				)
			)
		)
	) AS full_config
	FROM config_sets cs
	LEFT JOIN telemetry_settings ts ON cs.id = ts.config_set_id
	LEFT JOIN extensions e ON cs.id = e.config_set_id
	LEFT JOIN pipelines p ON cs.id = p.config_set_id
	WHERE cs.id = ?
	GROUP BY cs.id;
`

	var jsonConfig map[string]any
	err := f.db.QueryRow(query, id).Scan(&jsonConfig)
	if err != nil {
		return nil, err
	}
	return jsonConfig, nil
}

// DeleteConfig removes a configuration by ID
func (f *FrontendConfigRepository) DeleteConfig(ctx context.Context, id string) error {
	result, err := f.db.ExecContext(ctx, "DELETE FROM config WHERE ID = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no config found with ID: %s", id)
	}

	return nil
}

// UpdateConfig modifies an existing configuration by ID
func (f *FrontendConfigRepository) UpdateConfig(ctx context.Context, id string, configUpdateRequest ConfigUpsertRequest) error {
	query := `
		UPDATE config 
		SET Name = ?, Description = ?, Config = ?, TargetAgent = ?, UpdatedAt = ?
		WHERE ID = ?
	`

	result, err := f.db.ExecContext(ctx, query, configUpdateRequest.Name, configUpdateRequest.Description, configUpdateRequest.Config, configUpdateRequest.TargetAgent, time.Now(), id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no config found with ID: %s", id)
	}

	hostnames, err := f.getAgentHostnamesByConfigID(ctx, id)
	if err != nil {
		return fmt.Errorf("error fetching agent hostnames: %v", err)
	}

	jsonData, err := json.Marshal(configUpdateRequest)
	if err != nil {
		return fmt.Errorf("error marshalling config update request: %v", err)
	}

	for _, hostname := range hostnames {
		apiEndpoint := fmt.Sprintf("http://%s:443/agent/v1/config", hostname)

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, apiEndpoint, bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("Error creating request for hostname %s: %v", hostname, err)
		}

		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error making API call to hostname %s: %v", hostname, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("API call to hostname %s failed with status: %s", hostname, resp.Status)
		}
	}

	return nil
}

func (f *FrontendConfigRepository) getAgentHostnamesByConfigID(ctx context.Context, configID string) ([]string, error) {
	query := "SELECT hostname FROM agents WHERE configId = ?"
	rows, err := f.db.QueryContext(ctx, query, configID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hostnames []string
	for rows.Next() {
		var hostname string
		if err := rows.Scan(&hostname); err != nil {
			return nil, err
		}
		hostnames = append(hostnames, hostname)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return hostnames, nil
}
