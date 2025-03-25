package scraping

import (
	"site-availability/config"
	"testing"
)

func TestCheckAppStatus(t *testing.T) {
	// Define a mock App for testing
	app := config.App{
		Name:       "app1",
		Location:   "site1",
		Metric:     "up{instance=\"app1\"}",
		Prometheus: "prometheus1.app.url",
	}

	// Initialize the PrometheusChecker
	checker := &DefaultPrometheusChecker{}

	// Call CheckAppStatus to get the status
	status := CheckAppStatus(app, checker)

	// Expected status (this will depend on your mock Prometheus URLs logic)
	expectedStatus := "up" // or "down", based on your mock logic

	// Validate the status returned is correct
	if status.Status != expectedStatus {
		t.Errorf("Expected status %s, but got %s", expectedStatus, status.Status)
	}

	// Additional checks for app name, location, etc.
	if status.Name != app.Name {
		t.Errorf("Expected app name %s, but got %s", app.Name, status.Name)
	}
	if status.Location != app.Location {
		t.Errorf("Expected app location %s, but got %s", app.Location, status.Location)
	}
}
