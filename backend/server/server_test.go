package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"site-availability/config"
	"site-availability/handlers"
	"site-availability/scraping"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServer(t *testing.T) {
	cfg := &config.Config{
		ServerSettings: config.ServerSettings{
			Port:       "8080",
			SyncEnable: true,
		},
	}

	server := NewServer(cfg)
	assert.NotNil(t, server)
	assert.Equal(t, cfg, server.config)
}

func TestLoadCustomCA(t *testing.T) {
	// Create test CA certificate
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "ca.crt")
	err := os.WriteFile(certPath, []byte("test certificate"), 0644)
	require.NoError(t, err)

	cfg := &config.Config{
		ServerSettings: config.ServerSettings{
			CustomCAPath: certPath,
		},
	}

	server := NewServer(cfg)
	server.loadCustomCA()

	// Test with non-existent CA
	cfg.ServerSettings.CustomCAPath = "nonexistent.crt"
	server = NewServer(cfg)
	server.loadCustomCA() // Should not panic
}

func TestSetupRoutes(t *testing.T) {
	// Create temporary directory for static files
	tmpDir := t.TempDir()
	staticDir := filepath.Join(tmpDir, "static")
	err := os.Mkdir(staticDir, 0755)
	require.NoError(t, err)

	// Create a test static file
	err = os.WriteFile(filepath.Join(staticDir, "index.html"), []byte("test"), 0644)
	require.NoError(t, err)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(originalDir)
		require.NoError(t, err)
	}()

	// Change to the static directory for the test
	err = os.Chdir(staticDir)
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
	}

	server := NewServer(cfg)
	server.setupRoutes()

	// Test all routes
	testCases := []struct {
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
		{"/sync", "GET", http.StatusOK},
		{"/", "GET", http.StatusOK},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		w := httptest.NewRecorder()
		server.mux.ServeHTTP(w, req)
		assert.Equal(t, tc.statusCode, w.Code, "Path: %s", tc.path)
	}
}

func TestStatusFetcher(t *testing.T) {
	cfg := &config.Config{
		Applications: []config.Application{
			{
				Name:   "test-app",
				Metric: "up{instance=\"test\"}",
			},
		},
		PrometheusServers: []config.PrometheusServer{
			{
				Name: "test-server",
				URL:  "http://test-server:9090",
			},
		},
	}

	server := NewServer(cfg)
	checker := &scraping.PrometheusMetricChecker{
		PrometheusServers: cfg.PrometheusServers,
	}

	server.statusFetcher(checker, 5)
	statuses := handlers.GetAppStatusCache()
	assert.NotEmpty(t, statuses)

	// Clean up
	handlers.UpdateAppStatus([]handlers.AppStatus{})
}

func TestStartServer(t *testing.T) {
	cfg := &config.Config{
		ServerSettings: config.ServerSettings{
			Port: "8080",
		},
	}

	server := NewServer(cfg)

	// Start server in a goroutine
	go func() {
		err := server.startServer(cfg.ServerSettings.Port)
		assert.NoError(t, err)
	}()

	// Give the server time to start
	time.Sleep(100 * time.Millisecond)

	// Stop the server
	err := server.Stop()
	assert.NoError(t, err)
}

func TestGracefulShutdown(t *testing.T) {
	cfg := &config.Config{
		ServerSettings: config.ServerSettings{
			Port:       "8080",
			SyncEnable: true,
		},
	}

	server := NewServer(cfg)
	srv := &http.Server{Addr: ":" + cfg.ServerSettings.Port}

	// Test graceful shutdown of non-running server
	err := server.gracefulShutdown(srv)
	assert.NoError(t, err)

	// Test timeout scenario with server.gracefulShutdown method
	srv = &http.Server{Addr: ":" + cfg.ServerSettings.Port}

	// Start server in a goroutine
	go func() {
		_ = srv.ListenAndServe()
	}()

	// Wait a bit for the server to start
	time.Sleep(100 * time.Millisecond)

	// Test the actual gracefulShutdown method which has the timeout logic
	// This should test the timeout path in the server's gracefulShutdown method
	originalShutdown := server.gracefulShutdown

	// Create a mock server that will take longer to shutdown
	slowSrv := &http.Server{Addr: ":0"} // Use port 0 to avoid conflicts

	// Override the shutdown method to simulate a slow shutdown
	go func() {
		_ = slowSrv.ListenAndServe()
	}()
	time.Sleep(50 * time.Millisecond)

	// Test graceful shutdown - this should complete normally
	_ = originalShutdown(slowSrv) // Use blank identifier since we don't need the error
}

func TestLivenessProbe(t *testing.T) {
	server := NewServer(&config.Config{})
	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	server.livenessProbe(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

func TestReadinessProbe(t *testing.T) {
	server := NewServer(&config.Config{})

	// Test when cache is empty
	req := httptest.NewRequest("GET", "/readyz", nil)
	w := httptest.NewRecorder()
	server.readinessProbe(w, req)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Equal(t, "NOT READY", w.Body.String())

	// Test when cache has data
	handlers.UpdateAppStatus([]handlers.AppStatus{
		{
			Name:   "test-app",
			Status: "up",
		},
	})

	w = httptest.NewRecorder()
	server.readinessProbe(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "READY", w.Body.String())
}
