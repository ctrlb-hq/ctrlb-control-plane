package frontendconfig

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
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
func (f *FrontendConfigRepository) GetAllConfigs() ([]models.Config, error) {
	var configs []models.Config

	rows, err := f.db.Query("SELECT ID, Name, Description, Config, TargetAgent, CreatedAt, UpdatedAt FROM config")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var config models.Config
		var createdAt, updatedAt sql.NullTime

		// Scan data into struct fields
		if err := rows.Scan(&config.ID, &config.Name, &config.Description, &config.Config, &config.TargetAgent, &createdAt, &updatedAt); err != nil {
			return nil, err
		}

		// Set CreatedAt and UpdatedAt if valid
		if createdAt.Valid {
			config.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			config.UpdatedAt = updatedAt.Time
		}

		configs = append(configs, config)
	}

	// Handle any errors encountered during row iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return configs, nil
}

// CreateConfig inserts a new configuration into the database
func (f *FrontendConfigRepository) CreateConfig(config *models.Config) error {
	query := `
		INSERT INTO config (ID, Name, Description, Config, TargetAgent, CreatedAt, UpdatedAt)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := f.db.Exec(query, config.ID, config.Name, config.Description, config.Config, config.TargetAgent, config.CreatedAt, config.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert config: %w", err)
	}

	return nil
}

// GetConfig retrieves a specific configuration by ID
func (f *FrontendConfigRepository) GetConfig(id string) (*models.Config, error) {
	config := &models.Config{}
	query := "SELECT ID, Name, Description, Config, TargetAgent, CreatedAt, UpdatedAt FROM config WHERE ID = ?"

	// Use QueryRow for single row and handle missing entries explicitly
	err := f.db.QueryRow(query, id).Scan(
		&config.ID, &config.Name, &config.Description, &config.Config, &config.TargetAgent,
		&config.CreatedAt, &config.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("No config found with ID: %s", id)
			return nil, errors.New("no config found with specified ID")
		}
		log.Printf("Error retrieving config ID %s: %v", id, err)
		return nil, err
	}

	return config, nil
}

// DeleteConfig removes a configuration by ID
func (f *FrontendConfigRepository) DeleteConfig(id string) error {
	result, err := f.db.Exec("DELETE FROM config WHERE ID = ?", id)
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
func (f *FrontendConfigRepository) UpdateConfig(id string, configUpdateRequest ConfigUpsertRequest) error {
	query := `
		UPDATE config 
		SET Name = ?, Description = ?, Config = ?, TargetAgent = ?, UpdatedAt = ?
		WHERE ID = ?
	`

	result, err := f.db.Exec(query, configUpdateRequest.Name, configUpdateRequest.Description, configUpdateRequest.Config, configUpdateRequest.TargetAgent, time.Now(), id)
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
