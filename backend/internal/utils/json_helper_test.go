package utils_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
)

type SampleResponse struct {
	Message string `json:"message"`
}

type SampleRequest struct {
	Name string `json:"name"`
}

func TestWriteJSONResponse(t *testing.T) {
	recorder := httptest.NewRecorder()
	data := SampleResponse{Message: "Hello"}

	utils.WriteJSONResponse(recorder, http.StatusOK, data)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	if ct := recorder.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected content type application/json, got %s", ct)
	}

	var response SampleResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Message != data.Message {
		t.Errorf("expected message %q, got %q", data.Message, response.Message)
	}
}

func TestUnmarshalJSONRequest(t *testing.T) {
	payload := `{"name": "Admin"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(payload))
	var parsed SampleRequest

	err := utils.UnmarshalJSONRequest(req, &parsed)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if parsed.Name != "Admin" {
		t.Errorf("expected name 'Admin', got %q", parsed.Name)
	}
}

func TestSendJSONError(t *testing.T) {
	recorder := httptest.NewRecorder()
	errMsg := "Something went wrong"

	utils.SendJSONError(recorder, http.StatusBadRequest, errMsg)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}

	if ct := recorder.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected content type application/json, got %s", ct)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, errMsg) {
		t.Errorf("expected body to contain %q, got %s", errMsg, body)
	}
}
