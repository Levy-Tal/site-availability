package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"site-availability/authentication/hmac"
	"site-availability/config"
	"site-availability/labels"
	"site-availability/logging"
	"sort"
	"strings"
	"sync"
	"time"
)

// normalizeOriginURL normalizes an origin URL for consistent cache key usage
func normalizeOriginURL(originURL string) string {
	if originURL == "" {
		return ""
	}

	// Parse the URL to handle normalization
	parsed, err := url.Parse(originURL)
	if err != nil {
		// If URL parsing fails, return as-is but log warning
		logging.Logger.WithField("origin_url", originURL).Warn("Failed to parse origin URL for normalization")
		return strings.TrimSpace(strings.ToLower(originURL))
	}

	// Normalize: lowercase scheme and host, remove default ports, remove trailing slash
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)

	// Remove default ports
	if parsed.Scheme == "http" && strings.HasSuffix(parsed.Host, ":80") {
		parsed.Host = strings.TrimSuffix(parsed.Host, ":80")
	} else if parsed.Scheme == "https" && strings.HasSuffix(parsed.Host, ":443") {
		parsed.Host = strings.TrimSuffix(parsed.Host, ":443")
	}

	// Remove trailing slash from path
	if parsed.Path == "/" {
		parsed.Path = ""
	}

	return parsed.String()
}

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
	Status    *string `json:"status"` // "up", "down", "unavailable", or nil for no apps
}

// FilterParams represents both system field and label filters
type FilterParams struct {
	SystemFields map[string]string // location=siteA, origin_url=http://a.com
	Labels       map[string]string // env=prod, team=platform
}

// Simple cache with label manager integration
var (
	appStatusCache = make(map[string]map[string]map[string]AppStatus) // [origin_url][source][app_name]AppStatus
	locationCache  = make(map[string][]Location)
	cacheMutex     sync.RWMutex

	// Label manager for fast label queries
	labelManager = labels.NewLabelManager()

	// Performance metrics
	updateMetrics = struct {
		totalUpdates     int64
		totalAppsAdded   int64
		totalAppsSkipped int64
		totalErrors      int64
		avgDurationMs    int64
	}{}
)

