package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"site-availability/authentication/hmac"
	"site-availability/config"
	"site-availability/labels"
	"site-availability/logging"
	"strings"
	"sync"
	"time"
)

type AppStatus struct {
	Name      string            `json:"name"`
	Location  string            `json:"location"`
	Status    string            `json:"status"`
	Source    string            `json:"source"`
	OriginURL string            `json:"origin_url,omitempty"` // URL where app originally came from
	Labels    map[string]string `json:"labels,omitempty"`     // App labels (merged from app + source + server)
}

type StatusResponse struct {
	Locations []Location  `json:"locations"`
	Apps      []AppStatus `json:"apps"`
}

type Location struct {
	Name      string  `yaml:"name" json:"name"`
	Latitude  float64 `yaml:"latitude" json:"latitude"`
	Longitude float64 `yaml:"longitude" json:"longitude"`
	Source    string  `json:"source"`
}

// Simple cache with label manager integration
var (
	appStatusCache = make(map[string]map[string]AppStatus)
	locationCache  = make(map[string][]Location)
	cacheMutex     sync.RWMutex

	// Global deduplication to prevent circular scraping issues
	seenApps = make(map[string]bool)

	// Label manager for fast label queries
	labelManager = labels.NewLabelManager()

	// Simple performance counters
	cacheStats = struct {
		sync.RWMutex
		LabelQueries int64
		CacheHits    int64
		CacheMisses  int64
	}{}
)

// GetAppStatusCache returns a copy of the appStatusCache
func GetAppStatusCache() []AppStatus {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	var apps []AppStatus
	for _, sourceApps := range appStatusCache {
		for _, status := range sourceApps {
			apps = append(apps, status)
		}
	}
	return apps
}

// GetLocationCache returns a copy of all locations from all sources
func GetLocationCache() []Location {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	var locations []Location
	for _, sourceLocations := range locationCache {
		locations = append(locations, sourceLocations...)
	}
	return locations
}

// UpdateLocationCache updates the locationCache for a given source
func UpdateLocationCache(sourceName string, newLocations []Location) {
	logging.Logger.WithFields(map[string]interface{}{
		"count":  len(newLocations),
		"source": sourceName,
	}).Info("Updating location cache for source")

	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// If no locations provided, remove the source from cache
	if len(newLocations) == 0 {
		delete(locationCache, sourceName)
		return
	}

	// Set the source for all locations and update cache
	locations := make([]Location, len(newLocations))
	for i, loc := range newLocations {
		locations[i] = Location{
			Name:      loc.Name,
			Latitude:  loc.Latitude,
			Longitude: loc.Longitude,
			Source:    sourceName,
		}
		logging.Logger.WithFields(map[string]interface{}{
			"name":      loc.Name,
			"latitude":  loc.Latitude,
			"longitude": loc.Longitude,
			"source":    sourceName,
		}).Debug("Caching location")
	}

	locationCache[sourceName] = locations
}

// UpdateAppStatus updates the appStatusCache for a given source and merges labels
func UpdateAppStatus(sourceName string, newStatuses []AppStatus, source config.Source, serverSettings config.ServerSettings) {
	logging.Logger.WithFields(map[string]interface{}{
		"count":  len(newStatuses),
		"source": sourceName,
	}).Info("Updating app status cache for source")

	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// If no statuses provided, remove the source from cache
	if len(newStatuses) == 0 {
		delete(appStatusCache, sourceName)
		updateLabelManager()
		return
	}

	if _, ok := appStatusCache[sourceName]; !ok {
		appStatusCache[sourceName] = make(map[string]AppStatus)
	} else {
		// Clear existing statuses for this source
		appStatusCache[sourceName] = make(map[string]AppStatus)
	}

	for _, app := range newStatuses {
		// Only apply deduplication for apps with valid origin URLs to prevent circular scraping
		// Empty origin URLs are allowed (for direct prometheus scrapers, etc.)
		if app.OriginURL != "" {
			dedupKey := app.Name + "|" + app.Location + "|" + app.OriginURL

			// Skip if we've already seen this app from the same origin
			if seenApps[dedupKey] {
				logging.Logger.WithFields(map[string]interface{}{
					"app":        app.Name,
					"location":   app.Location,
					"origin_url": app.OriginURL,
					"source":     sourceName,
				}).Debug("Skipping duplicate app from circular scraping")
				continue
			}

			// Mark as seen
			seenApps[dedupKey] = true
		}

		// Merge labels: App labels > Source labels > Server labels
		app.Labels = labels.MergeLabels(serverSettings.Labels, source.Labels, app.Labels)

		logging.Logger.WithFields(map[string]interface{}{
			"app":         app.Name,
			"status":      app.Status,
			"location":    app.Location,
			"source":      app.Source,
			"origin_url":  app.OriginURL,
			"label_count": len(app.Labels),
		}).Debug("Caching app status with merged labels")
		appStatusCache[sourceName][app.Name] = app
	}

	// Update label manager for fast label queries
	updateLabelManager()
}

