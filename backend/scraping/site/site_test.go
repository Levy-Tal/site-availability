package site

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"site-availability/authentication/hmac"
	"site-availability/config"
	"site-availability/handlers"
	"site-availability/logging"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupSiteTest initializes logging for tests
func setupSiteTest() {
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
				Name:     "app1",
				Location: "location1",
				Status:   "up",
				Source:   "", // Will be set by scraper
			},
			{
				Name:     "app2",
				Location: "location2",
				Status:   "down",
				Source:   "", // Will be set by scraper
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
			_ = json.NewEncoder(w).Encode(expectedStatuses)
		}))
		defer server.Close()

		scraper := NewSiteScraper()
		source := config.Source{
			Name: "test-site",
			Type: "site",
			URL:  server.URL,
			// No token - no authentication
		}

		results, err := scraper.Scrape(source, 5*time.Second, 1, nil)
		require.NoError(t, err)
		require.Len(t, results, 2)

		// Verify the source was set correctly
		assert.Equal(t, "app1", results[0].Name)
		assert.Equal(t, "test-site", results[0].Source)
		assert.Equal(t, "app2", results[1].Name)
		assert.Equal(t, "test-site", results[1].Source)
	})

	t.Run("successful scrape with HMAC authentication", func(t *testing.T) {
		token := "test-secret-token"
		expectedStatuses := []handlers.AppStatus{
			{
				Name:     "authenticated-app",
				Location: "secure-location",
				Status:   "up",
				Source:   "",
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
			_ = json.NewEncoder(w).Encode(expectedStatuses)
		}))
		defer server.Close()

		scraper := NewSiteScraper()
		source := config.Source{
			Name:  "authenticated-site",
			Type:  "site",
			URL:   server.URL,
			Token: token,
		}

		results, err := scraper.Scrape(source, 5*time.Second, 1, nil)
		require.NoError(t, err)
		require.Len(t, results, 1)

		assert.Equal(t, "authenticated-app", results[0].Name)
		assert.Equal(t, "authenticated-site", results[0].Source)
	})

	t.Run("successful scrape with TLS config", func(t *testing.T) {
		expectedStatuses := []handlers.AppStatus{
			{
				Name:     "tls-app",
				Location: "secure-location",
				Status:   "up",
				Source:   "",
			},
		}

		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(expectedStatuses)
		}))
		defer server.Close()

		scraper := NewSiteScraper()
		source := config.Source{
			Name: "tls-site",
			Type: "site",
			URL:  server.URL,
		}

		// Use insecure TLS config for testing
		tlsConfig := &tls.Config{InsecureSkipVerify: true}

		results, err := scraper.Scrape(source, 5*time.Second, 1, tlsConfig)
		require.NoError(t, err)
		require.Len(t, results, 1)

		assert.Equal(t, "tls-app", results[0].Name)
		assert.Equal(t, "tls-site", results[0].Source)
	})

	t.Run("scrape with empty response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]handlers.AppStatus{})
		}))
		defer server.Close()

		scraper := NewSiteScraper()
		source := config.Source{
			Name: "empty-site",
			Type: "site",
			URL:  server.URL,
		}

		results, err := scraper.Scrape(source, 5*time.Second, 1, nil)
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("invalid request URL", func(t *testing.T) {
		scraper := NewSiteScraper()
		source := config.Source{
			Name: "invalid-site",
			Type: "site",
			URL:  "http://invalid url with spaces", // Invalid URL
		}

		results, err := scraper.Scrape(source, 5*time.Second, 1, nil)
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
			URL:  "http://localhost:99999", // Non-existent server
		}

		results, err := scraper.Scrape(source, 5*time.Second, 1, nil)
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
					URL:  server.URL,
				}

				results, err := scraper.Scrape(source, 5*time.Second, 1, nil)
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
			URL:  server.URL,
		}

		results, err := scraper.Scrape(source, 5*time.Second, 1, nil)
		require.NoError(t, err) // JSON errors are handled gracefully
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
			URL:  server.URL,
		}

		results, err := scraper.Scrape(source, 5*time.Second, 1, nil)
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
			URL:  server.URL,
		}

		// Use very short timeout
		results, err := scraper.Scrape(source, 10*time.Millisecond, 1, nil)
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
			_ = json.NewEncoder(w).Encode([]handlers.AppStatus{})
		}))
		defer server.Close()

		scraper := NewSiteScraper()
		source := config.Source{
			Name: "test-site",
			Type: "site",
			URL:  server.URL,
		}

		// Try with different maxParallel values - should always make only 1 request
		results, err := scraper.Scrape(source, 5*time.Second, 100, nil)
		require.NoError(t, err)
		assert.Empty(t, results)
		assert.Equal(t, 1, requestCount)
	})

	t.Run("verify URL construction", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify the exact URL path
			assert.Equal(t, "/sync", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]handlers.AppStatus{})
		}))
		defer server.Close()

		scraper := NewSiteScraper()
		source := config.Source{
			Name: "path-test-site",
			Type: "site",
			URL:  server.URL, // Should append /sync to this
		}

		results, err := scraper.Scrape(source, 5*time.Second, 1, nil)
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
			_ = json.NewEncoder(w).Encode([]handlers.AppStatus{})
		}))
		defer server.Close()

		scraper := NewSiteScraper()
		source := config.Source{
			Name: "timestamp-test-site",
			Type: "site",
			URL:  server.URL,
		}

		results, err := scraper.Scrape(source, 5*time.Second, 1, nil)
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("source name override for multiple apps", func(t *testing.T) {
		originalStatuses := []handlers.AppStatus{
			{
				Name:     "app1",
				Location: "location1",
				Status:   "up",
				Source:   "original-source", // This should be overridden
			},
			{
				Name:     "app2",
				Location: "location2",
				Status:   "down",
				Source:   "another-source", // This should also be overridden
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(originalStatuses)
		}))
		defer server.Close()

		scraper := NewSiteScraper()
		source := config.Source{
			Name: "override-test-site",
			Type: "site",
			URL:  server.URL,
		}

		results, err := scraper.Scrape(source, 5*time.Second, 1, nil)
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
				Name:     "special-app",
				Location: "special-location",
				Status:   "up",
				Source:   "",
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
			_ = json.NewEncoder(w).Encode(expectedStatuses)
		}))
		defer server.Close()

		scraper := NewSiteScraper()
		source := config.Source{
			Name:  "special-token-site",
			Type:  "site",
			URL:   server.URL,
			Token: token,
		}

		results, err := scraper.Scrape(source, 5*time.Second, 1, nil)
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "special-token-site", results[0].Source)
	})
}
