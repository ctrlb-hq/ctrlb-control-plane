package sessionManager

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	Email     string
	ExpiresAt time.Time
}

type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

// CreateSession generates a new session for the given email
func (sm *SessionManager) CreateSession(email string) (string, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sessionID := uuid.NewString()
	sm.sessions[sessionID] = &Session{
		Email:     email,
		ExpiresAt: time.Now().Add(24 * time.Hour), // Session expires after 24 hours
	}

	return sessionID, nil
}

// GetSession checks if a session exists and is valid
func (sm *SessionManager) GetSession(sessionID string) (*Session, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists || time.Now().After(session.ExpiresAt) {
		return nil, errors.New("invalid or expired session")
	}
	return session, nil
}

// DeleteSession removes a session
func (sm *SessionManager) DeleteSession(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, sessionID)
}
