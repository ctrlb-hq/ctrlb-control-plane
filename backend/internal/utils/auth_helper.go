package utils

import (
	"encoding/base64"
	"errors"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
)

func ValidateUserRegistrationRequest(request *models.UserRegisterRequest) error {
	if request.Name == "" {
		return errors.New("name cannot be empty")
	}
	if request.Email == "" {
		return errors.New("email cannot be empty")
	}
	if request.Password == "" {
		return errors.New("password cannot be empty")
	}
	return nil
}

func EncodeBasicAuth(email, password string) string {
	// Concatenate email and password with a colon
	auth := email + ":" + password

	// Base64 encode the resulting string
	encoded := base64.StdEncoding.EncodeToString([]byte(auth))

	return encoded
}
