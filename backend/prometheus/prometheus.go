package prometheus

import (
	"site-avilability/config"
	"site-avilability/handlers"
) // Ensure this import is present

// DefaultPrometheusChecker implements the Prometheus status checking
type DefaultPrometheusChecker struct{}

// Check checks the status of the app from Prometheus.
func (d *DefaultPrometheusChecker) Check(prometheusURL string, metric string) (int, error) {
	// Simulate Prometheus check (replace with actual logic)
	if prometheusURL == "prometheus1.app.url" {
		return 1, nil // Simulate an "up" status for testing.
	}
	return 0, nil // Simulate a "down" status for testing.
}

// CheckAppStatus now accepts a PrometheusChecker interface
func CheckAppStatus(app config.App, checker *DefaultPrometheusChecker) handlers.AppStatus {
	// Check status using the provided checker
	var status string
	for _, prometheusURL := range app.Prometheus {
		statusCode, err := checker.Check(prometheusURL, app.Metric)
		if err != nil {
			status = "down"
			break
		}
		if statusCode == 1 {
			status = "up"
			break
		}
	}
	if status == "" {
		status = "down"
	}

	return handlers.AppStatus{
		Name:     app.Name,
		Location: app.Location,
		Status:   status,
	}
}
