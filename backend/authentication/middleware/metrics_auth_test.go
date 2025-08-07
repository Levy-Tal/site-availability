package middleware

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"site-availability/config"
)

func TestMetricsAuthMiddleware_Disabled(t *testing.T) {
	// Create config with metrics auth disabled
	cfg := &config.Config{
		ServerSettings: config.ServerSettings{
			MetricsAuth: config.MetricsAuthConfig{
				Enabled: false,
			},
		},
	}

	middleware := NewMetricsAuthMiddleware(cfg)
	handler := middleware.RequireMetricsAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "success" {
		t.Errorf("Expected body 'success', got '%s'", w.Body.String())
	}
}

func TestMetricsAuthMiddleware_BasicAuth_Success(t *testing.T) {
	// Create config with basic auth enabled
	cfg := &config.Config{
		ServerSettings: config.ServerSettings{
			MetricsAuth: config.MetricsAuthConfig{
				Enabled:  true,
				Type:     "basic",
				Username: "prometheus",
				Password: "secret",
			},
		},
	}

	middleware := NewMetricsAuthMiddleware(cfg)
	handler := middleware.RequireMetricsAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	// Create request with valid basic auth
	credentials := base64.StdEncoding.EncodeToString([]byte("prometheus:secret"))
	req := httptest.NewRequest("GET", "/metrics", nil)
	req.Header.Set("Authorization", "Basic "+credentials)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "success" {
		t.Errorf("Expected body 'success', got '%s'", w.Body.String())
	}
}

func TestMetricsAuthMiddleware_BasicAuth_Failure(t *testing.T) {
	// Create config with basic auth enabled
	cfg := &config.Config{
		ServerSettings: config.ServerSettings{
			MetricsAuth: config.MetricsAuthConfig{
				Enabled:  true,
				Type:     "basic",
				Username: "prometheus",
				Password: "secret",
			},
		},
	}

	middleware := NewMetricsAuthMiddleware(cfg)
	handler := middleware.RequireMetricsAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	// Create request with invalid basic auth
	credentials := base64.StdEncoding.EncodeToString([]byte("prometheus:wrong"))
	req := httptest.NewRequest("GET", "/metrics", nil)
	req.Header.Set("Authorization", "Basic "+credentials)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	if !strings.Contains(w.Header().Get("WWW-Authenticate"), "Basic") {
		t.Errorf("Expected WWW-Authenticate header with Basic, got '%s'", w.Header().Get("WWW-Authenticate"))
	}
}

func TestMetricsAuthMiddleware_BasicAuth_NoHeader(t *testing.T) {
	// Create config with basic auth enabled
	cfg := &config.Config{
		ServerSettings: config.ServerSettings{
			MetricsAuth: config.MetricsAuthConfig{
				Enabled:  true,
				Type:     "basic",
				Username: "prometheus",
				Password: "secret",
			},
		},
	}

	middleware := NewMetricsAuthMiddleware(cfg)
	handler := middleware.RequireMetricsAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	// Create request without Authorization header
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	if !strings.Contains(w.Header().Get("WWW-Authenticate"), "Basic") {
		t.Errorf("Expected WWW-Authenticate header with Basic, got '%s'", w.Header().Get("WWW-Authenticate"))
	}
}

func TestMetricsAuthMiddleware_BearerToken_Success(t *testing.T) {
	// Create config with bearer token enabled
	cfg := &config.Config{
		ServerSettings: config.ServerSettings{
			MetricsAuth: config.MetricsAuthConfig{
				Enabled: true,
				Type:    "bearer",
				Token:   "test-token-123",
			},
		},
	}

	middleware := NewMetricsAuthMiddleware(cfg)
	handler := middleware.RequireMetricsAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	// Create request with valid bearer token
	req := httptest.NewRequest("GET", "/metrics", nil)
	req.Header.Set("Authorization", "Bearer test-token-123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "success" {
		t.Errorf("Expected body 'success', got '%s'", w.Body.String())
	}
}

func TestMetricsAuthMiddleware_BearerToken_Failure(t *testing.T) {
	// Create config with bearer token enabled
	cfg := &config.Config{
		ServerSettings: config.ServerSettings{
			MetricsAuth: config.MetricsAuthConfig{
				Enabled: true,
				Type:    "bearer",
				Token:   "test-token-123",
			},
		},
	}

	middleware := NewMetricsAuthMiddleware(cfg)
	handler := middleware.RequireMetricsAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	// Create request with invalid bearer token
	req := httptest.NewRequest("GET", "/metrics", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	if !strings.Contains(w.Header().Get("WWW-Authenticate"), "Bearer") {
		t.Errorf("Expected WWW-Authenticate header with Bearer, got '%s'", w.Header().Get("WWW-Authenticate"))
	}
}

func TestMetricsAuthMiddleware_BearerToken_NoHeader(t *testing.T) {
	// Create config with bearer token enabled
	cfg := &config.Config{
		ServerSettings: config.ServerSettings{
			MetricsAuth: config.MetricsAuthConfig{
				Enabled: true,
				Type:    "bearer",
				Token:   "test-token-123",
			},
		},
	}

	middleware := NewMetricsAuthMiddleware(cfg)
	handler := middleware.RequireMetricsAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	// Create request without Authorization header
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	if !strings.Contains(w.Header().Get("WWW-Authenticate"), "Bearer") {
		t.Errorf("Expected WWW-Authenticate header with Bearer, got '%s'", w.Header().Get("WWW-Authenticate"))
	}
}

func TestMetricsAuthMiddleware_InvalidType(t *testing.T) {
	// Create config with invalid auth type
	cfg := &config.Config{
		ServerSettings: config.ServerSettings{
			MetricsAuth: config.MetricsAuthConfig{
				Enabled: true,
				Type:    "invalid",
			},
		},
	}

	middleware := NewMetricsAuthMiddleware(cfg)
	handler := middleware.RequireMetricsAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}
