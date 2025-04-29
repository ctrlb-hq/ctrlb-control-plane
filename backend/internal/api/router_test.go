package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/agent"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/api"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/auth"
	frontendagent "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontend/agent"
	frontendnode "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontend/node"
	frontendpipeline "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/frontend/pipeline"
)

func setupMockHandler() *api.Handler {
	return &api.Handler{
		AgentHandler:            &agent.AgentHandler{},
		AuthHandler:             &auth.AuthHandler{},
		FrontendAgentHandler:    &frontendagent.FrontendAgentHandler{},
		FrontendPipelineHandler: &frontendpipeline.FrontendPipelineHandler{},
		FrontendNodeHandler:     &frontendnode.FrontendNodeHandler{},
	}
}

func TestRouter_AuthRoutes(t *testing.T) {
	handler := setupMockHandler()
	router := api.NewRouter(handler)

	tests := []struct {
		name       string
		method     string
		path       string
		payload    string
		expectCode int
	}{
		{
			name:       "POST register with invalid body",
			method:     http.MethodPost,
			path:       "/api/auth/v1/register",
			payload:    `{}`, // or you can even send invalid JSON like `"{invalid"`
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "POST login with invalid body",
			method:     http.MethodPost,
			path:       "/api/auth/v1/login",
			payload:    `{}`,
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "POST refresh with invalid body",
			method:     http.MethodPost,
			path:       "/api/auth/v1/refresh",
			payload:    `{}`,
			expectCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.payload))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.expectCode {
				t.Errorf("%s: expected status %d, got %d", tt.name, tt.expectCode, rec.Code)
			}
		})
	}
}


func TestRouter_AgentRoutes(t *testing.T) {
	handler := setupMockHandler()
	router := api.NewRouter(handler)

	req := httptest.NewRequest(http.MethodPost, "/api/agent/v1/agents", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestRouter_FrontendRoutes_Unauthorized(t *testing.T) {
	handler := setupMockHandler()
	router := api.NewRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/frontend/v2/agents", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected unauthorized 401, got %d", rec.Code)
	}
}

func TestRouter_MethodNotAllowed(t *testing.T) {
	handler := setupMockHandler()
	router := api.NewRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/v1/login", nil) // login expects POST
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 Method Not Found, got %d", rec.Code)
	}
}
