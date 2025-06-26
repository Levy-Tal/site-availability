package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"site-availability/config"
	"site-availability/handlers"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupServerTest clears the handlers cache for test isolation
func setupServerTest() {
	// Clear handlers cache to ensure test isolation
	handlers.UpdateAppStatus("test-source", []handlers.AppStatus{})
}

func TestNewServer(t *testing.T) {
	t.Run("create new server instance", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				Port:       "8080",
				SyncEnable: true,
			},
		}

		server := NewServer(cfg)
		assert.NotNil(t, server)
		assert.Equal(t, cfg, server.config)
		assert.NotNil(t, server.mux)
	})

	t.Run("create server with minimal config", func(t *testing.T) {
		cfg := &config.Config{}
		server := NewServer(cfg)
		assert.NotNil(t, server)
		assert.Equal(t, cfg, server.config)
	})
}

func TestSetupRoutes(t *testing.T) {
	setupServerTest()

	t.Run("setup routes with static directory", func(t *testing.T) {
		// Create temporary directory for static files
		tmpDir := t.TempDir()
		staticDir := filepath.Join(tmpDir, "static")
		err := os.Mkdir(staticDir, 0755)
		require.NoError(t, err)

		// Create a test static file
		err = os.WriteFile(filepath.Join(staticDir, "index.html"), []byte("test content"), 0644)
		require.NoError(t, err)

		// Change to the temporary directory
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() {
			err := os.Chdir(originalDir)
			require.NoError(t, err)
		}()

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				SyncEnable: true,
			},
			Scraping: config.ScrapingSettings{
				Interval:    "60s",
				Timeout:     "5s",
				MaxParallel: 5,
			},
			Documentation: config.Documentation{
				Title: "Test Docs",
				URL:   "https://test.example.com/docs",
			},
		}

		server := NewServer(cfg)
		server.setupRoutes()

		// Test API routes
		apiTestCases := []struct {
			path       string
			method     string
			statusCode int
		}{
			{"/api/status", "GET", http.StatusOK},
			{"/api/scrape-interval", "GET", http.StatusOK},
			{"/api/docs", "GET", http.StatusOK},
			{"/healthz", "GET", http.StatusOK},
			{"/readyz", "GET", http.StatusServiceUnavailable}, // Initially not ready
			{"/metrics", "GET", http.StatusOK},
			{"/sync", "GET", http.StatusOK}, // Sync is enabled
		}

		for _, tc := range apiTestCases {
			t.Run(tc.path, func(t *testing.T) {
				req := httptest.NewRequest(tc.method, tc.path, nil)
				w := httptest.NewRecorder()
				server.mux.ServeHTTP(w, req)
				assert.Equal(t, tc.statusCode, w.Code, "Path: %s", tc.path)
			})
		}

		// Test static file serving
		t.Run("static file serving", func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()
			server.mux.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		})
	})

	t.Run("setup routes without static directory", func(t *testing.T) {
		// Use a temporary directory that doesn't have a static subdirectory
		tmpDir := t.TempDir()
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() {
			err := os.Chdir(originalDir)
			require.NoError(t, err)
		}()

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				SyncEnable: false, // Test with sync disabled
			},
			Scraping: config.ScrapingSettings{
				Interval: "30s",
			},
		}

		server := NewServer(cfg)
		server.setupRoutes()

		// Test that root path works with fallback handler when static dir doesn't exist
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		server.mux.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// Test that unmatched routes fall back to static handler when sync is disabled
		// The server creates a fallback handler that returns 200 for any unmatched route
		req = httptest.NewRequest("GET", "/sync", nil)
		w = httptest.NewRecorder()
		server.mux.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestServerLifecycle(t *testing.T) {
	t.Run("graceful shutdown of non-running server", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				Port:       "0", // Use port 0 to avoid conflicts
				SyncEnable: true,
			},
		}

		server := NewServer(cfg)

		// Create a test HTTP server that's not actually running
		srv := &http.Server{
			Addr:    ":0",
			Handler: server.mux,
		}

		// Test graceful shutdown of non-running server
		err := server.gracefulShutdown(srv)
		assert.NoError(t, err)
	})

	t.Run("server start with immediate shutdown", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				Port: "0", // Use port 0 for dynamic port allocation
			},
		}

		server := NewServer(cfg)
		server.setupRoutes()

		// Create a test server that we can control
		srv := &http.Server{
			Addr:    ":0",
			Handler: server.mux,
		}

		// Test that graceful shutdown works
		err := server.gracefulShutdown(srv)
		assert.NoError(t, err)
	})
}

func TestHealthProbes(t *testing.T) {
	setupServerTest()

	t.Run("liveness probe", func(t *testing.T) {
		server := NewServer(&config.Config{})
		req := httptest.NewRequest("GET", "/healthz", nil)
		w := httptest.NewRecorder()

		server.livenessProbe(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "OK", w.Body.String())
	})

	t.Run("readiness probe when cache is empty", func(t *testing.T) {
		server := NewServer(&config.Config{})

		req := httptest.NewRequest("GET", "/readyz", nil)
		w := httptest.NewRecorder()
		server.readinessProbe(w, req)
		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		assert.Equal(t, "NOT READY", w.Body.String())
	})

	t.Run("readiness probe when cache has data", func(t *testing.T) {
		server := NewServer(&config.Config{})

		// Add test data to cache
		handlers.UpdateAppStatus("test-source", []handlers.AppStatus{
			{
				Name:     "test-app",
				Location: "test-location",
				Status:   "up",
				Source:   "test-source",
			},
		})

		req := httptest.NewRequest("GET", "/readyz", nil)
		w := httptest.NewRecorder()
		server.readinessProbe(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "OK", w.Body.String())

		// Clean up
		handlers.UpdateAppStatus("test-source", []handlers.AppStatus{})
	})
}

