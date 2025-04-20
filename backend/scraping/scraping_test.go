package scraping

import (
	"fmt"
	"os"
	"path/filepath"
	"site-availability/config"
	"testing"
)

// --------------------
// Mock implementation
// --------------------

type MockPrometheusChecker struct {
	mockResponse int
	mockError    error
}

func (m *MockPrometheusChecker) Check(prometheusURL string, promQLQuery string) (int, error) {
	if m.mockError != nil {
		return 0, m.mockError
	}
	return m.mockResponse, nil
}

// --------------------
// CheckAppStatus tests
// --------------------

func TestCheckAppStatus_Up(t *testing.T) {
	mockChecker := &MockPrometheusChecker{mockResponse: 1, mockError: nil}

	app := config.App{
		Name:       "app1",
		Location:   "site1",
		Metric:     "up{instance=\"app1\"}",
		Prometheus: "http://prometheus1.app.url",
	}

	status := CheckAppStatus(app, mockChecker)

	if status.Status != "up" {
		t.Errorf("Expected status 'up', but got %s", status.Status)
	}
	if status.Name != app.Name {
		t.Errorf("Expected app name %s, but got %s", app.Name, status.Name)
	}
	if status.Location != app.Location {
		t.Errorf("Expected app location %s, but got %s", app.Location, status.Location)
	}
}

func TestCheckAppStatus_Down(t *testing.T) {
	mockChecker := &MockPrometheusChecker{mockResponse: 0, mockError: nil}

	app := config.App{
		Name:       "app2",
		Location:   "site2",
		Metric:     "up{instance=\"app2\"}",
		Prometheus: "http://prometheus2.app.url",
	}

	status := CheckAppStatus(app, mockChecker)

	if status.Status != "down" {
		t.Errorf("Expected status 'down', but got %s", status.Status)
	}
}

func TestCheckAppStatus_Error(t *testing.T) {
	mockChecker := &MockPrometheusChecker{mockResponse: 0, mockError: ErrMockFailure}

	app := config.App{
		Name:       "app3",
		Location:   "site3",
		Metric:     "up{instance=\"app3\"}",
		Prometheus: "http://prometheus3.app.url",
	}

	status := CheckAppStatus(app, mockChecker)

	if status.Status != "unavailable" {
		t.Errorf("Expected status 'unavailable' on error, but got %s", status.Status)
	}
}

func TestCheckAppStatus_Unavailable(t *testing.T) {
	mockChecker := &MockPrometheusChecker{mockResponse: 0, mockError: fmt.Errorf("mock error")}

	app := config.App{
		Name:       "app5",
		Location:   "site5",
		Metric:     "up{instance=\"app5\"}",
		Prometheus: "http://prometheus5.app.url",
	}

	status := CheckAppStatus(app, mockChecker)

	if status.Status != "unavailable" {
		t.Errorf("Expected status 'unavailable' on error, but got %s", status.Status)
	}
}

var ErrMockFailure = &MockError{}

type MockError struct{}

func (e *MockError) Error() string {
	return "mock error"
}

// --------------------
// InitCertificate tests
// --------------------

func TestInitCertificate_Success(t *testing.T) {
	// Create a temporary valid CA cert file
	tmpFile := filepath.Join(t.TempDir(), "test-ca.pem")
	certContent := `-----BEGIN CERTIFICATE-----
MIID...FAKE...CERTIFICATE...CONTENT...==
-----END CERTIFICATE-----`
	if err := os.WriteFile(tmpFile, []byte(certContent), 0644); err != nil {
		t.Fatalf("Failed to write temp cert: %v", err)
	}

	// Set the env var to point to our temp cert
	os.Setenv("TEST_CA_CERT_PATH", tmpFile)
	defer os.Unsetenv("TEST_CA_CERT_PATH")

	// Should not panic or error (we can't easily assert the internal pool)
	InitCertificate("TEST_CA_CERT_PATH")
}

func TestInitCertificate_EmptyEnv(t *testing.T) {
	os.Unsetenv("EMPTY_CA_ENV")
	InitCertificate("EMPTY_CA_ENV") // Should silently skip
}

func TestInitCertificate_FileNotExist(t *testing.T) {
	// Set a non-existent path
	os.Setenv("INVALID_CA_PATH", "/non/existent/path.pem")
	defer os.Unsetenv("INVALID_CA_PATH")

	// Should log error but not crash
	InitCertificate("INVALID_CA_PATH")
}
