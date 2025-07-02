package labels

import (
	"site-availability/logging"
	"strings"
	"sync"
)

// AppInfo represents the minimal app information needed for label management
// This avoids circular imports with the handlers package
type AppInfo struct {
	Name      string
	Location  string
	Status    string
	Source    string
	OriginURL string
	Labels    map[string]string
}

// getUniqueID creates a unique identifier for an app by combining source and name
func (app AppInfo) getUniqueID() string {
	if app.Source == "" {
		return app.Name // Backward compatibility for apps without source
	}
	return app.Source + ":" + app.Name
}

// LabelManager manages field-to-app mappings for fast queries and provides
// label merging functionality for the application.
// It indexes both system fields (name, location, etc.) and user labels.
type LabelManager struct {
	// Performance optimization: map[field_name][field_value] -> []unique_app_ids
	// This enables O(1) lookups for "find all apps with field X=Y"
	// Works for both system fields and labels with "labels." prefix
	// unique_app_ids are in format "source:appname" to avoid conflicts
	appsByField map[string]map[string][]string
	mutex       sync.RWMutex
}

// NewLabelManager creates a new LabelManager instance
func NewLabelManager() *LabelManager {
	return &LabelManager{
		appsByField: make(map[string]map[string][]string),
	}
}

// MergeLabels combines labels from server, source, and app levels.
// Priority order: App labels (highest) > Source labels > Server labels (lowest)
// Returns a new map with all merged labels
func MergeLabels(serverLabels, sourceLabels, appLabels map[string]string) map[string]string {
	// Calculate total capacity for efficient allocation
	totalSize := len(serverLabels) + len(sourceLabels) + len(appLabels)
	merged := make(map[string]string, totalSize)

	// Apply in priority order (lower priority first, so higher can override)

	// 1. Server labels (lowest priority)
	for key, value := range serverLabels {
		merged[key] = value
	}

	// 2. Source labels (medium priority)
	for key, value := range sourceLabels {
		merged[key] = value
	}

	// 3. App labels (highest priority)
	for key, value := range appLabels {
		merged[key] = value
	}

	logging.Logger.WithFields(map[string]interface{}{
		"server_labels": len(serverLabels),
		"source_labels": len(sourceLabels),
		"app_labels":    len(appLabels),
		"merged_labels": len(merged),
	}).Debug("Merged labels from all levels")

	return merged
}

// UpdateAppLabels updates the internal field-to-app mapping for fast queries.
// This indexes both system fields and user labels for O(1) filtering.
// This should be called whenever the app status cache is updated.
func (lm *LabelManager) UpdateAppLabels(apps []AppInfo) {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	// Clear existing mappings
	lm.appsByField = make(map[string]map[string][]string)

	logging.Logger.WithField("app_count", len(apps)).Debug("Updating field-to-app mappings")

	// Build new mappings using unique identifiers
	for _, app := range apps {
		uniqueID := app.getUniqueID()

		// Index system fields for direct filtering (e.g., ?location=Hadera)
		systemFields := map[string]string{
			"name":       app.Name,
			"location":   app.Location,
			"status":     app.Status,
			"source":     app.Source,
			"origin_url": app.OriginURL,
		}

		for fieldName, fieldValue := range systemFields {
			if fieldValue != "" { // Only index non-empty values
				lm.indexField(fieldName, fieldValue, uniqueID)
			}
		}

		// Index user labels with "labels." prefix (e.g., ?labels.env=prod)
		for labelKey, labelValue := range app.Labels {
			if labelValue != "" { // Only index non-empty values
				lm.indexField("labels."+labelKey, labelValue, uniqueID)
			}
		}
	}

	// Log statistics for observability
	totalMappings := 0
	systemFieldCount := 0
	labelFieldCount := 0

	for fieldName, valueMap := range lm.appsByField {
		if strings.HasPrefix(fieldName, "labels.") {
			labelFieldCount++
		} else {
			systemFieldCount++
		}
		for _, appList := range valueMap {
			totalMappings += len(appList)
		}
	}

	logging.Logger.WithFields(map[string]interface{}{
		"total_fields":   len(lm.appsByField),
		"system_fields":  systemFieldCount,
		"label_fields":   labelFieldCount,
		"total_mappings": totalMappings,
	}).Debug("Field-to-app mappings updated with system fields and labels")
}

