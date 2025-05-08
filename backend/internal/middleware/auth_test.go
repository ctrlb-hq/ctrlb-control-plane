package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/middleware"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
	"github.com/golang-jwt/jwt/v5"
)

// Backup original function
var originalValidateJWT = utils.ValidateJWTFunc

func restoreValidateJWT() {
	utils.ValidateJWTFunc = originalValidateJWT
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	recorder := httptest.NewRecorder()

	middleware.AuthMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "Missing token") {
		t.Errorf("expected error message for missing token, got %s", recorder.Body.String())
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	defer restoreValidateJWT()
	utils.ValidateJWTFunc = func(tokenString, typ string) (string, error) {
		return "", errors.New("invalid token")
	}

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalidtoken")
	recorder := httptest.NewRecorder()

	middleware.AuthMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "Invalid token") {
		t.Errorf("expected error message for invalid token, got %s", recorder.Body.String())
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	defer restoreValidateJWT()
	utils.ValidateJWTFunc = func(tokenString, typ string) (string, error) {
		return "", jwt.ErrTokenExpired
	}

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer expiredtoken")
	recorder := httptest.NewRecorder()

	middleware.AuthMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "Token expired") {
		t.Errorf("expected error message for token expiration, got %s", recorder.Body.String())
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	defer restoreValidateJWT()

	email := "test@example.com"

	utils.ValidateJWTFunc = func(tokenString, typ string) (string, error) {
		return email, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer validtoken")
	recorder := httptest.NewRecorder()

	handlerCalled := false

	middleware.AuthMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true

		ctxEmail := r.Context().Value(middleware.EmailContextKey)
		if ctxEmail != email {
			t.Errorf("expected context email %q, got %q", email, ctxEmail)
		}
	})).ServeHTTP(recorder, req)

	if !handlerCalled {
		t.Error("expected next handler to be called")
	}
	if recorder.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", recorder.Code)
	}
}
