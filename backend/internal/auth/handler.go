package auth

import (
	"encoding/json"
	"errors"
	"log"
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
	var userRegisterRequest models.UserRegisterRequest

	// Step 1: Parse and decode the request body
	err := json.NewDecoder(r.Body).Decode(&userRegisterRequest)
	if err != nil {
		// Log the error for debugging purposes
		log.Printf("Error decoding request body: %v", err)

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
		log.Printf("Error decoding request body: %v", err)

		// Respond with a generic error message
		response := map[string]string{
			"error":   "invalid_request",
			"message": "The request body is invalid.",
		}
		utils.WriteJSONResponse(w, http.StatusBadRequest, response)
		return
	}

	// Step 2: Register the user
	err = a.AuthService.RegisterUser(&userRegisterRequest)
	if err != nil {
		// Log the actual error
		log.Printf("Error registering user: %v", err)

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
	var loginRequest LoginRequest
	err := json.NewDecoder(r.Body).Decode(&loginRequest)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Login and get session ID

	response, err := a.AuthService.Login(&loginRequest)
	token, err := a.AuthService.Login(&loginRequest)
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (a *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	response, err := a.AuthService.RefreshToken(req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "Invalid or expired refresh token" {
			status = http.StatusUnauthorized
		}
		utils.SendJSONError(w, status, err.Error())
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}
