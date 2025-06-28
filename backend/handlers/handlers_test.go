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

// setupTest clears the cache for test isolation
func setupTest() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	appStatusCache = make(map[string]map[string]AppStatus)
}

func TestAppStatusCache(t *testing.T) {
	t.Run("get empty cache", func(t *testing.T) {
		setupTest()

		cache := GetAppStatusCache()
		assert.Empty(t, cache)
	})

	t.Run("get non-empty cache", func(t *testing.T) {
		setupTest()

		testStatuses := []AppStatus{
			{Name: "app1", Location: "loc1", Status: "up", Source: "test-source"},
			{Name: "app2", Location: "loc2", Status: "down", Source: "test-source"},
		}
		UpdateAppStatus("test-source", testStatuses)

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
		assert.Equal(t, "test-source", app1.Source)

		// Verify app2
		app2, exists := statusMap["app2"]
		require.True(t, exists)
		assert.Equal(t, "loc2", app2.Location)
		assert.Equal(t, "down", app2.Status)
		assert.Equal(t, "test-source", app2.Source)
	})

	t.Run("update with new statuses", func(t *testing.T) {
		setupTest()

		newStatuses := []AppStatus{
			{Name: "app1", Location: "loc1", Status: "up", Source: "test-source"},
			{Name: "app2", Location: "loc2", Status: "down", Source: "test-source"},
		}
		UpdateAppStatus("test-source", newStatuses)

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
		assert.Equal(t, "test-source", app1.Source)

		// Verify app2
		app2, exists := statusMap["app2"]
		require.True(t, exists)
		assert.Equal(t, "loc2", app2.Location)
		assert.Equal(t, "down", app2.Status)
		assert.Equal(t, "test-source", app2.Source)
	})

	t.Run("update with empty statuses clears source", func(t *testing.T) {
		setupTest()

		// First, add some statuses
		initialStatuses := []AppStatus{
			{Name: "app1", Location: "loc1", Status: "up", Source: "test-source"},
		}
		UpdateAppStatus("test-source", initialStatuses)

		cache := GetAppStatusCache()
		require.Len(t, cache, 1)

		// Now clear the statuses for this source
		UpdateAppStatus("test-source", []AppStatus{})
		cache = GetAppStatusCache()
		assert.Empty(t, cache)
	})

	t.Run("multiple sources", func(t *testing.T) {
		setupTest()

		// Add statuses for source1
		source1Statuses := []AppStatus{
			{Name: "app1", Location: "loc1", Status: "up", Source: "source1"},
		}
		UpdateAppStatus("source1", source1Statuses)

		// Add statuses for source2
		source2Statuses := []AppStatus{
			{Name: "app2", Location: "loc2", Status: "down", Source: "source2"},
		}
		UpdateAppStatus("source2", source2Statuses)

		cache := GetAppStatusCache()
		require.Len(t, cache, 2)

		// Clear only source1
		UpdateAppStatus("source1", []AppStatus{})
		cache = GetAppStatusCache()
		require.Len(t, cache, 1)
		assert.Equal(t, "app2", cache[0].Name)
		assert.Equal(t, "source2", cache[0].Source)
	})
}

func TestIsAppStatusCacheEmpty(t *testing.T) {
	t.Run("empty cache", func(t *testing.T) {
		setupTest()
		assert.True(t, IsAppStatusCacheEmpty())
	})

	t.Run("non-empty cache", func(t *testing.T) {
		setupTest()

		UpdateAppStatus("test-source", []AppStatus{
			{Name: "app1", Location: "loc1", Status: "up", Source: "test-source"},
		})
		assert.False(t, IsAppStatusCacheEmpty())
	})

	t.Run("cache becomes empty after clearing", func(t *testing.T) {
		setupTest()

		// Add data
		UpdateAppStatus("test-source", []AppStatus{
			{Name: "app1", Location: "loc1", Status: "up", Source: "test-source"},
		})
		assert.False(t, IsAppStatusCacheEmpty())

		// Clear data
		UpdateAppStatus("test-source", []AppStatus{})
		assert.True(t, IsAppStatusCacheEmpty())
	})
}

