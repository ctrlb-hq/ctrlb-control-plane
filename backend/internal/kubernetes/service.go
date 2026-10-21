package kubernetes

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

var (
	ErrUnauthorized = errors.New("invalid installation token")
	ErrNotFound     = errors.New("installation or pipeline not found")
	ErrInvalid      = errors.New("invalid request")
)

const ConnectionTimeout = 35 * time.Second

type Service struct {
	repository *Repository
	now        func() time.Time
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository, now: time.Now}
}

// CanonicalConfig normalizes object key order and whitespace before hashing.
func CanonicalConfig(data []byte) (map[string]any, string, error) {
	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil || len(config) == 0 {
		return nil, "", fmt.Errorf("%w: pipeline must contain a configuration object", ErrInvalid)
	}
	canonical, err := json.Marshal(config)
	if err != nil {
		return nil, "", err
	}
	hash := sha256.Sum256(canonical)
	return config, "sha256:" + hex.EncodeToString(hash[:]), nil
}

func tokenHash(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (s *Service) Create(ctx context.Context, req CreateInstallationRequest, createdBy string) (*CreateInstallationResponse, error) {
	req.Name, req.ClusterName = strings.TrimSpace(req.Name), strings.TrimSpace(req.ClusterName)
	if req.Name == "" || req.ClusterName == "" || len(req.Name) > 253 || len(req.ClusterName) > 253 || req.PipelineID <= 0 {
		return nil, fmt.Errorf("%w: name, cluster_name, and a positive pipeline_id are required", ErrInvalid)
	}
	data, err := s.repository.PipelineConfig(ctx, req.PipelineID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	_, revision, err := CanonicalConfig(data)
	if err != nil {
		return nil, err
	}
	var idBytes [16]byte
	var tokenBytes [32]byte
	if _, err := rand.Read(idBytes[:]); err != nil {
		return nil, err
	}
	if _, err := rand.Read(tokenBytes[:]); err != nil {
		return nil, err
	}
	token := "ctrlb_k8s_" + base64.RawURLEncoding.EncodeToString(tokenBytes[:])
	installation := &Installation{
		ID: hex.EncodeToString(idBytes[:]), Name: req.Name, ClusterName: req.ClusterName,
		PipelineID: req.PipelineID, CreatedBy: createdBy, CreatedAt: s.now().Unix(),
		DesiredRevision: revision, Collectors: []Collector{}, tokenHash: tokenHash(token),
	}
	if err := s.repository.Create(ctx, installation); err != nil {
		return nil, err
	}
	return &CreateInstallationResponse{Installation: installation, Token: token}, nil
}

func (s *Service) CheckIn(ctx context.Context, installationID, token string, req CheckInRequest) (*CheckInResponse, error) {
	i, err := s.repository.Get(ctx, installationID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUnauthorized
	}
	if err != nil {
		return nil, err
	}
	if token == "" || subtle.ConstantTimeCompare([]byte(tokenHash(token)), []byte(i.tokenHash)) != 1 {
		return nil, ErrUnauthorized
	}
	if err := validateCheckIn(req); err != nil {
		return nil, err
	}
	config, revision, err := CanonicalConfig(i.configJSON)
	if err != nil {
		return nil, err
	}
	now := s.now().Unix()
	id, err := s.repository.CheckIn(ctx, i, req, now)
	if err != nil {
		return nil, err
	}
	response := &CheckInResponse{CollectorID: id, DesiredRevision: revision, PollIntervalSeconds: 10, ServerTime: now}
	if revision != req.AppliedRevision && revision != req.AttemptedRevision {
		response.Config = config
	}
	return response, nil
}

func validateCheckIn(req CheckInRequest) error {
	if strings.TrimSpace(req.NodeName) == "" || strings.TrimSpace(req.NodeName) != req.NodeName || len(req.NodeName) > 253 {
		return fmt.Errorf("%w: node_name is required and must be at most 253 characters", ErrInvalid)
	}
	switch req.EngineStatus {
	case "waiting", "running", "failed":
	default:
		return fmt.Errorf("%w: engine_status must be waiting, running, or failed", ErrInvalid)
	}
	for _, value := range []string{req.NodeUID, req.Namespace, req.PodName, req.PodUID, req.PodIP, req.Version, req.Platform} {
		if len(value) > 253 {
			return fmt.Errorf("%w: collector metadata must be at most 253 characters", ErrInvalid)
		}
	}
	if len(req.ConfigError) > 8192 || len(req.AppliedRevision) > 128 || len(req.AttemptedRevision) > 128 {
		return fmt.Errorf("%w: configuration error or revision exceeds its size limit", ErrInvalid)
	}
	if req.Metrics != nil {
		for _, value := range []*float64{req.Metrics.CPUPercent, req.Metrics.MemoryBytes, req.Metrics.LogsSentPerSecond,
			req.Metrics.MetricsSentPerSecond, req.Metrics.SpansSentPerSecond} {
			if value != nil && (*value < 0 || math.IsNaN(*value) || math.IsInf(*value, 0)) {
				return fmt.Errorf("%w: metrics must be finite and nonnegative", ErrInvalid)
			}
		}
	}
	return nil
}

func (s *Service) Get(ctx context.Context, id string) (*Installation, error) {
	i, err := s.repository.Get(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	_, i.DesiredRevision, err = CanonicalConfig(i.configJSON)
	if err != nil {
		return nil, err
	}
	i.Collectors, err = s.repository.Collectors(ctx, id)
	if err != nil {
		return nil, err
	}
	now := s.now().Unix()
	for index := range i.Collectors {
		c := &i.Collectors[index]
		c.ConnectionStatus = "connected"
		if now-c.LastSeen > int64(ConnectionTimeout/time.Second) {
			c.ConnectionStatus = "disconnected"
			i.Rollout.Disconnected++
		}
		switch {
		case c.AppliedRevision == i.DesiredRevision && c.EngineStatus == "running":
			c.RolloutStatus = "applied"
			i.Rollout.Applied++
		case c.AttemptedRevision == i.DesiredRevision && (c.ConfigError != "" || c.EngineStatus == "failed"):
			c.RolloutStatus = "failed"
			i.Rollout.Failed++
		default:
			c.RolloutStatus = "pending"
			i.Rollout.Pending++
		}
		i.Rollout.Total++
	}
	return i, nil
}

func (s *Service) List(ctx context.Context, pipelineID *int) ([]*Installation, error) {
	if pipelineID != nil {
		if _, err := s.repository.PipelineConfig(ctx, *pipelineID); errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		} else if err != nil {
			return nil, err
		}
	}
	ids, err := s.repository.ListIDs(ctx, pipelineID)
	if err != nil {
		return nil, err
	}
	installations := make([]*Installation, 0, len(ids))
	for _, id := range ids {
		i, err := s.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		installations = append(installations, i)
	}
	return installations, nil
}
