package labels

import (
	"site-availability/logging"
	"sync"
)

// AppInfo represents the minimal app information needed for label management
// This avoids circular imports with the handlers package
type AppInfo struct {
	Name   string
	Source string // New field to ensure uniqueness
	Labels map[string]string
}

// getUniqueID creates a unique identifier for an app by combining source and name
func (app AppInfo) getUniqueID() string {
	if app.Source == "" {
		return app.Name // Backward compatibility for apps without source
	}
	return app.Source + ":" + app.Name
}

// LabelManager manages label-to-app mappings for fast queries and provides
// label merging functionality for the application
type LabelManager struct {
	// Performance optimization: map[label_key][label_value] -> []unique_app_ids
	// This enables O(1) lookups for "find all apps with label X=Y"
	// unique_app_ids are in format "source:appname" to avoid conflicts
	appsByLabel map[string]map[string][]string
	mutex       sync.RWMutex
}

// NewLabelManager creates a new LabelManager instance
func NewLabelManager() *LabelManager {
	return &LabelManager{
		appsByLabel: make(map[string]map[string][]string),
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

// UpdateAppLabels updates the internal label-to-app mapping for fast label-based queries.
// This should be called whenever the app status cache is updated.
func (lm *LabelManager) UpdateAppLabels(apps []AppInfo) {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	// Clear existing mappings
	lm.appsByLabel = make(map[string]map[string][]string)

	logging.Logger.WithField("app_count", len(apps)).Debug("Updating label-to-app mappings")

	// Build new mappings using unique identifiers
	for _, app := range apps {
		uniqueID := app.getUniqueID()
		for labelKey, labelValue := range app.Labels {
			// Initialize key map if it doesn't exist
			if lm.appsByLabel[labelKey] == nil {
				lm.appsByLabel[labelKey] = make(map[string][]string)
			}

			// Add app unique ID to the mapping
			lm.appsByLabel[labelKey][labelValue] = append(
				lm.appsByLabel[labelKey][labelValue],
				uniqueID,
			)
		}
	}

	// Log statistics for observability
	totalMappings := 0
	for _, valueMap := range lm.appsByLabel {
		for _, appList := range valueMap {
			totalMappings += len(appList)
		}
	}

	logging.Logger.WithFields(map[string]interface{}{
		"label_keys":     len(lm.appsByLabel),
		"total_mappings": totalMappings,
	}).Debug("Label-to-app mappings updated with unique identifiers")
}

// FindAppsByLabel returns a list of unique app identifiers that have the specified label key-value pair.
// Returns unique identifiers in format "source:appname" to avoid conflicts between apps with same names.
// Returns an empty slice if no apps match the criteria.
func (lm *LabelManager) FindAppsByLabel(labelKey, labelValue string) []string {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()

	if valueMap, exists := lm.appsByLabel[labelKey]; exists {
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
		currentApps := lm.FindAppsByLabel(labelKey, labelValue)

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

	keys := make([]string, 0, len(lm.appsByLabel))
	for key := range lm.appsByLabel {
		keys = append(keys, key)
	}
	return keys
}

// GetLabelValues returns all available values for a specific label key
func (lm *LabelManager) GetLabelValues(labelKey string) []string {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()

	if valueMap, exists := lm.appsByLabel[labelKey]; exists {
		values := make([]string, 0, len(valueMap))
		for value := range valueMap {
			values = append(values, value)
		}
		return values
	}
	return []string{}
}

// GetStats returns statistics about the current label mappings
func (lm *LabelManager) GetStats() map[string]interface{} {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()

	totalMappings := 0
	totalValues := 0

	for _, valueMap := range lm.appsByLabel {
		totalValues += len(valueMap)
		for _, appList := range valueMap {
			totalMappings += len(appList)
		}
	}

	return map[string]interface{}{
		"label_keys":     len(lm.appsByLabel),
		"label_values":   totalValues,
		"total_mappings": totalMappings,
	}
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
