package prometheus

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"site-availability/config"
	"site-availability/handlers"
	"site-availability/logging"
	"sync"
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

// PrometheusScraper implements the Scraper interface for Prometheus.
type PrometheusScraper struct {
}

func NewPrometheusScraper() *PrometheusScraper {
	return &PrometheusScraper{}
}

func (p *PrometheusScraper) Scrape(source config.Source, timeout time.Duration, maxParallel int, tlsConfig *tls.Config) ([]handlers.AppStatus, []handlers.Location, error) {
	results := make([]handlers.AppStatus, len(source.Apps))
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

	for i, app := range source.Apps {
		sem <- struct{}{} // Acquire slot
		wg.Add(1)
		go func(i int, app config.App) {
			defer func() {
				<-sem
				wg.Done()
			}()
			statusCode, err := p.check(client, source.URL, app.Metric, source.Auth, source.Token)
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

			mu.Lock()
			results[i] = handlers.AppStatus{
				Name:     app.Name,
				Location: app.Location,
				Status:   status,
				Source:   source.Name,
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
