package scraping

import (
	"encoding/json"
	"fmt"
	"net/http"
	"site-availability/config"
	"site-availability/handlers"
	"time"
)

// DefaultPrometheusChecker implements the Prometheus status checking
type DefaultPrometheusChecker struct{}

// PrometheusResponse represents the response structure from Prometheus
type PrometheusResponse struct {
	Status string `json:"status"`
	Data   struct {
		Result []struct {
			Value []interface{} `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

// Check checks the status of the app from Prometheus using the URL from App.Prometheus and the PromQL query from App.Metric.
func (d *DefaultPrometheusChecker) Check(prometheusURL string, promQLQuery string) (int, error) {
	// Create the Prometheus API URL with the PromQL query
	url := fmt.Sprintf("%s/api/v1/query?query=%s", prometheusURL, promQLQuery)

	// Perform the HTTP request to Prometheus
	client := &http.Client{Timeout: 10 * time.Second} // Set a timeout for the request
	resp, err := client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to query Prometheus: %v", err)
	}
	defer resp.Body.Close()

	// Decode the Prometheus API response
	var promResp PrometheusResponse
	if err := json.NewDecoder(resp.Body).Decode(&promResp); err != nil {
		return 0, fmt.Errorf("failed to decode Prometheus response: %v", err)
	}

	// Check the status based on the Prometheus response
	if promResp.Status != "success" {
		return 0, fmt.Errorf("Prometheus query failed: %s", promResp.Status)
	}

	// If there are results, the app is considered "up", otherwise it's "down"
	if len(promResp.Data.Result) > 0 {
		return 1, nil // Status is "up"
	}

	// If no result, consider the app "down"
	return 0, nil // Status is "down"
}

// CheckAppStatus now accepts a PrometheusChecker interface
func CheckAppStatus(app config.App, checker *DefaultPrometheusChecker) handlers.AppStatus {
	// Check status using the provided checker
	var status string
	// Using the Prometheus URL (app.Prometheus) and PromQL query (app.Metric)
	statusCode, err := checker.Check(app.Prometheus, app.Metric)
	if err != nil {
		status = "down"
	} else if statusCode == 1 {
		status = "up"
	} else {
		status = "down"
	}

	return handlers.AppStatus{
		Name:     app.Name,
		Location: app.Location,
		Status:   status,
	}
}