// updateLabelManager updates the label manager with current app data
func updateLabelManager() {
	// Get all current apps
	var apps []labels.AppInfo
	for sourceName, sourceApps := range appStatusCache {
		for _, app := range sourceApps {
			apps = append(apps, labels.AppInfo{
				Name:   app.Name,
				Source: sourceName, // Include source to ensure uniqueness
				Labels: app.Labels,
			})
		}
	}

	// Update label manager
	labelManager.UpdateAppLabels(apps)

	logging.Logger.WithFields(map[string]interface{}{
		"total_apps": len(apps),
		"label_keys": len(labelManager.GetLabelKeys()),
	}).Debug("Updated label manager")
}

// GetAppStatus handles the /api/status endpoint
func GetAppStatus(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	logging.Logger.Debug("Handling /api/status request")

	// Parse query parameters for label filtering
	labelFilters := parseLabelFilters(r.URL.Query())

	apps := GetAppStatusCache()
	locations := GetLocationCache()

	// Apply label filters if any were specified
	filteredApps, filteredCount := filterAppsByLabels(apps, labelFilters)

	// Add server's own locations from config with empty source
	serverLocations := convertToHandlersLocation(cfg.Locations)
	locations = append(locations, serverLocations...)

	response := StatusResponse{
		Locations: locations,
		Apps:      filteredApps,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logging.Logger.WithError(err).Error("Failed to encode status response")
		http.Error(w, "Failed to encode status", http.StatusInternalServerError)
		return
	}

	logging.Logger.WithFields(map[string]interface{}{
		"apps":           len(response.Apps),
		"locations":      len(response.Locations),
		"label_filters":  labelFilters,
		"filtered_count": filteredCount,
	}).Debug("Status response sent")
}

// convertToHandlersLocation converts config.Locations to handlers.Location
func convertToHandlersLocation(configLocations []config.Location) []Location {
	logging.Logger.Debug("Converting config locations to handler format")

	var locations []Location
	for _, loc := range configLocations {
		logging.Logger.WithFields(map[string]interface{}{
			"name":      loc.Name,
			"latitude":  loc.Latitude,
			"longitude": loc.Longitude,
		}).Debug("Processing location")

		locations = append(locations, Location{
			Name:      loc.Name,
			Latitude:  loc.Latitude,
			Longitude: loc.Longitude,
			Source:    "", // Empty source indicates this server's locations
		})
	}
	return locations
}

// parseLabelFilters extracts label filters from query parameters
// Supports format: ?labels.key1=value1&labels.key2=value2
func parseLabelFilters(queryParams url.Values) map[string]string {
	labelFilters := make(map[string]string)

	for key, values := range queryParams {
		// Check if this is a label filter parameter
		if strings.HasPrefix(key, "labels.") && len(values) > 0 {
			// Extract the label key (remove "labels." prefix)
			labelKey := strings.TrimPrefix(key, "labels.")
			if labelKey != "" && values[0] != "" {
				labelFilters[labelKey] = values[0] // Use first value if multiple provided
			}
		}
	}

	logging.Logger.WithField("label_filters", labelFilters).Debug("Parsed label filters from query parameters")
	return labelFilters
}

// filterAppsByLabels filters apps using the label manager for fast queries
// Returns filtered apps and the count of apps that were filtered out
func filterAppsByLabels(apps []AppStatus, labelFilters map[string]string) ([]AppStatus, int) {
	start := time.Now()

	// Update stats
	cacheStats.Lock()
	cacheStats.LabelQueries++
	cacheStats.Unlock()

	if len(labelFilters) == 0 {
		return apps, 0 // No filters, return all apps
	}

	// Use LabelManager for fast filtering - returns unique IDs like "source:appname"
	matchingAppIDs := labelManager.FindAppsByLabels(labelFilters)

	// Convert unique IDs back to full AppStatus objects
	// Use source:name as key to handle apps with same names from different sources
	idToApp := make(map[string]AppStatus, len(apps))
	for _, app := range apps {
		uniqueID := app.Source + ":" + app.Name
		idToApp[uniqueID] = app
	}

	filteredApps := make([]AppStatus, 0, len(matchingAppIDs))
	for _, appID := range matchingAppIDs {
		if app, exists := idToApp[appID]; exists {
			filteredApps = append(filteredApps, app)
		}
	}

	filteredCount := len(apps) - len(filteredApps)
	duration := time.Since(start)

	logging.Logger.WithFields(map[string]interface{}{
		"total_apps":    len(apps),
		"filtered_apps": len(filteredApps),
		"filtered_out":  filteredCount,
		"duration_μs":   duration.Microseconds(),
		"label_filters": labelFilters,
	}).Debug("Applied label filters using label manager")

	return filteredApps, filteredCount
}

