package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"site-availability/handlers"
	"site-availability/logging"
)

func TestMain(m *testing.M) {
	// Initialize global logger before tests run
	logging.Init()
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

func TestGetServerPort(t *testing.T) {
	os.Setenv("PORT", ":9090")
	defer os.Unsetenv("PORT")

	port := getServerPort()
	if port != ":9090" {
		t.Errorf("Expected port :9090, got %s", port)
	}
}

func TestGetServerPort_Default(t *testing.T) {
	os.Unsetenv("PORT")

	port := getServerPort()
	if port != ":8080" {
		t.Errorf("Expected default port :8080, got %s", port)
	}
}
