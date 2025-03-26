package scraping

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"site-availability/config"
	"site-availability/handlers"
	"time"
)

// PrometheusResponse represents the structure of the Prometheus API response.
type PrometheusResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"` // Timestamp and metric value
		} `json:"result"`
	} `json:"data"`
}

// DefaultPrometheusChecker checks Prometheus metrics.
type DefaultPrometheusChecker struct{}

// Check queries Prometheus and extracts the metric value.
func (d *DefaultPrometheusChecker) Check(prometheusURL string, promQLQuery string) (int, error) {
	// Properly encode the query
	encodedQuery := url.QueryEscape(promQLQuery)

	// Construct the full URL
	fullURL := fmt.Sprintf("%s/api/v1/query?query=%s", prometheusURL, encodedQuery)

	// Perform the HTTP request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(fullURL)
	if err != nil {
		return 0, fmt.Errorf("failed to query Prometheus: %v", err)
	}
	defer resp.Body.Close()

	// Decode the Prometheus API response
	var promResp PrometheusResponse
	if err := json.NewDecoder(resp.Body).Decode(&promResp); err != nil {
		return 0, fmt.Errorf("failed to decode Prometheus response: %v", err)
	}

	// Check if the response is successful
	if promResp.Status != "success" {
		return 0, fmt.Errorf("prometheus query %s failed: %s", promQLQuery, promResp.Status)
	}

	// If there are no results, consider the app "down"
	if len(promResp.Data.Result) == 0 {
		return 0, fmt.Errorf("prometheus query %s did not return any result", promQLQuery)
	}

	// Extract the metric value (second element in `value` array)
	value, ok := promResp.Data.Result[0].Value[1].(string)
	if !ok {
		return 0, fmt.Errorf("unexpected response format: value is not a string")
	}

	// Convert value to integer (assuming "1" means up and "0" means down)
	if value == "1" {
		return 1, nil // App is "up"
	}

	return 0, nil // App is "down"
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
