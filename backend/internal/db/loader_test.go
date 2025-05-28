package database_test

import (
	"database/sql"
	"strings"
	"testing"
	"testing/fstest"

	database "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/db"
	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := database.DBInit(":memory:")
	if err != nil {
		t.Fatalf("failed to initialize test database: %v", err)
	}
	return db
}

func TestLoadSchemasFromDirectory_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	mockSchemasFS := fstest.MapFS{
		"testcomponent.json": &fstest.MapFile{
			Data: []byte(`{
				"title": "Test Component",
				"type": "object",
				"properties": {
					"field": {"type": "string"}
				}
			}`),
		},
		"ignore.txt": &fstest.MapFile{
			Data: []byte("not a json file"),
		},
	}

	mockUISchemasFS := fstest.MapFS{
		"testcomponent.json": &fstest.MapFile{
			Data: []byte(`{
				"ui:order": ["field"]
			}`),
		},
	}

	typeMapping := map[string]string{
		"testcomponent": "receiver",
	}
	signalMapping := map[string][]string{
		"testcomponent": {"metrics", "logs"},
	}

	err := database.LoadSchemasFromDirectory(db, mockSchemasFS, mockUISchemasFS, typeMapping, signalMapping)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	row := db.QueryRow(`SELECT name, type, display_name, supported_signals, ui_schema_json FROM component_schemas WHERE name = ?`, "testcomponent")

	var name, typ, displayName, supportedSignals, uiSchema string
	if err := row.Scan(&name, &typ, &displayName, &supportedSignals, &uiSchema); err != nil {
		t.Fatalf("failed querying inserted component: %v", err)
	}

	if name != "testcomponent" {
		t.Errorf("expected name 'testcomponent', got %s", name)
	}
	if typ != "receiver" {
		t.Errorf("expected type 'receiver', got %s", typ)
	}
	if displayName != "Test Component" {
		t.Errorf("expected display name 'Test Component', got %s", displayName)
	}
	if !strings.Contains(supportedSignals, "metrics") || !strings.Contains(supportedSignals, "logs") {
		t.Errorf("expected supported_signals to include 'metrics' and 'logs', got %s", supportedSignals)
	}
	if !strings.Contains(uiSchema, `"ui:order"`) {
		t.Errorf("expected ui_schema_json to include 'ui:order', got %s", uiSchema)
	}
}

func TestLoadSchemasFromDirectory_InvalidJSON(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	mockSchemasFS := fstest.MapFS{
		"badcomponent.json": &fstest.MapFile{
			Data: []byte(`{ invalid json}`),
		},
	}
	mockUISchemasFS := fstest.MapFS{
		"badcomponent.json": &fstest.MapFile{
			Data: []byte(`{}`), // valid or dummy ui schema
		},
	}

	err := database.LoadSchemasFromDirectory(db, mockSchemasFS, mockUISchemasFS, map[string]string{}, map[string][]string{})
	if err == nil {
		t.Fatal("expected error due to invalid JSON, but got nil")
	}
	if !strings.Contains(err.Error(), "invalid JSON") {
		t.Errorf("expected error to mention 'invalid JSON', got %v", err)
	}
}
