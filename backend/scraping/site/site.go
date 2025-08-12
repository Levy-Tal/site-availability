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

// SiteConfig represents the configuration for Site sources
type SiteConfig struct {
	URL   string `yaml:"url"`
	Token string `yaml:"token"`
}

// SiteScraper implements the scraping.Source interface for scraping other sites.
type SiteScraper struct {
	directScrapedSites []string // URLs of sites directly scraped by this server
}

func NewSiteScraper() *SiteScraper {
	return &SiteScraper{
		directScrapedSites: []string{},
	}
}

// SetDirectScrapedSites sets the list of site URLs that this server directly scrapes
// This is used for circular scraping prevention
func (s *SiteScraper) SetDirectScrapedSites(siteURLs []string) {
	s.directScrapedSites = siteURLs
}

// ValidateConfig validates the Site-specific configuration
func (s *SiteScraper) ValidateConfig(source config.Source) error {
	siteCfg, err := config.DecodeConfig[SiteConfig](source.Config, source.Name)
	if err != nil {
		return err
	}

	// Validate required fields
	if siteCfg.URL == "" {
		return fmt.Errorf("site source %s: missing 'url'", source.Name)
	}

	return nil
}

// Scrape fetches the status of all apps and locations from a remote site using the /sync endpoint.
// Since site scraping involves a single request, the maxParallel parameter is not used.
// Circular prevention is handled automatically using the configured directScrapedSites.
func (s *SiteScraper) Scrape(source config.Source, serverSettings config.ServerSettings, timeout time.Duration, maxParallel int, tlsConfig *tls.Config) ([]handlers.AppStatus, []handlers.Location, error) {
	return s.ScrapeWithCircularPrevention(source, serverSettings, timeout, maxParallel, tlsConfig, s.directScrapedSites)
}

// ScrapeWithCircularPrevention is like Scrape but includes circular scraping prevention logic
func (s *SiteScraper) ScrapeWithCircularPrevention(source config.Source, serverSettings config.ServerSettings, timeout time.Duration, maxParallel int, tlsConfig *tls.Config, directScrapedSites []string) ([]handlers.AppStatus, []handlers.Location, error) {
	// Decode the source-specific config
	siteCfg, err := config.DecodeConfig[SiteConfig](source.Config, source.Name)
	if err != nil {
		return nil, nil, err
	}

	url := fmt.Sprintf("%s/sync", siteCfg.URL)
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
	if siteCfg.Token != "" {
		validator := hmac.NewValidator(siteCfg.Token)
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

	// Apply circular scraping prevention filters before processing
	filteredApps := make([]handlers.AppStatus, 0, len(response.Apps))
	appsSkipped := 0

	// Convert directScrapedSites to map for O(1) lookup
	directSitesMap := make(map[string]bool)
	for _, site := range directScrapedSites {
		if site != "" {
			directSitesMap[site] = true
		}
	}

	for _, app := range response.Apps {
		// Rule: Drop apps where origin_url matches any site that this server directly scrapes
		// BUT keep apps where origin_url matches the site we're currently scraping from (siteCfg.URL)
		// ALSO drop apps where origin_url matches our own host_url (these are stale copies of our own apps)

		isDirectScraped := directSitesMap[app.OriginURL]
		isCurrentSite := app.OriginURL == siteCfg.URL
		isOwnServer := app.OriginURL == serverSettings.HostURL

		if app.OriginURL != "" && ((isDirectScraped && !isCurrentSite) || isOwnServer) {
			reason := "third-party directly scraped site"
			if isOwnServer {
				reason = "own server (stale copy)"
			}
			logging.Logger.WithFields(map[string]interface{}{
				"app":          app.Name,
				"location":     app.Location,
				"origin_url":   app.OriginURL,
				"source":       source.Name,
				"scraped_from": siteCfg.URL,
				"reason":       reason,
			}).Debug("Skipping app to prevent circular scraping")
			appsSkipped++
			continue
		}

		// Set the source name to this scraper's source
		app.Source = source.Name
		// Preserve the original OriginURL from the remote site (don't override it)
		// The OriginURL should indicate where the app originally came from, not where we scraped it from
		filteredApps = append(filteredApps, app)
	}

	// Replace the original apps with filtered apps
	response.Apps = filteredApps

	logging.Logger.WithFields(map[string]interface{}{
		"source":       source.Name,
		"total_apps":   len(response.Apps) + appsSkipped,
		"kept_apps":    len(response.Apps),
		"skipped_apps": appsSkipped,
	}).Debug("Applied circular scraping prevention filters")

	// Set correct source for locations
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
