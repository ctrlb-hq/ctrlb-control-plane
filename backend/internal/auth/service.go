package auth

import (
	"errors"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	AuthRepository *AuthRepository
}

func NewAuthService(authRepository *AuthRepository) *AuthService {
	return &AuthService{
		AuthRepository: authRepository,
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

	// Generate token
	token, err := utils.GenerateJWT(request.Email)
	if err != nil {
		return "", err
	}
	return token, nil
}
