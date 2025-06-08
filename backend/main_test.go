package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"site-availability/config"
	"site-availability/handlers"
	"site-availability/logging"
)

func TestMain(m *testing.M) {
	// Initialize global logger before tests run
	if err := logging.Init(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	code := m.Run()
	os.Exit(code)
}

func TestLivenessProbe(t *testing.T) {
	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	livenessProbe(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestReadinessProbe_EmptyCache(t *testing.T) {
	handlers.UpdateAppStatus([]handlers.AppStatus{}) // Clear cache

	req := httptest.NewRequest("GET", "/readyz", nil)
	w := httptest.NewRecorder()
	readinessProbe(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", resp.StatusCode)
	}
}

func TestReadinessProbe_WithData(t *testing.T) {
	handlers.UpdateAppStatus([]handlers.AppStatus{
		{Name: "TestApp", Status: "up"},
	})

	req := httptest.NewRequest("GET", "/readyz", nil)
	w := httptest.NewRecorder()
	readinessProbe(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestGetEnv(t *testing.T) {
	// Test with environment variable set
	os.Setenv("TEST_ENV_VAR", "test_value")
	defer os.Unsetenv("TEST_ENV_VAR")

	if value := getEnv("TEST_ENV_VAR", "default"); value != "test_value" {
		t.Errorf("Expected 'test_value', got %s", value)
	}

	// Test with environment variable not set
	if value := getEnv("NONEXISTENT_VAR", "default"); value != "default" {
		t.Errorf("Expected 'default', got %s", value)
	}

	// Test with empty environment variable
	os.Setenv("EMPTY_VAR", "")
	defer os.Unsetenv("EMPTY_VAR")

	if value := getEnv("EMPTY_VAR", "default"); value != "default" {
		t.Errorf("Expected 'default', got %s", value)
	}
}

func TestStartServer(t *testing.T) {
	// Create a test configuration
	testCfg := &config.Config{
		ServerSettings: config.ServerSettings{
			Port: "8080",
		},
	}

	// Set up routes before starting the server
	setupRoutes()

	// Start server in a goroutine
	go startServer(testCfg.ServerSettings.Port)

	// Give the server more time to start
	time.Sleep(500 * time.Millisecond)

	// Test the server is running by making a request to the liveness probe
	resp, err := http.Get("http://localhost:8080/healthz")
	if err != nil {
		t.Errorf("Failed to connect to server: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}
