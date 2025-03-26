package auth

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
)

type AuthHandler struct {
	AuthService *AuthService
}

func NewAuthHandler(authService *AuthService) *AuthHandler {
	return &AuthHandler{
		AuthService: authService,
	}
}

func (a *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	utils.Logger.Info("Starting user registration")
	var userRegisterRequest models.UserRegisterRequest

	// Step 1: Parse and decode the request body
	err := utils.UnmarshalJSONRequest(r, &userRegisterRequest)
	if err != nil {
		// Log the error for debugging purposes
		utils.Logger.Error(fmt.Sprintf("Error decoding request body: %v", err))

		// Respond with a generic error message
		response := map[string]string{
			"error":   "invalid_request",
			"message": "The request body is invalid.",
		}
		utils.WriteJSONResponse(w, http.StatusBadRequest, response)
		return
	}
	err = utils.ValidateUserRegistrationRequest(&userRegisterRequest)
	if err != nil {
		// Log the error for debugging purposes
		utils.Logger.Error(fmt.Sprintf("Error decoding request body: %v", err))

		// Respond with a generic error message
		response := map[string]string{
			"error":   "invalid_request",
			"message": "The request body is invalid.",
		}
		utils.WriteJSONResponse(w, http.StatusBadRequest, response)
		return
	}

	if userRegisterRequest.Role == "" {
		userRegisterRequest.Role = "user"
	}

	utils.Logger.Info(fmt.Sprintf("Registering user with email: %s and role: %s", userRegisterRequest.Email, userRegisterRequest.Role))
	// Step 2: Register the user
	err = a.AuthService.RegisterUser(&userRegisterRequest)
	if err != nil {
		// Log the actual error
		utils.Logger.Error(fmt.Sprintf("Error registering user: %v", err))

		// Send a generic error message to the client
		response := map[string]string{
			"error":   "registration_failed",
			"message": "Unable to register user.",
		}

		// Determine error type and respond with the appropriate status code
		if errors.Is(err, utils.ErrUserAlreadyExists) {
			utils.WriteJSONResponse(w, http.StatusConflict, response)
		} else {
			utils.WriteJSONResponse(w, http.StatusInternalServerError, response)
		}
		return
	}

	// Step 3: Success response
	response := map[string]string{
		"message": "User registered successfully",
	}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (a *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	utils.Logger.Info("Starting login process")
	var loginRequest models.LoginRequest

	if err := utils.UnmarshalJSONRequest(r, &loginRequest); err != nil {
		utils.Logger.Error(fmt.Sprintf("Error decoding login request body: %v", err))
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := utils.ValidateUserLoginRequest(&loginRequest); err != nil {
		utils.Logger.Error(fmt.Sprintf("Error validating login request: %v", err))
		utils.WriteJSONResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	utils.Logger.Info(fmt.Sprintf("Processing login request for user: %s", loginRequest.Email))
	response, err := a.AuthService.Login(&loginRequest)
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("Login failed for user %s: %v", loginRequest.Email, err))
		utils.WriteJSONResponse(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	utils.Logger.Info(fmt.Sprintf("Login successful for user: %s", loginRequest.Email))
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (a *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	utils.Logger.Info("Starting token refresh process")
	var req RefreshTokenRequest

	if err := utils.UnmarshalJSONRequest(r, &req); err != nil {
		utils.Logger.Error(fmt.Sprintf("Error decoding refresh token request: %v", err))
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if req.RefreshToken == "" {
		utils.Logger.Error("Empty refresh token provided")
		utils.SendJSONError(w, http.StatusBadRequest, "Refresh token is required")
		return
	}

	response, err := a.AuthService.RefreshToken(req)
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("Token refresh failed: %v", err))
		status := http.StatusInternalServerError
		if err.Error() == "Invalid or expired refresh token" {
			status = http.StatusUnauthorized
		}
		utils.SendJSONError(w, status, err.Error())
		return
	}

	utils.Logger.Info("Token refresh successful")
	utils.WriteJSONResponse(w, http.StatusOK, response)
}
