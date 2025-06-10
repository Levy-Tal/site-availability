package scraping

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"site-availability/config"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetScrapeTimeout(t *testing.T) {
	// Test setting timeout
	timeout := 5 * time.Second
	SetScrapeTimeout(timeout)
	assert.Equal(t, timeout, defaultScrapeTimeout)
	assert.NotNil(t, httpClient)
	assert.Equal(t, timeout, httpClient.Timeout)

	// Test updating timeout
	newTimeout := 10 * time.Second
	SetScrapeTimeout(newTimeout)
	assert.Equal(t, newTimeout, defaultScrapeTimeout)
	assert.Equal(t, newTimeout, httpClient.Timeout)
}

func TestInitCertificate(t *testing.T) {
	// Create temporary CA certificate
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "ca.crt")
	err := os.WriteFile(certPath, []byte("test certificate"), 0644)
	require.NoError(t, err)

	// Test with valid certificate
	os.Setenv("TEST_CA_PATH", certPath)
	defer os.Unsetenv("TEST_CA_PATH")

	InitCertificate("TEST_CA_PATH")
	assert.NotNil(t, httpClient)
	assert.NotNil(t, httpClient.Transport)

	// Test with empty environment variable
	os.Unsetenv("TEST_CA_PATH")
	InitCertificate("TEST_CA_PATH")
	assert.NotNil(t, httpClient)

	// Test with invalid certificate path
	os.Setenv("TEST_CA_PATH", "nonexistent.crt")
	InitCertificate("TEST_CA_PATH")
	assert.NotNil(t, httpClient)
}

