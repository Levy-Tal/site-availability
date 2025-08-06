package session

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSession_Struct(t *testing.T) {
	now := time.Now()
	session := &Session{
		ID:         "test-session-id",
		Username:   "testuser",
		IsAdmin:    true,
		Roles:      []string{"admin", "user"},
		Groups:     []string{"admins", "developers"},
		AuthMethod: "local",
		CreatedAt:  now,
		ExpiresAt:  now.Add(time.Hour),
	}

	assert.Equal(t, "test-session-id", session.ID)
	assert.Equal(t, "testuser", session.Username)
	assert.True(t, session.IsAdmin)
	assert.Equal(t, []string{"admin", "user"}, session.Roles)
	assert.Equal(t, []string{"admins", "developers"}, session.Groups)
	assert.Equal(t, "local", session.AuthMethod)
	assert.Equal(t, now, session.CreatedAt)
	assert.Equal(t, now.Add(time.Hour), session.ExpiresAt)
}

func TestNewManager(t *testing.T) {
	timeout := 30 * time.Minute
	manager := NewManager(timeout)

	assert.NotNil(t, manager)
	assert.Equal(t, timeout, manager.timeout)
	assert.NotNil(t, manager.sessions)
	assert.Equal(t, 0, len(manager.sessions))
}

func TestManager_CreateSession(t *testing.T) {
	manager := NewManager(1 * time.Hour)

	// Test successful session creation
	session, err := manager.CreateSession("testuser", true, []string{"admin"}, []string{"admins"}, "local")
	require.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, "testuser", session.Username)
	assert.True(t, session.IsAdmin)
	assert.Equal(t, []string{"admin"}, session.Roles)
	assert.Equal(t, []string{"admins"}, session.Groups)
	assert.Equal(t, "local", session.AuthMethod)
	assert.NotEmpty(t, session.ID)
	assert.True(t, session.CreatedAt.Before(time.Now().Add(time.Second)))
	assert.True(t, session.ExpiresAt.After(time.Now().Add(59*time.Minute)))

	// Verify session is stored
	assert.Equal(t, 1, manager.GetSessionCount())
}

func TestManager_ValidateSession(t *testing.T) {
	manager := NewManager(1 * time.Hour)

	// Create a session
	session, err := manager.CreateSession("testuser", false, []string{"user"}, []string{"users"}, "oidc")
	require.NoError(t, err)

	// Test valid session
	validSession, isValid := manager.ValidateSession(session.ID)
	assert.True(t, isValid)
	assert.Equal(t, session, validSession)

	// Test invalid session ID
	invalidSession, isValid := manager.ValidateSession("invalid-id")
	assert.False(t, isValid)
	assert.Nil(t, invalidSession)

	// Test expired session
	expiredManager := NewManager(-1 * time.Hour) // Negative timeout for immediate expiration
	expiredSession, err := expiredManager.CreateSession("testuser", false, []string{"user"}, []string{"users"}, "local")
	require.NoError(t, err)

	// Wait a moment to ensure expiration
	time.Sleep(10 * time.Millisecond)

	expiredValidSession, isValid := expiredManager.ValidateSession(expiredSession.ID)
	assert.False(t, isValid)
	assert.Nil(t, expiredValidSession)
}

func TestManager_RefreshSession(t *testing.T) {
	manager := NewManager(1 * time.Hour)

	// Create a session
	session, err := manager.CreateSession("testuser", false, []string{"user"}, []string{"users"}, "local")
	require.NoError(t, err)

	originalExpiresAt := session.ExpiresAt

	// Wait a moment to ensure time difference
	time.Sleep(10 * time.Millisecond)

	// Test successful refresh
	refreshed := manager.RefreshSession(session.ID)
	assert.True(t, refreshed)

	// Verify session expiration was extended
	validSession, isValid := manager.ValidateSession(session.ID)
	assert.True(t, isValid)
	assert.True(t, validSession.ExpiresAt.After(originalExpiresAt))

	// Test refresh of non-existent session
	refreshed = manager.RefreshSession("non-existent-id")
	assert.False(t, refreshed)

	// Test refresh of expired session
	expiredManager := NewManager(-1 * time.Hour)
	expiredSession, err := expiredManager.CreateSession("testuser", false, []string{"user"}, []string{"users"}, "local")
	require.NoError(t, err)

	// Wait a moment to ensure expiration
	time.Sleep(10 * time.Millisecond)

	refreshed = expiredManager.RefreshSession(expiredSession.ID)
	assert.False(t, refreshed)
}

func TestManager_DeleteSession(t *testing.T) {
	manager := NewManager(1 * time.Hour)

	// Create a session
	session, err := manager.CreateSession("testuser", false, []string{"user"}, []string{"users"}, "local")
	require.NoError(t, err)

	assert.Equal(t, 1, manager.GetSessionCount())

	// Delete the session
	manager.DeleteSession(session.ID)

	// Verify session is deleted
	assert.Equal(t, 0, manager.GetSessionCount())

	// Verify session is no longer valid
	validSession, isValid := manager.ValidateSession(session.ID)
	assert.False(t, isValid)
	assert.Nil(t, validSession)

	// Test deleting non-existent session (should not panic)
	manager.DeleteSession("non-existent-id")
}

