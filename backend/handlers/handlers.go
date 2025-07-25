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
	appStatusCache = make(map[string]map[string]AppStatus)
	locationCache  = make(map[string][]Location)
	cacheMutex     sync.RWMutex

	// Global deduplication to prevent circular scraping issues
	seenApps = make(map[string]bool)

	// Label manager for fast label queries
	labelManager = labels.NewLabelManager()
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

		// Clear deduplication entries for this source to allow fresh apps
		// Remove all entries that belong to this source (by origin URL)
		for dedupKey := range seenApps {
			// Extract origin URL from dedupKey (format: appName|location|originURL)
			parts := strings.Split(dedupKey, "|")
			if len(parts) == 3 {
				originURL := parts[2]
				// Check if this origin URL matches any of the apps from this source
				for _, app := range newStatuses {
					if app.OriginURL == originURL {
						delete(seenApps, dedupKey)
						break
					}
				}
			}
		}
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
	// Get all current apps with full system field information
	var apps []labels.AppInfo
	for _, sourceApps := range appStatusCache {
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

	appStatusCache = make(map[string]map[string]AppStatus)
	locationCache = make(map[string][]Location)
	seenApps = make(map[string]bool) // Reset deduplication map
	labelManager = labels.NewLabelManager()
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