func TestGetAppStatus(t *testing.T) {
	t.Run("successful response with data", func(t *testing.T) {
		setupTest()

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
			{Name: "app1", Location: "test location", Status: "up", Source: "test-source"},
		}
		UpdateAppStatus("test-source", testStatuses)

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
		assert.Equal(t, "test-source", response.Apps[0].Source)
	})

	t.Run("empty cache response", func(t *testing.T) {
		setupTest()

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
	})

	t.Run("JSON encoding error", func(t *testing.T) {
		setupTest()

		// Create a response writer that fails on Write
		w := &failingResponseWriter{httptest.NewRecorder()}
		req := httptest.NewRequest("GET", "/api/status", nil)
		cfg := &config.Config{}

		GetAppStatus(w, req, cfg)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestConvertToHandlersLocation(t *testing.T) {
	t.Run("convert multiple locations", func(t *testing.T) {
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
	})

	t.Run("convert empty slice", func(t *testing.T) {
		locations := convertToHandlersLocation([]config.Location{})
		assert.Empty(t, locations)
	})
}

func TestGetScrapeInterval(t *testing.T) {
	t.Run("valid interval", func(t *testing.T) {
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
	})

	t.Run("invalid interval format", func(t *testing.T) {
		cfg := &config.Config{
			Scraping: config.ScrapingSettings{
				Interval: "invalid",
			},
		}

		req := httptest.NewRequest("GET", "/api/scrape-interval", nil)
		w := httptest.NewRecorder()

		GetScrapeInterval(w, req, cfg)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("different valid intervals", func(t *testing.T) {
		testCases := []struct {
			interval string
			expected int64
		}{
			{"30s", 30000},
			{"2m", 120000},
			{"1h", 3600000},
		}

		for _, tc := range testCases {
			t.Run(tc.interval, func(t *testing.T) {
				cfg := &config.Config{
					Scraping: config.ScrapingSettings{
						Interval: tc.interval,
					},
				}

				req := httptest.NewRequest("GET", "/api/scrape-interval", nil)
				w := httptest.NewRecorder()

				GetScrapeInterval(w, req, cfg)

				assert.Equal(t, http.StatusOK, w.Code)
				var response map[string]int64
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tc.expected, response["scrape_interval_ms"])
			})
		}
	})
}

func TestGetDocs(t *testing.T) {
	t.Run("valid docs configuration", func(t *testing.T) {
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
	})

	t.Run("empty docs configuration", func(t *testing.T) {
		cfg := &config.Config{
			Documentation: config.Documentation{
				Title: "",
				URL:   "",
			},
		}

		req := httptest.NewRequest("GET", "/api/docs", nil)
		w := httptest.NewRecorder()

		GetDocs(w, req, cfg)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]string
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, "", response["docs_title"])
		assert.Equal(t, "", response["docs_url"])
	})
}

func TestHandleSyncRequest(t *testing.T) {
	t.Run("sync disabled", func(t *testing.T) {
		setupTest()

		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				SyncEnable: false,
			},
		}

		req := httptest.NewRequest("GET", "/sync", nil)
		w := httptest.NewRecorder()

		HandleSyncRequest(w, req, cfg)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("sync enabled without token", func(t *testing.T) {
		setupTest()

		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				SyncEnable: true,
			},
		}

		// Set up test data
		testStatuses := []AppStatus{
			{Name: "app1", Location: "loc1", Status: "up", Source: "test-source"},
		}
		UpdateAppStatus("test-source", testStatuses)

		req := httptest.NewRequest("GET", "/sync", nil)
		w := httptest.NewRecorder()

		HandleSyncRequest(w, req, cfg)
		assert.Equal(t, http.StatusOK, w.Code)

		var response StatusResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		require.Len(t, response.Apps, 1)
		assert.Equal(t, "app1", response.Apps[0].Name)
		// Should have no locations since we didn't add any to the location cache
		assert.Len(t, response.Locations, 0)
	})

	t.Run("sync enabled with token but no signature", func(t *testing.T) {
		setupTest()

		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				SyncEnable: true,
				Token:      "test-token",
			},
		}

		req := httptest.NewRequest("GET", "/sync", nil)
		w := httptest.NewRecorder()

		HandleSyncRequest(w, req, cfg)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("sync enabled with token and invalid signature", func(t *testing.T) {
		setupTest()

		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				SyncEnable: true,
				Token:      "test-token",
			},
		}

		req := httptest.NewRequest("GET", "/sync", nil)
		req.Header.Set("X-HMAC-Signature", "invalid-signature")
		w := httptest.NewRecorder()

		HandleSyncRequest(w, req, cfg)
		// This will return 401 because we can't easily generate a valid HMAC signature in the test
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// failingResponseWriter is a custom ResponseWriter that fails on Write
type failingResponseWriter struct {
	*httptest.ResponseRecorder
}

func (w *failingResponseWriter) Write([]byte) (int, error) {
	return 0, assert.AnError
}