func TestManager_GetSessionCount(t *testing.T) {
	manager := NewManager(1 * time.Hour)

	// Initially no sessions
	assert.Equal(t, 0, manager.GetSessionCount())

	// Create multiple sessions
	session1, err := manager.CreateSession("user1", false, []string{"user"}, []string{"users"}, "local")
	require.NoError(t, err)

	session2, err := manager.CreateSession("user2", true, []string{"admin"}, []string{"admins"}, "oidc")
	require.NoError(t, err)

	assert.Equal(t, 2, manager.GetSessionCount())

	// Delete one session
	manager.DeleteSession(session1.ID)
	assert.Equal(t, 1, manager.GetSessionCount())

	// Delete the other session
	manager.DeleteSession(session2.ID)
	assert.Equal(t, 0, manager.GetSessionCount())
}

func TestGenerateSessionID(t *testing.T) {
	// Test multiple session ID generations
	sessionIDs := make(map[string]bool)
	for i := 0; i < 100; i++ {
		sessionID, err := generateSessionID()
		require.NoError(t, err)
		assert.NotEmpty(t, sessionID)
		assert.Len(t, sessionID, 64) // 32 bytes = 64 hex characters
		assert.False(t, sessionIDs[sessionID], "Session ID should be unique")
		sessionIDs[sessionID] = true
	}
}

func TestParseTimeout(t *testing.T) {
	tests := []struct {
		name        string
		timeoutStr  string
		expected    time.Duration
		expectError bool
	}{
		{
			name:        "empty string defaults to 12 hours",
			timeoutStr:  "",
			expected:    12 * time.Hour,
			expectError: false,
		},
		{
			name:        "valid hours",
			timeoutStr:  "6h",
			expected:    6 * time.Hour,
			expectError: false,
		},
		{
			name:        "valid minutes",
			timeoutStr:  "30m",
			expected:    30 * time.Minute,
			expectError: false,
		},
		{
			name:        "valid seconds",
			timeoutStr:  "45s",
			expected:    45 * time.Second,
			expectError: false,
		},
		{
			name:        "valid combination",
			timeoutStr:  "1h30m",
			expected:    1*time.Hour + 30*time.Minute,
			expectError: false,
		},
		{
			name:        "invalid format",
			timeoutStr:  "invalid",
			expected:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTimeout(tt.timeoutStr)
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, time.Duration(0), result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestManager_CleanupExpiredSessions(t *testing.T) {
	// Create manager with short timeout for testing
	manager := NewManager(10 * time.Millisecond)

	// Create a session
	session, err := manager.CreateSession("testuser", false, []string{"user"}, []string{"users"}, "local")
	require.NoError(t, err)

	// Verify session exists
	assert.Equal(t, 1, manager.GetSessionCount())

	// Wait for session to expire
	time.Sleep(20 * time.Millisecond)

	// The cleanup routine runs every 5 minutes, so we'll manually trigger cleanup
	// by calling ValidateSession which should delete expired sessions
	validSession, isValid := manager.ValidateSession(session.ID)
	assert.False(t, isValid)
	assert.Nil(t, validSession)

	// Verify session was cleaned up
	assert.Equal(t, 0, manager.GetSessionCount())
}

func TestManager_ConcurrentAccess(t *testing.T) {
	manager := NewManager(1 * time.Hour)
	const numGoroutines = 10
	const sessionsPerGoroutine = 10

	// Test concurrent session creation
	done := make(chan bool, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < sessionsPerGoroutine; j++ {
				username := fmt.Sprintf("user-%d-%d", id, j)
				session, err := manager.CreateSession(username, false, []string{"user"}, []string{"users"}, "local")
				assert.NoError(t, err)
				assert.NotNil(t, session)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify all sessions were created
	expectedCount := numGoroutines * sessionsPerGoroutine
	assert.Equal(t, expectedCount, manager.GetSessionCount())
}

func TestManager_SessionExpiration(t *testing.T) {
	// Test with very short timeout
	manager := NewManager(1 * time.Millisecond)

	// Create session
	session, err := manager.CreateSession("testuser", false, []string{"user"}, []string{"users"}, "local")
	require.NoError(t, err)

	// Session should be valid immediately
	validSession, isValid := manager.ValidateSession(session.ID)
	assert.True(t, isValid)
	assert.Equal(t, session, validSession)

	// Wait for session to expire
	time.Sleep(10 * time.Millisecond)

	// Session should now be expired
	validSession, isValid = manager.ValidateSession(session.ID)
	assert.False(t, isValid)
	assert.Nil(t, validSession)
}
