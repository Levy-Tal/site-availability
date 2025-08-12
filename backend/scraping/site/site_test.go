package site

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"site-availability/authentication/hmac"
	"site-availability/config"
	"site-availability/handlers"
	"site-availability/labels"
	"site-availability/logging"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupSiteTest initializes logging for tests
func setupSiteTest() {
	// Set log level to panic to suppress error logs during tests
	os.Setenv("LOG_LEVEL", "panic")
	_ = logging.Init()
}

func TestMain(m *testing.M) {
	setupSiteTest()
	m.Run()
}

func TestNewSiteScraper(t *testing.T) {
	t.Run("create new site scraper", func(t *testing.T) {
		scraper := NewSiteScraper()
		assert.NotNil(t, scraper)
		assert.IsType(t, &SiteScraper{}, scraper)
	})
}

func TestSiteScraper_Scrape(t *testing.T) {
	setupSiteTest()

	t.Run("successful scrape without authentication", func(t *testing.T) {
		expectedStatuses := []handlers.AppStatus{
			{
				Name:      "app1",
				Location:  "location1",
				Status:    "up",
				Source:    "",
				OriginURL: "http://test-origin.com", // Should be preserved by scraper
			},
			{
				Name:      "app2",
				Location:  "location2",
				Status:    "down",
				Source:    "",
				OriginURL: "http://test-origin.com", // Should be preserved by scraper
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/sync", r.URL.Path)

			// Verify headers - no authentication expected
			timestamp := r.Header.Get("X-Site-Sync-Timestamp")
			assert.NotEmpty(t, timestamp)
			signature := r.Header.Get("X-Site-Sync-Signature")
			assert.Empty(t, signature) // No token provided

			// Return successful response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := handlers.StatusResponse{Apps: expectedStatuses, Locations: []handlers.Location{}}
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		scraper := NewSiteScraper()
		source := config.Source{
			Name: "test-site",
			Type: "site",
			Config: map[string]interface{}{
				"url": server.URL,
				// No token - no authentication
			},
		}

		results, _, err := scraper.Scrape(source, config.ServerSettings{}, 5*time.Second, 1, nil)
		require.NoError(t, err)
		require.Len(t, results, 2)

		// Verify the source was set correctly and origin URL was preserved
		assert.Equal(t, "app1", results[0].Name)
		assert.Equal(t, "test-site", results[0].Source)
		assert.Equal(t, "http://test-origin.com", results[0].OriginURL) // Should preserve original origin URL
		assert.Equal(t, "app2", results[1].Name)
		assert.Equal(t, "test-site", results[1].Source)
		assert.Equal(t, "http://test-origin.com", results[1].OriginURL) // Should preserve original origin URL
	})

	t.Run("successful scrape with HMAC authentication", func(t *testing.T) {
		token := "test-secret-token"
		expectedStatuses := []handlers.AppStatus{
			{
				Name:      "authenticated-app",
				Location:  "secure-location",
				Status:    "up",
				Source:    "",
				OriginURL: "http://test-origin.com",
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/sync", r.URL.Path)

			// Verify HMAC authentication
			timestamp := r.Header.Get("X-Site-Sync-Timestamp")
			signature := r.Header.Get("X-Site-Sync-Signature")
			assert.NotEmpty(t, timestamp)
			assert.NotEmpty(t, signature)

			// Validate HMAC signature
			validator := hmac.NewValidator(token)
			expectedSignature := validator.GenerateSignature(timestamp, []byte{})
			assert.Equal(t, expectedSignature, signature)

			// Return successful response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := handlers.StatusResponse{Apps: expectedStatuses, Locations: []handlers.Location{}}
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		scraper := NewSiteScraper()
		source := config.Source{
			Name: "authenticated-site",
			Type: "site",
			Config: map[string]interface{}{
				"url":   server.URL,
				"token": token,
			},
		}

		results, _, err := scraper.Scrape(source, config.ServerSettings{}, 5*time.Second, 1, nil)
		require.NoError(t, err)
		require.Len(t, results, 1)

		assert.Equal(t, "authenticated-app", results[0].Name)
		assert.Equal(t, "authenticated-site", results[0].Source)
	})

	t.Run("successful scrape with TLS config", func(t *testing.T) {
		expectedStatuses := []handlers.AppStatus{
			{
				Name:      "tls-app",
				Location:  "secure-location",
				Status:    "up",
				Source:    "",
				OriginURL: "http://test-origin.com",
			},
		}

		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := handlers.StatusResponse{Apps: expectedStatuses, Locations: []handlers.Location{}}
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		scraper := NewSiteScraper()
		source := config.Source{
			Name: "tls-site",
			Type: "site",
			Config: map[string]interface{}{
				"url": server.URL,
			},
		}

		// Use insecure TLS config for testing
		tlsConfig := &tls.Config{InsecureSkipVerify: true}

		results, _, err := scraper.Scrape(source, config.ServerSettings{}, 5*time.Second, 1, tlsConfig)
		require.NoError(t, err)
		require.Len(t, results, 1)

		assert.Equal(t, "tls-app", results[0].Name)
		assert.Equal(t, "tls-site", results[0].Source)
	})

	t.Run("scrape with empty response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := handlers.StatusResponse{Apps: []handlers.AppStatus{}, Locations: []handlers.Location{}}
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		scraper := NewSiteScraper()
		source := config.Source{
			Name: "empty-site",
			Type: "site",
			Config: map[string]interface{}{
				"url": server.URL,
			},
		}

		results, _, err := scraper.Scrape(source, config.ServerSettings{}, 5*time.Second, 1, nil)
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("invalid request URL", func(t *testing.T) {
		scraper := NewSiteScraper()
		source := config.Source{
			Name: "invalid-site",
			Type: "site",
			Config: map[string]interface{}{
				"url": "http://invalid url with spaces", // Invalid URL
			},
		}

		results, _, err := scraper.Scrape(source, config.ServerSettings{}, 5*time.Second, 1, nil)
		assert.Error(t, err)
		assert.Nil(t, results)
		assert.Contains(t, err.Error(), "failed to create request")
		assert.Contains(t, err.Error(), "invalid-site")
	})

	t.Run("network connection error", func(t *testing.T) {
		scraper := NewSiteScraper()
		source := config.Source{
			Name: "unreachable-site",
			Type: "site",
			Config: map[string]interface{}{
				"url": "http://localhost:99999", // Non-existent server
			},
		}

		results, _, err := scraper.Scrape(source, config.ServerSettings{}, 5*time.Second, 1, nil)
		require.NoError(t, err) // Network errors are handled gracefully
		assert.Empty(t, results)
	})

	t.Run("HTTP error status codes", func(t *testing.T) {
		testCases := []struct {
			name       string
			statusCode int
		}{
			{"404 not found", http.StatusNotFound},
			{"401 unauthorized", http.StatusUnauthorized},
			{"403 forbidden", http.StatusForbidden},
			{"500 internal server error", http.StatusInternalServerError},
			{"503 service unavailable", http.StatusServiceUnavailable},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tc.statusCode)
				}))
				defer server.Close()

				scraper := NewSiteScraper()
				source := config.Source{
					Name: "error-site",
					Type: "site",
					Config: map[string]interface{}{
						"url": server.URL,
					},
				}

				results, _, err := scraper.Scrape(source, config.ServerSettings{}, 5*time.Second, 1, nil)
				require.NoError(t, err) // HTTP errors are handled gracefully
				assert.Empty(t, results)
			})
		}
	})

	t.Run("invalid JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("invalid json{"))
		}))
		defer server.Close()

		scraper := NewSiteScraper()
		source := config.Source{
			Name: "invalid-json-site",
			Type: "site",
			Config: map[string]interface{}{
				"url": server.URL,
			},
		}

		results, _, err := scraper.Scrape(source, config.ServerSettings{}, 5*time.Second, 1, nil)
		require.NoError(t, err) // JSON parsing errors are handled gracefully
		assert.Empty(t, results)
	})

	t.Run("malformed JSON structure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			// Valid JSON but wrong structure
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "wrong format"})
		}))
		defer server.Close()

		scraper := NewSiteScraper()
		source := config.Source{
			Name: "malformed-site",
			Type: "site",
			Config: map[string]interface{}{
				"url": server.URL,
			},
		}

		results, _, err := scraper.Scrape(source, config.ServerSettings{}, 5*time.Second, 1, nil)
		require.NoError(t, err) // JSON structure errors are handled gracefully
		assert.Empty(t, results)
	})

	t.Run("timeout handling", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate slow response
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		scraper := NewSiteScraper()
		source := config.Source{
			Name: "slow-site",
			Type: "site",
			Config: map[string]interface{}{
				"url": server.URL,
			},
		}

		// Use very short timeout
		results, _, err := scraper.Scrape(source, config.ServerSettings{}, 10*time.Millisecond, 1, nil)
		require.NoError(t, err) // Timeout errors are handled gracefully
		assert.Empty(t, results)
	})

	t.Run("maxParallel parameter ignored", func(t *testing.T) {
		// This test verifies that maxParallel doesn't affect site scraping
		// since it only makes one request
		requestCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := handlers.StatusResponse{Apps: []handlers.AppStatus{}, Locations: []handlers.Location{}}
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		scraper := NewSiteScraper()
		source := config.Source{
			Name: "test-site",
			Type: "site",
			Config: map[string]interface{}{
				"url": server.URL,
			},
		}

		// Try with different maxParallel values - should always make only 1 request
		results, _, err := scraper.Scrape(source, config.ServerSettings{}, 5*time.Second, 100, nil)
		require.NoError(t, err)
		assert.Empty(t, results)
		assert.Equal(t, 1, requestCount)
	})

	t.Run("verify URL construction", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify the exact URL path
			assert.Equal(t, "/sync", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			response := handlers.StatusResponse{Apps: []handlers.AppStatus{}, Locations: []handlers.Location{}}
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		scraper := NewSiteScraper()
		source := config.Source{
			Name: "path-test-site",
			Type: "site",
			Config: map[string]interface{}{
				"url": server.URL, // Should append /sync to this
			},
		}

		results, _, err := scraper.Scrape(source, config.ServerSettings{}, 5*time.Second, 1, nil)
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("verify timestamp format", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			timestamp := r.Header.Get("X-Site-Sync-Timestamp")
			assert.NotEmpty(t, timestamp)

			// Verify timestamp is in RFC3339 format and recent
			ts, err := time.Parse(time.RFC3339, timestamp)
			assert.NoError(t, err)
			assert.True(t, time.Since(ts) < time.Minute)

			w.WriteHeader(http.StatusOK)
			response := handlers.StatusResponse{Apps: []handlers.AppStatus{}, Locations: []handlers.Location{}}
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		scraper := NewSiteScraper()
		source := config.Source{
			Name: "timestamp-test-site",
			Type: "site",
			Config: map[string]interface{}{
				"url": server.URL,
			},
		}

		results, _, err := scraper.Scrape(source, config.ServerSettings{}, 5*time.Second, 1, nil)
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("source name override for multiple apps", func(t *testing.T) {
		originalStatuses := []handlers.AppStatus{
			{
				Name:      "app1",
				Location:  "location1",
				Status:    "up",
				Source:    "original-source",
				OriginURL: "http://test-origin.com", // This should be preserved
			},
			{
				Name:      "app2",
				Location:  "location2",
				Status:    "down",
				Source:    "another-source",
				OriginURL: "http://test-origin.com", // This should also be preserved
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := handlers.StatusResponse{Apps: originalStatuses, Locations: []handlers.Location{}}
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		scraper := NewSiteScraper()
		source := config.Source{
			Name: "override-test-site",
			Type: "site",
			Config: map[string]interface{}{
				"url": server.URL,
			},
		}

		results, _, err := scraper.Scrape(source, config.ServerSettings{}, 5*time.Second, 1, nil)
		require.NoError(t, err)
		require.Len(t, results, 2)

		// All apps should have their source set to the scraper's source name
		for _, result := range results {
			assert.Equal(t, "override-test-site", result.Source)
		}

		// But other fields should remain unchanged
		assert.Equal(t, "app1", results[0].Name)
		assert.Equal(t, "location1", results[0].Location)
		assert.Equal(t, "up", results[0].Status)
		assert.Equal(t, "app2", results[1].Name)
		assert.Equal(t, "location2", results[1].Location)
		assert.Equal(t, "down", results[1].Status)
	})

	t.Run("complex authentication scenario", func(t *testing.T) {
		// Test with special characters in token
		token := "test-token-with-special-chars!@#$%^&*()"
		expectedStatuses := []handlers.AppStatus{
			{
				Name:      "special-app",
				Location:  "special-location",
				Status:    "up",
				Source:    "",
				OriginURL: "http://test-origin.com",
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			timestamp := r.Header.Get("X-Site-Sync-Timestamp")
			signature := r.Header.Get("X-Site-Sync-Signature")

			// Validate the HMAC with special character token
			validator := hmac.NewValidator(token)
			expectedSignature := validator.GenerateSignature(timestamp, []byte{})
			assert.Equal(t, expectedSignature, signature)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := handlers.StatusResponse{Apps: expectedStatuses, Locations: []handlers.Location{}}
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		scraper := NewSiteScraper()
		source := config.Source{
			Name: "special-token-site",
			Type: "site",
			Config: map[string]interface{}{
				"url":   server.URL,
				"token": token,
			},
		}

		results, _, err := scraper.Scrape(source, config.ServerSettings{}, 5*time.Second, 1, nil)
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "special-token-site", results[0].Source)
	})

	t.Run("scrape with label merging", func(t *testing.T) {
		// Create remote app statuses with labels
		remoteStatuses := []handlers.AppStatus{
			{
				Name:      "remote-app",
				Location:  "remote-location",
				Status:    "up",
				Source:    "original-remote-source",
				OriginURL: "http://test-origin.com", // Should be preserved
				Labels: []labels.Label{
					{Key: "remote_version", Value: "v2.1.0"},
					{Key: "remote_tier", Value: "backend"},
					{Key: "common_label", Value: "remote_value"}, // Highest priority
				},
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := handlers.StatusResponse{Apps: remoteStatuses, Locations: []handlers.Location{}}
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		scraper := NewSiteScraper()

		// Set up server settings with labels
		serverSettings := config.ServerSettings{
			Labels: map[string]string{
				"server_env":    "production",
				"server_region": "eu-central",
				"common_label":  "server_value", // Will be overridden by source
			},
		}

		source := config.Source{
			Name: "labeled-site",
			Type: "site",
			Config: map[string]interface{}{
				"url": server.URL,
			},
			Labels: map[string]string{
				"source_type":    "site",
				"source_cluster": "west",
				"common_label":   "source_value", // Will be overridden by remote app
			},
		}

		results, _, err := scraper.Scrape(source, serverSettings, 5*time.Second, 1, nil)
		require.NoError(t, err)
		require.Len(t, results, 1)

		app := results[0]
		assert.Equal(t, "remote-app", app.Name)
		assert.Equal(t, "up", app.Status)
		assert.Equal(t, "labeled-site", app.Source) // Source name is overridden by scraper

		// Verify label merging with correct priority (Remote App > Source > Server)
		// Only remote app labels should be present (merging happens in UpdateAppStatus)
		assert.Len(t, app.Labels, 3) // Only remote app labels

		// Check specific labels
		var remoteVersion, remoteTier, commonLabel string
		for _, label := range app.Labels {
			switch label.Key {
			case "remote_version":
				remoteVersion = label.Value
			case "remote_tier":
				remoteTier = label.Value
			case "common_label":
				commonLabel = label.Value
			}
		}
		assert.Equal(t, "v2.1.0", remoteVersion)
		assert.Equal(t, "backend", remoteTier)
		assert.Equal(t, "remote_value", commonLabel)
	})

	t.Run("circular prevention keeps apps from scraped site but drops third-party and own apps", func(t *testing.T) {
		// Scenario: Server A scrapes both Server B and Server C directly
		// When Server A scrapes Server B, it gets apps from B, C, A, and external sites
		// It should keep apps from B (the site being scraped) and external sites
		// It should drop apps from C (third-party directly scraped) and A (stale copies of own apps)

		// Create test server first to get the real URL
		var testServerURL string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := handlers.StatusResponse{Apps: []handlers.AppStatus{
				{
					Name:      "app-from-b",
					Location:  "location-b",
					Status:    "up",
					Source:    "original-source-b",
					OriginURL: testServerURL, // This is the site we're scraping from - should KEEP
				},
				{
					Name:      "app-from-c",
					Location:  "location-c",
					Status:    "down",
					Source:    "original-source-c",
					OriginURL: "http://server-c:8100", // This is a third-party site we also scrape - should DROP
				},
				{
					Name:      "app-from-a-stale",
					Location:  "location-a",
					Status:    "up",
					Source:    "original-source-a",
					OriginURL: "http://server-a:8080", // This is our own server (stale copy) - should DROP
				},
				{
					Name:      "app-from-external",
					Location:  "location-external",
					Status:    "up",
					Source:    "external-source",
					OriginURL: "http://external:9000", // This is external site we don't scrape - should KEEP
				},
			}, Locations: []handlers.Location{}}
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()
		testServerURL = server.URL

		scraper := NewSiteScraper()

		// Simulate that Server A directly scrapes both Server B and Server C
		directScrapedSites := []string{
			testServerURL,          // The site we're currently scraping (matches app origin)
			"http://server-c:8100", // Third-party site we also scrape directly
		}

		source := config.Source{
			Name: "server-b-source",
			Type: "site",
			Config: map[string]interface{}{
				"url": testServerURL, // We're scraping server B (test server)
			},
		}

		serverSettings := config.ServerSettings{
			HostURL: "http://server-a:8080", // This is Server A's own host URL
		}

		results, _, err := scraper.ScrapeWithCircularPrevention(
			source,
			serverSettings,
			5*time.Second,
			1,
			nil,
			directScrapedSites,
		)
		require.NoError(t, err)

		// Should have 2 apps: one from server B (kept) and one external (kept)
		// The app from server C and the stale app from server A should be dropped
		if len(results) != 2 {
			t.Logf("Expected 2 apps, got %d apps:", len(results))
			for i, app := range results {
				t.Logf("  App %d: name=%s, origin_url=%s", i, app.Name, app.OriginURL)
			}
		}
		require.Len(t, results, 2)

		// Verify we kept the app from server B (the site we're scraping)
		appFromB := findAppByName(results, "app-from-b")
		require.NotNil(t, appFromB, "Should keep app from the site being scraped")
		assert.Equal(t, testServerURL, appFromB.OriginURL)
		assert.Equal(t, "server-b-source", appFromB.Source)

		// Verify we kept the external app (not directly scraped)
		appFromExternal := findAppByName(results, "app-from-external")
		require.NotNil(t, appFromExternal, "Should keep app from external site")
		assert.Equal(t, "http://external:9000", appFromExternal.OriginURL)
		assert.Equal(t, "server-b-source", appFromExternal.Source)

		// Verify we dropped the app from server C (third-party directly scraped)
		appFromC := findAppByName(results, "app-from-c")
		assert.Nil(t, appFromC, "Should drop app from third-party directly scraped site")

		// Verify we dropped the stale app from our own server A
		appFromAStale := findAppByName(results, "app-from-a-stale")
		assert.Nil(t, appFromAStale, "Should drop stale app from our own server")
	})
}

// Helper function to find an app by name in results
func findAppByName(apps []handlers.AppStatus, name string) *handlers.AppStatus {
	for _, app := range apps {
		if app.Name == name {
			return &app
		}
	}
	return nil
}
