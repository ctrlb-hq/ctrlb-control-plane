package frontendnode

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestGetComponents(t *testing.T) {
	// Create a new SQL mock
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewFrontendNodeRepository(db)

	tests := []struct {
		name               string
		componentType      string
		mockRows           *sqlmock.Rows
		expectedComponents *[]ComponentInfo
		expectError        bool
	}{
		{
			name:          "Success - All Components",
			componentType: "",
			mockRows: sqlmock.NewRows([]string{"name", "display_name", "supported_signals", "type"}).
				AddRow("button", "Button", "click,hover", "input").
				AddRow("textbox", "Text Box", "input,focus", "input"),
			expectedComponents: &[]ComponentInfo{
				{
					Name:             "button",
					DisplayName:      "Button",
					Type:             "input",
					SupportedSignals: []string{"click", "hover"},
				},
				{
					Name:             "textbox",
					DisplayName:      "Text Box",
					Type:             "input",
					SupportedSignals: []string{"input", "focus"},
				},
			},
			expectError: false,
		},
		{
			name:          "Return All Receivers",
			componentType: "receiver",
			mockRows: sqlmock.NewRows([]string{"name", "display_name", "supported_signals", "type"}).
				AddRow("awscloudwatch_receiver", "AWS CloudWatch Receiver Configuration", "logs", "receiver").
				AddRow("awscloudwatchmetrics_receiver", "AWS CloudWatch Metrics Receiver Configuration", "metrics", "receiver").
				AddRow("azuremonitor_receiver", "Azure Monitor Receiver Configuration", "metrics", "receiver").
				AddRow("filelog_receiver", "Filelog Receiver Configuration", "logs", "receiver").
				AddRow("googlecloudmonitoring_receiver", "Google Cloud Monitoring Receiver Configuration", "metrics", "receiver").
				AddRow("hostmetrics_receiver", "Host Metrics Receiver Configuration", "metrics", "receiver").
				AddRow("otlp_receiver", "OTLP Receiver Configuration", "traces,metrics,logs", "receiver"),
			expectedComponents: &[]ComponentInfo{
				{"awscloudwatch_receiver", "AWS CloudWatch Receiver Configuration", "receiver", []string{"logs"}},
				{"awscloudwatchmetrics_receiver", "AWS CloudWatch Metrics Receiver Configuration", "receiver", []string{"metrics"}},
				{"azuremonitor_receiver", "Azure Monitor Receiver Configuration", "receiver", []string{"metrics"}},
				{"filelog_receiver", "Filelog Receiver Configuration", "receiver", []string{"logs"}},
				{"googlecloudmonitoring_receiver", "Google Cloud Monitoring Receiver Configuration", "receiver", []string{"metrics"}},
				{"hostmetrics_receiver", "Host Metrics Receiver Configuration", "receiver", []string{"metrics"}},
				{"otlp_receiver", "OTLP Receiver Configuration", "receiver", []string{"traces", "metrics", "logs"}},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up expectations
			if tt.componentType == "" {
				mock.ExpectQuery("SELECT name, display_name, supported_signals, type FROM component_schemas").
					WillReturnRows(tt.mockRows)
			} else {
				mock.ExpectQuery("SELECT name, display_name, supported_signals, type FROM component_schemas WHERE type = ?").
					WithArgs(tt.componentType).
					WillReturnRows(tt.mockRows)
			}

			// Call the method
			components, err := repo.GetComponents(tt.componentType)

			// Assert results
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, components)
				assert.Equal(t, *tt.expectedComponents, *components)
			}
		})
	}
}

func TestGetComponentSchemaByName(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewFrontendNodeRepository(db)

	tests := []struct {
		name          string
		componentName string
		mockSchema    string
		expectError   bool
	}{
		{
			name:          "Success",
			componentName: "button",
			mockSchema:    `{"type": "button", "properties": {"label": "string"}}`,
			expectError:   false,
		},
		{
			name:          "Not Found",
			componentName: "nonexistent",
			mockSchema:    "",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := sqlmock.NewRows([]string{"schema_json"})
			if !tt.expectError {
				rows.AddRow(tt.mockSchema)
			}

			mock.ExpectQuery("SELECT schema_json FROM component_schemas WHERE name = ?").
				WithArgs(tt.componentName).
				WillReturnRows(rows)

			schema, err := repo.GetComponentSchemaByName(tt.componentName)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, schema)
			}
		})
	}
}
