package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"site-availability/config"
	"site-availability/labels"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTest clears the cache for test isolation
func setupTest() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	appStatusCache = make(map[string]map[string]AppStatus)
	locationCache = make(map[string][]Location)
	seenApps = make(map[string]bool) // Reset deduplication map
	labelManager = labels.NewLabelManager()
}

// Helper function to create mock source and server settings for tests
func getMockSourceAndSettings(sourceName string) (config.Source, config.ServerSettings) {
	source := config.Source{
		Name:   sourceName,
		Type:   "test",
		URL:    "http://test.example.com",
		Labels: map[string]string{"source_env": "test", "source_type": "mock"},
	}

	serverSettings := config.ServerSettings{
		Labels: map[string]string{"server_env": "test", "server_region": "us-west"},
	}

	return source, serverSettings
}

// Helper function to call UpdateAppStatus with mock parameters
func updateAppStatusTest(sourceName string, statuses []AppStatus) {
	source, serverSettings := getMockSourceAndSettings(sourceName)
	UpdateAppStatus(sourceName, statuses, source, serverSettings)
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
			{Name: "app1", Location: "loc1", Status: "up", Source: "test-source", OriginURL: "http://test-origin.com"},
			{Name: "app2", Location: "loc2", Status: "down", Source: "test-source", OriginURL: "http://test-origin.com"},
		}
		updateAppStatusTest("test-source", testStatuses)

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
			{Name: "app1", Location: "loc1", Status: "up", Source: "test-source", OriginURL: "http://test-origin.com"},
			{Name: "app2", Location: "loc2", Status: "down", Source: "test-source", OriginURL: "http://test-origin.com"},
		}
		updateAppStatusTest("test-source", newStatuses)

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
			{Name: "app1", Location: "loc1", Status: "up", Source: "test-source", OriginURL: "http://test-origin.com"},
		}
		updateAppStatusTest("test-source", initialStatuses)

		cache := GetAppStatusCache()
		require.Len(t, cache, 1)

		// Now clear the statuses for this source
		updateAppStatusTest("test-source", []AppStatus{})
		cache = GetAppStatusCache()
		assert.Empty(t, cache)
	})

	t.Run("multiple sources", func(t *testing.T) {
		setupTest()

		// Add statuses for source1
		source1Statuses := []AppStatus{
			{Name: "app1", Location: "loc1", Status: "up", Source: "source1", OriginURL: "http://test-origin.com"},
		}
		updateAppStatusTest("source1", source1Statuses)

		// Add statuses for source2
		source2Statuses := []AppStatus{
			{Name: "app2", Location: "loc2", Status: "down", Source: "source2", OriginURL: "http://test-origin.com"},
		}
		updateAppStatusTest("source2", source2Statuses)

		cache := GetAppStatusCache()
		require.Len(t, cache, 2)

		// Clear only source1
		updateAppStatusTest("source1", []AppStatus{})
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

		updateAppStatusTest("test-source", []AppStatus{
			{Name: "app1", Location: "loc1", Status: "up", Source: "test-source", OriginURL: "http://test-origin.com"},
		})
		assert.False(t, IsAppStatusCacheEmpty())
	})

	t.Run("cache becomes empty after clearing", func(t *testing.T) {
		setupTest()

		// Add data
		updateAppStatusTest("test-source", []AppStatus{
			{Name: "app1", Location: "loc1", Status: "up", Source: "test-source", OriginURL: "http://test-origin.com"},
		})
		assert.False(t, IsAppStatusCacheEmpty())

		// Clear data
		updateAppStatusTest("test-source", []AppStatus{})
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
			{Name: "app1", Location: "test location", Status: "up", Source: "test-source", OriginURL: "http://test-origin.com"},
		}
		updateAppStatusTest("test-source", testStatuses)

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
			{Name: "app1", Location: "loc1", Status: "up", Source: "test-source", OriginURL: "http://test-origin.com"},
		}
		updateAppStatusTest("test-source", testStatuses)

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

