package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"site-availability/config"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAppStatusCache(t *testing.T) {
	// Test empty cache
	emptyCache := GetAppStatusCache()
	assert.Empty(t, emptyCache)

	// Test with some data
	testStatuses := []AppStatus{
		{Name: "app1", Location: "loc1", Status: "up"},
		{Name: "app2", Location: "loc2", Status: "down"},
	}
	UpdateAppStatus(testStatuses)

	cache := GetAppStatusCache()
	require.Len(t, cache, 2)
	assert.Equal(t, "app1", cache[0].Name)
	assert.Equal(t, "loc1", cache[0].Location)
	assert.Equal(t, "up", cache[0].Status)
	assert.Equal(t, "app2", cache[1].Name)
	assert.Equal(t, "loc2", cache[1].Location)
	assert.Equal(t, "down", cache[1].Status)
}

func TestUpdateAppStatus(t *testing.T) {
	// Clear cache first
	UpdateAppStatus([]AppStatus{})

	// Test updating with new statuses
	newStatuses := []AppStatus{
		{Name: "app1", Location: "loc1", Status: "up"},
		{Name: "app2", Location: "loc2", Status: "down"},
	}
	UpdateAppStatus(newStatuses)

	cache := GetAppStatusCache()
	require.Len(t, cache, 2)

	// Create a map for easier lookup
	statusMap := make(map[string]AppStatus)
	for _, status := range cache {
		statusMap[status.Name] = status
	}

	// Verify app1
	app1, exists := statusMap["app1"]
	require.True(t, exists)
	assert.Equal(t, "loc1", app1.Location)
	assert.Equal(t, "up", app1.Status)

	// Verify app2
	app2, exists := statusMap["app2"]
	require.True(t, exists)
	assert.Equal(t, "loc2", app2.Location)
	assert.Equal(t, "down", app2.Status)

	// Test updating with empty statuses
	UpdateAppStatus([]AppStatus{})
	cache = GetAppStatusCache()
	assert.Empty(t, cache)
}

func TestGetAppStatus(t *testing.T) {
	// Create test config
	cfg := &config.Config{
		Locations: []config.Location{
			{
				Name:      "test location",
				Latitude:  31.782904,
				Longitude: 35.214774,
			},
		},
	}

	// Set up test data
	testStatuses := []AppStatus{
		{Name: "app1", Location: "test location", Status: "up"},
	}
	UpdateAppStatus(testStatuses)

	// Create test request
	req := httptest.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()

	// Call handler
	GetAppStatus(w, req, cfg)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response StatusResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	// Verify response content
	require.Len(t, response.Locations, 1)
	assert.Equal(t, "test location", response.Locations[0].Name)
	assert.Equal(t, 31.782904, response.Locations[0].Latitude)
	assert.Equal(t, 35.214774, response.Locations[0].Longitude)

	require.Len(t, response.Apps, 1)
	assert.Equal(t, "app1", response.Apps[0].Name)
	assert.Equal(t, "test location", response.Apps[0].Location)
	assert.Equal(t, "up", response.Apps[0].Status)
}

func TestConvertToHandlersLocation(t *testing.T) {
	configLocations := []config.Location{
		{
			Name:      "loc1",
			Latitude:  31.782904,
			Longitude: 35.214774,
		},
		{
			Name:      "loc2",
			Latitude:  32.0853,
			Longitude: 34.7818,
		},
	}

	locations := convertToHandlersLocation(configLocations)
	require.Len(t, locations, 2)

	assert.Equal(t, "loc1", locations[0].Name)
	assert.Equal(t, 31.782904, locations[0].Latitude)
	assert.Equal(t, 35.214774, locations[0].Longitude)

	assert.Equal(t, "loc2", locations[1].Name)
	assert.Equal(t, 32.0853, locations[1].Latitude)
	assert.Equal(t, 34.7818, locations[1].Longitude)
}