func TestPrometheusMetricChecker_Check(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check authentication
		if auth := r.Header.Get("Authorization"); auth != "" {
			if auth != "Bearer test-token" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}

		// Return test response
		response := PrometheusResponse{
			Status: "success",
			Data: struct {
				ResultType string `json:"resultType"`
				Result     []struct {
					Metric map[string]string `json:"metric"`
					Value  []interface{}     `json:"value"`
				} `json:"result"`
			}{
				ResultType: "vector",
				Result: []struct {
					Metric map[string]string `json:"metric"`
					Value  []interface{}     `json:"value"`
				}{
					{
						Metric: map[string]string{"instance": "test"},
						Value:  []interface{}{float64(time.Now().Unix()), "1"},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer server.Close()

	// Test cases
	tests := []struct {
		name           string
		prometheusURL  string
		promQLQuery    string
		servers        []config.PrometheusServer
		expectedStatus int
		expectError    bool
	}{
		{
			name:          "Successful query",
			prometheusURL: server.URL,
			promQLQuery:   "up{instance='test'}",
			servers: []config.PrometheusServer{
				{
					Name:  "test",
					URL:   server.URL,
					Auth:  "bearer",
					Token: "test-token",
				},
			},
			expectedStatus: 1,
			expectError:    false,
		},
		{
			name:          "Authentication failure",
			prometheusURL: server.URL,
			promQLQuery:   "up{instance='test'}",
			servers: []config.PrometheusServer{
				{
					Name:  "test",
					URL:   server.URL,
					Auth:  "bearer",
					Token: "wrong-token",
				},
			},
			expectedStatus: 0,
			expectError:    true,
		},
		{
			name:           "No authentication",
			prometheusURL:  server.URL,
			promQLQuery:    "up{instance='test'}",
			servers:        []config.PrometheusServer{},
			expectedStatus: 1,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := &PrometheusMetricChecker{
				PrometheusServers: tt.servers,
			}

			status, err := checker.Check(tt.prometheusURL, tt.promQLQuery, tt.servers)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, status)
			}
		})
	}
}

func TestCheckAppStatus(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := PrometheusResponse{
			Status: "success",
			Data: struct {
				ResultType string `json:"resultType"`
				Result     []struct {
					Metric map[string]string `json:"metric"`
					Value  []interface{}     `json:"value"`
				} `json:"result"`
			}{
				ResultType: "vector",
				Result: []struct {
					Metric map[string]string `json:"metric"`
					Value  []interface{}     `json:"value"`
				}{
					{
						Metric: map[string]string{"instance": "test"},
						Value:  []interface{}{float64(time.Now().Unix()), "1"},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer server.Close()

	// Test cases
	tests := []struct {
		name           string
		app            config.Application
		servers        []config.PrometheusServer
		expectedStatus string
	}{
		{
			name: "Valid app with matching server",
			app: config.Application{
				Name:       "test-app",
				Location:   "test-location",
				Metric:     "up{instance='test'}",
				Prometheus: "test-server",
			},
			servers: []config.PrometheusServer{
				{
					Name: "test-server",
					URL:  server.URL,
				},
			},
			expectedStatus: "up",
		},
		{
			name: "App with non-existent server",
			app: config.Application{
				Name:       "test-app",
				Location:   "test-location",
				Metric:     "up{instance='test'}",
				Prometheus: "nonexistent",
			},
			servers: []config.PrometheusServer{
				{
					Name: "test-server",
					URL:  server.URL,
				},
			},
			expectedStatus: "unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := &PrometheusMetricChecker{
				PrometheusServers: tt.servers,
			}

			status := CheckAppStatus(tt.app, tt.servers, checker)
			assert.Equal(t, tt.app.Name, status.Name)
			assert.Equal(t, tt.app.Location, status.Location)
			assert.Equal(t, tt.expectedStatus, status.Status)
		})
	}
}

func TestParallelScrapeAppStatuses(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := PrometheusResponse{
			Status: "success",
			Data: struct {
				ResultType string `json:"resultType"`
				Result     []struct {
					Metric map[string]string `json:"metric"`
					Value  []interface{}     `json:"value"`
				} `json:"result"`
			}{
				ResultType: "vector",
				Result: []struct {
					Metric map[string]string `json:"metric"`
					Value  []interface{}     `json:"value"`
				}{
					{
						Metric: map[string]string{"instance": "test"},
						Value:  []interface{}{float64(time.Now().Unix()), "1"},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer server.Close()

	// Test data
	apps := []config.Application{
		{
			Name:       "app1",
			Location:   "loc1",
			Metric:     "up{instance='test'}",
			Prometheus: "test-server",
		},
		{
			Name:       "app2",
			Location:   "loc2",
			Metric:     "up{instance='test'}",
			Prometheus: "test-server",
		},
	}

	servers := []config.PrometheusServer{
		{
			Name: "test-server",
			URL:  server.URL,
		},
	}

	checker := &PrometheusMetricChecker{
		PrometheusServers: servers,
	}

	// Test with different maxParallelScrapes values
	maxParallelScrapes := 2
	results := ParallelScrapeAppStatuses(apps, servers, checker, maxParallelScrapes)

	// Verify results
	require.Len(t, results, 2)
	assert.Equal(t, "app1", results[0].Name)
	assert.Equal(t, "loc1", results[0].Location)
	assert.Equal(t, "up", results[0].Status)
	assert.Equal(t, "app2", results[1].Name)
	assert.Equal(t, "loc2", results[1].Location)
	assert.Equal(t, "up", results[1].Status)
}

func TestPrometheusMetricChecker_Check_ErrorCases(t *testing.T) {
	// Create test server that returns error responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return error response
		response := PrometheusResponse{
			Status: "error",
			Data: struct {
				ResultType string `json:"resultType"`
				Result     []struct {
					Metric map[string]string `json:"metric"`
					Value  []interface{}     `json:"value"`
				} `json:"result"`
			}{},
		}

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer server.Close()

	checker := &PrometheusMetricChecker{}

	// Test error cases
	_, err := checker.Check(server.URL, "invalid query", nil)
	assert.Error(t, err)

	// Test with empty result
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := PrometheusResponse{
			Status: "success",
			Data: struct {
				ResultType string `json:"resultType"`
				Result     []struct {
					Metric map[string]string `json:"metric"`
					Value  []interface{}     `json:"value"`
				} `json:"result"`
			}{
				ResultType: "vector",
				Result: []struct {
					Metric map[string]string `json:"metric"`
					Value  []interface{}     `json:"value"`
				}{},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer server.Close()

	_, err = checker.Check(server.URL, "empty result query", nil)
	assert.Error(t, err)
}
