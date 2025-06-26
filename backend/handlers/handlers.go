package handlers

import (
	"encoding/json"
	"net/http"
	"site-availability/authentication/hmac"
	"site-availability/config"
	"site-availability/logging"
	"sync"
	"time"
)

type AppStatus struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	Status   string `json:"status"`
	Source   string `json:"source"`
}

type StatusResponse struct {
	Locations []Location  `json:"locations"`
	Apps      []AppStatus `json:"apps"`
}

type Location struct {
	Name      string  `yaml:"name" json:"name"`
	Latitude  float64 `yaml:"latitude" json:"latitude"`
	Longitude float64 `yaml:"longitude" json:"longitude"`
}

var (
	appStatusCache = make(map[string]map[string]AppStatus)
	cacheMutex     sync.RWMutex
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

// UpdateAppStatus updates the appStatusCache for a given source
func UpdateAppStatus(sourceName string, newStatuses []AppStatus) {
	logging.Logger.WithFields(map[string]interface{}{
		"count":  len(newStatuses),
		"source": sourceName,
	}).Info("Updating app status cache for source")

	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// If no statuses provided, remove the source from cache
	if len(newStatuses) == 0 {
		delete(appStatusCache, sourceName)
		return
	}

	if _, ok := appStatusCache[sourceName]; !ok {
		appStatusCache[sourceName] = make(map[string]AppStatus)
	} else {
		// Clear existing statuses for this source
		appStatusCache[sourceName] = make(map[string]AppStatus)
	}

	for _, app := range newStatuses {
		logging.Logger.WithFields(map[string]interface{}{
			"app":      app.Name,
			"status":   app.Status,
			"location": app.Location,
			"source":   app.Source,
		}).Debug("Caching app status")
		appStatusCache[sourceName][app.Name] = app
	}
}

// GetAppStatus handles the /api/status endpoint
func GetAppStatus(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	logging.Logger.Debug("Handling /api/status request")

	apps := GetAppStatusCache()

	response := StatusResponse{
		Locations: convertToHandlersLocation(cfg.Locations),
		Apps:      apps,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logging.Logger.WithError(err).Error("Failed to encode status response")
		http.Error(w, "Failed to encode status", http.StatusInternalServerError)
		return
	}

	logging.Logger.WithFields(map[string]interface{}{
		"apps":      len(response.Apps),
		"locations": len(response.Locations),
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
		})
	}
	return locations
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

	// Return current statuses from the global cache
	statuses := GetAppStatusCache()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(statuses); err != nil {
		logging.Logger.WithError(err).Error("Failed to encode sync response")
		http.Error(w, "Failed to encode sync response", http.StatusInternalServerError)
		return
	}
}