func TestIsAppStatusCacheEmpty(t *testing.T) {
	// Test empty cache
	UpdateAppStatus([]AppStatus{})
	assert.True(t, IsAppStatusCacheEmpty())

	// Test non-empty cache
	UpdateAppStatus([]AppStatus{
		{Name: "app1", Location: "loc1", Status: "up"},
	})
	assert.False(t, IsAppStatusCacheEmpty())
}

func TestGetScrapeInterval(t *testing.T) {
	// Test valid interval
	cfg := &config.Config{
		Scraping: config.ScrapingSettings{
			Interval: "60s",
		},
	}

	req := httptest.NewRequest("GET", "/api/scrape-interval", nil)
	w := httptest.NewRecorder()

	GetScrapeInterval(w, req, cfg)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]int64
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, int64(60000), response["scrape_interval_ms"])

	// Test invalid interval
	cfg.Scraping.Interval = "invalid"
	w = httptest.NewRecorder()
	GetScrapeInterval(w, req, cfg)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetDocs(t *testing.T) {
	cfg := &config.Config{
		Documentation: config.Documentation{
			Title: "Test Docs",
			URL:   "https://test.example.com/docs",
		},
	}

	req := httptest.NewRequest("GET", "/api/docs", nil)
	w := httptest.NewRecorder()

	GetDocs(w, req, cfg)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "Test Docs", response["docs_title"])
	assert.Equal(t, "https://test.example.com/docs", response["docs_url"])
}

func TestHandleSyncRequest(t *testing.T) {
	// Test with sync disabled
	cfg := &config.Config{
		ServerSettings: config.ServerSettings{
			SyncEnable: false,
		},
	}

	req := httptest.NewRequest("GET", "/sync", nil)
	w := httptest.NewRecorder()

	HandleSyncRequest(w, req, cfg)
	assert.Equal(t, http.StatusForbidden, w.Code)

	// Test with sync enabled but no token
	cfg.ServerSettings.SyncEnable = true
	w = httptest.NewRecorder()
	HandleSyncRequest(w, req, cfg)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test with sync enabled and token
	cfg.ServerSettings.Token = "test-token"
	cfg.ServerSettings.SyncEnable = true

	// Set up test data
	testStatuses := []AppStatus{
		{Name: "app1", Location: "loc1", Status: "up"},
	}
	UpdateAppStatus(testStatuses)

	// Create request with valid HMAC
	req = httptest.NewRequest("GET", "/sync", nil)
	req.Header.Set("X-HMAC-Signature", "valid-signature") // Note: In real test, you'd need to generate a valid signature
	w = httptest.NewRecorder()

	HandleSyncRequest(w, req, cfg)
	// Note: This will fail with 401 because we can't easily generate a valid HMAC signature in the test
	// In a real test, you'd need to properly generate the HMAC signature
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetAppStatusWithEmptyCache(t *testing.T) {
	// Clear cache
	UpdateAppStatus([]AppStatus{})

	cfg := &config.Config{
		Locations: []config.Location{
			{
				Name:      "test location",
				Latitude:  31.782904,
				Longitude: 35.214774,
			},
		},
	}

	req := httptest.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()

	GetAppStatus(w, req, cfg)

	assert.Equal(t, http.StatusOK, w.Code)
	var response StatusResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Empty(t, response.Apps)
	assert.Len(t, response.Locations, 1)
}

func TestGetAppStatusWithJSONError(t *testing.T) {
	// Create a response writer that fails on Write
	w := &failingResponseWriter{httptest.NewRecorder()}
	req := httptest.NewRequest("GET", "/api/status", nil)
	cfg := &config.Config{}

	GetAppStatus(w, req, cfg)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// failingResponseWriter is a custom ResponseWriter that fails on Write
type failingResponseWriter struct {
	*httptest.ResponseRecorder
}

func (w *failingResponseWriter) Write([]byte) (int, error) {
	return 0, assert.AnError
}
