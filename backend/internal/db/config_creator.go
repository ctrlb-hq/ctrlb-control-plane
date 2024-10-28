package database

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/constants"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
)

func createDefaultConfig(db *sql.DB) error {
	if checkExistingDefault(db) {
		log.Println("Default config already exist")
		return nil
	}

	config := defaultFBConfig()
	err := insertConfigToTable(config, db)
	if err != nil {
		return err
	}
	config = defaultOTELConfig()
	err = insertConfigToTable(config, db)
	if err != nil {
		return err
	}
	log.Println("Default Config created")
	return nil
}

func checkExistingDefault(db *sql.DB) bool {
	var exists bool
	_ = db.QueryRow(`SELECT EXISTS(SELECT 1 FROM config WHERE Name = "default")`).Scan(&exists)

	return exists
}

func defaultFBConfig() models.Config {
	configString :=
		`
service:
    http_server: "on"

pipeline:
    inputs:
        - name: dummy
          Interval_sec: 5


    outputs:
        - name: stdout
          match: "*"
`
	constants.DEFAULT_CONFIG_FB_ID = utils.CreateNewUUID()
	config := models.Config{
		ID:          constants.DEFAULT_CONFIG_FB_ID,
		Name:        "default",
		Description: "Default Config for Fluent Bit Agent",
		Config:      configString,
		TargetAgent: "fluent-bit",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	return config
}

func defaultOTELConfig() models.Config {
	otelConfigString :=
		`
receivers:
otlp:
	protocols:
	http:
	grpc:

exporters:
logging:
	logLevel: info

service:
pipelines:
	traces:
	receivers: [otlp]
	exporters: [logging]
`

	constants.DEFAULT_CONFIG_OTEL_ID = utils.CreateNewUUID()
	config := models.Config{
		ID:          constants.DEFAULT_CONFIG_OTEL_ID,
		Name:        "default",
		Description: "Default Config for OTEL Agent",
		Config:      otelConfigString,
		TargetAgent: "otel",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	return config
}

func insertConfigToTable(config models.Config, db *sql.DB) error {
	_, err := db.Exec("INSERT INTO config (ID, Name, Description, Config, TargetAgent, CreatedAt, UpdatedAt) VALUES (?, ?, ?, ?, ?, ?, ?)", config.ID, config.Name, config.Description, config.Config, config.TargetAgent, config.CreatedAt, config.UpdatedAt)

	if err != nil {
		return errors.New("error encountered while adding default config" + err.Error())
	}
	return nil
}
