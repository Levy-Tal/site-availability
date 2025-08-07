package authhandlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"site-availability/authentication/middleware"
	"site-availability/authentication/session"
	"site-availability/config"
	"site-availability/logging"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAuthHandlers(t *testing.T) {
	// Initialize logger for tests
	err := logging.Init()
	require.NoError(t, err)

	// Create test config
	cfg := &config.Config{
		ServerSettings: config.ServerSettings{
			HostURL:           "http://localhost:8080",
			SessionTimeout:    "1h",
			TrustProxyHeaders: false,
			LocalAdmin: config.LocalAdminConfig{
				Enabled:  true,
				Username: "admin",
				Password: "password",
			},
		},
	}

	// Create session manager
	sessionTimeout, _ := session.ParseTimeout("1h")
	sessionManager := session.NewManager(sessionTimeout)

	t.Run("create auth handlers successfully", func(t *testing.T) {
		handlers, err := NewAuthHandlers(cfg, sessionManager)
		require.NoError(t, err)
		assert.NotNil(t, handlers)
		assert.NotNil(t, handlers.config)
		assert.NotNil(t, handlers.sessionManager)
		assert.NotNil(t, handlers.localAuth)
		assert.NotNil(t, handlers.oidcAuth)
	})
}

func TestHandleLogin(t *testing.T) {
	// Initialize logger for tests
	err := logging.Init()
	require.NoError(t, err)

	// Create test config with local admin enabled
	cfg := &config.Config{
		ServerSettings: config.ServerSettings{
			HostURL:           "http://localhost:8080",
			SessionTimeout:    "1h",
			TrustProxyHeaders: false,
			LocalAdmin: config.LocalAdminConfig{
				Enabled:  true,
				Username: "admin",
				Password: "password",
			},
		},
	}

	sessionTimeout, _ := session.ParseTimeout("1h")
	sessionManager := session.NewManager(sessionTimeout)
	handlers, err := NewAuthHandlers(cfg, sessionManager)
	require.NoError(t, err)

	t.Run("successful login", func(t *testing.T) {
		loginReq := LoginRequest{
			Username: "admin",
			Password: "password",
		}

		body, err := json.Marshal(loginReq)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handlers.HandleLogin(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response LoginResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.Equal(t, "Login successful", response.Message)

		// Check that session cookie was set
		cookies := w.Result().Cookies()
		var sessionCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "session_id" {
				sessionCookie = cookie
				break
			}
		}
		assert.NotNil(t, sessionCookie)
		assert.NotEmpty(t, sessionCookie.Value)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		loginReq := LoginRequest{
			Username: "admin",
			Password: "wrongpassword",
		}

		body, err := json.Marshal(loginReq)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handlers.HandleLogin(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "Invalid credentials", response.Error)
	})

	t.Run("invalid request method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
		w := httptest.NewRecorder()

		handlers.HandleLogin(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestHandleUser(t *testing.T) {
	// Initialize logger for tests
	err := logging.Init()
	require.NoError(t, err)

	// Create test config
	cfg := &config.Config{
		ServerSettings: config.ServerSettings{
			HostURL:           "http://localhost:8080",
			SessionTimeout:    "1h",
			TrustProxyHeaders: false,
			LocalAdmin: config.LocalAdminConfig{
				Enabled:  true,
				Username: "admin",
				Password: "password",
			},
		},
	}

	sessionTimeout, _ := session.ParseTimeout("1h")
	sessionManager := session.NewManager(sessionTimeout)
	handlers, err := NewAuthHandlers(cfg, sessionManager)
	require.NoError(t, err)

	t.Run("get user info with valid session", func(t *testing.T) {
		// Create a session first
		sessionInfo, err := sessionManager.CreateSession("admin", true, []string{"admin"}, []string{"admins"}, "local")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/auth/user", nil)
		w := httptest.NewRecorder()

		// Add user to context
		ctx := context.WithValue(req.Context(), middleware.UserContextKey, sessionInfo)
		req = req.WithContext(ctx)

		handlers.HandleUser(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response UserResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "admin", response.User.Username)
		assert.True(t, response.User.IsAdmin)
		assert.Equal(t, "local", response.User.AuthMethod)
		assert.Equal(t, []string{"admin"}, response.User.Roles)
		assert.Equal(t, []string{"admins"}, response.User.Groups)
		assert.NotEmpty(t, response.Session.ExpiresAt)
	})

	t.Run("get user info without session", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/user", nil)
		w := httptest.NewRecorder()

		handlers.HandleUser(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestHandleLogout(t *testing.T) {
	// Initialize logger for tests
	err := logging.Init()
	require.NoError(t, err)

	// Create test config
	cfg := &config.Config{
		ServerSettings: config.ServerSettings{
			HostURL:           "http://localhost:8080",
			SessionTimeout:    "1h",
			TrustProxyHeaders: false,
		},
	}

	sessionTimeout, _ := session.ParseTimeout("1h")
	sessionManager := session.NewManager(sessionTimeout)
	handlers, err := NewAuthHandlers(cfg, sessionManager)
	require.NoError(t, err)

	t.Run("successful logout with session", func(t *testing.T) {
		// Create a session first
		sessionInfo, err := sessionManager.CreateSession("admin", true, []string{"admin"}, []string{"admins"}, "local")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
		w := httptest.NewRecorder()

		// Add session cookie
		req.AddCookie(&http.Cookie{
			Name:  "session_id",
			Value: sessionInfo.ID,
		})

		handlers.HandleLogout(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))
		assert.Equal(t, "Logout successful", response["message"])

		// Check that session cookie was deleted
		cookies := w.Result().Cookies()
		var sessionCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "session_id" {
				sessionCookie = cookie
				break
			}
		}
		assert.NotNil(t, sessionCookie)
		assert.Equal(t, -1, sessionCookie.MaxAge) // Cookie should be deleted
	})
}

func TestHandleAuthConfig(t *testing.T) {
	// Initialize logger for tests
	err := logging.Init()
	require.NoError(t, err)

	t.Run("auth config with local admin enabled", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				HostURL:           "http://localhost:8080",
				SessionTimeout:    "1h",
				TrustProxyHeaders: false,
				LocalAdmin: config.LocalAdminConfig{
					Enabled:  true,
					Username: "admin",
					Password: "password",
				},
			},
		}

		sessionTimeout, _ := session.ParseTimeout("1h")
		sessionManager := session.NewManager(sessionTimeout)
		handlers, err := NewAuthHandlers(cfg, sessionManager)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/auth/config", nil)
		w := httptest.NewRecorder()

		handlers.HandleAuthConfig(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["auth_enabled"].(bool))
		authMethods := response["auth_methods"].([]interface{})
		assert.Contains(t, authMethods, "local")
	})
}

func TestGetSessionCount(t *testing.T) {
	// Initialize logger for tests
	err := logging.Init()
	require.NoError(t, err)

	cfg := &config.Config{
		ServerSettings: config.ServerSettings{
			HostURL:           "http://localhost:8080",
			SessionTimeout:    "1h",
			TrustProxyHeaders: false,
		},
	}

	sessionTimeout, _ := session.ParseTimeout("1h")
	sessionManager := session.NewManager(sessionTimeout)
	handlers, err := NewAuthHandlers(cfg, sessionManager)
	require.NoError(t, err)

	t.Run("get session count", func(t *testing.T) {
		// Initially should be 0
		assert.Equal(t, 0, handlers.GetSessionCount())

		// Create a session
		_, err := sessionManager.CreateSession("testuser", false, []string{}, []string{}, "local")
		require.NoError(t, err)

		// Should be 1
		assert.Equal(t, 1, handlers.GetSessionCount())
	})
}
