package auth

import (
	"errors"
	"fmt"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

type AuthRepositoryInterface interface {
	RegisterUser(user User) error
	Login(email string) (*User, error)
	UserExists(email string) bool
}

type AuthServiceInterface interface {
	RegisterUser(request *models.UserRegisterRequest) (*UserResponse, error)
	Login(request *models.LoginRequest) (*UserResponse, error)
	RefreshToken(req RefreshTokenRequest) (any, error)
}


type AuthService struct {
	AuthRepository AuthRepositoryInterface
}

func NewAuthService(authRepositoryInterface AuthRepositoryInterface) *AuthService {
	return &AuthService{
		AuthRepository: authRepositoryInterface,
	}
}

func (a *AuthService) RegisterUser(request *models.UserRegisterRequest) (*UserResponse, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
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
		return nil, err
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

	response := &UserResponse{
		Name:         user.Name,
		Email:        user.Email,
		Role:         user.Role,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Message:      "User registered successfully",
	}
	return response, nil
}

// Login handles user login and returns both access and refresh tokens
func (a *AuthService) Login(request *models.LoginRequest) (*UserResponse, error) {
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
	response := &UserResponse{
		Name:         user.Name,
		Email:        user.Email,
		Role:         user.Role,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Message:      "Login successfully",
	}

	// Return both tokens
	return response, nil
}

// Login handles user login and returns both access and refresh tokens
func (a *AuthService) RefreshToken(req RefreshTokenRequest) (any, error) {

	// Validate the refresh token
	email, err := utils.ValidateJWT(req.RefreshToken, "refresh")
	if err != nil {
		return nil, errors.New("invalid or expired refresh token")
	}

	if !a.AuthRepository.UserExists(email) {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Generate a new access token
	accessToken, err := utils.GenerateAccessToken(email)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}
	response := map[string]string{
		"access_token":  accessToken,
		"refresh_token": req.RefreshToken,
	}
	return response, nil
}
