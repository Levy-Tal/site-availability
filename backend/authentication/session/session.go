package session

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"site-availability/logging"
)

// Session represents a user session
type Session struct {
	ID         string    `json:"id"`
	Username   string    `json:"username"`
	IsAdmin    bool      `json:"is_admin"`
	Roles      []string  `json:"roles"`
	Groups     []string  `json:"groups"`
	AuthMethod string    `json:"auth_method"` // "local" or "oidc"
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}

// Manager handles session storage and management
type Manager struct {
	sessions map[string]*Session
	mutex    sync.RWMutex
	timeout  time.Duration
}

// NewManager creates a new session manager
func NewManager(sessionTimeout time.Duration) *Manager {
	manager := &Manager{
		sessions: make(map[string]*Session),
		timeout:  sessionTimeout,
	}

	// Start cleanup routine
	go manager.cleanupExpiredSessions()

	return manager
}

// CreateSession creates a new session for the user
func (m *Manager) CreateSession(username string, isAdmin bool, roles, groups []string, authMethod string) (*Session, error) {
	logging.Logger.WithFields(map[string]interface{}{
		"username":    username,
		"is_admin":    isAdmin,
		"roles":       roles,
		"groups":      groups,
		"auth_method": authMethod,
	}).Debug("Creating new session")

	sessionID, err := generateSessionID()
	if err != nil {
		logging.Logger.WithError(err).Error("Failed to generate session ID")
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	now := time.Now()
	session := &Session{
		ID:         sessionID,
		Username:   username,
		IsAdmin:    isAdmin,
		Roles:      roles,
		Groups:     groups,
		AuthMethod: authMethod,
		CreatedAt:  now,
		ExpiresAt:  now.Add(m.timeout),
	}

	m.mutex.Lock()
	m.sessions[sessionID] = session
	m.mutex.Unlock()

	logging.Logger.WithFields(map[string]interface{}{
		"session_id":  "****", // Mask session ID for security
		"username":    username,
		"auth_method": authMethod,
		"expires_at":  session.ExpiresAt,
		"timeout":     m.timeout,
	}).Info("Session created successfully")

	return session, nil
}

// ValidateSession validates a session ID and returns the session if valid
func (m *Manager) ValidateSession(sessionID string) (*Session, bool) {
	logging.Logger.WithField("session_id", "****").Debug("Validating session")

	m.mutex.RLock()
	session, exists := m.sessions[sessionID]
	m.mutex.RUnlock()

	if !exists {
		logging.Logger.WithField("session_id", "****").Debug("Session not found")
		return nil, false
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		logging.Logger.WithFields(map[string]interface{}{
			"session_id": "****", // Mask session ID for security
			"expires_at": session.ExpiresAt,
			"now":        time.Now(),
		}).Debug("Session expired, deleting")
		m.DeleteSession(sessionID)
		return nil, false
	}

	logging.Logger.WithFields(map[string]interface{}{
		"session_id": "****", // Mask session ID for security
		"username":   session.Username,
		"expires_at": session.ExpiresAt,
	}).Debug("Session validated successfully")

	return session, true
}

// RefreshSession extends the session expiration time
func (m *Manager) RefreshSession(sessionID string) bool {
	logging.Logger.WithField("session_id", "****").Debug("Refreshing session")

	m.mutex.Lock()
	defer m.mutex.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		logging.Logger.WithField("session_id", "****").Debug("Session not found for refresh")
		return false
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		logging.Logger.WithField("session_id", "****").Debug("Session expired during refresh, deleting")
		delete(m.sessions, sessionID)
		return false
	}

	// Extend expiration
	oldExpiresAt := session.ExpiresAt
	session.ExpiresAt = time.Now().Add(m.timeout)

	logging.Logger.WithFields(map[string]interface{}{
		"session_id":     "****", // Mask session ID for security
		"old_expires_at": oldExpiresAt,
		"new_expires_at": session.ExpiresAt,
	}).Debug("Session refreshed successfully")

	return true
}

// DeleteSession removes a session
func (m *Manager) DeleteSession(sessionID string) {
	m.mutex.Lock()
	delete(m.sessions, sessionID)
	m.mutex.Unlock()
}

// GetSessionCount returns the number of active sessions
func (m *Manager) GetSessionCount() int {
	m.mutex.RLock()
	count := len(m.sessions)
	m.mutex.RUnlock()
	return count
}

// cleanupExpiredSessions runs periodically to remove expired sessions
func (m *Manager) cleanupExpiredSessions() {
	ticker := time.NewTicker(5 * time.Minute) // Cleanup every 5 minutes
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		m.mutex.Lock()
		for sessionID, session := range m.sessions {
			if now.After(session.ExpiresAt) {
				delete(m.sessions, sessionID)
			}
		}
		m.mutex.Unlock()
	}
}

// generateSessionID creates a cryptographically secure session ID
func generateSessionID() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// ParseTimeout parses a timeout string like "12h" or "30m" into time.Duration
func ParseTimeout(timeoutStr string) (time.Duration, error) {
	if timeoutStr == "" {
		return 12 * time.Hour, nil // Default to 12 hours
	}

	duration, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return 0, fmt.Errorf("invalid session timeout format: %w", err)
	}

	return duration, nil
}
