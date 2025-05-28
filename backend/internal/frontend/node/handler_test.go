package frontendnode

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

// MockFrontendAgentService is a mock implementation of FrontendAgentServiceInterface
type MockFrontendAgentService struct {
	GetComponentsFunc              func(componentType string) (*[]ComponentInfo, error)
	GetComponentSchemaByNameFunc   func(componentName string) (any, error)
	GetComponentUISchemaByNameFunc func(componentName string) (any, error)
}

func (m *MockFrontendAgentService) GetComponents(componentType string) (*[]ComponentInfo, error) {
	return m.GetComponentsFunc(componentType)
}

func (m *MockFrontendAgentService) GetComponentSchemaByName(componentName string) (any, error) {
	return m.GetComponentSchemaByNameFunc(componentName)
}

func (m *MockFrontendAgentService) GetComponentUISchemaByName(componentName string) (any, error) {
	return m.GetComponentUISchemaByNameFunc(componentName)
}

// TestFrontendNodeHandler_GetComponent tests GetComponent endpoint
func TestFrontendNodeHandler_GetComponent(t *testing.T) {
	tests := []struct {
		name              string
		componentType     string
		mockServiceOutput *[]ComponentInfo
		mockServiceError  error
		expectedStatus    int
	}{
		{
			name:              "Valid receiver type",
			componentType:     "receiver",
			mockServiceOutput: &[]ComponentInfo{{Name: "comp1", Type: "receiver"}},
			mockServiceError:  nil,
			expectedStatus:    http.StatusOK,
		},
		{
			name:              "Invalid type",
			componentType:     "invalid",
			mockServiceOutput: nil,
			mockServiceError:  nil,
			expectedStatus:    http.StatusBadRequest,
		},
		{
			name:              "Internal server error",
			componentType:     "receiver",
			mockServiceOutput: nil,
			mockServiceError:  errors.New("internal error"),
			expectedStatus:    http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockFrontendAgentService{
				GetComponentsFunc: func(componentType string) (*[]ComponentInfo, error) {
					return tt.mockServiceOutput, tt.mockServiceError
				},
			}

			handler := NewFrontendNodeHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/components?type="+tt.componentType, nil)
			w := httptest.NewRecorder()

			handler.GetComponent(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

// TestFrontendNodeHandler_GetComponentSchema tests GetComponentSchema endpoint
func TestFrontendNodeHandler_GetComponentSchema(t *testing.T) {
	tests := []struct {
		name                 string
		componentName        string
		mockServiceOutput    any
		mockServiceError     error
		expectedStatus       int
		expectedErrorMessage string
	}{
		{
			name:              "Schema found",
			componentName:     "comp1",
			mockServiceOutput: map[string]string{"field1": "value1"},
			mockServiceError:  nil,
			expectedStatus:    http.StatusOK,
		},
		{
			name:                 "Schema not found (sql.ErrNoRows)",
			componentName:        "comp2",
			mockServiceOutput:    nil,
			mockServiceError:     sql.ErrNoRows,
			expectedStatus:       http.StatusOK,
			expectedErrorMessage: "Schema not found",
		},
		{
			name:                 "Internal server error",
			componentName:        "comp3",
			mockServiceOutput:    nil,
			mockServiceError:     errors.New("internal error"),
			expectedStatus:       http.StatusInternalServerError,
			expectedErrorMessage: "internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockFrontendAgentService{
				GetComponentSchemaByNameFunc: func(componentName string) (any, error) {
					return tt.mockServiceOutput, tt.mockServiceError
				},
			}

			handler := NewFrontendNodeHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/schema/"+tt.componentName, nil)
			req = mux.SetURLVars(req, map[string]string{"name": tt.componentName})
			w := httptest.NewRecorder()

			handler.GetComponentSchema(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedErrorMessage != "" {
				var responseBody map[string]any
				json.NewDecoder(resp.Body).Decode(&responseBody)

				errorMessage, ok := responseBody["error"].(string)
				if !ok || errorMessage != tt.expectedErrorMessage {
					t.Errorf("expected error message %q, got %q", tt.expectedErrorMessage, errorMessage)
				}
			}
		})
	}
}

func TestFrontendNodeHandler_GetComponentUISchema(t *testing.T) {
	tests := []struct {
		name                 string
		componentName        string
		mockServiceOutput    any
		mockServiceError     error
		expectedStatus       int
		expectedErrorMessage string
	}{
		{
			name:              "UI Schema found",
			componentName:     "comp1",
			mockServiceOutput: map[string]string{"ui:order": "field1"},
			mockServiceError:  nil,
			expectedStatus:    http.StatusOK,
		},
		{
			name:                 "UI Schema not found (sql.ErrNoRows)",
			componentName:        "comp2",
			mockServiceOutput:    nil,
			mockServiceError:     sql.ErrNoRows,
			expectedStatus:       http.StatusOK,
			expectedErrorMessage: "UI Schema not found",
		},
		{
			name:                 "Internal server error",
			componentName:        "comp3",
			mockServiceOutput:    nil,
			mockServiceError:     errors.New("internal error"),
			expectedStatus:       http.StatusInternalServerError,
			expectedErrorMessage: "internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockFrontendAgentService{
				GetComponentUISchemaByNameFunc: func(componentName string) (any, error) {
					return tt.mockServiceOutput, tt.mockServiceError
				},
			}

			handler := NewFrontendNodeHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/ui-schema/"+tt.componentName, nil)
			req = mux.SetURLVars(req, map[string]string{"name": tt.componentName})
			w := httptest.NewRecorder()

			handler.GetComponentUISchema(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedErrorMessage != "" {
				var responseBody map[string]any
				json.NewDecoder(resp.Body).Decode(&responseBody)

				errorMessage, ok := responseBody["error"].(string)
				if !ok || errorMessage != tt.expectedErrorMessage {
					t.Errorf("expected error message %q, got %q", tt.expectedErrorMessage, errorMessage)
				}
			}
		})
	}
}
