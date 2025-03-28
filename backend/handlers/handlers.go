package handlers

import (
	"encoding/json"
	"net/http"
	"site-availability/config"
	"sync"
)

type AppStatus struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	Status   string `json:"status"`
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
	appStatusCache = make(map[string]AppStatus)
	cacheMutex     sync.RWMutex
)

// UpdateAppStatus updates the appStatusCache
func UpdateAppStatus(newStatuses []AppStatus) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	appStatusCache = make(map[string]AppStatus) // Reset the cache
	for _, app := range newStatuses {
		appStatusCache[app.Name] = app
	}
}

func GetAppStatus(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	// Convert appStatusCache map to slice
	var apps []AppStatus
	for _, status := range appStatusCache {
		apps = append(apps, status)
	}

	// Construct response with preloaded config
	response := StatusResponse{
		Locations: convertToHandlersLocation(cfg.Locations), // Convert Locations to StatusResponse format
		Apps:      apps,
	}

	// Encode response as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode status", http.StatusInternalServerError)
	}
}

// Helper function to convert config.Locations to handlers.Location
func convertToHandlersLocation(configLocations []config.Location) []Location {
	var locations []Location
	for _, loc := range configLocations {
		locations = append(locations, Location{
			Name:      loc.Name,      // Convert Name to lowercase if needed
			Latitude:  loc.Latitude,  // Copy latitude
			Longitude: loc.Longitude, // Copy longitude
		})
	}
	return locations
}

// IsAppStatusCacheEmpty checks if the app status cache is empty
func IsAppStatusCacheEmpty() bool {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	return len(appStatusCache) == 0
}

