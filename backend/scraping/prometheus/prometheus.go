package prometheus

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"site-availability/config"
	"site-availability/handlers"
	"site-availability/labels"
	"site-availability/logging"
	"sync"
	"time"
)

// PrometheusConfig represents the configuration for Prometheus sources
type PrometheusConfig struct {
	URL   string          `yaml:"url"`
	Token string          `yaml:"token"`
	Auth  string          `yaml:"auth"`
	Apps  []PrometheusApp `yaml:"apps"`
}

// PrometheusApp represents an app configuration for Prometheus
type PrometheusApp struct {
	Name     string            `yaml:"name"`
	Location string            `yaml:"location"`
	Metric   string            `yaml:"metric"`
	Labels   map[string]string `yaml:"labels"`
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

// PrometheusScraper implements the Scraper interface for Prometheus.
type PrometheusScraper struct {
}

func NewPrometheusScraper() *PrometheusScraper {
	return &PrometheusScraper{}
}

// ValidateConfig validates the Prometheus-specific configuration
func (p *PrometheusScraper) ValidateConfig(source config.Source) error {
	promCfg, err := config.DecodeConfig[PrometheusConfig](source.Config, source.Name)
	if err != nil {
		return err
	}

	// Validate required fields
	if promCfg.URL == "" {
		return fmt.Errorf("prometheus source %s: missing 'url'", source.Name)
	}

	// Validate apps
	if len(promCfg.Apps) == 0 {
		return fmt.Errorf("prometheus source %s: at least one app is required", source.Name)
	}

	appNames := make(map[string]bool)
	for _, app := range promCfg.Apps {
		if app.Name == "" {
			return fmt.Errorf("prometheus source %s: app name is required", source.Name)
		}
		if _, exists := appNames[app.Name]; exists {
			return fmt.Errorf("prometheus source %s: duplicate app name %q", source.Name, app.Name)
		}
		appNames[app.Name] = true

		if app.Metric == "" {
			return fmt.Errorf("prometheus source %s: app %s missing 'metric'", source.Name, app.Name)
		}
		if app.Location == "" {
			return fmt.Errorf("prometheus source %s: app %s missing 'location'", source.Name, app.Name)
		}
	}

	return nil
}

func (p *PrometheusScraper) Scrape(source config.Source, serverSettings config.ServerSettings, timeout time.Duration, maxParallel int, tlsConfig *tls.Config) ([]handlers.AppStatus, []handlers.Location, error) {
	// Decode the source-specific config
	promCfg, err := config.DecodeConfig[PrometheusConfig](source.Config, source.Name)
	if err != nil {
		return nil, nil, err
	}

	results := make([]handlers.AppStatus, len(promCfg.Apps))
	sem := make(chan struct{}, maxParallel)
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Shared HTTP client with optional TLS config
	client := &http.Client{Timeout: timeout}
	if tlsConfig != nil {
		client.Transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}
	}

	for i, app := range promCfg.Apps {
		sem <- struct{}{} // Acquire slot
		wg.Add(1)
		go func(i int, app PrometheusApp) {
			defer func() {
				<-sem
				wg.Done()
			}()
			statusCode, err := p.check(client, promCfg.URL, app.Metric, promCfg.Auth, promCfg.Token)
			status := "unavailable"

			if err != nil {
				// If check function failed, log as warning and mark app as unavailable
				logging.Logger.WithFields(map[string]interface{}{
					"app":    app.Name,
					"source": source.Name,
					"error":  err.Error(),
				}).Warn("Application status check failed - marking as unavailable")
			} else {
				if statusCode == 1 {
					status = "up"
					logging.Logger.WithField("app", app.Name).Debug("Application is UP")
				} else {
					status = "down"
					logging.Logger.WithField("app", app.Name).Debug("Application is DOWN")
				}
			}

			// Convert map[string]string labels to []Label format
			var appLabels []labels.Label
			for key, value := range app.Labels {
				appLabels = append(appLabels, labels.Label{Key: key, Value: value})
			}

			mu.Lock()
			results[i] = handlers.AppStatus{
				Name:      app.Name,
				Location:  app.Location,
				Status:    status,
				Source:    source.Name,
				OriginURL: serverSettings.HostURL, // Use host URL as origin for deduplication
				Labels:    appLabels,              // App labels only - source/server labels added in UpdateAppStatus
			}
			mu.Unlock()
		}(i, app)
	}
	wg.Wait()

	// Always return success - check errors are handled by marking apps as unavailable
	// Prometheus scraper returns nil for locations since it only provides app statuses
	return results, nil, nil
}

func (p *PrometheusScraper) check(client *http.Client, prometheusURL, promQLQuery, auth, token string) (int, error) {
	encodedQuery := url.QueryEscape(promQLQuery)
	fullURL := fmt.Sprintf("%s/api/v1/query?query=%s", prometheusURL, encodedQuery)

	logging.Logger.WithFields(map[string]interface{}{
		"url":    fullURL,
		"metric": promQLQuery,
		"source": "prometheusScraper.check",
	}).Debug("Querying Prometheus")

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication if credentials are available
	if auth != "" && token != "" {
		switch auth {
		case "bearer":
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		case "basic":
			req.Header.Set("Authorization", fmt.Sprintf("Basic %s", token))
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		logging.Logger.WithError(err).WithField("url", fullURL).Error("Failed to query Prometheus")
		return 0, fmt.Errorf("failed to query Prometheus: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		logging.Logger.WithFields(map[string]interface{}{
			"url":         fullURL,
			"auth_method": auth,
			"status_code": resp.StatusCode,
		}).Error("Authentication failed for Prometheus server")
		return 0, fmt.Errorf("authentication failed for Prometheus server %s using %s auth", fullURL, auth)
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

	if len(promResp.Data.Result[0].Value) < 2 {
		logging.Logger.Error("Unexpected response format: value array too short")
		return 0, fmt.Errorf("unexpected response format: value array too short")
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