func TestRouteHandlers(t *testing.T) {
	setupServerTest()

	t.Run("api status endpoint", func(t *testing.T) {
		cfg := &config.Config{
			Locations: []config.Location{
				{
					Name:      "test location",
					Latitude:  31.782904,
					Longitude: 35.214774,
				},
			},
		}

		server := NewServer(cfg)
		server.setupRoutes()

		req := httptest.NewRequest("GET", "/api/status", nil)
		w := httptest.NewRecorder()
		server.mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	})

	t.Run("api scrape interval endpoint", func(t *testing.T) {
		cfg := &config.Config{
			Scraping: config.ScrapingSettings{
				Interval: "60s",
			},
		}

		server := NewServer(cfg)
		server.setupRoutes()

		req := httptest.NewRequest("GET", "/api/scrape-interval", nil)
		w := httptest.NewRecorder()
		server.mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	})

	t.Run("api docs endpoint", func(t *testing.T) {
		cfg := &config.Config{
			Documentation: config.Documentation{
				Title: "Test Documentation",
				URL:   "https://test.example.com/docs",
			},
		}

		server := NewServer(cfg)
		server.setupRoutes()

		req := httptest.NewRequest("GET", "/api/docs", nil)
		w := httptest.NewRecorder()
		server.mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	})

	t.Run("metrics endpoint", func(t *testing.T) {
		server := NewServer(&config.Config{})
		server.setupRoutes()

		req := httptest.NewRequest("GET", "/metrics", nil)
		w := httptest.NewRecorder()
		server.mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("sync endpoint when enabled", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				SyncEnable: true,
			},
		}

		server := NewServer(cfg)
		server.setupRoutes()

		req := httptest.NewRequest("GET", "/sync", nil)
		w := httptest.NewRecorder()
		server.mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	})

	t.Run("sync endpoint when disabled", func(t *testing.T) {
		// Create a temporary directory without static files to test the fallback
		tmpDir := t.TempDir()
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() {
			err := os.Chdir(originalDir)
			require.NoError(t, err)
		}()

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				SyncEnable: false,
			},
		}

		server := NewServer(cfg)
		server.setupRoutes()

		req := httptest.NewRequest("GET", "/sync", nil)
		w := httptest.NewRecorder()
		server.mux.ServeHTTP(w, req)

		// When sync is disabled and no static dir exists, the fallback handler returns 200
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestServerBehavior(t *testing.T) {
	setupServerTest()

	t.Run("static file fallback behavior", func(t *testing.T) {
		// Create a temporary directory without static files
		tmpDir := t.TempDir()
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() {
			err := os.Chdir(originalDir)
			require.NoError(t, err)
		}()

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		// Provide proper configuration to avoid 500 errors
		cfg := &config.Config{
			Scraping: config.ScrapingSettings{
				Interval: "60s", // Provide valid interval
			},
			Documentation: config.Documentation{
				Title: "Test Docs",
				URL:   "https://test.example.com/docs",
			},
		}

		server := NewServer(cfg)
		server.setupRoutes()

		// Test that specific API paths work correctly
		validPaths := map[string]int{
			"/api/status":          http.StatusOK,
			"/api/scrape-interval": http.StatusOK,
			"/api/docs":            http.StatusOK,
			"/healthz":             http.StatusOK,
			"/readyz":              http.StatusServiceUnavailable, // Cache is empty
			"/metrics":             http.StatusOK,
		}

		for path, expectedStatus := range validPaths {
			t.Run(path, func(t *testing.T) {
				req := httptest.NewRequest("GET", path, nil)
				w := httptest.NewRecorder()
				server.mux.ServeHTTP(w, req)
				assert.Equal(t, expectedStatus, w.Code)
			})
		}

		// Test that unmatched paths fall back to static handler
		// When static directory doesn't exist, the fallback returns 200
		fallbackPaths := []string{
			"/",
			"/invalid",
			"/api/invalid",
			"/api/status/invalid",
			"/some/deep/path",
		}

		for _, path := range fallbackPaths {
			t.Run(path+" (fallback)", func(t *testing.T) {
				req := httptest.NewRequest("GET", path, nil)
				w := httptest.NewRecorder()
				server.mux.ServeHTTP(w, req)
				// The fallback handler returns 200 when no static directory exists
				assert.Equal(t, http.StatusOK, w.Code)
			})
		}
	})

	t.Run("method handling for endpoints", func(t *testing.T) {
		server := NewServer(&config.Config{})
		server.setupRoutes()

		// Test that endpoints accept different HTTP methods
		// Note: The handlers don't specifically check HTTP methods in most cases

		t.Run("POST /healthz", func(t *testing.T) {
			req := httptest.NewRequest("POST", "/healthz", nil)
			w := httptest.NewRecorder()
			server.mux.ServeHTTP(w, req)
			// The liveness handler doesn't check method, so it returns 200
			assert.Equal(t, http.StatusOK, w.Code)
		})

		t.Run("POST /readyz", func(t *testing.T) {
			req := httptest.NewRequest("POST", "/readyz", nil)
			w := httptest.NewRecorder()
			server.mux.ServeHTTP(w, req)
			// The readiness handler doesn't check method, but depends on cache state
			// Since cache is empty, it returns 503 regardless of method
			assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		})

		t.Run("PUT /api/status", func(t *testing.T) {
			req := httptest.NewRequest("PUT", "/api/status", nil)
			w := httptest.NewRecorder()
			server.mux.ServeHTTP(w, req)
			// The status handler doesn't check method, so it returns 200
			assert.Equal(t, http.StatusOK, w.Code)
		})
	})
}
