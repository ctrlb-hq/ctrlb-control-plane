package managed

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/adapters"
)

var revisionPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

type Engine interface {
	Validate(context.Context, string) error
	Start(context.Context, string) error
	Stop(context.Context) error
	Status() adapters.EngineStatus
}

type Status struct {
	EngineStatus      string `json:"engine_status"`
	AppliedRevision   string `json:"applied_revision"`
	AttemptedRevision string `json:"attempted_revision"`
	ConfigError       string `json:"config_error"`
}

type Manager struct {
	opMu           sync.Mutex
	mu             sync.RWMutex
	engine         Engine
	dir            string
	store          stateStore
	state          savedState
	active         string
	configPath     string
	recoverAt      time.Time
	recoverBackoff time.Duration
	recoveryError  string
	// save is replaceable in tests to exercise disk failure after engine startup.
	save func(savedState) error
}

func NewManager(dir, installationID, nodeName string, engine Engine) (*Manager, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("create state directory: %w", err)
	}
	store := stateStore{path: filepath.Join(dir, "state.json")}
	return &Manager{
		engine: engine, dir: dir, store: store, save: store.save,
		state: savedState{InstallationID: installationID, NodeName: nodeName},
	}, nil
}

// Restore starts the saved working config before the control plane is needed.
// An unusable cache leaves empty revisions so check-in requests fresh config.
func (m *Manager) Restore(ctx context.Context) error {
	m.opMu.Lock()
	defer m.opMu.Unlock()
	state, err := m.store.load()
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return m.cacheFailure(fmt.Errorf("read saved configuration: %w", err))
	}
	if state.InstallationID != m.state.InstallationID || state.NodeName != m.state.NodeName {
		return m.cacheFailure(errors.New("saved configuration belongs to another installation or node"))
	}
	if state.WorkingConfig == nil || !revisionPattern.MatchString(state.AppliedRevision) {
		return nil
	}
	if revision, err := configRevision(state.WorkingConfig); err != nil || revision != state.AppliedRevision ||
		(state.AttemptedRevision != "" && !revisionPattern.MatchString(state.AttemptedRevision)) {
		return m.cacheFailure(errors.New("saved configuration revision does not match its contents"))
	}
	path, err := writeConfig(m.dir, state.WorkingConfig)
	if err != nil {
		return m.cacheFailure(err)
	}
	if err = m.engine.Validate(ctx, path); err == nil {
		err = m.engine.Start(ctx, path)
	}
	if err != nil {
		os.Remove(path)
		return m.cacheFailure(fmt.Errorf("start saved configuration: %w", err))
	}
	m.mu.Lock()
	m.state, m.active, m.configPath = state, state.AppliedRevision, path
	m.mu.Unlock()
	return nil
}

func (m *Manager) cacheFailure(err error) error {
	m.mu.Lock()
	m.state.ConfigError = err.Error()
	m.mu.Unlock()
	return err
}

func (m *Manager) Apply(ctx context.Context, revision string, config map[string]any) error {
	m.opMu.Lock()
	defer m.opMu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if !revisionPattern.MatchString(revision) || config == nil {
		return errors.New("control plane returned an invalid revision or missing configuration")
	}
	m.mu.RLock()
	old, oldPath := m.state, m.configPath
	m.mu.RUnlock()
	if actual, err := configRevision(config); err != nil || actual != revision {
		return m.fail(old, revision, errors.New("desired revision does not match its configuration contents"))
	}
	if revision == old.AttemptedRevision {
		return nil
	}
	path, err := writeConfig(m.dir, config)
	if err != nil {
		return m.fail(old, revision, fmt.Errorf("stage configuration: %w", err))
	}
	removeCandidate := true
	defer func() {
		if removeCandidate {
			os.Remove(path)
		}
	}()
	if err := m.engine.Validate(ctx, path); err != nil {
		if ctx.Err() != nil {
			return err
		}
		return m.fail(old, revision, fmt.Errorf("validate configuration: %w", err))
	}
	m.mu.Lock()
	m.active = ""
	m.mu.Unlock()
	if err := m.stop(); err != nil {
		if ctx.Err() != nil {
			return err
		}
		return m.fail(old, revision, err)
	}
	if err := m.engine.Start(ctx, path); err != nil {
		if ctx.Err() != nil {
			// Shutdown during an update keeps the old durable state and does not
			// suppress this desired revision when the replacement pod checks in.
			return errors.Join(ctx.Err(), m.stop())
		}
		return m.rollback(old, oldPath, revision, fmt.Errorf("start configuration: %w", err))
	}
	if !m.engine.Status().Running {
		return m.rollback(old, oldPath, revision, errors.New("Collector stopped before configuration could be committed"))
	}
	next := savedState{
		InstallationID: old.InstallationID, NodeName: old.NodeName,
		WorkingConfig: config, AppliedRevision: revision, AttemptedRevision: revision,
	}
	if err := m.save(next); err != nil {
		return m.rollback(old, oldPath, revision, fmt.Errorf("persist working configuration: %w", err))
	}
	m.mu.Lock()
	m.state, m.active, m.configPath = next, revision, path
	m.recoveryError, m.recoverAt, m.recoverBackoff = "", time.Time{}, 0
	m.mu.Unlock()
	removeCandidate = false
	if oldPath != "" {
		os.Remove(oldPath)
	}
	return nil
}