func TestParseLabelFilters(t *testing.T) {
	t.Run("no label filters", func(t *testing.T) {
		queryParams := url.Values{}
		queryParams.Set("app", "test-app")
		queryParams.Set("location", "test-location")

		filters := parseLabelFilters(queryParams)
		assert.Empty(t, filters)
	})

	t.Run("single label filter", func(t *testing.T) {
		queryParams := url.Values{}
		queryParams.Set("labels.env", "production")

		filters := parseLabelFilters(queryParams)
		assert.Len(t, filters, 1)
		assert.Equal(t, "production", filters["env"])
	})

	t.Run("multiple label filters", func(t *testing.T) {
		queryParams := url.Values{}
		queryParams.Set("labels.env", "production")
		queryParams.Set("labels.tier", "backend")
		queryParams.Set("labels.team", "platform")

		filters := parseLabelFilters(queryParams)
		assert.Len(t, filters, 3)
		assert.Equal(t, "production", filters["env"])
		assert.Equal(t, "backend", filters["tier"])
		assert.Equal(t, "platform", filters["team"])
	})

	t.Run("mixed parameters", func(t *testing.T) {
		queryParams := url.Values{}
		queryParams.Set("app", "test-app") // Non-label parameter
		queryParams.Set("labels.env", "staging")
		queryParams.Set("location", "test-location") // Non-label parameter
		queryParams.Set("labels.version", "v1.2.3")

		filters := parseLabelFilters(queryParams)
		assert.Len(t, filters, 2)
		assert.Equal(t, "staging", filters["env"])
		assert.Equal(t, "v1.2.3", filters["version"])
	})

	t.Run("empty label values ignored", func(t *testing.T) {
		queryParams := url.Values{}
		queryParams.Set("labels.env", "production")
		queryParams.Set("labels.empty", "") // Empty value should be ignored

		filters := parseLabelFilters(queryParams)
		assert.Len(t, filters, 1)
		assert.Equal(t, "production", filters["env"])
		assert.NotContains(t, filters, "empty")
	})

	t.Run("multiple values uses first", func(t *testing.T) {
		queryParams := url.Values{}
		queryParams.Add("labels.env", "production")
		queryParams.Add("labels.env", "staging") // Second value should be ignored

		filters := parseLabelFilters(queryParams)
		assert.Len(t, filters, 1)
		assert.Equal(t, "production", filters["env"]) // First value used
	})
}

func TestFilterAppsByLabels(t *testing.T) {
	setupTest()

	// Create test apps with various labels
	testApps := []AppStatus{
		{
			Name:      "app1",
			Location:  "loc1",
			Status:    "up",
			Source:    "test",
			OriginURL: "http://test-origin.com",
			Labels: map[string]string{
				"env":     "production",
				"tier":    "backend",
				"team":    "platform",
				"version": "v1.0.0",
			},
		},
		{
			Name:      "app2",
			Location:  "loc2",
			Status:    "down",
			Source:    "test",
			OriginURL: "http://test-origin.com",
			Labels: map[string]string{
				"env":     "staging",
				"tier":    "frontend",
				"team":    "platform",
				"version": "v1.1.0",
			},
		},
		{
			Name:      "app3",
			Location:  "loc3",
			Status:    "up",
			Source:    "test",
			OriginURL: "http://test-origin.com",
			Labels: map[string]string{
				"env":     "production",
				"tier":    "frontend",
				"team":    "security",
				"version": "v2.0.0",
			},
		},
		{
			Name:      "app4",
			Location:  "loc4",
			Status:    "unavailable",
			Source:    "test",
			OriginURL: "http://test-origin.com",
			Labels:    map[string]string{}, // No labels
		},
	}

	// Populate the cache and label manager with test data
	updateAppStatusTest("test", testApps)

	t.Run("no filters returns all apps", func(t *testing.T) {
		filters := map[string]string{}
		filteredApps, filteredCount := filterAppsByLabels(testApps, filters)

		assert.Len(t, filteredApps, 4)
		assert.Equal(t, 0, filteredCount)
		assert.Equal(t, testApps, filteredApps)
	})

	t.Run("single filter matches multiple apps", func(t *testing.T) {
		filters := map[string]string{"env": "production"}
		filteredApps, filteredCount := filterAppsByLabels(testApps, filters)

		assert.Len(t, filteredApps, 2)
		assert.Equal(t, 2, filteredCount)
		assert.Equal(t, "app1", filteredApps[0].Name)
		assert.Equal(t, "app3", filteredApps[1].Name)
	})

	t.Run("single filter matches single app", func(t *testing.T) {
		filters := map[string]string{"env": "staging"}
		filteredApps, filteredCount := filterAppsByLabels(testApps, filters)

		assert.Len(t, filteredApps, 1)
		assert.Equal(t, 3, filteredCount)
		assert.Equal(t, "app2", filteredApps[0].Name)
	})

	t.Run("multiple filters narrow down results", func(t *testing.T) {
		filters := map[string]string{
			"env":  "production",
			"tier": "backend",
		}
		filteredApps, filteredCount := filterAppsByLabels(testApps, filters)

		assert.Len(t, filteredApps, 1)
		assert.Equal(t, 3, filteredCount)
		assert.Equal(t, "app1", filteredApps[0].Name)
	})

	t.Run("filter with no matches", func(t *testing.T) {
		filters := map[string]string{"env": "development"}
		filteredApps, filteredCount := filterAppsByLabels(testApps, filters)

		assert.Len(t, filteredApps, 0)
		assert.Equal(t, 4, filteredCount)
	})

	t.Run("filter excludes apps without required labels", func(t *testing.T) {
		filters := map[string]string{"team": "platform"}
		filteredApps, filteredCount := filterAppsByLabels(testApps, filters)

		assert.Len(t, filteredApps, 2)
		assert.Equal(t, 2, filteredCount)
		assert.Equal(t, "app1", filteredApps[0].Name)
		assert.Equal(t, "app2", filteredApps[1].Name)
	})

	t.Run("complex multi-label filter", func(t *testing.T) {
		filters := map[string]string{
			"env":     "production",
			"tier":    "frontend",
			"team":    "security",
			"version": "v2.0.0",
		}
		filteredApps, filteredCount := filterAppsByLabels(testApps, filters)

		assert.Len(t, filteredApps, 1)
		assert.Equal(t, 3, filteredCount)
		assert.Equal(t, "app3", filteredApps[0].Name)
	})
}

