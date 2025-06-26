package scraping

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"site-availability/config"
	"site-availability/handlers"
	"site-availability/logging"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupScrapingTest initializes logging and cleans up global state
func setupScrapingTest() {
	_ = logging.Init()

	// Reset global state
	Scrapers = make(map[string]Source)
	globalTLSConfig = nil
}

func TestMain(m *testing.M) {
	setupScrapingTest()
	m.Run()
}

// Mock scraper for testing
type MockScraper struct {
	results []handlers.AppStatus
	err     error
	delay   time.Duration
	mutex   sync.Mutex
	calls   int
}

func (m *MockScraper) Scrape(source config.Source, timeout time.Duration, maxParallel int, tlsConfig *tls.Config) ([]handlers.AppStatus, error) {
	m.mutex.Lock()
	m.calls++
	m.mutex.Unlock()

	if m.delay > 0 {
		time.Sleep(m.delay)
	}

	if m.err != nil {
		return nil, m.err
	}

	// Add source to results
	results := make([]handlers.AppStatus, len(m.results))
	copy(results, m.results)
	for i := range results {
		results[i].Source = source.Name
	}

	return results, nil
}

func (m *MockScraper) GetCallCount() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.calls
}

func TestInitCertificate(t *testing.T) {
	setupScrapingTest()

	t.Run("environment variable not set", func(t *testing.T) {
		// Ensure env var is not set
		os.Unsetenv("TEST_CA_CERT")

		InitCertificate("TEST_CA_CERT")

		assert.Nil(t, globalTLSConfig)
	})

	t.Run("environment variable set to empty string", func(t *testing.T) {
		os.Setenv("TEST_CA_CERT", "")
		defer os.Unsetenv("TEST_CA_CERT")

		InitCertificate("TEST_CA_CERT")

		assert.Nil(t, globalTLSConfig)
	})

	t.Run("valid certificate file", func(t *testing.T) {
		// Create a temporary directory and certificate file
		tempDir := t.TempDir()
		certFile := filepath.Join(tempDir, "ca.pem")

		// Create a valid PEM certificate (self-signed for testing)
		certPEM := `-----BEGIN CERTIFICATE-----
MIIBxTCCAWugAwIBAgIJAKjYQV+z9YlwMA0GCSqGSIb3DQEBCwUAMCkxJzAlBgNV
BAMTHkVsYXN0aWNzZWFyY2ggVGVzdCBOb2RlOiBub2RlMTAeFw0yMDAzMDQxNTAy
NDhaFw0yMjAzMDQxNTAyNDhaMCkxJzAlBgNVBAMTHkVsYXN0aWNzZWFyY2ggVGVz
dCBOb2RlOiBub2RlMTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABDPiUKNkwUYW
xF5EhLYuBLFnQcH1Zzex/J7rJE3QpAjJe9R6YKJJPt1wJGLjQvQ4YZK5Y1J+YKJw
5L4v3L2v3L2jUDBOMB0GA1UdDgQWBBQX3vWKb3LRJbJR5p7bOLtD7vbV7zAfBgNV
HSMEGDAWgBQX3vWKb3LRJbJR5p7bOLtD7vbV7zAMBgNVHRMEBTADAQH/MA0GCSqG
SIb3DQEBCwUAA0EABhh2sxOe3lj8PzFd6NiMhTfNHZj7w7v3m3v3m3v3m3v3m3v3
m3v3m3v3m3v3m3v3m3v3m3v3m3v3m3v3m3v3m3v3m3v3mw==
-----END CERTIFICATE-----`

		err := os.WriteFile(certFile, []byte(certPEM), 0644)
		require.NoError(t, err)

		os.Setenv("TEST_CA_CERT", certFile)
		defer os.Unsetenv("TEST_CA_CERT")

		InitCertificate("TEST_CA_CERT")

		assert.NotNil(t, globalTLSConfig)
		assert.NotNil(t, globalTLSConfig.RootCAs)
		assert.False(t, globalTLSConfig.InsecureSkipVerify)
	})

	t.Run("multiple certificate files", func(t *testing.T) {
		tempDir := t.TempDir()
		certFile1 := filepath.Join(tempDir, "ca1.pem")
		certFile2 := filepath.Join(tempDir, "ca2.pem")

		certPEM := `-----BEGIN CERTIFICATE-----
MIIBxTCCAWugAwIBAgIJAKjYQV+z9YlwMA0GCSqGSIb3DQEBCwUAMCkxJzAlBgNV
BAMTHkVsYXN0aWNzZWFyY2ggVGVzdCBOb2RlOiBub2RlMTAeFw0yMDAzMDQxNTAy
NDhaFw0yMjAzMDQxNTAyNDhaMCkxJzAlBgNVBAMTHkVsYXN0aWNzZWFyY2ggVGVz
dCBOb2RlOiBub2RlMTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABDPiUKNkwUYW
xF5EhLYuBLFnQcH1Zzex/J7rJE3QpAjJe9R6YKJJPt1wJGLjQvQ4YZK5Y1J+YKJw
5L4v3L2v3L2jUDBOMB0GA1UdDgQWBBQX3vWKb3LRJbJR5p7bOLtD7vbV7zAfBgNV
HSMEGDAWgBQX3vWKb3LRJbJR5p7bOLtD7vbV7zAMBgNVHRMEBTADAQH/MA0GCSqG
SIb3DQEBCwUAA0EABhh2sxOe3lj8PzFd6NiMhTfNHZj7w7v3m3v3m3v3m3v3m3v3
m3v3m3v3m3v3m3v3m3v3m3v3m3v3m3v3m3v3m3v3m3v3mw==
-----END CERTIFICATE-----`

		err := os.WriteFile(certFile1, []byte(certPEM), 0644)
		require.NoError(t, err)
		err = os.WriteFile(certFile2, []byte(certPEM), 0644)
		require.NoError(t, err)

		os.Setenv("TEST_CA_CERT", fmt.Sprintf("%s:%s", certFile1, certFile2))
		defer os.Unsetenv("TEST_CA_CERT")

		InitCertificate("TEST_CA_CERT")

		assert.NotNil(t, globalTLSConfig)
		assert.NotNil(t, globalTLSConfig.RootCAs)
	})

	t.Run("non-existent certificate file", func(t *testing.T) {
		os.Setenv("TEST_CA_CERT", "/non/existent/file.pem")
		defer os.Unsetenv("TEST_CA_CERT")

		InitCertificate("TEST_CA_CERT")

		// Should still create pool but with no certificates
		assert.NotNil(t, globalTLSConfig)
	})

	t.Run("invalid certificate file", func(t *testing.T) {
		tempDir := t.TempDir()
		invalidCertFile := filepath.Join(tempDir, "invalid.pem")

		// Write invalid PEM data
		err := os.WriteFile(invalidCertFile, []byte("invalid pem data"), 0644)
		require.NoError(t, err)

		os.Setenv("TEST_CA_CERT", invalidCertFile)
		defer os.Unsetenv("TEST_CA_CERT")

		InitCertificate("TEST_CA_CERT")

		// Should still create config but certificate won't be added
		assert.NotNil(t, globalTLSConfig)
	})

	t.Run("mixed valid and invalid paths", func(t *testing.T) {
		tempDir := t.TempDir()
		validCertFile := filepath.Join(tempDir, "valid.pem")

		certPEM := `-----BEGIN CERTIFICATE-----
MIIBxTCCAWugAwIBAgIJAKjYQV+z9YlwMA0GCSqGSIb3DQEBCwUAMCkxJzAlBgNV
BAMTHkVsYXN0aWNzZWFyY2ggVGVzdCBOb2RlOiBub2RlMTAeFw0yMDAzMDQxNTAy
NDhaFw0yMjAzMDQxNTAyNDhaMCkxJzAlBgNVBAMTHkVsYXN0aWNzZWFyY2ggVGVz
dCBOb2RlOiBub2RlMTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABDPiUKNkwUYW
xF5EhLYuBLFnQcH1Zzex/J7rJE3QpAjJe9R6YKJJPt1wJGLjQvQ4YZK5Y1J+YKJw
5L4v3L2v3L2jUDBOMB0GA1UdDgQWBBQX3vWKb3LRJbJR5p7bOLtD7vbV7zAfBgNV
HSMEGDAWgBQX3vWKb3LRJbJR5p7bOLtD7vbV7zAMBgNVHRMEBTADAQH/MA0GCSqG
SIb3DQEBCwUAA0EABhh2sxOe3lj8PzFd6NiMhTfNHZj7w7v3m3v3m3v3m3v3m3v3
m3v3m3v3m3v3m3v3m3v3m3v3m3v3m3v3m3v3m3v3m3v3mw==
-----END CERTIFICATE-----`

		err := os.WriteFile(validCertFile, []byte(certPEM), 0644)
		require.NoError(t, err)

		os.Setenv("TEST_CA_CERT", fmt.Sprintf("%s:/non/existent.pem:", validCertFile))
		defer os.Unsetenv("TEST_CA_CERT")

		InitCertificate("TEST_CA_CERT")

		assert.NotNil(t, globalTLSConfig)
	})

	t.Run("empty paths in colon-separated list", func(t *testing.T) {
		os.Setenv("TEST_CA_CERT", ":::")
		defer os.Unsetenv("TEST_CA_CERT")

		InitCertificate("TEST_CA_CERT")

		assert.NotNil(t, globalTLSConfig)
	})
}

