package auth_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/auth"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
)

// MockAuthService mocks the AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) RegisterUser(req *models.UserRegisterRequest) (*auth.UserResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*auth.UserResponse), args.Error(1)
}

func (m *MockAuthService) Login(req *models.LoginRequest) (*auth.UserResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.UserResponse), args.Error(1)
}

func (m *MockAuthService) RefreshToken(req auth.RefreshTokenRequest) (any, error) {
	args := m.Called(req)
	return args.Get(0), args.Error(1)
}

func TestAuthHandler_Register(t *testing.T) {
	mockSvc := new(MockAuthService)
	handler := auth.NewAuthHandler(mockSvc)

	requestBody := models.UserRegisterRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "password123",
	}
	jsonBody, _ := json.Marshal(requestBody)

	r := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	expectedResponse := &auth.UserResponse{
		Name:         "Test User",
		Email:        "test@example.com",
		Role:         "user",
		AccessToken:  "dummy-access-token",
		RefreshToken: "dummy-refresh-token",
		Message:      "User registered successfully",
	}

	mockSvc.On("RegisterUser", mock.Anything).Return(expectedResponse, nil)

	handler.Register(w, r)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var gotResp auth.UserResponse
	err := json.NewDecoder(resp.Body).Decode(&gotResp)
	assert.NoError(t, err)

	assert.Equal(t, expectedResponse.Email, gotResp.Email)
	assert.Equal(t, expectedResponse.Name, gotResp.Name)
	assert.Equal(t, expectedResponse.Role, gotResp.Role)
	assert.Equal(t, expectedResponse.Message, gotResp.Message)
	assert.NotEmpty(t, gotResp.AccessToken)
	assert.NotEmpty(t, gotResp.RefreshToken)

	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	mockSvc := new(MockAuthService)
	handler := auth.NewAuthHandler(mockSvc)

	loginReq := models.LoginRequest{
		Email:    "login@example.com",
		Password: "password",
	}
	jsonBody, _ := json.Marshal(loginReq)
	r := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mockSvc.On("Login", mock.Anything).Return(&auth.UserResponse{
		Email: loginReq.Email,
		Name:  "Login User",
		Role:  "admin",
	}, nil)

	handler.Login(w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Login_Failure(t *testing.T) {
	mockSvc := new(MockAuthService)
	handler := auth.NewAuthHandler(mockSvc)

	loginReq := models.LoginRequest{
		Email:    "login@example.com",
		Password: "wrongpassword",
	}
	jsonBody, _ := json.Marshal(loginReq)
	r := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mockSvc.On("Login", mock.Anything).Return(nil, errors.New("invalid credentials"))

	handler.Login(w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	mockSvc := new(MockAuthService)
	handler := auth.NewAuthHandler(mockSvc)

	refreshReq := auth.RefreshTokenRequest{
		RefreshToken: "dummy-refresh-token",
	}
	jsonBody, _ := json.Marshal(refreshReq)
	r := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBuffer(jsonBody))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mockSvc.On("RefreshToken", mock.Anything).Return(map[string]string{
		"access_token":  "new-access-token",
		"refresh_token": refreshReq.RefreshToken,
	}, nil)

	handler.RefreshToken(w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockSvc.AssertExpectations(t)
}
