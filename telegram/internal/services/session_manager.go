package services

import (
	"sync"
	"time"
	"ung-telegram/internal/models"
)

// SessionManager manages user conversation states
type SessionManager struct {
	sessions map[int64]*models.Session
	users    map[int64]*models.User
	mu       sync.RWMutex
}

// NewSessionManager creates a new session manager
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[int64]*models.Session),
		users:    make(map[int64]*models.User),
	}
}

// GetSession retrieves a session for a user
func (sm *SessionManager) GetSession(telegramID int64) *models.Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if session, exists := sm.sessions[telegramID]; exists {
		return session
	}

	return nil
}

// SetSession stores a session for a user
func (sm *SessionManager) SetSession(session *models.Session) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session.UpdatedAt = time.Now()
	sm.sessions[session.TelegramID] = session
}

// ClearSession removes a session
func (sm *SessionManager) ClearSession(telegramID int64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.sessions, telegramID)
}

// GetUser retrieves a user
func (sm *SessionManager) GetUser(telegramID int64) *models.User {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.users[telegramID]
}

// SetUser stores a user
func (sm *SessionManager) SetUser(user *models.User) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.users[user.TelegramID] = user
}

// IsAuthenticated checks if user is authenticated
func (sm *SessionManager) IsAuthenticated(telegramID int64) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	user, exists := sm.users[telegramID]
	return exists && user.APIToken != ""
}