func TestGetAppStatusWithLabelFiltering(t *testing.T) {
	setupTest()

	// Set up test data with labels
	testStatuses := []AppStatus{
		{
			Name:      "prod-api",
			Location:  "us-east",
			Status:    "up",
			Source:    "prometheus",
			OriginURL: "http://test-origin.com",
			Labels: map[string]string{
				"env":  "production",
				"tier": "backend",
				"team": "platform",
			},
		},
		{
			Name:      "staging-web",
			Location:  "us-west",
			Status:    "down",
			Source:    "prometheus",
			OriginURL: "http://test-origin.com",
			Labels: map[string]string{
				"env":  "staging",
				"tier": "frontend",
				"team": "product",
			},
		},
		{
			Name:      "prod-worker",
			Location:  "eu-central",
			Status:    "up",
			Source:    "prometheus",
			OriginURL: "http://test-origin.com",
			Labels: map[string]string{
				"env":  "production",
				"tier": "worker",
				"team": "platform",
			},
		},
	}
	updateAppStatusTest("prometheus", testStatuses)

	cfg := &config.Config{
		Locations: []config.Location{},
	}

	t.Run("no label filters returns all apps", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/status", nil)
		w := httptest.NewRecorder()

		GetAppStatus(w, req, cfg)

		assert.Equal(t, http.StatusOK, w.Code)
		var response StatusResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.Len(t, response.Apps, 3)
	})

	t.Run("filter by environment", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/status?labels.env=production", nil)
		w := httptest.NewRecorder()

		GetAppStatus(w, req, cfg)

		assert.Equal(t, http.StatusOK, w.Code)
		var response StatusResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.Len(t, response.Apps, 2)

		// Verify correct apps are returned
		appNames := []string{response.Apps[0].Name, response.Apps[1].Name}
		assert.Contains(t, appNames, "prod-api")
		assert.Contains(t, appNames, "prod-worker")
	})

	t.Run("filter by multiple labels", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/status?labels.env=production&labels.team=platform", nil)
		w := httptest.NewRecorder()

		GetAppStatus(w, req, cfg)

		assert.Equal(t, http.StatusOK, w.Code)
		var response StatusResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.Len(t, response.Apps, 2)

		// Verify correct apps are returned
		appNames := []string{response.Apps[0].Name, response.Apps[1].Name}
		assert.Contains(t, appNames, "prod-api")
		assert.Contains(t, appNames, "prod-worker")
	})

	t.Run("filter with no matches", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/status?labels.env=development", nil)
		w := httptest.NewRecorder()

		GetAppStatus(w, req, cfg)

		assert.Equal(t, http.StatusOK, w.Code)
		var response StatusResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.Len(t, response.Apps, 0)
	})
}

func TestHandleSyncRequestWithLabelFiltering(t *testing.T) {
	setupTest()

	// Set up test data with labels
	testStatuses := []AppStatus{
		{
			Name:      "sync-app1",
			Location:  "location1",
			Status:    "up",
			Source:    "site-sync",
			OriginURL: "http://test-origin.com",
			Labels: map[string]string{
				"env":     "production",
				"service": "api",
			},
		},
		{
			Name:      "sync-app2",
			Location:  "location2",
			Status:    "down",
			Source:    "site-sync",
			OriginURL: "http://test-origin.com",
			Labels: map[string]string{
				"env":     "staging",
				"service": "worker",
			},
		},
	}
	updateAppStatusTest("site-sync", testStatuses)

	cfg := &config.Config{
		ServerSettings: config.ServerSettings{
			SyncEnable: true,
		},
		Locations: []config.Location{},
	}

	t.Run("sync without filters returns all apps", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/sync", nil)
		w := httptest.NewRecorder()

		HandleSyncRequest(w, req, cfg)

		assert.Equal(t, http.StatusOK, w.Code)
		var response StatusResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.Len(t, response.Apps, 2)
	})

	t.Run("sync with label filter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/sync?labels.env=production", nil)
		w := httptest.NewRecorder()

		HandleSyncRequest(w, req, cfg)

		assert.Equal(t, http.StatusOK, w.Code)
		var response StatusResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.Len(t, response.Apps, 1)
		assert.Equal(t, "sync-app1", response.Apps[0].Name)
		assert.Equal(t, "production", response.Apps[0].Labels["env"])
	})
}

