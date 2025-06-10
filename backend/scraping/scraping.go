package scraping

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"site-availability/config"
	"site-availability/handlers"
	"site-availability/logging"
	"strings"
	"sync"
	"time"
)

var httpClient *http.Client

var defaultScrapeTimeout = 10 * time.Second

func SetScrapeTimeout(timeout time.Duration) {
	defaultScrapeTimeout = timeout
	if httpClient != nil {
		httpClient.Timeout = defaultScrapeTimeout
	} else {
		httpClient = &http.Client{Timeout: defaultScrapeTimeout}
	}
}

// InitCertificate loads CA certificates from the file paths listed in the given environment variable name.
// The environment variable value may contain multiple file paths separated by ":".
func InitCertificate(envVarName string) {
	caPath := os.Getenv(envVarName)
	if caPath == "" {
		logging.Logger.WithField("env", envVarName).Info("Env var not set. Using default HTTP client.")
		httpClient = &http.Client{Timeout: defaultScrapeTimeout}
		return
	}

	caCertPool := x509.NewCertPool()
	paths := strings.Split(caPath, ":")

	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}

		certData, err := os.ReadFile(path)
		if err != nil {
			logging.Logger.WithError(err).WithField("path", path).Error("Failed to read CA certificate")
			continue
		}

		if ok := caCertPool.AppendCertsFromPEM(certData); !ok {
			logging.Logger.WithField("path", path).Error("Failed to append CA certificate")
			continue
		}
	}

	tlsConfig := &tls.Config{
		RootCAs:            caCertPool,
		InsecureSkipVerify: false, // Only set true if you know what you're doing
	}

	httpClient = &http.Client{
		Timeout: defaultScrapeTimeout,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	logging.Logger.WithField("env", envVarName).Info("Custom CA certificates loaded successfully")
}

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

// MetricChecker defines an interface for checking Prometheus metrics.
type MetricChecker interface {
	Check(prometheusURL string, promQLQuery string, prometheusServers []config.PrometheusServer) (int, error)
}

// PrometheusMetricChecker implements the MetricChecker interface for Prometheus.
type PrometheusMetricChecker struct {
	PrometheusServers []config.PrometheusServer
}

// Check queries Prometheus and extracts the metric value.
func (c *PrometheusMetricChecker) Check(prometheusURL string, promQLQuery string, prometheusServers []config.PrometheusServer) (int, error) {
	encodedQuery := url.QueryEscape(promQLQuery)
	fullURL := fmt.Sprintf("%s/api/v1/query?query=%s", prometheusURL, encodedQuery)

	logging.Logger.WithFields(map[string]interface{}{
		"url":    fullURL,
		"metric": promQLQuery,
		"source": "scraping.Check",
	}).Debug("Querying Prometheus")

	client := httpClient
	if client == nil {
		client = &http.Client{Timeout: defaultScrapeTimeout}
	}

	// Create request
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication if credentials are available
	var authMethod string
	if c.PrometheusServers != nil {
		logging.Logger.WithFields(map[string]interface{}{
			"servers_count":  len(c.PrometheusServers),
			"prometheus_url": prometheusURL,
		}).Debug("Checking credentials for Prometheus server")

		// Find the Prometheus server from the URL
		for _, server := range c.PrometheusServers {
			if server.URL == prometheusURL {
				if server.Auth != "" && server.Token != "" {
					authMethod = server.Auth
					logging.Logger.WithFields(map[string]interface{}{
						"prometheus": server.Name,
						"auth_type":  server.Auth,
					}).Debug("Found matching credentials for Prometheus server")

					switch server.Auth {
					case "bearer":
						req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", server.Token))
					case "basic":
						req.Header.Set("Authorization", fmt.Sprintf("Basic %s", server.Token))
					}
				}
				break
			}
		}
	} else {
		logging.Logger.Debug("No credentials available for authentication")
	}

	resp, err := client.Do(req)
	if err != nil {
		logging.Logger.WithError(err).WithField("url", fullURL).Error("Failed to query Prometheus")
		return 0, fmt.Errorf("failed to query Prometheus: %v", err)
	}
	defer resp.Body.Close()

	// Check for authentication errors
	if resp.StatusCode == http.StatusUnauthorized {
		logging.Logger.WithFields(map[string]interface{}{
			"url":         fullURL,
			"auth_method": authMethod,
			"status_code": resp.StatusCode,
		}).Error("Authentication failed for Prometheus server")

		return 0, fmt.Errorf("authentication failed for Prometheus server %s using %s auth",
			fullURL, authMethod)
	}

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

// CheckAppStatus checks the status of an application using the provided metric checker.
func CheckAppStatus(app config.Application, prometheusServers []config.PrometheusServer, checker MetricChecker) handlers.AppStatus {
	logging.Logger.WithFields(map[string]interface{}{
		"app":        app.Name,
		"location":   app.Location,
		"prometheus": app.Prometheus,
		"metric":     app.Metric,
	}).Debug("Checking application status")

	// Find the Prometheus server URL by name
	var prometheusURL string
	for _, server := range prometheusServers {
		if server.Name == app.Prometheus {
			prometheusURL = server.URL
			break
		}
	}

	if prometheusURL == "" {
		logging.Logger.WithField("prometheus", app.Prometheus).Error("Prometheus server not found")
		return handlers.AppStatus{
			Name:     app.Name,
			Location: app.Location,
			Status:   "unavailable",
		}
	}

	statusCode, err := checker.Check(prometheusURL, app.Metric, prometheusServers)

	status := "unavailable" // Default to "unavailable" if there's an error
	if err != nil {
		logging.Logger.WithFields(map[string]interface{}{
			"app":   app.Name,
			"error": err.Error(),
		}).Warn("Application status check failed")
	} else if statusCode == 1 {
		status = "up"
		logging.Logger.WithField("app", app.Name).Debug("Application is UP")
	} else {
		status = "down"
		logging.Logger.WithField("app", app.Name).Debug("Application is DOWN")
	}

	return handlers.AppStatus{
		Name:     app.Name,
		Location: app.Location,
		Status:   status,
	}
}

func ParallelScrapeAppStatuses(apps []config.Application, prometheusServers []config.PrometheusServer, checker *PrometheusMetricChecker, maxParallelScrapes int) []handlers.AppStatus {
	results := make([]handlers.AppStatus, len(apps))
	sem := make(chan struct{}, maxParallelScrapes)
	var wg sync.WaitGroup

	for i, app := range apps {
		sem <- struct{}{} // Acquire slot
		wg.Add(1)         // One more goroutine to wait for

		go func(i int, app config.Application) {
			defer func() {
				<-sem     // Release slot
				wg.Done() // Mark as done
			}()

			logging.Logger.Debugf("Checking app status: name=%s, url=%s", app.Name, app.Prometheus)
			results[i] = CheckAppStatus(app, prometheusServers, checker)
			logging.Logger.Debugf("App status fetched: name=%s, status=%s", results[i].Name, results[i].Status)
		}(i, app)
	}

	wg.Wait() // Wait for all goroutines to finish
	return results
}
