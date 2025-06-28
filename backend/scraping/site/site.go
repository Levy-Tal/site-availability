package site

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"site-availability/authentication/hmac"
	"site-availability/config"
	"site-availability/handlers"
	"site-availability/logging"
	"time"
)

// SiteScraper implements the scraping.Source interface for scraping other sites.
type SiteScraper struct {
}

func NewSiteScraper() *SiteScraper {
	return &SiteScraper{}
}

// Scrape fetches the status of all apps and locations from a remote site using the /sync endpoint.
// Since site scraping involves a single request, the maxParallel parameter is not used.
func (s *SiteScraper) Scrape(source config.Source, timeout time.Duration, maxParallel int, tlsConfig *tls.Config) ([]handlers.AppStatus, []handlers.Location, error) {
	url := fmt.Sprintf("%s/sync", source.URL)
	client := &http.Client{Timeout: timeout}

	if tlsConfig != nil {
		client.Transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// Request creation errors are code issues, return them
		return nil, nil, fmt.Errorf("failed to create request for site %s: %w", source.Name, err)
	}

	// Generate timestamp for the request
	timestamp := time.Now().Format(time.RFC3339)
	req.Header.Set("X-Site-Sync-Timestamp", timestamp)

	// Generate HMAC signature if token is provided
	if source.Token != "" {
		validator := hmac.NewValidator(source.Token)
		// For GET request, body is empty
		signature := validator.GenerateSignature(timestamp, []byte{})
		req.Header.Set("X-Site-Sync-Signature", signature)

		logging.Logger.WithFields(map[string]interface{}{
			"source":    source.Name,
			"url":       url,
			"timestamp": timestamp,
		}).Debug("Generated HMAC signature for site sync request")
	} else {
		logging.Logger.WithFields(map[string]interface{}{
			"source": source.Name,
			"url":    url,
		}).Warn("No token provided for site sync - proceeding without HMAC authentication")
	}

	resp, err := client.Do(req)
	if err != nil {
		// Network request failed - log warning and return empty results
		logging.Logger.WithFields(map[string]interface{}{
			"source": source.Name,
			"url":    url,
			"error":  err.Error(),
		}).Warn("Failed to sync with site - no apps will be available from this source")
		return []handlers.AppStatus{}, []handlers.Location{}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// HTTP status errors are typically network/server issues, handle gracefully
		logging.Logger.WithFields(map[string]interface{}{
			"source":      source.Name,
			"url":         url,
			"status_code": resp.StatusCode,
		}).Warn("Site sync failed with non-200 status - no apps will be available from this source")
		return []handlers.AppStatus{}, []handlers.Location{}, nil
	}

	var response handlers.StatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		// JSON parsing errors could indicate version mismatch, handle gracefully
		logging.Logger.WithFields(map[string]interface{}{
			"source": source.Name,
			"url":    url,
			"error":  err.Error(),
		}).Warn("Failed to decode sync response from site - no apps will be available from this source")
		return []handlers.AppStatus{}, []handlers.Location{}, nil
	}

	logging.Logger.WithFields(map[string]interface{}{
		"source":         source.Name,
		"app_count":      len(response.Apps),
		"location_count": len(response.Locations),
	}).Info("Successfully received app statuses and locations from remote site")

	// Add the source name to each status and location to ensure correct identification
	for i := range response.Apps {
		response.Apps[i].Source = source.Name
	}
	for i := range response.Locations {
		response.Locations[i].Source = source.Name
	}

	logging.Logger.WithFields(map[string]interface{}{
		"source":         source.Name,
		"app_count":      len(response.Apps),
		"location_count": len(response.Locations),
	}).Debug("Returning app statuses and locations from remote site")

	return response.Apps, response.Locations, nil
}
