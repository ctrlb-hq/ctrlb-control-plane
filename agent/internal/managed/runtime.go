package managed

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/pkg/logger"
)

type checkInRequest struct {
	NodeName  string `json:"node_name"`
	NodeUID   string `json:"node_uid,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	PodName   string `json:"pod_name,omitempty"`
	PodUID    string `json:"pod_uid,omitempty"`
	PodIP     string `json:"pod_ip,omitempty"`
	Version   string `json:"version,omitempty"`
	Platform  string `json:"platform"`
	Status
	Metrics map[string]float64 `json:"metrics"`
}

type checkInResponse struct {
	DesiredRevision     string         `json:"desired_revision"`
	Config              map[string]any `json:"config"`
	PollIntervalSeconds int            `json:"poll_interval_seconds"`
}

type Runtime struct {
	config  Config
	manager *Manager
	client  *http.Client
	metrics *MetricsSampler
	url     string
}

func New(cfg Config, engine Engine) (*Runtime, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	manager, err := NewManager(cfg.StateDir, cfg.InstallationID, cfg.NodeName, engine)
	if err != nil {
		return nil, err
	}
	endpoint, err := url.JoinPath(cfg.BackendURL, "api", "agent", "v2", "installations", cfg.InstallationID, "check-in")
	if err != nil {
		return nil, err
	}
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 10 * time.Second
	}
	if cfg.ListenAddress == "" {
		cfg.ListenAddress = ":3421"
	}
	if cfg.MetricsURL == "" {
		cfg.MetricsURL = "http://127.0.0.1:8888/metrics"
	}
	return &Runtime{
		config: cfg, manager: manager, url: endpoint,
		client:  &http.Client{Timeout: 15 * time.Second, CheckRedirect: rejectRedirect},
		metrics: &MetricsSampler{url: cfg.MetricsURL, client: &http.Client{Timeout: 2 * time.Second, CheckRedirect: rejectRedirect}},
	}, nil
}

// Refusing redirects also prevents an installation token from reaching another
// origin, including a redirect to a different port on the same hostname.
func rejectRedirect(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }

func (r *Runtime) Run(ctx context.Context) error {
	if err := r.manager.Restore(ctx); err != nil {
		logger.Logger.Sugar().Warnf("Saved Collector configuration could not start: %v", err)
	}
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := r.manager.Close(stopCtx); err != nil {
			logger.Logger.Sugar().Errorf("Collector shutdown failed: %v", err)
		}
	}()
	listener, err := net.Listen("tcp", r.config.ListenAddress)
	if err != nil {
		return fmt.Errorf("listen for health probes: %w", err)
	}
	server := &http.Server{Handler: r.Handler(), ReadHeaderTimeout: 5 * time.Second, WriteTimeout: 5 * time.Second}
	serverErrors := make(chan error, 1)
	go func() { serverErrors <- server.Serve(listener) }()
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(stopCtx)
	}()

	backoff := time.Second
	for {
		interval, err := r.CheckIn(ctx)
		if err != nil {
			logger.Logger.Sugar().Warnf("Collector check-in failed: %v", err)
			interval = backoff
			backoff = min(backoff*2, 30*time.Second)
		} else {
			backoff = time.Second
		}
		timer := time.NewTimer(jitter(interval))
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil
		case err := <-serverErrors:
			timer.Stop()
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}
			return fmt.Errorf("health server: %w", err)
		case <-timer.C:
		}
	}
}

func jitter(interval time.Duration) time.Duration {
	return time.Duration(float64(interval) * (0.8 + rand.Float64()*0.4))
}

func (r *Runtime) CheckIn(ctx context.Context) (time.Duration, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	previous := r.manager.Status()
	if err := r.manager.Recover(ctx); err != nil {
		logger.Logger.Sugar().Warnf("Collector recovery failed: %v", err)
	}
	if previous.EngineStatus != "running" && r.manager.Status().EngineStatus == "running" {
		r.metrics.Reset()
	}
	token, err := r.config.token()
	if err != nil {
		return 0, err
	}
	metrics, _ := r.metrics.Sample(ctx)
	body, err := json.Marshal(checkInRequest{
		NodeName: r.config.NodeName, NodeUID: r.config.NodeUID, Namespace: r.config.Namespace,
		PodName: r.config.PodName, PodUID: r.config.PodUID, PodIP: r.config.PodIP,
		Version: r.config.Version, Platform: runtime.GOOS + "/" + runtime.GOARCH,
		Status: r.manager.Status(), Metrics: metrics,
	})
	if err != nil {
		return 0, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.url, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := r.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("send check-in: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("control plane returned HTTP %d", resp.StatusCode)
	}
	const responseLimit = 16 << 20
	data, err := io.ReadAll(io.LimitReader(resp.Body, responseLimit+1))
	if err != nil {
		return 0, err
	}
	if len(data) > responseLimit {
		return 0, errors.New("control-plane response exceeds size limit")
	}
	var response checkInResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return 0, fmt.Errorf("decode check-in response: %w", err)
	}
	if !revisionPattern.MatchString(response.DesiredRevision) {
		return 0, errors.New("control plane returned an invalid desired revision")
	}
	status := r.manager.Status()
	updated := false
	if response.DesiredRevision != status.AppliedRevision && response.DesiredRevision != status.AttemptedRevision {
		if response.Config == nil {
			return 0, errors.New("changed desired revision arrived without its configuration")
		}
		if err := r.manager.Apply(ctx, response.DesiredRevision, response.Config); err != nil {
			logger.Logger.Sugar().Warnf("Collector configuration update failed: %v", err)
		}
		r.metrics.Reset()
		updated = true
	}
	interval := r.config.PollInterval
	if response.PollIntervalSeconds > 0 {
		interval = time.Duration(min(max(response.PollIntervalSeconds, 5), 20)) * time.Second
	}
	if updated {
		// Report the result promptly, then resume the ordinary check-in cadence.
		interval = 100 * time.Millisecond
	}
	return interval, nil
}

func (r *Runtime) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"status":"alive"}`)
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, _ *http.Request) {
		status := r.manager.Status()
		w.Header().Set("Content-Type", "application/json")
		if status.EngineStatus != "running" {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		// Probe output contains no configuration or credentials.
		json.NewEncoder(w).Encode(map[string]string{"status": strings.ToLower(status.EngineStatus)})
	})
	return mux
}