// indexField adds an app to the field index
func (lm *LabelManager) indexField(fieldName, fieldValue, uniqueID string) {
	if lm.appsByField[fieldName] == nil {
		lm.appsByField[fieldName] = make(map[string][]string)
	}

	lm.appsByField[fieldName][fieldValue] = append(
		lm.appsByField[fieldName][fieldValue],
		uniqueID,
	)
}

// FindAppsByLabel returns a list of unique app identifiers that have the specified label key-value pair.
// Returns unique identifiers in format "source:appname" to avoid conflicts between apps with same names.
// Returns an empty slice if no apps match the criteria.
func (lm *LabelManager) FindAppsByLabel(labelKey, labelValue string) []string {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()

	if valueMap, exists := lm.appsByField[labelKey]; exists {
		if appIDs, exists := valueMap[labelValue]; exists {
			// Return a copy to prevent external modification
			result := make([]string, len(appIDs))
			copy(result, appIDs)
			return result
		}
	}

	return []string{}
}

// FindAppsByLabels returns a list of unique app identifiers that have ALL the specified label key-value pairs.
// Returns unique identifiers in format "source:appname" to avoid conflicts between apps with same names.
// This is useful for complex filtering like "team=A AND environment=prod"
func (lm *LabelManager) FindAppsByLabels(labels map[string]string) []string {
	if len(labels) == 0 {
		return []string{}
	}

	lm.mutex.RLock()
	defer lm.mutex.RUnlock()

	var candidateApps []string
	isFirst := true

	// For each required label, find matching apps
	for labelKey, labelValue := range labels {
		// Look up apps with this specific label/value pair (inline to avoid nested locking)
		var currentApps []string
		if valueMap, exists := lm.appsByField[labelKey]; exists {
			if appIDs, exists := valueMap[labelValue]; exists {
				// Return a copy to prevent external modification
				currentApps = make([]string, len(appIDs))
				copy(currentApps, appIDs)
			}
		}

		if isFirst {
			// First iteration: start with all apps that match this label
			candidateApps = currentApps
			isFirst = false
		} else {
			// Subsequent iterations: intersect with previous results
			candidateApps = intersectSlices(candidateApps, currentApps)
		}

		// Early exit if no apps match current label
		if len(candidateApps) == 0 {
			break
		}
	}

	return candidateApps
}

// GetLabelKeys returns all available label keys across all apps
func (lm *LabelManager) GetLabelKeys() []string {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()

	keys := make([]string, 0, len(lm.appsByField))
	for key := range lm.appsByField {
		keys = append(keys, key)
	}
	return keys
}

// GetLabelValues returns all available values for a specific label key
func (lm *LabelManager) GetLabelValues(labelKey string) []string {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()

	if valueMap, exists := lm.appsByField[labelKey]; exists {
		values := make([]string, 0, len(valueMap))
		for value := range valueMap {
			values = append(values, value)
		}
		return values
	}
	return []string{}
}

// intersectSlices returns elements that exist in both slices
func intersectSlices(slice1, slice2 []string) []string {
	if len(slice1) == 0 || len(slice2) == 0 {
		return []string{}
	}

	// Create a map for O(1) lookups
	set := make(map[string]bool)
	for _, item := range slice1 {
		set[item] = true
	}

	var intersection []string
	for _, item := range slice2 {
		if set[item] {
			intersection = append(intersection, item)
			// Remove from set to avoid duplicates
			delete(set, item)
		}
	}

	return intersection
}
