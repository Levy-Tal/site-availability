package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"site-availability/authentication/middleware"
	"site-availability/authentication/rbac"
	"site-availability/config"
	"site-availability/logging"
)

// GetLabelsWithAuthz handles the /api/labels endpoint with authorization filtering
func GetLabelsWithAuthz(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	logging.Logger.Debug("Handling /api/labels request with authorization")

	// Get user permissions from context
	userPermissions, hasPermissions := middleware.GetPermissionsFromContext(r)

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
		// Return values for the specific label key, filtered by user permissions
		allLabelValues := labelManager.GetLabelValues("labels." + requestedLabelKey)

		var filteredValues []string
		if hasPermissions && !userPermissions.HasFullAccess {
			// Filter values based on user permissions
			if labelPerm, hasAccess := userPermissions.AllowedLabels[requestedLabelKey]; hasAccess {
				// Only return values the user is allowed to see
				for _, value := range allLabelValues {
					for _, allowedValue := range labelPerm.AllowedValues {
						if value == allowedValue {
							filteredValues = append(filteredValues, value)
							break
						}
					}
				}
			}
			// If user doesn't have permission for this label, return empty array
		} else {
			// Admin or no auth - return all values
			filteredValues = allLabelValues
		}

		response := LabelsResponse{
			Labels: filteredValues,
		}

		writeJSONResponse(w, response, "label values")

		logging.Logger.WithFields(map[string]interface{}{
			"label_key":       requestedLabelKey,
			"total_values":    len(allLabelValues),
			"filtered_values": len(filteredValues),
			"has_permissions": hasPermissions,
			"is_admin":        userPermissions.HasFullAccess,
		}).Debug("Filtered label values response sent")
		return
	}

	// Default behavior: return label keys the user has access to
	labelKeys := labelManager.GetLabelKeys()

	var userLabels []string
	for _, key := range labelKeys {
		if strings.HasPrefix(key, "labels.") {
			// Remove "labels." prefix to return clean label key
			userLabel := strings.TrimPrefix(key, "labels.")

			// Check if user has permission to access this label
			if hasPermissions && !userPermissions.HasFullAccess {
				// Only include labels the user has permission for
				if _, hasAccess := userPermissions.AllowedLabels[userLabel]; hasAccess {
					userLabels = append(userLabels, userLabel)
				}
			} else {
				// Admin or no auth - include all labels
				userLabels = append(userLabels, userLabel)
			}
		}
	}

	response := LabelsResponse{
		Labels: userLabels,
	}

	writeJSONResponse(w, response, "labels")

	logging.Logger.WithFields(map[string]interface{}{
		"total_labels":    len(labelKeys),
		"filtered_labels": len(userLabels),
		"has_permissions": hasPermissions,
		"is_admin":        userPermissions.HasFullAccess,
	}).Debug("Filtered labels response sent")
}

// GetAppsWithAuthz handles the /api/apps endpoint with authorization filtering
func GetAppsWithAuthz(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	logging.Logger.Debug("Handling /api/apps request with authorization")

	// Parse all query parameters for filtering (including location)
	filters := parseFilters(r.URL.Query())

	// Get user permissions from context
	userPermissions, hasPermissions := middleware.GetPermissionsFromContext(r)

	apps := GetAppStatusCache()

	// Apply authorization filters if user doesn't have full access
	if hasPermissions && !userPermissions.HasFullAccess {
		// Create authorizer to filter apps
		authorizer := rbac.NewAuthorizer(cfg)

		var authorizedApps []AppStatus
		for _, app := range apps {
			if authorizer.CanAccessApp(userPermissions, app.Labels) {
				authorizedApps = append(authorizedApps, app)
			}
		}
		apps = authorizedApps
	}

	// Apply regular filters if any were specified
	if len(filters) > 0 {
		var filteredCount int
		apps, filteredCount = filterApps(apps, filters)
		logging.Logger.WithFields(map[string]interface{}{
			"filters":        filters,
			"filtered_count": filteredCount,
		}).Debug("Applied regular filters to apps")
	}

	response := AppsResponse{
		Apps: apps,
	}

	writeJSONResponse(w, response, "apps")

	logging.Logger.WithFields(map[string]interface{}{
		"apps":            len(response.Apps),
		"filters":         filters,
		"has_permissions": hasPermissions,
		"is_admin":        userPermissions.HasFullAccess,
	}).Debug("Filtered apps response sent")
}

// GetLocationsWithAuthz handles the /api/locations endpoint with authorization filtering
func GetLocationsWithAuthz(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	logging.Logger.Debug("Handling /api/locations request with authorization")

	// Parse query parameters for filtering
	filters := parseFilters(r.URL.Query())

	// Get user permissions from context
	userPermissions, hasPermissions := middleware.GetPermissionsFromContext(r)

	// Get all apps first
	allApps := GetAppStatusCache()

	// Apply authorization filtering first if user doesn't have full access
	var authorizedApps []AppStatus
	if hasPermissions && !userPermissions.HasFullAccess {
		// Create authorizer to filter apps
		authorizer := rbac.NewAuthorizer(cfg)

		for _, app := range allApps {
			if authorizer.CanAccessApp(userPermissions, app.Labels) {
				authorizedApps = append(authorizedApps, app)
			}
		}
	} else {
		// Admin or no auth - use all apps
		authorizedApps = allApps
	}

	// Apply regular filters if specified
	var filteredApps []AppStatus
	var filteredCount int
	if len(filters) > 0 {
		filteredApps, filteredCount = filterApps(authorizedApps, filters)
		logging.Logger.WithFields(map[string]interface{}{
			"filters":        filters,
			"filtered_count": filteredCount,
			"filtered_apps":  len(filteredApps),
		}).Debug("Applied filters to apps for location status calculation")
	} else {
		filteredApps = authorizedApps
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

	// If filtering is applied (either auth or regular), only return locations that have apps matching the filter
	if len(filters) > 0 || (hasPermissions && !userPermissions.HasFullAccess) {
		locationsWithApps := make([]Location, 0)
		locationHasApps := make(map[string]bool)

		// Check which locations have matching apps
		for _, app := range filteredApps {
			locationHasApps[app.Location] = true
		}

		// Only include locations that have matching apps
		for _, location := range locations {
			if locationHasApps[location.Name] {
				locationsWithApps = append(locationsWithApps, location)
			}
		}
		locations = locationsWithApps
	}

	response := LocationStatusResponse{
		Locations: locations,
	}

	writeJSONResponse(w, response, "locations")

	logging.Logger.WithFields(map[string]interface{}{
		"total_apps":      len(allApps),
		"authorized_apps": len(authorizedApps),
		"filtered_apps":   len(filteredApps),
		"locations":       len(response.Locations),
		"filters":         filters,
		"has_permissions": hasPermissions,
		"is_admin":        userPermissions.HasFullAccess,
	}).Debug("Filtered locations response sent")
}

// writeJSONResponse is a helper function to write JSON responses
func writeJSONResponse(w http.ResponseWriter, response interface{}, responseType string) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logging.Logger.WithError(err).Errorf("Failed to encode %s response", responseType)
		http.Error(w, fmt.Sprintf("Failed to encode %s", responseType), http.StatusInternalServerError)
		return
	}
}
