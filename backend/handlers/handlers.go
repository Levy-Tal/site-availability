package handlers

import (
	"encoding/json"
	"net/http"
	"sync"
)

// AppStatus stores the status of a particular app
type AppStatus struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	Status   string `json:"status"` // "up" or "down"
}

var (
	appStatusCache = make(map[string]AppStatus)
	cacheMutex     sync.RWMutex
)

// GetAppStatus returns the status of all apps in the cache
func GetAppStatus(w http.ResponseWriter, r *http.Request) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	// Return the cached app statuses as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(appStatusCache); err != nil {
		http.Error(w, "Failed to encode status", http.StatusInternalServerError)
	}
}

// UpdateAppStatusCache updates the cache with the app's status
func UpdateAppStatusCache(appName string, status AppStatus) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	appStatusCache[appName] = status
}