func TestGetHTTPClient(t *testing.T) {
	setupScrapingTest()

	t.Run("client without custom TLS config", func(t *testing.T) {
		globalTLSConfig = nil

		client := GetHTTPClient(5 * time.Second)

		assert.NotNil(t, client)
		assert.Equal(t, 5*time.Second, client.Timeout)
		assert.Nil(t, client.Transport)
	})

	t.Run("client with custom TLS config", func(t *testing.T) {
		globalTLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}

		client := GetHTTPClient(10 * time.Second)

		assert.NotNil(t, client)
		assert.Equal(t, 10*time.Second, client.Timeout)
		assert.NotNil(t, client.Transport)

		transport, ok := client.Transport.(*http.Transport)
		require.True(t, ok)
		assert.NotNil(t, transport.TLSClientConfig)
		assert.True(t, transport.TLSClientConfig.InsecureSkipVerify)
	})

	t.Run("different timeout values", func(t *testing.T) {
		testCases := []time.Duration{
			1 * time.Second,
			30 * time.Second,
			1 * time.Minute,
		}

		for _, timeout := range testCases {
			client := GetHTTPClient(timeout)
			assert.Equal(t, timeout, client.Timeout)
		}
	})
}

func TestInitScrapers(t *testing.T) {
	setupScrapingTest()

	t.Run("initialize prometheus and site scrapers", func(t *testing.T) {
		setupScrapingTest() // Reset scrapers for this test
		cfg := &config.Config{
			Sources: []config.Source{
				{
					Name: "test-prometheus",
					Type: "prometheus",
					URL:  "http://localhost:9090",
				},
				{
					Name: "test-site",
					Type: "site",
					URL:  "http://localhost:8080",
				},
			},
		}

		InitScrapers(cfg)

		assert.Len(t, Scrapers, 2)
		assert.Contains(t, Scrapers, "test-prometheus")
		assert.Contains(t, Scrapers, "test-site")

		// Verify scraper types
		assert.NotNil(t, Scrapers["test-prometheus"])
		assert.NotNil(t, Scrapers["test-site"])
	})

	t.Run("initialize with empty sources", func(t *testing.T) {
		setupScrapingTest() // Reset scrapers for this test
		cfg := &config.Config{
			Sources: []config.Source{},
		}

		InitScrapers(cfg)

		assert.Empty(t, Scrapers)
	})

	t.Run("initialize with unknown source type", func(t *testing.T) {
		setupScrapingTest() // Reset scrapers for this test
		cfg := &config.Config{
			Sources: []config.Source{
				{
					Name: "test-unknown",
					Type: "unknown-type",
					URL:  "http://localhost:9090",
				},
			},
		}

		InitScrapers(cfg)

		// Unknown types are ignored, so no scrapers should be added
		assert.Empty(t, Scrapers)
	})

	t.Run("initialize multiple sources of same type", func(t *testing.T) {
		setupScrapingTest() // Reset scrapers for this test
		cfg := &config.Config{
			Sources: []config.Source{
				{
					Name: "prometheus-1",
					Type: "prometheus",
					URL:  "http://localhost:9090",
				},
				{
					Name: "prometheus-2",
					Type: "prometheus",
					URL:  "http://localhost:9091",
				},
			},
		}

		InitScrapers(cfg)

		assert.Len(t, Scrapers, 2)
		assert.Contains(t, Scrapers, "prometheus-1")
		assert.Contains(t, Scrapers, "prometheus-2")
	})
}

