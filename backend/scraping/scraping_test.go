package scraping

import (
	"site-availability/config"
	"testing"
)

// MockPrometheusChecker implements PrometheusChecker for testing.
type MockPrometheusChecker struct {
	mockResponse int
	mockError    error
}

// Check simulates a Prometheus response.
func (m *MockPrometheusChecker) Check(prometheusURL string, promQLQuery string) (int, error) {
	if m.mockError != nil {
		return 0, m.mockError
	}
	return m.mockResponse, nil
}

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

	if status.Status != "down" {
		t.Errorf("Expected status 'down' on error, but got %s", status.Status)
	}
}

// ErrMockFailure is a mock error for testing.
var ErrMockFailure = &MockError{}

type MockError struct{}

func (e *MockError) Error() string {
	return "mock error"
}