// GetAppStatusCache returns a copy of the appStatusCache
func GetAppStatusCache() []AppStatus {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	var apps []AppStatus
	for _, originApps := range appStatusCache {
		for _, sourceApps := range originApps {
			for _, status := range sourceApps {
				apps = append(apps, status)
			}
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

// UpdateLocationCache updates the locationCache for a given source with conflict resolution
func UpdateLocationCache(sourceName string, newLocations []Location, configuredLocations []config.Location) {
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

	// Filter out locations that conflict with server's configured locations (applies to all sources)
	// Create a map of configured location names for O(1) lookup
	configuredLocationNames := make(map[string]bool)
	for _, configLoc := range configuredLocations {
		configuredLocationNames[configLoc.Name] = true
	}

	// Filter out conflicting locations
	var locationsToCache []Location
	locationsDropped := 0
	for _, loc := range newLocations {
		if configuredLocationNames[loc.Name] {
			logging.Logger.WithFields(map[string]interface{}{
				"location_name": loc.Name,
				"source":        sourceName,
			}).Debug("Dropping scraped location that conflicts with server's configured location")
			locationsDropped++
			continue
		}
		locationsToCache = append(locationsToCache, loc)
	}

	logging.Logger.WithFields(map[string]interface{}{
		"source":            sourceName,
		"total_scraped":     len(newLocations),
		"locations_kept":    len(locationsToCache),
		"locations_dropped": locationsDropped,
	}).Info("Applied location conflict filtering for source")

	// Set the source for all locations and update cache
	locations := make([]Location, len(locationsToCache))
	for i, loc := range locationsToCache {
		locations[i] = Location{
			Name:      loc.Name,
			Latitude:  loc.Latitude,
			Longitude: loc.Longitude,
			Source:    sourceName,
			Status:    loc.Status,
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

// UpdateAppStatusResult contains the result of an update operation
type UpdateAppStatusResult struct {
	AppsAdded   int
	AppsSkipped int
	Error       error
}

// UpdateAppStatus updates the appStatusCache for a given source and merges labels
// Returns result with statistics and any errors encountered
func UpdateAppStatus(sourceName string, newStatuses []AppStatus, source config.Source, serverSettings config.ServerSettings) UpdateAppStatusResult {
	start := time.Now()

	// Input validation
	if sourceName == "" {
		err := fmt.Errorf("source name cannot be empty")
		logging.Logger.WithError(err).Error("Invalid input to UpdateAppStatus")
		updateMetrics.totalErrors++
		return UpdateAppStatusResult{Error: err}
	}

	if serverSettings.HostURL == "" {
		err := fmt.Errorf("server host_url cannot be empty")
		logging.Logger.WithError(err).Error("Invalid server settings in UpdateAppStatus")
		updateMetrics.totalErrors++
		return UpdateAppStatusResult{Error: err}
	}

	logging.Logger.WithFields(map[string]interface{}{
		"count":  len(newStatuses),
		"source": sourceName,
	}).Info("Updating app status cache for source")

	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	var result UpdateAppStatusResult

	// If no statuses provided, remove the source from all origin_url caches
	if len(newStatuses) == 0 {
		for originURL := range appStatusCache {
			delete(appStatusCache[originURL], sourceName)
			// Clean up empty origin_url entries
			if len(appStatusCache[originURL]) == 0 {
				delete(appStatusCache, originURL)
			}
		}
		updateLabelManager()
		updateMetrics.totalUpdates++

		duration := time.Since(start)
		updateMetrics.avgDurationMs = (updateMetrics.avgDurationMs + duration.Milliseconds()) / 2

		logging.Logger.WithFields(map[string]interface{}{
			"source":      sourceName,
			"duration_ms": duration.Milliseconds(),
		}).Info("Removed source from cache (empty status list)")

		return result
	}

	// Group apps by normalized origin_url
	appsByOrigin := make(map[string][]AppStatus)

	// Process each app with comprehensive validation and error handling
	for i, app := range newStatuses {
		// Validate app data
		if app.Name == "" {
			logging.Logger.WithFields(map[string]interface{}{
				"source":    sourceName,
				"app_index": i,
			}).Warn("Skipping app with empty name")
			result.AppsSkipped++
			continue
		}

		if app.Location == "" {
			logging.Logger.WithFields(map[string]interface{}{
				"source":   sourceName,
				"app_name": app.Name,
			}).Warn("Skipping app with empty location")
			result.AppsSkipped++
			continue
		}

		// Validate origin_url - MANDATORY
		if app.OriginURL == "" {
			logging.Logger.WithFields(map[string]interface{}{
				"source":   sourceName,
				"app_name": app.Name,
			}).Error("Skipping app with empty origin_url - this is mandatory")
			result.AppsSkipped++
			continue
		}

		// Validate status
		validStatuses := map[string]bool{"up": true, "down": true, "unavailable": true}
		if !validStatuses[app.Status] {
			logging.Logger.WithFields(map[string]interface{}{
				"source":   sourceName,
				"app_name": app.Name,
				"status":   app.Status,
			}).Warn("Invalid app status, treating as unavailable")
			app.Status = "unavailable"
		}

		// Merge labels with error handling
		if app.Labels == nil {
			app.Labels = make(map[string]string)
		}

		mergedLabels := labels.MergeLabels(serverSettings.Labels, source.Labels, app.Labels)
		if mergedLabels == nil {
			logging.Logger.WithFields(map[string]interface{}{
				"app":    app.Name,
				"source": sourceName,
			}).Warn("Label merge failed, using app labels only")
			mergedLabels = app.Labels
		}
		app.Labels = mergedLabels

		// Final validation before caching
		if len(app.Name) > 255 {
			logging.Logger.WithFields(map[string]interface{}{
				"app":         app.Name[:50] + "...",
				"source":      sourceName,
				"name_length": len(app.Name),
			}).Warn("App name is too long, truncating")
			app.Name = app.Name[:255]
		}

		// Group by normalized origin_url
		normalizedOriginURL := normalizeOriginURL(app.OriginURL)
		appsByOrigin[normalizedOriginURL] = append(appsByOrigin[normalizedOriginURL], app)

		logging.Logger.WithFields(map[string]interface{}{
			"app":                   app.Name,
			"status":                app.Status,
			"location":              app.Location,
			"source":                app.Source,
			"origin_url":            app.OriginURL,
			"normalized_origin_url": normalizedOriginURL,
			"label_count":           len(app.Labels),
		}).Debug("Processed app for caching")
	}

	// Now update the cache: replace entire source for each origin_url
	for normalizedOriginURL, apps := range appsByOrigin {
		// Initialize origin_url cache if needed
		if _, ok := appStatusCache[normalizedOriginURL]; !ok {
			appStatusCache[normalizedOriginURL] = make(map[string]map[string]AppStatus)
		}

		// Replace entire source cache for this origin_url
		appStatusCache[normalizedOriginURL][sourceName] = make(map[string]AppStatus)

		for _, app := range apps {
			appStatusCache[normalizedOriginURL][sourceName][app.Name] = app
		}

		logging.Logger.WithFields(map[string]interface{}{
			"origin_url": normalizedOriginURL,
			"source":     sourceName,
			"app_count":  len(apps),
		}).Debug("Updated cache for origin_url and source")

		result.AppsAdded += len(apps)
	}

	// Update label manager for fast label queries
	updateLabelManager()

	// Update metrics
	updateMetrics.totalUpdates++
	updateMetrics.totalAppsAdded += int64(result.AppsAdded)
	updateMetrics.totalAppsSkipped += int64(result.AppsSkipped)

	duration := time.Since(start)
	updateMetrics.avgDurationMs = (updateMetrics.avgDurationMs + duration.Milliseconds()) / 2

	// Count total apps for this source across all origin_urls
	totalAppsForSource := 0
	for _, originApps := range appStatusCache {
		if sourceApps, exists := originApps[sourceName]; exists {
			totalAppsForSource += len(sourceApps)
		}
	}

	logging.Logger.WithFields(map[string]interface{}{
		"source":       sourceName,
		"apps_added":   result.AppsAdded,
		"apps_skipped": result.AppsSkipped,
		"duration_ms":  duration.Milliseconds(),
		"total_apps":   totalAppsForSource,
	}).Info("Successfully updated app status cache for source")

	return result
}

// GetUpdateMetrics returns current performance metrics (for monitoring/debugging)
func GetUpdateMetrics() map[string]interface{} {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	return map[string]interface{}{
		"total_updates":      updateMetrics.totalUpdates,
		"total_apps_added":   updateMetrics.totalAppsAdded,
		"total_apps_skipped": updateMetrics.totalAppsSkipped,
		"total_errors":       updateMetrics.totalErrors,
		"avg_duration_ms":    updateMetrics.avgDurationMs,
		"cache_origin_urls":  len(appStatusCache),
	}
}

// updateLabelManager updates the label manager with current app data
func updateLabelManager() {
	// Get all current apps with full system field information
	var apps []labels.AppInfo
	for _, originApps := range appStatusCache {
		for _, sourceApps := range originApps {
			for _, app := range sourceApps {
				apps = append(apps, labels.AppInfo{
					Name:      app.Name,
					Location:  app.Location,
					Status:    app.Status,
					Source:    app.Source,
					OriginURL: app.OriginURL,
					Labels:    app.Labels,
				})
			}
		}
	}

	// Update label manager with both system fields and user labels
	labelManager.UpdateAppLabels(apps)

	logging.Logger.WithFields(map[string]interface{}{
		"total_apps":   len(apps),
		"total_fields": len(labelManager.GetLabelKeys()),
	}).Debug("Updated label manager with system fields and labels")
}

// calculateLocationStatus calculates the status of a location based on its apps
// Returns: "up" if all apps are up, "down" if any app is down, "unavailable" if any app is unavailable but none down, nil if no apps
func calculateLocationStatus(locationName string, apps []AppStatus) *string {
	appsInLocation := make([]AppStatus, 0)
	for _, app := range apps {
		if app.Location == locationName {
			appsInLocation = append(appsInLocation, app)
		}
	}

	if len(appsInLocation) == 0 {
		return nil // No apps means location status is null
	}

	hasDown := false
	hasUnavailable := false
	allUp := true

	for _, app := range appsInLocation {
		switch app.Status {
		case "down":
			hasDown = true
			allUp = false
		case "unavailable":
			hasUnavailable = true
			allUp = false
		case "up":
			// Keep allUp true
		default:
			// Unknown status treated as unavailable
			hasUnavailable = true
			allUp = false
		}
	}

	if allUp {
		status := "up"
		return &status
	} else if hasDown {
		status := "down"
		return &status
	} else if hasUnavailable {
		status := "unavailable"
		return &status
	}

	status := "unavailable" // Default fallback
	return &status
}

// GetLocationsWithStatus returns locations with calculated status
func GetLocationsWithStatus() []Location {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	var locations []Location
	apps := GetAppStatusCache()

	// Get all locations from cache
	for _, sourceLocations := range locationCache {
		for _, location := range sourceLocations {
			// Calculate status for this location
			status := calculateLocationStatus(location.Name, apps)
			locationWithStatus := Location{
				Name:      location.Name,
				Latitude:  location.Latitude,
				Longitude: location.Longitude,
				Source:    location.Source,
				Status:    status,
			}
			locations = append(locations, locationWithStatus)
		}
	}

	return locations
}

// convertToHandlersLocation converts config.Locations to handlers.Location
func convertToHandlersLocation(configLocations []config.Location) []Location {
	logging.Logger.Debug("Converting config locations to handler format")

	var locations []Location
	apps := GetAppStatusCache()

	for _, loc := range configLocations {
		logging.Logger.WithFields(map[string]interface{}{
			"name":      loc.Name,
			"latitude":  loc.Latitude,
			"longitude": loc.Longitude,
		}).Debug("Processing location")

		// Calculate status for this location
		status := calculateLocationStatus(loc.Name, apps)

		locations = append(locations, Location{
			Name:      loc.Name,
			Latitude:  loc.Latitude,
			Longitude: loc.Longitude,
			Source:    "", // Empty source indicates this server's locations
			Status:    status,
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

// parseFilters extracts both system field and label filters from query parameters
// Returns a single map where system fields use their direct names (location, status, etc.)
// and user labels use "labels." prefix (labels.env, labels.team, etc.)
// Special handling for status: multiple status values are joined with "|" for OR logic
// Supports: ?location=siteA&origin_url=http://a.com&labels.env=production&status=up&status=down
func parseFilters(queryParams url.Values) map[string]string {
	filters := make(map[string]string)
	unrecognizedParams := make([]string, 0)

	// Define allowed system fields for filtering
	allowedSystemFields := map[string]bool{
		"name":       true,
		"location":   true,
		"status":     true,
		"source":     true,
		"origin_url": true,
	}

	for key, values := range queryParams {
		if len(values) == 0 || values[0] == "" {
			continue
		}

		if strings.HasPrefix(key, "labels.") {
			// Label filter: labels.env=production -> store as "labels.env"
			filters[key] = values[0]
		} else if allowedSystemFields[key] {
			// Special handling for status field to support OR logic
			if key == "status" && len(values) > 1 {
				// Multiple status values: join with "|" for OR logic
				// Remove empty values
				validStatuses := make([]string, 0, len(values))
				for _, status := range values {
					if status != "" {
						validStatuses = append(validStatuses, status)
					}
				}
				if len(validStatuses) > 0 {
					filters[key] = strings.Join(validStatuses, "|")
				}
			} else {
				// System field filter: location=siteA -> store as "location"
				filters[key] = values[0]
			}
		} else {
			// Check for common mistakes and log them
			if strings.HasPrefix(key, "label[") && strings.HasSuffix(key, "]") {
				// Extract key from label[key] format
				labelKey := strings.TrimSuffix(strings.TrimPrefix(key, "label["), "]")
				unrecognizedParams = append(unrecognizedParams, fmt.Sprintf("'%s' (should be 'labels.%s')", key, labelKey))
			} else {
				unrecognizedParams = append(unrecognizedParams, fmt.Sprintf("'%s'", key))
			}
		}
	}

	if len(unrecognizedParams) > 0 {
		logging.Logger.WithFields(map[string]interface{}{
			"unrecognized_params": unrecognizedParams,
			"recognized_filters":  filters,
		}).Warn("Some query parameters were not recognized as valid filters")
	}

	logging.Logger.WithFields(map[string]interface{}{
		"filters": filters,
	}).Debug("Parsed unified filters from query parameters")

	return filters
}

// filterApps applies unified filtering using the LabelManager for O(1) performance
// Works for both system fields (name, location, etc.) and user labels (labels.env, etc.)
// Special handling: status field supports OR logic when multiple values are provided (separated by "|")
func filterApps(apps []AppStatus, filters map[string]string) ([]AppStatus, int) {
	start := time.Now()

	if len(filters) == 0 {
		return apps, 0 // No filters
	}

	// Handle status OR logic specially
	var statusFilteredApps []AppStatus
	statusFilter, hasStatusFilter := filters["status"]

	if hasStatusFilter && strings.Contains(statusFilter, "|") {
		// Multiple status values - use OR logic
		statusValues := strings.Split(statusFilter, "|")
		statusSet := make(map[string]bool)
		for _, status := range statusValues {
			statusSet[status] = true
		}

		// Filter apps by OR logic for status
		for _, app := range apps {
			if statusSet[app.Status] {
				statusFilteredApps = append(statusFilteredApps, app)
			}
		}

		// Remove status from filters for remaining AND logic processing
		remainingFilters := make(map[string]string)
		for k, v := range filters {
			if k != "status" {
				remainingFilters[k] = v
			}
		}

		// If no other filters, return status-filtered apps
		if len(remainingFilters) == 0 {
			filteredCount := len(apps) - len(statusFilteredApps)
			duration := time.Since(start)

			logging.Logger.WithFields(map[string]interface{}{
				"total_apps":     len(apps),
				"filtered_apps":  len(statusFilteredApps),
				"filtered_out":   filteredCount,
				"duration_μs":    duration.Microseconds(),
				"status_filters": statusValues,
			}).Debug("Applied status OR filter")

			return statusFilteredApps, filteredCount
		}

		// Apply remaining filters with AND logic to status-filtered apps
		if len(remainingFilters) > 0 {
			// Use LabelManager for remaining filters
			matchingAppIDs := labelManager.FindAppsByLabels(remainingFilters)

			// Convert status-filtered apps to ID map
			idToApp := make(map[string]AppStatus)
			for _, app := range statusFilteredApps {
				uniqueID := app.Source + ":" + app.Name
				idToApp[uniqueID] = app
			}

			// Intersect status-filtered apps with label-filtered apps
			finalApps := make([]AppStatus, 0)
			for _, appID := range matchingAppIDs {
				if app, exists := idToApp[appID]; exists {
					finalApps = append(finalApps, app)
				}
			}

			// Sort filtered apps by name for deterministic ordering
			sort.Slice(finalApps, func(i, j int) bool {
				return finalApps[i].Name < finalApps[j].Name
			})

			filteredCount := len(apps) - len(finalApps)
			duration := time.Since(start)

			logging.Logger.WithFields(map[string]interface{}{
				"total_apps":        len(apps),
				"status_filtered":   len(statusFilteredApps),
				"final_filtered":    len(finalApps),
				"filtered_out":      filteredCount,
				"duration_μs":       duration.Microseconds(),
				"status_filters":    statusValues,
				"remaining_filters": remainingFilters,
			}).Debug("Applied status OR filter with additional AND filters")

			return finalApps, filteredCount
		}
	}

	// Standard filtering using LabelManager (AND logic for all filters)
	matchingAppIDs := labelManager.FindAppsByLabels(filters)

	// Convert unique IDs back to full AppStatus objects
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

	// Sort filtered apps by name for deterministic ordering
	sort.Slice(filteredApps, func(i, j int) bool {
		return filteredApps[i].Name < filteredApps[j].Name
	})

	filteredCount := len(apps) - len(filteredApps)
	duration := time.Since(start)

	logging.Logger.WithFields(map[string]interface{}{
		"total_apps":    len(apps),
		"filtered_apps": len(filteredApps),
		"filtered_out":  filteredCount,
		"duration_μs":   duration.Microseconds(),
		"filters":       filters,
	}).Debug("Applied unified filters using field manager")

	return filteredApps, filteredCount
}

// ResetCacheForTesting resets all global cache state for test isolation
func ResetCacheForTesting() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	appStatusCache = make(map[string]map[string]map[string]AppStatus)
	locationCache = make(map[string][]Location)

	labelManager = labels.NewLabelManager()

	// Reset metrics
	updateMetrics.totalUpdates = 0
	updateMetrics.totalAppsAdded = 0
	updateMetrics.totalAppsSkipped = 0
	updateMetrics.totalErrors = 0
	updateMetrics.avgDurationMs = 0
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

	// Count total apps across all origins and sources
	totalApps := 0
	for _, originApps := range appStatusCache {
		for _, sourceApps := range originApps {
			totalApps += len(sourceApps)
		}
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

	// Parse query parameters for both system field and label filtering
	filters := parseFilters(r.URL.Query())

	// Return current statuses and locations from the global cache
	apps := GetAppStatusCache()
	locations := GetLocationCache()

	// Apply all filters if any were specified
	filteredApps, filteredCount := filterApps(apps, filters)

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
		"system_filters": filters,
		"filtered_count": filteredCount,
	}).Debug("Sync response sent")
}

// LocationStatusResponse represents the response for /api/locations
type LocationStatusResponse struct {
	Locations []Location `json:"locations"`
}

// AppsResponse represents the response for /api/apps
type AppsResponse struct {
	Apps []AppStatus `json:"apps"`
}

// LabelsResponse represents the response for /api/labels
type LabelsResponse struct {
	Labels []string `json:"labels"`
}

// GetLocations handles the /api/locations endpoint
func GetLocations(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	logging.Logger.Debug("Handling /api/locations request")

	// Parse query parameters for filtering
	filters := parseFilters(r.URL.Query())

	// Get all apps first
	allApps := GetAppStatusCache()

	// Apply filters if any were specified
	var filteredApps []AppStatus
	var filteredCount int

	if len(filters) > 0 {
		filteredApps, filteredCount = filterApps(allApps, filters)
		logging.Logger.WithFields(map[string]interface{}{
			"filters":        filters,
			"filtered_count": filteredCount,
			"filtered_apps":  len(filteredApps),
		}).Debug("Applied filters to apps for location status calculation")
	} else {
		filteredApps = allApps
	}

	// Get all locations but calculate status based on filtered apps only
	locations := GetLocationsWithStatus()

	// Recalculate status for each location based on filtered apps
	for i := range locations {
		locations[i].Status = calculateLocationStatus(locations[i].Name, filteredApps)
	}

	// Add server's own locations from config with status calculated from filtered apps
	serverLocations := convertToHandlersLocation(cfg.Locations)
	for i := range serverLocations {
		serverLocations[i].Status = calculateLocationStatus(serverLocations[i].Name, filteredApps)
	}
	locations = append(locations, serverLocations...)

	// If filtering is applied, only return locations that have apps matching the filter
	if len(filters) > 0 {
		locationsWithApps := make([]Location, 0)
		locationHasApps := make(map[string]bool)

		// Mark locations that have filtered apps
		for _, app := range filteredApps {
			locationHasApps[app.Location] = true
		}

		// Only include locations that have matching apps or explicitly include all locations
		for _, location := range locations {
			if locationHasApps[location.Name] || location.Status == nil {
				locationsWithApps = append(locationsWithApps, location)
			}
		}
		locations = locationsWithApps
	}

	response := LocationStatusResponse{
		Locations: locations,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logging.Logger.WithError(err).Error("Failed to encode locations response")
		http.Error(w, "Failed to encode locations", http.StatusInternalServerError)
		return
	}

	logging.Logger.WithFields(map[string]interface{}{
		"locations":      len(response.Locations),
		"filters":        filters,
		"filtered_count": filteredCount,
	}).Debug("Locations response sent with filtered status calculation")
}

// GetApps handles the /api/apps endpoint
func GetApps(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	logging.Logger.Debug("Handling /api/apps request")

	// Parse all query parameters for filtering (including location)
	filters := parseFilters(r.URL.Query())

	apps := GetAppStatusCache()

	// Apply all filters if any were specified
	if len(filters) > 0 {
		var filteredCount int
		apps, filteredCount = filterApps(apps, filters)
		logging.Logger.WithFields(map[string]interface{}{
			"filters":        filters,
			"filtered_count": filteredCount,
		}).Debug("Applied filters to apps")
	}

	response := AppsResponse{
		Apps: apps,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logging.Logger.WithError(err).Error("Failed to encode apps response")
		http.Error(w, "Failed to encode apps", http.StatusInternalServerError)
		return
	}

	logging.Logger.WithFields(map[string]interface{}{
		"apps":    len(response.Apps),
		"filters": filters,
	}).Debug("Apps response sent")
}

// GetLabels handles the /api/labels endpoint
func GetLabels(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	logging.Logger.Debug("Handling /api/labels request")

	// Check if a specific label name is requested (e.g., /api/labels?env)
	queryParams := r.URL.Query()

	// Find the first non-empty query parameter as the label name
	var requestedLabelKey string
	for key, values := range queryParams {
		if len(values) > 0 && values[0] != "" {
			requestedLabelKey = key
			break
		} else if len(values) == 0 || values[0] == "" {
			// Handle case where parameter exists but has no value (e.g., ?env)
			requestedLabelKey = key
			break
		}
	}

	if requestedLabelKey != "" {
		// Return all values for the specific label key
		labelValues := labelManager.GetLabelValues("labels." + requestedLabelKey)

		response := LabelsResponse{
			Labels: labelValues,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logging.Logger.WithError(err).Error("Failed to encode label values response")
			http.Error(w, "Failed to encode label values", http.StatusInternalServerError)
			return
		}

		logging.Logger.WithFields(map[string]interface{}{
			"label_key":    requestedLabelKey,
			"label_values": len(response.Labels),
		}).Debug("Label values response sent")
		return
	}

	// Default behavior: return all available label keys
	labelKeys := labelManager.GetLabelKeys()

	// Filter out system fields, only return user labels
	userLabels := make([]string, 0)
	for _, key := range labelKeys {
		if strings.HasPrefix(key, "labels.") {
			// Remove "labels." prefix to return clean label key
			userLabel := strings.TrimPrefix(key, "labels.")
			userLabels = append(userLabels, userLabel)
		}
	}

	response := LabelsResponse{
		Labels: userLabels,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logging.Logger.WithError(err).Error("Failed to encode labels response")
		http.Error(w, "Failed to encode labels", http.StatusInternalServerError)
		return
	}

	logging.Logger.WithFields(map[string]interface{}{
		"labels": len(response.Labels),
	}).Debug("Labels response sent")
}