func (m *Manager) stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return m.engine.Stop(ctx)
}

func (m *Manager) rollback(old savedState, oldPath, revision string, cause error) error {
	if err := m.stop(); err != nil {
		return m.fail(old, revision, errors.Join(cause, fmt.Errorf("stop failed configuration: %w", err)))
	}
	if old.WorkingConfig != nil && oldPath != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := m.engine.Start(ctx, oldPath); err != nil {
			cause = errors.Join(cause, fmt.Errorf("restore previous configuration: %w", err))
		} else {
			m.mu.Lock()
			m.active = old.AppliedRevision
			m.mu.Unlock()
		}
	}
	return m.fail(old, revision, cause)
}

func (m *Manager) fail(old savedState, revision string, cause error) error {
	old.AttemptedRevision, old.ConfigError = revision, limitError(cause.Error())
	if err := m.save(old); err != nil {
		cause = errors.Join(cause, fmt.Errorf("persist failed attempt: %w", err))
		old.ConfigError = limitError(cause.Error())
	}
	m.mu.Lock()
	m.state = old
	m.mu.Unlock()
	return cause
}

func (m *Manager) Status() Status {
	m.mu.RLock()
	defer m.mu.RUnlock()
	engine := m.engine.Status()
	status := Status{EngineStatus: "waiting", AttemptedRevision: m.state.AttemptedRevision, ConfigError: m.state.ConfigError}
	if engine.Running {
		status.EngineStatus, status.AppliedRevision = "running", m.active
	} else if engine.Error != "" {
		status.EngineStatus = "failed"
		if status.ConfigError != "" {
			status.ConfigError += "; "
		}
		status.ConfigError += engine.Error
	} else if m.state.AttemptedRevision != "" {
		status.EngineStatus = "failed"
		if status.ConfigError == "" {
			status.ConfigError = "Collector is not running"
		}
	}
	if !engine.Running && m.recoveryError != "" {
		status.EngineStatus = "failed"
		if status.ConfigError != "" {
			status.ConfigError += "; "
		}
		status.ConfigError += m.recoveryError
	}
	status.ConfigError = limitError(status.ConfigError)
	return status
}

// Recover retries only a previously committed working configuration. A failed
// candidate is never restarted here, and retries remain bounded during outages.
func (m *Manager) Recover(ctx context.Context) error {
	m.opMu.Lock()
	defer m.opMu.Unlock()
	m.mu.RLock()
	state, oldPath, next := m.state, m.configPath, m.recoverAt
	m.mu.RUnlock()
	if m.engine.Status().Running || state.WorkingConfig == nil || time.Now().Before(next) {
		return nil
	}
	m.mu.Lock()
	if m.recoverBackoff == 0 {
		m.recoverBackoff = time.Second
	}
	m.recoverAt = time.Now().Add(m.recoverBackoff)
	m.recoverBackoff = min(m.recoverBackoff*2, 30*time.Second)
	m.active = ""
	m.mu.Unlock()
	path, err := writeConfig(m.dir, state.WorkingConfig)
	if err == nil {
		err = m.stop()
	}
	if err == nil {
		err = m.engine.Start(ctx, path)
	}
	if err != nil {
		os.Remove(path)
		m.mu.Lock()
		m.recoveryError = limitError(fmt.Sprintf("restart working Collector configuration: %v", err))
		m.mu.Unlock()
		return err
	}
	m.mu.Lock()
	m.active, m.configPath, m.recoveryError = state.AppliedRevision, path, ""
	m.recoverBackoff, m.recoverAt = 0, time.Time{}
	m.mu.Unlock()
	if oldPath != "" {
		os.Remove(oldPath)
	}
	return nil
}

func limitError(message string) string {
	if len(message) <= 8000 {
		return message
	}
	message = message[:8000]
	for !utf8.ValidString(message) {
		message = message[:len(message)-1]
	}
	return message
}

func (m *Manager) Close(ctx context.Context) error {
	m.opMu.Lock()
	defer m.opMu.Unlock()
	err := m.engine.Stop(ctx)
	if err == nil && m.configPath != "" {
		os.Remove(m.configPath)
	}
	return err
}
