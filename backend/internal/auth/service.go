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
		Role:     "user",
	}

	err = a.AuthRepository.RegisterUser(user)
	if err != nil {
		return err
	}
	return nil
}

// Login handles user login and returns both access and refresh tokens
func (a *AuthService) Login(request *LoginRequest) (*LoginResponse, error) {
	user, err := a.AuthRepository.Login(request.Email)
	if err != nil {
		return nil, err
	}

	// Compare the provided password with the hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Generate access token (short-lived)
	accessToken, err := utils.GenerateAccessToken(request.Email)
	if err != nil {
		return nil, err
	}

	// Generate refresh token (long-lived)
	refreshToken, err := utils.GenerateRefreshToken(request.Email)
	if err != nil {
		return nil, err
	}

	// Return both tokens
	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Message:      "Login successful",
	}, nil
}

// Login handles user login and returns both access and refresh tokens
func (a *AuthService) RefreshToken(req RefreshTokenRequest) (interface{}, error) {

	// Validate the refresh token
	email, err := utils.ValidateJWT(req.RefreshToken)
	if err != nil {
		return nil, errors.New("invalid or expired refresh token")
	}

	// Generate a new access token
	accessToken, err := utils.GenerateAccessToken(email)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}
	response := map[string]string{
		"access_token": accessToken,
	}
	return response, nil
}
