package kubernetes

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/middleware"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
	"github.com/gorilla/mux"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) Manifest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	var req ManifestRequest
	if err := decodeRequest(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	data, filename, err := h.service.Manifest(r.Context(), mux.Vars(r)["installation_id"], req)
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/yaml")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (h *Handler) CreatePipeline(w http.ResponseWriter, r *http.Request) {
	var req CreatePipelineRequest
	if err := decodeRequest(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	email, _ := r.Context().Value(middleware.EmailContextKey).(string)
	response, err := h.service.CreatePipeline(r.Context(), req, email)
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	utils.WriteJSONResponse(w, http.StatusCreated, response)
}

func (h *Handler) CreateInstallation(w http.ResponseWriter, r *http.Request) {
	var req CreateInstallationRequest
	if err := decodeRequest(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	email, _ := r.Context().Value(middleware.EmailContextKey).(string)
	response, err := h.service.Create(r.Context(), req, email)
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	utils.WriteJSONResponse(w, http.StatusCreated, response)
}

func (h *Handler) CheckIn(w http.ResponseWriter, r *http.Request) {
	token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok || token == "" || strings.TrimSpace(token) != token {
		writeError(w, ErrUnauthorized)
		return
	}
	var req CheckInRequest
	if err := decodeRequest(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	response, err := h.service.CheckIn(r.Context(), mux.Vars(r)["installation_id"], token, req)
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (h *Handler) GetInstallation(w http.ResponseWriter, r *http.Request) {
	response, err := h.service.Get(r.Context(), mux.Vars(r)["installation_id"])
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (h *Handler) ListInstallations(w http.ResponseWriter, r *http.Request) {
	h.list(w, r, nil)
}

func (h *Handler) GetPipelineInstallations(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil || id <= 0 {
		writeError(w, ErrInvalid)
		return
	}
	h.list(w, r, &id)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request, pipelineID *int) {
	response, err := h.service.List(r.Context(), pipelineID)
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func decodeRequest(w http.ResponseWriter, r *http.Request, value any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(value); err != nil {
		return ErrInvalid
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		return ErrInvalid
	}
	return nil
}

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	message := "Kubernetes management request failed"
	switch {
	case errors.Is(err, ErrUnauthorized):
		status, message = http.StatusUnauthorized, ErrUnauthorized.Error()
	case errors.Is(err, ErrNotFound):
		status, message = http.StatusNotFound, ErrNotFound.Error()
	case errors.Is(err, ErrInvalid):
		status, message = http.StatusBadRequest, err.Error()
	}
	utils.SendJSONError(w, status, message)
}