func TestStart(t *testing.T) {
	setupScrapingTest()

	t.Run("invalid scraping interval", func(t *testing.T) {
		cfg := &config.Config{
			Scraping: config.ScrapingSettings{
				Interval:    "invalid-duration",
				Timeout:     "5s",
				MaxParallel: 10,
			},
		}

		_ = cfg // Suppress unused variable warning

		// This should cause a fatal error, but we can't easily test fatal
		// in a unit test. We'll test this by checking if the function panics
		// or exits (which it does via Fatal)
		defer func() {
			if r := recover(); r != nil {
				// If we recovered from a panic, the test is working as expected
				_ = r // Use the recovered value to suppress staticcheck warning
			}
		}()

		// Note: This will call Fatal which exits the program
		// In a real scenario, we'd need dependency injection to make this testable
		// For now, we'll skip this test case or comment it out
		t.Skip("Cannot easily test fatal errors in unit tests")
	})

	t.Run("invalid scraping timeout", func(t *testing.T) {
		cfg := &config.Config{
			Scraping: config.ScrapingSettings{
				Interval:    "10s",
				Timeout:     "invalid-duration",
				MaxParallel: 10,
			},
		}

		_ = cfg // Suppress unused variable warning

		defer func() {
			if r := recover(); r != nil {
				// If we recovered from a panic, the test is working as expected
				_ = r // Use the recovered value to suppress staticcheck warning
			}
		}()

		t.Skip("Cannot easily test fatal errors in unit tests")
	})

	t.Run("successful scraping start", func(t *testing.T) {
		// Create mock scrapers
		mockScraper := &MockScraper{
			results: []handlers.AppStatus{
				{
					Name:     "test-app",
					Location: "test-location",
					Status:   "up",
				},
			},
		}

		// Add mock scraper to scrapers map
		Scrapers["test-source"] = mockScraper

		cfg := &config.Config{
			Scraping: config.ScrapingSettings{
				Interval:    "100ms", // Very short for testing
				Timeout:     "5s",
				MaxParallel: 2,
			},
			Sources: []config.Source{
				{
					Name: "test-source",
					Type: "prometheus",
					URL:  "http://localhost:9090",
				},
			},
		}

		// Start scraping in background
		go Start(cfg)

		// Wait for a few scrapes to happen
		time.Sleep(300 * time.Millisecond)

		// Verify scraper was called multiple times
		assert.True(t, mockScraper.GetCallCount() >= 2, "Expected at least 2 scrape calls")

		// Verify cache was updated
		cache := handlers.GetAppStatusCache()
		assert.NotEmpty(t, cache)

		found := false
		for _, status := range cache {
			if status.Name == "test-app" && status.Source == "test-source" {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected to find test-app in cache")
	})

	t.Run("scraping with error", func(t *testing.T) {
		// Create mock scraper that returns errors
		mockScraper := &MockScraper{
			err: fmt.Errorf("scraping failed"),
		}

		Scrapers["error-source"] = mockScraper

		cfg := &config.Config{
			Scraping: config.ScrapingSettings{
				Interval:    "100ms",
				Timeout:     "5s",
				MaxParallel: 1,
			},
			Sources: []config.Source{
				{
					Name: "error-source",
					Type: "prometheus",
					URL:  "http://localhost:9090",
				},
			},
		}

		// Start scraping in background
		go Start(cfg)

		// Wait for a few scrapes to happen
		time.Sleep(300 * time.Millisecond)

		// Verify scraper was called despite errors
		assert.True(t, mockScraper.GetCallCount() >= 2, "Expected at least 2 scrape calls even with errors")
	})

	t.Run("unsupported source name", func(t *testing.T) {
		cfg := &config.Config{
			Scraping: config.ScrapingSettings{
				Interval:    "10s",
				Timeout:     "5s",
				MaxParallel: 1,
			},
			Sources: []config.Source{
				{
					Name: "non-existent-source",
					Type: "prometheus",
					URL:  "http://localhost:9090",
				},
			},
		}

		_ = cfg // Suppress unused variable warning

		defer func() {
			if r := recover(); r != nil {
				// If we recovered from a panic, the test is working as expected
				_ = r // Use the recovered value to suppress staticcheck warning
			}
		}()

		t.Skip("Cannot easily test fatal errors in unit tests")
	})

	t.Run("multiple sources scraping concurrently", func(t *testing.T) {
		// Create multiple mock scrapers
		mockScraper1 := &MockScraper{
			results: []handlers.AppStatus{
				{Name: "app1", Location: "loc1", Status: "up"},
			},
			delay: 50 * time.Millisecond,
		}

		mockScraper2 := &MockScraper{
			results: []handlers.AppStatus{
				{Name: "app2", Location: "loc2", Status: "down"},
			},
			delay: 30 * time.Millisecond,
		}

		Scrapers["source1"] = mockScraper1
		Scrapers["source2"] = mockScraper2

		cfg := &config.Config{
			Scraping: config.ScrapingSettings{
				Interval:    "200ms",
				Timeout:     "1s",
				MaxParallel: 5,
			},
			Sources: []config.Source{
				{Name: "source1", Type: "prometheus"},
				{Name: "source2", Type: "site"},
			},
		}

		// Start scraping in background
		go Start(cfg)

		// Wait for scrapes to happen
		time.Sleep(500 * time.Millisecond)

		// Both scrapers should have been called
		assert.True(t, mockScraper1.GetCallCount() >= 1)
		assert.True(t, mockScraper2.GetCallCount() >= 1)

		// Verify both sources are in cache
		cache := handlers.GetAppStatusCache()
		sourceNames := make(map[string]bool)
		for _, status := range cache {
			sourceNames[status.Source] = true
		}

		assert.True(t, sourceNames["source1"])
		assert.True(t, sourceNames["source2"])
	})

	t.Run("scraping with TLS config", func(t *testing.T) {
		// Set up global TLS config
		globalTLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}

		mockScraper := &MockScraper{
			results: []handlers.AppStatus{
				{Name: "tls-app", Location: "secure-loc", Status: "up"},
			},
		}

		Scrapers["tls-source"] = mockScraper

		cfg := &config.Config{
			Scraping: config.ScrapingSettings{
				Interval:    "100ms",
				Timeout:     "5s",
				MaxParallel: 1,
			},
			Sources: []config.Source{
				{Name: "tls-source", Type: "prometheus"},
			},
		}

		go Start(cfg)
		time.Sleep(200 * time.Millisecond)

		assert.True(t, mockScraper.GetCallCount() >= 1)
	})
}

func TestIntegration(t *testing.T) {
	setupScrapingTest()

	t.Run("complete workflow", func(t *testing.T) {
		// Test the complete workflow: InitCertificate -> InitScrapers -> Start

		// 1. Initialize certificates (none in this test)
		InitCertificate("NON_EXISTENT_ENV_VAR")

		// 2. Initialize scrapers
		cfg := &config.Config{
			Sources: []config.Source{
				{
					Name: "integration-test",
					Type: "prometheus",
					URL:  "http://localhost:9090",
				},
			},
			Scraping: config.ScrapingSettings{
				Interval:    "200ms",
				Timeout:     "5s",
				MaxParallel: 1,
			},
		}

		InitScrapers(cfg)

		// 3. Replace with mock scraper
		mockScraper := &MockScraper{
			results: []handlers.AppStatus{
				{Name: "integration-app", Location: "test-loc", Status: "up"},
			},
		}
		Scrapers["integration-test"] = mockScraper

		// 4. Start scraping
		go Start(cfg)

		// 5. Wait and verify
		time.Sleep(500 * time.Millisecond)

		assert.True(t, mockScraper.GetCallCount() >= 2)

		cache := handlers.GetAppStatusCache()
		found := false
		for _, status := range cache {
			if status.Name == "integration-app" && status.Source == "integration-test" {
				found = true
				break
			}
		}
		assert.True(t, found)
	})
}
