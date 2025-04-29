package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/middleware"
)

func TestCorsMiddleware_PreflightOptionsRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	recorder := httptest.NewRecorder()

	called := false
	handler := middleware.CorsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	handler.ServeHTTP(recorder, req)

	// Should not call next handler for OPTIONS
	if called {
		t.Error("expected handler not to be called for OPTIONS request")
	}

	if recorder.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", recorder.Code)
	}

	checkCorsHeaders(t, recorder)
}

func TestCorsMiddleware_NormalRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()

	handlerCalled := false
	handler := middleware.CorsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	}))

	handler.ServeHTTP(recorder, req)

	if !handlerCalled {
		t.Error("expected handler to be called for normal request")
	}

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", recorder.Code)
	}

	checkCorsHeaders(t, recorder)
}

func checkCorsHeaders(t *testing.T, recorder *httptest.ResponseRecorder) {
	t.Helper()
	if origin := recorder.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
		t.Errorf("expected Access-Control-Allow-Origin '*', got %q", origin)
	}
	if methods := recorder.Header().Get("Access-Control-Allow-Methods"); methods == "" {
		t.Error("expected Access-Control-Allow-Methods header to be set")
	}
	if headers := recorder.Header().Get("Access-Control-Allow-Headers"); headers == "" {
		t.Error("expected Access-Control-Allow-Headers header to be set")
	}
}