func TestLabelManagerIntegration(t *testing.T) {
	t.Run("label manager updates with cache", func(t *testing.T) {
		setupTest()

		testApps := []AppStatus{
			{
				Name:      "web-app",
				Location:  "loc1",
				Status:    "up",
				Source:    "test-source",
				OriginURL: "http://test-origin.com",
				Labels:    map[string]string{"env": "prod", "tier": "frontend"},
			},
			{
				Name:      "api-app",
				Location:  "loc2",
				Status:    "up",
				Source:    "test-source",
				OriginURL: "http://test-origin.com",
				Labels:    map[string]string{"env": "prod", "tier": "backend"},
			},
			{
				Name:      "staging-app",
				Location:  "loc3",
				Status:    "up",
				Source:    "test-source",
				OriginURL: "http://test-origin.com",
				Labels:    map[string]string{"env": "staging", "tier": "frontend"},
			},
		}

		updateAppStatusTest("test-source", testApps)

		// Test label filtering
		apps := GetAppStatusCache()
		filteredApps, filteredCount := filterAppsByLabels(apps, map[string]string{"env": "prod"})

		assert.Len(t, filteredApps, 2, "Should find 2 prod apps")
		assert.Equal(t, 1, filteredCount, "Should filter out 1 app")

		// Verify the right apps were returned
		prodApps := make(map[string]bool)
		for _, app := range filteredApps {
			prodApps[app.Name] = true
		}
		assert.True(t, prodApps["web-app"])
		assert.True(t, prodApps["api-app"])
		assert.False(t, prodApps["staging-app"])
	})

	t.Run("multiple label filters", func(t *testing.T) {
		setupTest()

		testApps := []AppStatus{
			{
				Name:      "web-prod",
				Location:  "loc1",
				Status:    "up",
				Source:    "test-source",
				OriginURL: "http://test-origin.com",
				Labels:    map[string]string{"env": "prod", "tier": "frontend"},
			},
			{
				Name:      "api-prod",
				Location:  "loc2",
				Status:    "up",
				Source:    "test-source",
				OriginURL: "http://test-origin.com",
				Labels:    map[string]string{"env": "prod", "tier": "backend"},
			},
			{
				Name:      "web-staging",
				Location:  "loc3",
				Status:    "up",
				Source:    "test-source",
				OriginURL: "http://test-origin.com",
				Labels:    map[string]string{"env": "staging", "tier": "frontend"},
			},
		}

		updateAppStatusTest("test-source", testApps)

		apps := GetAppStatusCache()
		filteredApps, _ := filterAppsByLabels(apps, map[string]string{
			"env":  "prod",
			"tier": "frontend",
		})

		assert.Len(t, filteredApps, 1, "Should find exactly 1 app matching both labels")
		assert.Equal(t, "web-prod", filteredApps[0].Name)
	})

	t.Run("cache update removes old labels", func(t *testing.T) {
		setupTest()

		// Initial apps
		updateAppStatusTest("test-source", []AppStatus{
			{
				Name:      "app1",
				Location:  "loc1",
				Status:    "up",
				Source:    "test-source",
				OriginURL: "http://test-origin.com",
				Labels:    map[string]string{"env": "old"},
			},
		})

		// Update with new apps
		updateAppStatusTest("test-source", []AppStatus{
			{
				Name:      "app2",
				Location:  "loc2",
				Status:    "up",
				Source:    "test-source",
				OriginURL: "http://test-origin.com",
				Labels:    map[string]string{"env": "new"},
			},
		})

		apps := GetAppStatusCache()

		// Should only find the new app
		filteredApps, _ := filterAppsByLabels(apps, map[string]string{"env": "new"})
		assert.Len(t, filteredApps, 1)
		assert.Equal(t, "app2", filteredApps[0].Name)

		// Old app should not be found
		filteredApps, _ = filterAppsByLabels(apps, map[string]string{"env": "old"})
		assert.Empty(t, filteredApps)
	})
}
