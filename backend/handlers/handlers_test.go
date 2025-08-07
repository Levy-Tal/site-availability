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
	appStatusCache = make(map[string]map[string]map[string]AppStatus)
	locationCache = make(map[string][]Location)
	labelManager = labels.NewLabelManager()
}

// Helper function to create mock source and server settings for tests
func getMockSourceAndSettings(sourceName string) (config.Source, config.ServerSettings) {
	source := config.Source{
		Name: sourceName,
		Type: "test",
		Config: map[string]interface{}{
			"url": "http://test.example.com",
		},
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
	// Ensure HostURL is set for validation
	if serverSettings.HostURL == "" {
		serverSettings.HostURL = "https://test-server.com"
	}

	// Ensure all test apps have OriginURL set (required for new validation)
	for i := range statuses {
		if statuses[i].OriginURL == "" {
			statuses[i].OriginURL = "https://test-origin.com"
		}
	}

	_ = UpdateAppStatus(sourceName, statuses, source, serverSettings)
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

func TestGetApps(t *testing.T) {
	t.Run("successful response with all apps", func(t *testing.T) {
		setupTest()

		// Set up test data
		testStatuses := []AppStatus{
			{Name: "app1", Location: "test location", Status: "up", Source: "test-source", OriginURL: "http://test-origin.com"},
			{Name: "app2", Location: "another location", Status: "down", Source: "test-source"},
		}
		updateAppStatusTest("test-source", testStatuses)

		// Create test request
		req := httptest.NewRequest("GET", "/api/apps", nil)
		w := httptest.NewRecorder()

		// Call handler
		GetApps(w, req, &config.Config{})

		// Check response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response AppsResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		// Verify response content
		require.Len(t, response.Apps, 2)
		// Create a map to check apps without relying on order
		appMap := make(map[string]string)
		for _, app := range response.Apps {
			appMap[app.Name] = app.Status
		}
		assert.Equal(t, "up", appMap["app1"])
		assert.Equal(t, "down", appMap["app2"])
	})

	t.Run("filter apps by location", func(t *testing.T) {
		setupTest()

		// Set up test data
		testStatuses := []AppStatus{
			{Name: "app1", Location: "location1", Status: "up", Source: "test-source"},
			{Name: "app2", Location: "location2", Status: "down", Source: "test-source"},
			{Name: "app3", Location: "location1", Status: "unavailable", Source: "test-source"},
		}
		updateAppStatusTest("test-source", testStatuses)

		// Create test request with location filter
		req := httptest.NewRequest("GET", "/api/apps?location=location1", nil)
		w := httptest.NewRecorder()

		// Call handler
		GetApps(w, req, &config.Config{})

		// Check response
		assert.Equal(t, http.StatusOK, w.Code)
		var response AppsResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		// Should only return apps from location1
		require.Len(t, response.Apps, 2)
		for _, app := range response.Apps {
			assert.Equal(t, "location1", app.Location)
		}
	})

	t.Run("empty cache response", func(t *testing.T) {
		setupTest()

		req := httptest.NewRequest("GET", "/api/apps", nil)
		w := httptest.NewRecorder()

		GetApps(w, req, &config.Config{})

		assert.Equal(t, http.StatusOK, w.Code)
		var response AppsResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.Empty(t, response.Apps)
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
		assert.Equal(t, 0, locations[0].Up)
		assert.Equal(t, 0, locations[0].Down)
		assert.Equal(t, 0, locations[0].Unavailable)

		assert.Equal(t, "loc2", locations[1].Name)
		assert.Equal(t, 32.0853, locations[1].Latitude)
		assert.Equal(t, 34.7818, locations[1].Longitude)
		assert.Equal(t, 0, locations[1].Up)
		assert.Equal(t, 0, locations[1].Down)
		assert.Equal(t, 0, locations[1].Unavailable)
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
			Labels: []labels.Label{
				{Key: "env", Value: "production"},
				{Key: "tier", Value: "backend"},
				{Key: "team", Value: "platform"},
				{Key: "version", Value: "v1.0.0"},
			},
		},
		{
			Name:      "app2",
			Location:  "loc2",
			Status:    "down",
			Source:    "test",
			OriginURL: "http://test-origin.com",
			Labels: []labels.Label{
				{Key: "env", Value: "staging"},
				{Key: "tier", Value: "frontend"},
				{Key: "team", Value: "platform"},
				{Key: "version", Value: "v1.1.0"},
			},
		},
		{
			Name:      "app3",
			Location:  "loc3",
			Status:    "up",
			Source:    "test",
			OriginURL: "http://test-origin.com",
			Labels: []labels.Label{
				{Key: "env", Value: "production"},
				{Key: "tier", Value: "frontend"},
				{Key: "team", Value: "security"},
				{Key: "version", Value: "v2.0.0"},
			},
		},
		{
			Name:      "app4",
			Location:  "loc4",
			Status:    "unavailable",
			Source:    "test",
			OriginURL: "http://test-origin.com",
			Labels:    []labels.Label{}, // No labels
		},
	}

	// Populate the cache and label manager with test data
	updateAppStatusTest("test", testApps)

	t.Run("no filters returns all apps", func(t *testing.T) {
		filters := map[string]string{}
		filteredApps, filteredCount := filterApps(testApps, filters)

		assert.Len(t, filteredApps, 4)
		assert.Equal(t, 0, filteredCount)
		assert.Equal(t, testApps, filteredApps)
	})

	t.Run("single filter matches multiple apps", func(t *testing.T) {
		filters := map[string]string{"labels.env": "production"}
		filteredApps, filteredCount := filterApps(testApps, filters)

		assert.Len(t, filteredApps, 2)
		assert.Equal(t, 2, filteredCount)
		assert.Equal(t, "app1", filteredApps[0].Name)
		assert.Equal(t, "app3", filteredApps[1].Name)
	})

	t.Run("single filter matches single app", func(t *testing.T) {
		filters := map[string]string{"labels.env": "staging"}
		filteredApps, filteredCount := filterApps(testApps, filters)

		assert.Len(t, filteredApps, 1)
		assert.Equal(t, 3, filteredCount)
		assert.Equal(t, "app2", filteredApps[0].Name)
	})

	t.Run("multiple filters narrow down results", func(t *testing.T) {
		filters := map[string]string{
			"labels.env":  "production",
			"labels.tier": "backend",
		}
		filteredApps, filteredCount := filterApps(testApps, filters)

		assert.Len(t, filteredApps, 1)
		assert.Equal(t, 3, filteredCount)
		assert.Equal(t, "app1", filteredApps[0].Name)
	})

	t.Run("filter with no matches", func(t *testing.T) {
		filters := map[string]string{"labels.env": "development"}
		filteredApps, filteredCount := filterApps(testApps, filters)

		assert.Len(t, filteredApps, 0)
		assert.Equal(t, 4, filteredCount)
	})

	t.Run("filter excludes apps without required labels", func(t *testing.T) {
		filters := map[string]string{"labels.team": "platform"}
		filteredApps, filteredCount := filterApps(testApps, filters)

		assert.Len(t, filteredApps, 2)
		assert.Equal(t, 2, filteredCount)
		assert.Equal(t, "app1", filteredApps[0].Name)
		assert.Equal(t, "app2", filteredApps[1].Name)
	})

	t.Run("complex multi-label filter", func(t *testing.T) {
		filters := map[string]string{
			"labels.env":     "production",
			"labels.tier":    "frontend",
			"labels.team":    "security",
			"labels.version": "v2.0.0",
		}
		filteredApps, filteredCount := filterApps(testApps, filters)

		assert.Len(t, filteredApps, 1)
		assert.Equal(t, 3, filteredCount)
		assert.Equal(t, "app3", filteredApps[0].Name)
	})
}

func TestGetAppStatusWithLabelFiltering_DISABLED(t *testing.T) {
	t.Skip("TODO: Update this test to use new endpoints")
}
func TestFilterAppsBySpecificLabel_ShouldNotReturnAppsWithoutLabel(t *testing.T) {
	ResetCacheForTesting()

	// Create test apps: some with version label, some without
	testApps := []AppStatus{
		{
			Name:      "app-with-version",
			Location:  "test-location",
			Status:    "up",
			Source:    "test-source",
			OriginURL: "https://test-origin.com",
			Labels:    []labels.Label{{Key: "version", Value: "v1.0"}, {Key: "environment", Value: "prod"}},
		},
		{
			Name:      "app-without-version",
			Location:  "test-location",
			Status:    "up",
			Source:    "test-source",
			OriginURL: "https://test-origin.com",
			Labels:    []labels.Label{{Key: "environment", Value: "prod"}}, // No version label
		},
		{
			Name:      "app-with-different-version",
			Location:  "test-location",
			Status:    "up",
			Source:    "test-source",
			OriginURL: "https://test-origin.com",
			Labels:    []labels.Label{{Key: "version", Value: "v2.0"}, {Key: "environment", Value: "prod"}},
		},
		{
			Name:      "app-with-empty-version",
			Location:  "test-location",
			Status:    "up",
			Source:    "test-source",
			OriginURL: "https://test-origin.com",
			Labels:    []labels.Label{{Key: "version", Value: ""}, {Key: "environment", Value: "prod"}}, // Empty version
		},
	}

	// Update the cache with test apps
	_ = UpdateAppStatus("test-source", testApps, config.Source{}, config.ServerSettings{HostURL: "https://test-server.com"})

	t.Run("filter by version=v1.0 should only return apps with that exact label", func(t *testing.T) {
		filters := map[string]string{
			"labels.version": "v1.0",
		}

		allApps := GetAppStatusCache()
		filteredApps, _ := filterApps(allApps, filters)

		// Should only return the one app with version=v1.0
		assert.Equal(t, 1, len(filteredApps), "Should only return 1 app with version=v1.0")
		assert.Equal(t, "app-with-version", filteredApps[0].Name, "Should return the app with version=v1.0")

		// Verify it has the correct label
		var found bool
		for _, label := range filteredApps[0].Labels {
			if label.Key == "version" {
				assert.Equal(t, "v1.0", label.Value, "Returned app should have version=v1.0")
				found = true
				break
			}
		}
		assert.True(t, found, "Returned app should have a version label")
	})

	t.Run("filter by version=v2.0 should only return apps with that exact label", func(t *testing.T) {
		filters := map[string]string{
			"labels.version": "v2.0",
		}

		allApps := GetAppStatusCache()
		filteredApps, _ := filterApps(allApps, filters)

		// Should only return the one app with version=v2.0
		assert.Equal(t, 1, len(filteredApps), "Should only return 1 app with version=v2.0")
		assert.Equal(t, "app-with-different-version", filteredApps[0].Name, "Should return the app with version=v2.0")
	})

	t.Run("filter by non-existent version should return no apps", func(t *testing.T) {
		filters := map[string]string{
			"labels.version": "v3.0",
		}

		allApps := GetAppStatusCache()
		filteredApps, _ := filterApps(allApps, filters)

		// Should return no apps
		assert.Equal(t, 0, len(filteredApps), "Should return no apps with non-existent version")
	})

	t.Run("apps without version label should never be returned when filtering by version", func(t *testing.T) {
		filters := map[string]string{
			"labels.version": "v1.0",
		}

		allApps := GetAppStatusCache()
		filteredApps, _ := filterApps(allApps, filters)

		// Check that none of the returned apps are missing the version label
		for _, app := range filteredApps {
			var version string
			var hasVersion bool
			for _, label := range app.Labels {
				if label.Key == "version" {
					version = label.Value
					hasVersion = true
					break
				}
			}
			assert.True(t, hasVersion, "App %s should have version label when filtering by version", app.Name)
			assert.NotEmpty(t, version, "App %s should have non-empty version label", app.Name)
			assert.Equal(t, "v1.0", version, "App %s should have version=v1.0", app.Name)
		}
	})
}

func TestRealWorldBug_VersionLabelFiltering(t *testing.T) {
	ResetCacheForTesting()

	// Create test apps matching the user's actual data
	testApps := []AppStatus{
		{
			Name:      "app1",
			Location:  "Hadera",
			Status:    "unavailable",
			Source:    "prom1",
			OriginURL: "http://prometheus:9090",
			Labels: []labels.Label{
				{Key: "app_type", Value: "web-service"},
				{Key: "criticality", Value: "low"},
				{Key: "datacenter", Value: "dev"},
				{Key: "environment", Value: "development"},
				{Key: "importance", Value: "low"},
				{Key: "owner", Value: "dev-team"},
				{Key: "region", Value: "local"},
				{Key: "service", Value: "dev-monitoring"},
				{Key: "team", Value: "development"},
				{Key: "tier", Value: "backend"},
				{Key: "version", Value: "v1.0"}, // HAS version label
			},
		},
		{
			Name:      "app2",
			Location:  "Hadera",
			Status:    "unavailable",
			Source:    "prom2",
			OriginURL: "http://prometheus2:9090",
			Labels: []labels.Label{
				{Key: "app_type", Value: "web-service"},
				{Key: "beta", Value: "true"},
				{Key: "criticality", Value: "low"},
				{Key: "datacenter", Value: "dev"},
				{Key: "environment", Value: "development"},
				{Key: "importance", Value: "low"},
				{Key: "owner", Value: "dev-team"},
				{Key: "region", Value: "local"},
				{Key: "service", Value: "secondary-monitoring"},
				{Key: "team", Value: "development"},
				{Key: "tier", Value: "testing"},
				// NO version label
			},
		},
		{
			Name:      "app7",
			Location:  "Hadera",
			Status:    "unavailable",
			Source:    "prom2",
			OriginURL: "http://prometheus2:9090",
			Labels: []labels.Label{
				{Key: "app_type", Value: "test-service"},
				{Key: "beta", Value: "true"},
				{Key: "criticality", Value: "low"},
				{Key: "datacenter", Value: "dev"},
				{Key: "environment", Value: "development"},
				{Key: "important", Value: "low"},
				{Key: "inverted", Value: "true"},
				{Key: "owner", Value: "qa-team"},
				{Key: "region", Value: "local"},
				{Key: "service", Value: "secondary-monitoring"},
				{Key: "team", Value: "development"},
				{Key: "tier", Value: "testing"},
				// NO version label
			},
		},
	}

	// Update the cache with test apps
	_ = UpdateAppStatus("prom1", []AppStatus{testApps[0]}, config.Source{}, config.ServerSettings{HostURL: "https://test-server.com"})
	_ = UpdateAppStatus("prom2", []AppStatus{testApps[1], testApps[2]}, config.Source{}, config.ServerSettings{HostURL: "https://test-server.com"})

	t.Run("filter by location=Hadera and labels.version=v1.0 should only return app1", func(t *testing.T) {
		filters := map[string]string{
			"location":       "Hadera",
			"labels.version": "v1.0",
		}

		allApps := GetAppStatusCache()
		filteredApps, _ := filterApps(allApps, filters)

		// Should only return app1 which has both location=Hadera AND version=v1.0
		assert.Equal(t, 1, len(filteredApps), "Should only return 1 app with location=Hadera AND version=v1.0")
		assert.Equal(t, "app1", filteredApps[0].Name, "Should return only app1")

		// Verify the returned app has the correct labels
		assert.Equal(t, "Hadera", filteredApps[0].Location, "Returned app should have location=Hadera")

		// Find version label
		var version string
		for _, label := range filteredApps[0].Labels {
			if label.Key == "version" {
				version = label.Value
				break
			}
		}
		assert.Equal(t, "v1.0", version, "Returned app should have version=v1.0")

		// Verify apps without version label are NOT returned
		for _, app := range filteredApps {
			var hasVersion bool
			for _, label := range app.Labels {
				if label.Key == "version" {
					hasVersion = true
					break
				}
			}
			assert.True(t, hasVersion, "App %s should have version label when filtering by version", app.Name)
		}
	})

	t.Run("filter by only location=Hadera should return all 3 apps", func(t *testing.T) {
		filters := map[string]string{
			"location": "Hadera",
		}

		allApps := GetAppStatusCache()
		filteredApps, _ := filterApps(allApps, filters)

		// Should return all 3 apps in Hadera
		assert.Equal(t, 3, len(filteredApps), "Should return all 3 apps in Hadera")
	})

	t.Run("filter by only labels.version=v1.0 should return only app1", func(t *testing.T) {
		filters := map[string]string{
			"labels.version": "v1.0",
		}

		allApps := GetAppStatusCache()
		filteredApps, _ := filterApps(allApps, filters)

		// Should only return app1
		assert.Equal(t, 1, len(filteredApps), "Should only return 1 app with version=v1.0")
		assert.Equal(t, "app1", filteredApps[0].Name, "Should return only app1")
	})
}

// TestCircularScrapingPrevention tests all the comprehensive scenarios we discussed
func TestCircularScrapingPrevention(t *testing.T) {
	// Server configuration
	serverSettings := config.ServerSettings{
		HostURL: "https://site-a.com",
		Labels:  map[string]string{"server_env": "test"},
	}

	t.Run("UpdateAppStatus no longer filters apps - filtering moved to site scraper", func(t *testing.T) {
		setupTest()

		apps := []AppStatus{
			{Name: "app1", Location: "us-east", Status: "up", OriginURL: "https://site-a.com"},   // Should be kept (no filtering in UpdateAppStatus)
			{Name: "app2", Location: "us-west", Status: "up", OriginURL: "https://external.com"}, // Should be kept
		}

		result := UpdateAppStatus("site-source", apps, config.Source{Type: "site"}, serverSettings)

		assert.Equal(t, 2, result.AppsAdded, "Expected 2 apps added")
		assert.Equal(t, 0, result.AppsSkipped, "Expected 0 apps skipped")
		assert.Nil(t, result.Error, "Expected no error")

		cache := GetAppStatusCache()
		assert.Len(t, cache, 2, "Expected 2 apps in cache")
	})

	t.Run("Rule 1: Non-site scrapers keep apps with own host_url", func(t *testing.T) {
		setupTest()

		apps := []AppStatus{
			{Name: "app1", Location: "us-east", Status: "up", OriginURL: "https://site-a.com"},   // Should be kept (not a site scraper)
			{Name: "app2", Location: "us-west", Status: "up", OriginURL: "https://external.com"}, // Should be kept
		}

		result := UpdateAppStatus("prometheus-source", apps, config.Source{Type: "prometheus"}, serverSettings)

		assert.Equal(t, 2, result.AppsAdded, "Expected 2 apps added")
		assert.Equal(t, 0, result.AppsSkipped, "Expected 0 apps skipped")
		assert.Nil(t, result.Error, "Expected no error")

		cache := GetAppStatusCache()
		assert.Len(t, cache, 2, "Expected 2 apps in cache")
	})

	t.Run("All apps passed to UpdateAppStatus are kept (no filtering)", func(t *testing.T) {
		setupTest()

		apps := []AppStatus{
			{Name: "app1", Location: "us-east", Status: "up", OriginURL: "https://site-b.com"},   // Should be kept (no filtering in UpdateAppStatus)
			{Name: "app2", Location: "us-west", Status: "up", OriginURL: "https://site-c.com"},   // Should be kept
			{Name: "app3", Location: "us-west", Status: "up", OriginURL: "https://site-d.com"},   // Should be kept
			{Name: "app4", Location: "us-west", Status: "up", OriginURL: "https://external.com"}, // Should be kept
		}

		result := UpdateAppStatus("site-source", apps, config.Source{}, serverSettings)

		assert.Equal(t, 4, result.AppsAdded, "Expected 4 apps added")
		assert.Equal(t, 0, result.AppsSkipped, "Expected 0 apps skipped")
		assert.Nil(t, result.Error, "Expected no error")

		cache := GetAppStatusCache()
		assert.Len(t, cache, 4, "Expected 4 apps in cache")
	})

	t.Run("Multi-path scenario: Same app from different sources creates separate entries", func(t *testing.T) {
		setupTest()

		// Apps from different sources with same name but different source
		appsFromSiteB := []AppStatus{
			{Name: "shared-app", Location: "us-east", Status: "up", OriginURL: "https://site-d.com", Source: "site-b"},
		}
		result1 := UpdateAppStatus("site-b-source", appsFromSiteB, config.Source{}, serverSettings)

		// Same app name from different source
		appsFromSiteC := []AppStatus{
			{Name: "shared-app", Location: "us-east", Status: "down", OriginURL: "https://site-d.com", Source: "site-c"},
		}
		result2 := UpdateAppStatus("site-c-source", appsFromSiteC, config.Source{}, serverSettings)

		// Both should be kept since they come from different sources
		assert.Equal(t, 1, result1.AppsAdded, "Apps from site-b should be kept")
		assert.Equal(t, 0, result1.AppsSkipped, "No apps should be skipped")
		assert.Equal(t, 1, result2.AppsAdded, "Apps from site-c should be kept")
		assert.Equal(t, 0, result2.AppsSkipped, "No apps should be skipped")

		cache := GetAppStatusCache()
		assert.Len(t, cache, 2, "Both apps should be kept (different sources)")
	})

	t.Run("Site scraper preserves origin URLs from scraped apps", func(t *testing.T) {
		setupTest()

		// Simulate what Site A gets when scraping Site E (not directly scraped)
		// Site E has apps from various origins
		appsFromSiteE := []AppStatus{
			{Name: "prom-app", Location: "us-east", Status: "up", OriginURL: "https://site-e.com", Source: "site-e"},    // From Site E's prometheus
			{Name: "scraped-app", Location: "us-west", Status: "up", OriginURL: "https://site-f.com", Source: "site-e"}, // Site E scraped from Site F
		}

		// The site scraper should preserve these origin URLs
		result := UpdateAppStatus("site-e-source", appsFromSiteE, config.Source{}, serverSettings)

		assert.Equal(t, 2, result.AppsAdded, "Expected 2 apps added")
		assert.Equal(t, 0, result.AppsSkipped, "Expected 0 apps skipped")

		cache := GetAppStatusCache()
		assert.Len(t, cache, 2, "Expected 2 apps in cache")

		originURLs := make(map[string]string)
		for _, app := range cache {
			originURLs[app.Name] = app.OriginURL
		}

		// Verify origin URLs are preserved
		assert.Equal(t, "https://site-e.com", originURLs["prom-app"], "Prometheus app origin should be preserved")
		assert.Equal(t, "https://site-f.com", originURLs["scraped-app"], "Scraped app origin should be preserved")
	})

	t.Run("Prometheus and HTTP scrapers use host_url as origin", func(t *testing.T) {
		setupTest()

		// Simulate prometheus scraper apps (these should have host_url as origin)
		promApps := []AppStatus{
			{Name: "prom-app1", Location: "us-east", Status: "up", OriginURL: "https://site-a.com", Source: "prometheus"},
			{Name: "prom-app2", Location: "us-west", Status: "down", OriginURL: "https://site-a.com", Source: "prometheus"},
		}

		// Simulate HTTP scraper apps (these should also have host_url as origin)
		httpApps := []AppStatus{
			{Name: "http-app1", Location: "eu-west", Status: "up", OriginURL: "https://site-a.com", Source: "http"},
		}

		result1 := UpdateAppStatus("prometheus-source", promApps, config.Source{Type: "prometheus"}, serverSettings)
		result2 := UpdateAppStatus("http-source", httpApps, config.Source{Type: "http"}, serverSettings)

		// Prometheus and HTTP apps should be kept (self-loop prevention only applies to site scrapers)
		assert.Equal(t, 2, result1.AppsAdded, "Prometheus apps should be kept")
		assert.Equal(t, 0, result1.AppsSkipped, "No prometheus apps should be skipped")
		assert.Equal(t, 1, result2.AppsAdded, "HTTP apps should be kept")
		assert.Equal(t, 0, result2.AppsSkipped, "No HTTP apps should be skipped")

		cache := GetAppStatusCache()
		assert.Len(t, cache, 3, "All apps should be kept")
	})

	t.Run("Prometheus apps are always kept in UpdateAppStatus", func(t *testing.T) {
		setupTest()

		// Use a different server with different host_url
		differentServerSettings := config.ServerSettings{
			HostURL: "https://different-site.com",
			Labels:  map[string]string{"server_env": "test"},
		}

		// Simulate prometheus scraper apps with the different host_url
		promApps := []AppStatus{
			{Name: "prom-app1", Location: "us-east", Status: "up", OriginURL: "https://different-site.com", Source: "prometheus"},
		}

		result := UpdateAppStatus("prometheus-source", promApps, config.Source{}, differentServerSettings)

		// All prometheus apps should be kept (no filtering in UpdateAppStatus)
		assert.Equal(t, 1, result.AppsAdded, "App should be kept (no filtering in UpdateAppStatus)")
		assert.Equal(t, 0, result.AppsSkipped, "No apps should be skipped")
	})

	t.Run("Complex scenario: All apps passed to UpdateAppStatus are kept", func(t *testing.T) {
		setupTest()

		// Step 1: Apps from Site B
		appsFromSiteB := []AppStatus{
			{Name: "site-b-own-app", Location: "us-east", Status: "up", OriginURL: "https://site-b.com", Source: "site-b"},
			{Name: "shared-from-d", Location: "us-west", Status: "up", OriginURL: "https://site-d.com", Source: "site-b"},
		}
		result1 := UpdateAppStatus("site-b-source", appsFromSiteB, config.Source{}, serverSettings)

		// Step 2: Apps from Site C
		appsFromSiteC := []AppStatus{
			{Name: "site-c-own-app", Location: "eu-west", Status: "down", OriginURL: "https://site-c.com", Source: "site-c"},
			{Name: "shared-from-d", Location: "us-west", Status: "down", OriginURL: "https://site-d.com", Source: "site-c"}, // Same app, different status
			{Name: "another-from-d", Location: "asia", Status: "up", OriginURL: "https://site-d.com", Source: "site-c"},
		}
		result2 := UpdateAppStatus("site-c-source", appsFromSiteC, config.Source{}, serverSettings)

		// All apps should be kept (no filtering in UpdateAppStatus)
		assert.Equal(t, 2, result1.AppsAdded, "Site B apps should be kept")
		assert.Equal(t, 0, result1.AppsSkipped, "No Site B apps should be skipped")
		assert.Equal(t, 3, result2.AppsAdded, "Site C apps should be kept")
		assert.Equal(t, 0, result2.AppsSkipped, "No Site C apps should be skipped")

		cache := GetAppStatusCache()
		assert.Len(t, cache, 5, "All apps should be kept")
	})

	t.Run("Valid scenario: Site A scrapes Site E (not directly scraped)", func(t *testing.T) {
		setupTest()

		// Site A scrapes Site E (which is NOT in directScrapedSites)
		appsFromSiteE := []AppStatus{
			{Name: "site-e-app", Location: "us-east", Status: "up", OriginURL: "https://site-e.com", Source: "site-e"},
			{Name: "e-scraped-from-f", Location: "us-west", Status: "down", OriginURL: "https://site-f.com", Source: "site-e"},
		}

		result := UpdateAppStatus("site-e-source", appsFromSiteE, config.Source{}, serverSettings)

		assert.Equal(t, 2, result.AppsAdded, "Expected 2 apps added")
		assert.Equal(t, 0, result.AppsSkipped, "Expected 0 apps skipped")
		assert.Nil(t, result.Error, "Expected no error")

		cache := GetAppStatusCache()
		assert.Len(t, cache, 2, "Expected 2 apps in cache")

		// Verify origin URLs are preserved
		origins := make(map[string]string)
		for _, app := range cache {
			origins[app.Name] = app.OriginURL
		}

		assert.Equal(t, "https://site-e.com", origins["site-e-app"], "Site E app origin should be preserved")
		assert.Equal(t, "https://site-f.com", origins["e-scraped-from-f"], "Site F app origin should be preserved")
	})

	t.Run("Edge case: Apps with empty origin URL should be rejected", func(t *testing.T) {
		setupTest()

		apps := []AppStatus{
			{Name: "no-origin-app", Location: "us-east", Status: "up", OriginURL: "", Source: "legacy"},
			{Name: "nil-origin-app", Location: "us-west", Status: "down", Source: "legacy"},                                    // No OriginURL field
			{Name: "valid-app", Location: "us-central", Status: "up", OriginURL: "https://valid-origin.com", Source: "legacy"}, // Valid app
		}

		result := UpdateAppStatus("legacy-source", apps, config.Source{}, serverSettings)

		assert.Equal(t, 1, result.AppsAdded, "Expected 1 app added (only the valid one)")
		assert.Equal(t, 2, result.AppsSkipped, "Expected 2 apps skipped (empty OriginURL)")
		assert.Nil(t, result.Error, "Expected no error")

		cache := GetAppStatusCache()
		assert.Len(t, cache, 1, "Expected 1 app in cache (only valid one)")
		assert.Equal(t, "valid-app", cache[0].Name, "Only the valid app should be in cache")
	})

	t.Run("Input validation: Invalid server settings", func(t *testing.T) {
		setupTest()

		apps := []AppStatus{
			{Name: "test-app", Location: "us-east", Status: "up", OriginURL: "https://test.com", Source: "test"},
		}

		// Test empty source name
		result1 := UpdateAppStatus("", apps, config.Source{}, serverSettings)
		assert.NotNil(t, result1.Error, "Expected error for empty source name")
		assert.Equal(t, 0, result1.AppsAdded, "No apps should be added with invalid input")

		// Test empty host URL
		invalidServerSettings := config.ServerSettings{HostURL: ""}
		result2 := UpdateAppStatus("test-source", apps, config.Source{}, invalidServerSettings)
		assert.NotNil(t, result2.Error, "Expected error for empty host URL")
		assert.Equal(t, 0, result2.AppsAdded, "No apps should be added with invalid input")
	})

	t.Run("Input validation: Invalid app data", func(t *testing.T) {
		setupTest()

		apps := []AppStatus{
			{Name: "", Location: "us-east", Status: "up", OriginURL: "https://test.com", Source: "test"},                    // Empty name
			{Name: "valid-app", Location: "", Status: "up", OriginURL: "https://test.com", Source: "test"},                  // Empty location
			{Name: "invalid-status", Location: "us-east", Status: "unknown", OriginURL: "https://test.com", Source: "test"}, // Invalid status
			{Name: "valid-app2", Location: "us-west", Status: "down", OriginURL: "https://test.com", Source: "test"},        // Valid app
		}

		result := UpdateAppStatus("test-source", apps, config.Source{}, serverSettings)

		assert.Equal(t, 2, result.AppsAdded, "Expected 2 valid apps added") // invalid-status gets corrected to unavailable
		assert.Equal(t, 2, result.AppsSkipped, "Expected 2 apps skipped for validation issues")
		assert.Nil(t, result.Error, "Expected no error")

		cache := GetAppStatusCache()
		assert.Len(t, cache, 2, "Expected 2 apps in cache")

		// Verify the invalid status was corrected
		for _, app := range cache {
			if app.Name == "invalid-status" {
				assert.Equal(t, "unavailable", app.Status, "Invalid status should be corrected to unavailable")
			}
		}
	})

	t.Run("Performance metrics tracking", func(t *testing.T) {
		setupTest()

		apps := []AppStatus{
			{Name: "test-app", Location: "us-east", Status: "up", OriginURL: "https://external.com", Source: "test"},
		}

		// Get initial metrics
		initialMetrics := GetUpdateMetrics()
		initialUpdates := initialMetrics["total_updates"].(int64)
		initialAdded := initialMetrics["total_apps_added"].(int64)

		result := UpdateAppStatus("test-source", apps, config.Source{}, serverSettings)

		assert.Equal(t, 1, result.AppsAdded, "Expected 1 app added")
		assert.Nil(t, result.Error, "Expected no error")

		// Check updated metrics
		newMetrics := GetUpdateMetrics()
		assert.Equal(t, initialUpdates+1, newMetrics["total_updates"].(int64), "Total updates should increment")
		assert.Equal(t, initialAdded+1, newMetrics["total_apps_added"].(int64), "Total apps added should increment")
	})
}
