package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/google/uuid"
)

func GenerateAgentName(typ string, version string, hostname string) string {
	return fmt.Sprintf("%s_%s@%s", typ, version, hostname)
}

var ErrUserAlreadyExists = errors.New("user already exists")

type ErrorResponse struct {
	Error string `json:"error"`
}

func WriteJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	jsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	_, err = w.Write(jsonData)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func UnmarshalJSONRequest(r *http.Request, v interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return err
	}
	defer r.Body.Close()
	return nil
}

func UnmarshalJSONResponse(r *http.Response, v interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return err
	}
	defer r.Body.Close()
	return nil
}

func SendJSONError(w http.ResponseWriter, statusCode int, errMsg string) {
	errResp := ErrorResponse{
		Error: errMsg,
	}

	// Set the response content type
	w.Header().Set("Content-Type", "application/json")

	// Set the response status code
	w.WriteHeader(statusCode)

	// Marshal the error response struct to JSON
	jsonData, err := json.Marshal(errResp)
	if err != nil {
		log.Printf("Failed to marshal error response to JSON: %s", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content", errMsg)

	// Write the JSON response to the response writer
	_, err = w.Write(jsonData)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func CreateNewUUID() string {
	return uuid.New().String()
}

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
