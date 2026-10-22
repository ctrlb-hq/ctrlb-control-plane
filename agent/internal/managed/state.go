package managed

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type savedState struct {
	InstallationID    string         `json:"installation_id"`
	NodeName          string         `json:"node_name"`
	WorkingConfig     map[string]any `json:"working_config,omitempty"`
	AppliedRevision   string         `json:"applied_revision,omitempty"`
	AttemptedRevision string         `json:"attempted_revision,omitempty"`
	ConfigError       string         `json:"config_error,omitempty"`
}

type stateStore struct{ path string }

func (s stateStore) load() (savedState, error) {
	var state savedState
	data, err := os.ReadFile(s.path)
	if err != nil {
		return state, err
	}
	err = json.Unmarshal(data, &state)
	return state, err
}

func (s stateStore) save(state savedState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return atomicWrite(s.path, data)
}

func atomicWrite(path string, data []byte) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".ctrlb-state-*")
	if err != nil {
		return err
	}
	temporary := file.Name()
	defer os.Remove(temporary)
	defer file.Close()
	if err := file.Chmod(0600); err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporary, path); err != nil {
		return err
	}
	dir, err := os.Open(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer dir.Close()
	if err := dir.Sync(); err != nil {
		return fmt.Errorf("sync state directory: %w", err)
	}
	return nil
}

func writeConfig(dir string, config map[string]any) (string, error) {
	data, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	file, err := os.CreateTemp(dir, ".collector-*.json")
	if err != nil {
		return "", err
	}
	path := file.Name()
	if err := file.Chmod(0600); err != nil {
		file.Close()
		os.Remove(path)
		return "", err
	}
	_, writeErr := file.Write(data)
	err = file.Close()
	if writeErr != nil || err != nil {
		os.Remove(path)
		return "", fmt.Errorf("write collector config: %w", errors.Join(writeErr, err))
	}
	return path, nil
}

func configRevision(config map[string]any) (string, error) {
	data, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("sha256:%x", sha256.Sum256(data)), nil
}
