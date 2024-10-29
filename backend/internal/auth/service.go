package auth

import (
	"errors"

	sessionManager "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/auth/session-manager"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	AuthRepository *AuthRepository
	SessionManager *sessionManager.SessionManager
}

func NewAuthService(authRepository *AuthRepository, sessionManager *sessionManager.SessionManager) *AuthService {
	return &AuthService{
		AuthRepository: authRepository,
		SessionManager: sessionManager,
	}
}

func (a *AuthService) RegisterUser(request *models.UserRegisterRequest) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Create a user model instance with the hashed password
	user := User{
		Email:    request.Email,
		Name:     request.Name,
		Password: string(hashedPassword), // Store hashed password
	}

	err = a.AuthRepository.RegisterUser(user)
	if err != nil {
		return err
	}
	return nil
}

func (a *AuthService) Login(request *LoginRequest) (string, error) {
	user, err := a.AuthRepository.Login(request.Email)
	if err != nil {
		return "", err
	}

	// Compare the provided password with the hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	// Generate session
	sessionID, err := a.SessionManager.CreateSession(user.Email)
	if err != nil {
		return "", errors.New("could not create session")
	}

	return sessionID, nil
}

func (a *AuthService) Logout(sessionID string) error {
	a.SessionManager.DeleteSession(sessionID)
	return nil
}
