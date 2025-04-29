package auth

import (
	"testing"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// MockAuthRepository mocks the AuthRepositoryInterface
type MockAuthRepository struct {
	mock.Mock
}

func (m *MockAuthRepository) RegisterUser(user User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockAuthRepository) Login(email string) (*User, error) {
	args := m.Called(email)
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockAuthRepository) UserExists(email string) bool {
	args := m.Called(email)
	return args.Bool(0)
}

func TestAuthService_RegisterUser(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	svc := NewAuthService(mockRepo)

	req := &models.UserRegisterRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "password123",
	}

	mockRepo.On("RegisterUser", mock.Anything).Return(nil)

	resp, err := svc.RegisterUser(req)
	assert.NoError(t, err)
	assert.Equal(t, req.Email, resp.Email)
	assert.Equal(t, req.Name, resp.Name)
	assert.Equal(t, "user", resp.Role)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Login_Success(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	svc := NewAuthService(mockRepo)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	existingUser := &User{
		Email:    "login@example.com",
		Name:     "Login User",
		Password: string(hashedPassword),
		Role:     "admin",
	}
	mockRepo.On("Login", "login@example.com").Return(existingUser, nil)

	req := &models.LoginRequest{
		Email:    "login@example.com",
		Password: "password123",
	}

	resp, err := svc.Login(req)
	assert.NoError(t, err)
	assert.Equal(t, existingUser.Email, resp.Email)
	assert.Equal(t, existingUser.Name, resp.Name)
	assert.Equal(t, existingUser.Role, resp.Role)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	svc := NewAuthService(mockRepo)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	existingUser := &User{
		Email:    "wrongpass@example.com",
		Name:     "Wrong Pass User",
		Password: string(hashedPassword),
		Role:     "user",
	}
	mockRepo.On("Login", "wrongpass@example.com").Return(existingUser, nil)

	req := &models.LoginRequest{
		Email:    "wrongpass@example.com",
		Password: "wrongpassword",
	}

	_, err := svc.Login(req)
	assert.Error(t, err)
	assert.Equal(t, "invalid credentials", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestAuthService_RefreshToken(t *testing.T) {
	mockRepo := new(MockAuthRepository)

	// Mock the user existence
	mockRepo.On("UserExists", "test@example.com").Return(true)
}