// GetCacheStats returns simple cache statistics
func GetCacheStats() map[string]interface{} {
	cacheStats.RLock()
	defer cacheStats.RUnlock()

	cacheMutex.RLock()
	totalApps := 0
	for _, sourceApps := range appStatusCache {
		totalApps += len(sourceApps)
	}
	cacheMutex.RUnlock()

	return map[string]interface{}{
		"total_apps":    totalApps,
		"label_queries": cacheStats.LabelQueries,
		"label_keys":    len(labelManager.GetLabelKeys()),
		"label_manager": labelManager.GetStats(),
	}
}

// ResetCacheForTesting resets all global cache state for test isolation
func ResetCacheForTesting() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	appStatusCache = make(map[string]map[string]AppStatus)
	locationCache = make(map[string][]Location)
	seenApps = make(map[string]bool) // Reset deduplication map
	labelManager = labels.NewLabelManager()

	cacheStats.Lock()
	cacheStats.LabelQueries = 0
	cacheStats.CacheHits = 0
	cacheStats.CacheMisses = 0
	cacheStats.Unlock()
}

// IsAppStatusCacheEmpty checks if the app status cache is empty
func IsAppStatusCacheEmpty() bool {
	logging.Logger.Debug("Checking if app status cache is empty")

	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	// Check if there are no sources or all sources are empty
	if len(appStatusCache) == 0 {
		logging.Logger.WithField("empty", true).Debug("Cache is empty - no sources")
		return true
	}

	// Count total apps across all sources
	totalApps := 0
	for _, sourceApps := range appStatusCache {
		totalApps += len(sourceApps)
	}

	empty := totalApps == 0
	logging.Logger.WithField("empty", empty).Debug("Checking if app status cache is empty")
	return empty
}

func GetScrapeInterval(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	logging.Logger.Debug("Handling /api/scrape-interval request")

	// Parse the scrape interval string into a time.Duration
	duration, err := time.ParseDuration(cfg.Scraping.Interval)
	if err != nil {
		logging.Logger.WithError(err).Error("Invalid scrape interval format")
		http.Error(w, "Invalid scrape interval format", http.StatusInternalServerError)
		return
	}

	// Convert the duration to milliseconds
	intervalInMs := duration.Milliseconds()

	// Create the response
	response := map[string]int64{
		"scrape_interval_ms": intervalInMs,
	}

	// Encode the response as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logging.Logger.WithError(err).Error("Failed to encode scrape interval response")
		http.Error(w, "Failed to encode scrape interval", http.StatusInternalServerError)
		return
	}

	logging.Logger.WithField("scrape_interval_ms", intervalInMs).Debug("Scrape interval response sent")
}

func GetDocs(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	logging.Logger.Debug("Handling /api/docs request")

	// Create the response
	response := map[string]string{
		"docs_title": cfg.Documentation.Title,
		"docs_url":   cfg.Documentation.URL,
	}

	// Encode the response as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logging.Logger.WithError(err).Error("Failed to encode docs response")
		http.Error(w, "Failed to encode docs", http.StatusInternalServerError)
		return
	}

	logging.Logger.WithField("docs response:", response).Debug("Docs response sent")
}

// HandleSyncRequest handles the /sync endpoint
func HandleSyncRequest(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	logging.Logger.Debug("Handling /sync request")

	// Check if sync is enabled
	if !cfg.ServerSettings.SyncEnable {
		http.Error(w, "Sync is disabled", http.StatusForbidden)
		return
	}

	// Validate the request using the server's token if available
	if cfg.ServerSettings.Token != "" {
		validator := hmac.NewValidator(cfg.ServerSettings.Token)
		if !validator.ValidateRequest(r) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	// Parse query parameters for label filtering
	labelFilters := parseLabelFilters(r.URL.Query())

	// Return current statuses and locations from the global cache
	apps := GetAppStatusCache()
	locations := GetLocationCache()

	// Apply label filters if any were specified
	filteredApps, filteredCount := filterAppsByLabels(apps, labelFilters)

	// Add server's own locations from config with empty source
	serverLocations := convertToHandlersLocation(cfg.Locations)
	locations = append(locations, serverLocations...)

	response := StatusResponse{
		Locations: locations,
		Apps:      filteredApps,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logging.Logger.WithError(err).Error("Failed to encode sync response")
		http.Error(w, "Failed to encode sync response", http.StatusInternalServerError)
		return
	}

	logging.Logger.WithFields(map[string]interface{}{
		"apps":           len(response.Apps),
		"locations":      len(response.Locations),
		"label_filters":  labelFilters,
		"filtered_count": filteredCount,
	}).Debug("Sync response sent")
}
