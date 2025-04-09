package scraping

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"site-availability/config"
	"site-availability/handlers"
	"site-availability/logging"
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

// PrometheusChecker defines an interface for checking Prometheus metrics.
type PrometheusChecker interface {
	Check(prometheusURL string, promQLQuery string) (int, error)
}

// DefaultPrometheusChecker checks Prometheus metrics.
type DefaultPrometheusChecker struct{}

// Check queries Prometheus and extracts the metric value.
func (d *DefaultPrometheusChecker) Check(prometheusURL string, promQLQuery string) (int, error) {
	encodedQuery := url.QueryEscape(promQLQuery)
	fullURL := fmt.Sprintf("%s/api/v1/query?query=%s", prometheusURL, encodedQuery)

	logging.Logger.WithFields(map[string]interface{}{
		"url":    fullURL,
		"metric": promQLQuery,
		"source": "scraping.Check",
	}).Debug("Querying Prometheus")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(fullURL)
	if err != nil {
		logging.Logger.WithError(err).WithField("url", fullURL).Error("Failed to query Prometheus")
		return 0, fmt.Errorf("failed to query Prometheus: %v", err)
	}
	defer resp.Body.Close()

	var promResp PrometheusResponse
	if err := json.NewDecoder(resp.Body).Decode(&promResp); err != nil {
		logging.Logger.WithError(err).Error("Failed to decode Prometheus response")
		return 0, fmt.Errorf("failed to decode Prometheus response: %v", err)
	}

	if promResp.Status != "success" {
		logging.Logger.WithField("status", promResp.Status).Error("Prometheus query did not succeed")
		return 0, fmt.Errorf("prometheus query %s failed: %s", promQLQuery, promResp.Status)
	}

	if len(promResp.Data.Result) == 0 {
		logging.Logger.WithField("metric", promQLQuery).Warn("Prometheus query returned no results")
		return 0, fmt.Errorf("prometheus query %s did not return any result", promQLQuery)
	}

	value, ok := promResp.Data.Result[0].Value[1].(string)
	if !ok {
		logging.Logger.Error("Unexpected response format: metric value is not a string")
		return 0, fmt.Errorf("unexpected response format: value is not a string")
	}

	logging.Logger.WithFields(map[string]interface{}{
		"value":  value,
		"metric": promQLQuery,
	}).Debug("Prometheus metric value retrieved")

	if value == "1" {
		return 1, nil
	}

	return 0, nil
}

// CheckAppStatus now accepts a PrometheusChecker interface
func CheckAppStatus(app config.App, checker PrometheusChecker) handlers.AppStatus {
	logging.Logger.WithFields(map[string]interface{}{
		"app":        app.Name,
		"location":   app.Location,
		"prometheus": app.Prometheus,
		"metric":     app.Metric,
	}).Debug("Checking application status")

	statusCode, err := checker.Check(app.Prometheus, app.Metric)

	status := "down"
	if err != nil {
		logging.Logger.WithFields(map[string]interface{}{
			"app":   app.Name,
			"error": err.Error(),
		}).Warn("Application status check failed")
	} else if statusCode == 1 {
		status = "up"
		logging.Logger.WithField("app", app.Name).Debug("Application is UP")
	} else {
		logging.Logger.WithField("app", app.Name).Debug("Application is DOWN")
	}

	return handlers.AppStatus{
		Name:     app.Name,
		Location: app.Location,
		Status:   status,
	}
}
