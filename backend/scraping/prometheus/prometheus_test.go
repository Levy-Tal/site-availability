package prometheus

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"site-availability/config"
	"site-availability/logging"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupPrometheusTest initializes logging for tests
func setupPrometheusTest() {
	_ = logging.Init()
}

func TestMain(m *testing.M) {
	setupPrometheusTest()
	m.Run()
}

func TestNewPrometheusScraper(t *testing.T) {
	t.Run("create new prometheus scraper", func(t *testing.T) {
		scraper := NewPrometheusScraper()
		assert.NotNil(t, scraper)
		assert.IsType(t, &PrometheusScraper{}, scraper)
	})
}

func TestPrometheusScraper_Scrape(t *testing.T) {
	setupPrometheusTest()

	t.Run("successful scrape with up status", func(t *testing.T) {
		// Create mock Prometheus server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.URL.Path, "/api/v1/query")
			assert.Contains(t, r.URL.RawQuery, "query=")

			// Return successful response with value 1 (up)
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
							Value:  []interface{}{1234567890.0, "1"},
						},
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		scraper := NewPrometheusScraper()
		source := config.Source{
			Name: "test-prometheus",
			Type: "prometheus",
			URL:  server.URL,
			Apps: []config.App{
				{
					Name:     "test-app",
					Location: "test-location",
					Metric:   `up{instance="test"}`,
				},
			},
		}

		statuses, _, err := scraper.Scrape(source, config.ServerSettings{}, 5*time.Second, 1, nil)
		require.NoError(t, err)
		require.Len(t, statuses, 1)

		assert.Equal(t, "test-app", statuses[0].Name)
		assert.Equal(t, "test-location", statuses[0].Location)
		assert.Equal(t, "up", statuses[0].Status)
		assert.Equal(t, "test-prometheus", statuses[0].Source)
	})

	t.Run("successful scrape with down status", func(t *testing.T) {
		// Create mock Prometheus server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Return successful response with value 0 (down)
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
							Value:  []interface{}{1234567890.0, "0"},
						},
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		scraper := NewPrometheusScraper()
		source := config.Source{
			Name: "test-prometheus",
			URL:  server.URL,
			Apps: []config.App{
				{
					Name:     "test-app",
					Location: "test-location",
					Metric:   `up{instance="test"}`,
				},
			},
		}

		statuses, _, err := scraper.Scrape(source, config.ServerSettings{}, 5*time.Second, 1, nil)
		require.NoError(t, err)
		require.Len(t, statuses, 1)
		assert.Equal(t, "down", statuses[0].Status)
	})

	t.Run("scrape with unavailable status due to error", func(t *testing.T) {
		// Create mock server that returns errors
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		scraper := NewPrometheusScraper()
		source := config.Source{
			Name: "test-prometheus",
			URL:  server.URL,
			Apps: []config.App{
				{
					Name:     "test-app",
					Location: "test-location",
					Metric:   `up{instance="test"}`,
				},
			},
		}

		statuses, _, err := scraper.Scrape(source, config.ServerSettings{}, 5*time.Second, 1, nil)
		require.NoError(t, err) // Scrape method always returns success
		require.Len(t, statuses, 1)
		assert.Equal(t, "unavailable", statuses[0].Status)
	})

	t.Run("scrape multiple apps concurrently", func(t *testing.T) {
		// Create mock server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate different responses based on query
			query := r.URL.Query().Get("query")
			value := "1"
			if query != "" && query != `up{instance="test1"}` {
				value = "0"
			}

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
							Value:  []interface{}{1234567890.0, value},
						},
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		scraper := NewPrometheusScraper()
		source := config.Source{
			Name: "test-prometheus",
			URL:  server.URL,
			Apps: []config.App{
				{
					Name:     "app1",
					Location: "location1",
					Metric:   `up{instance="test1"}`,
				},
				{
					Name:     "app2",
					Location: "location2",
					Metric:   `up{instance="test2"}`,
				},
			},
		}

		statuses, _, err := scraper.Scrape(source, config.ServerSettings{}, 5*time.Second, 2, nil)
		require.NoError(t, err)
		require.Len(t, statuses, 2)

		// Verify both apps were processed
		assert.Contains(t, []string{"app1", "app2"}, statuses[0].Name)
		assert.Contains(t, []string{"app1", "app2"}, statuses[1].Name)
	})

	t.Run("scrape with TLS config", func(t *testing.T) {
		// Create HTTPS mock server
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
							Value:  []interface{}{1234567890.0, "1"},
						},
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		scraper := NewPrometheusScraper()
		source := config.Source{
			Name: "test-prometheus",
			URL:  server.URL,
			Apps: []config.App{
				{
					Name:     "test-app",
					Location: "test-location",
					Metric:   `up{instance="test"}`,
				},
			},
		}

		// Use insecure TLS config for testing
		tlsConfig := &tls.Config{InsecureSkipVerify: true}

		statuses, _, err := scraper.Scrape(source, config.ServerSettings{}, 5*time.Second, 1, tlsConfig)
		require.NoError(t, err)
		require.Len(t, statuses, 1)
		assert.Equal(t, "up", statuses[0].Status)
	})

	t.Run("scrape with empty apps", func(t *testing.T) {
		scraper := NewPrometheusScraper()
		source := config.Source{
			Name: "test-prometheus",
			URL:  "http://localhost:9090",
			Apps: []config.App{}, // Empty apps
		}

		statuses, _, err := scraper.Scrape(source, config.ServerSettings{}, 5*time.Second, 1, nil)
		require.NoError(t, err)
		assert.Empty(t, statuses)
	})

	t.Run("scrape with label merging", func(t *testing.T) {
		// Create mock Prometheus server
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
							Value:  []interface{}{1234567890.0, "1"},
						},
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		scraper := NewPrometheusScraper()

		// Set up server settings with labels
		serverSettings := config.ServerSettings{
			Labels: map[string]string{
				"server_env":    "test",
				"server_region": "us-west",
				"common_label":  "server_value", // Will be overridden by source
			},
		}

		source := config.Source{
			Name: "test-prometheus",
			URL:  server.URL,
			Labels: map[string]string{
				"source_type":  "prometheus",
				"source_team":  "platform",
				"common_label": "source_value", // Will be overridden by app
			},
			Apps: []config.App{
				{
					Name:     "labeled-app",
					Location: "test-location",
					Metric:   `up{instance="test"}`,
					Labels: map[string]string{
						"app_version":  "v1.2.3",
						"app_tier":     "frontend",
						"common_label": "app_value", // Highest priority
					},
				},
			},
		}

		statuses, _, err := scraper.Scrape(source, serverSettings, 5*time.Second, 1, nil)
		require.NoError(t, err)
		require.Len(t, statuses, 1)

		app := statuses[0]
		assert.Equal(t, "labeled-app", app.Name)
		assert.Equal(t, "up", app.Status)

		// Verify label merging with correct priority (App > Source > Server)
		expectedLabels := map[string]string{
			// Only app labels should be present (merging happens in UpdateAppStatus)
			"app_version":  "v1.2.3",
			"app_tier":     "frontend",
			"common_label": "app_value",
		}

		assert.Equal(t, expectedLabels, app.Labels)
		assert.Len(t, app.Labels, 3) // Only app labels
	})
}

func TestPrometheusScraper_Check(t *testing.T) {
	setupPrometheusTest()

	t.Run("successful query with up status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify the URL format
			assert.Contains(t, r.URL.Path, "/api/v1/query")
			assert.Contains(t, r.URL.RawQuery, "query=")

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
							Value:  []interface{}{1234567890.0, "1"},
						},
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		scraper := NewPrometheusScraper()
		client := &http.Client{Timeout: 5 * time.Second}

		result, err := scraper.check(client, server.URL, `up{instance="test"}`, "", "")
		require.NoError(t, err)
		assert.Equal(t, 1, result)
	})

	t.Run("successful query with down status", func(t *testing.T) {
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
							Value:  []interface{}{1234567890.0, "0"},
						},
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		scraper := NewPrometheusScraper()
		client := &http.Client{Timeout: 5 * time.Second}

		result, err := scraper.check(client, server.URL, `up{instance="test"}`, "", "")
		require.NoError(t, err)
		assert.Equal(t, 0, result)
	})

	t.Run("authentication with bearer token", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify bearer token
			auth := r.Header.Get("Authorization")
			assert.Equal(t, "Bearer test-token", auth)

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
							Value:  []interface{}{1234567890.0, "1"},
						},
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		scraper := NewPrometheusScraper()
		client := &http.Client{Timeout: 5 * time.Second}

		result, err := scraper.check(client, server.URL, `up{instance="test"}`, "bearer", "test-token")
		require.NoError(t, err)
		assert.Equal(t, 1, result)
	})

	t.Run("authentication with basic token", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify basic auth header
			auth := r.Header.Get("Authorization")
			assert.Equal(t, "Basic test-token", auth)

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
							Value:  []interface{}{1234567890.0, "1"},
						},
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		scraper := NewPrometheusScraper()
		client := &http.Client{Timeout: 5 * time.Second}

		result, err := scraper.check(client, server.URL, `up{instance="test"}`, "basic", "test-token")
		require.NoError(t, err)
		assert.Equal(t, 1, result)
	})

	t.Run("authentication with empty auth but token provided", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Should not have authorization header
			auth := r.Header.Get("Authorization")
			assert.Empty(t, auth)

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
							Value:  []interface{}{1234567890.0, "1"},
						},
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		scraper := NewPrometheusScraper()
		client := &http.Client{Timeout: 5 * time.Second}

		result, err := scraper.check(client, server.URL, `up{instance="test"}`, "", "test-token")
		require.NoError(t, err)
		assert.Equal(t, 1, result)
	})

	t.Run("invalid request URL", func(t *testing.T) {
		scraper := NewPrometheusScraper()
		client := &http.Client{Timeout: 5 * time.Second}

		// Use invalid URL with spaces to trigger request creation error
		_, err := scraper.check(client, "http://invalid url with spaces", `up{instance="test"}`, "", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create request")
	})

	t.Run("connection error", func(t *testing.T) {
		scraper := NewPrometheusScraper()
		client := &http.Client{Timeout: 5 * time.Second}

		// Use non-existent server
		_, err := scraper.check(client, "http://localhost:99999", `up{instance="test"}`, "", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to query Prometheus")
	})

	t.Run("authentication failure - 401 response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		scraper := NewPrometheusScraper()
		client := &http.Client{Timeout: 5 * time.Second}

		_, err := scraper.check(client, server.URL, `up{instance="test"}`, "bearer", "invalid-token")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "authentication failed")
		assert.Contains(t, err.Error(), "bearer")
	})

	t.Run("invalid JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte("invalid json{"))
		}))
		defer server.Close()

		scraper := NewPrometheusScraper()
		client := &http.Client{Timeout: 5 * time.Second}

		_, err := scraper.check(client, server.URL, `up{instance="test"}`, "", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode Prometheus response")
	})

	t.Run("prometheus query failure status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := PrometheusResponse{
				Status: "error",
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		scraper := NewPrometheusScraper()
		client := &http.Client{Timeout: 5 * time.Second}

		_, err := scraper.check(client, server.URL, `up{instance="test"}`, "", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "prometheus query")
		assert.Contains(t, err.Error(), "failed")
	})

	t.Run("empty results from prometheus", func(t *testing.T) {
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
					}{}, // Empty results
				},
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		scraper := NewPrometheusScraper()
		client := &http.Client{Timeout: 5 * time.Second}

		_, err := scraper.check(client, server.URL, `up{instance="test"}`, "", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "did not return any result")
	})

	t.Run("invalid value array too short", func(t *testing.T) {
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
							Value:  []interface{}{1234567890.0}, // Missing value
						},
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		scraper := NewPrometheusScraper()
		client := &http.Client{Timeout: 5 * time.Second}

		_, err := scraper.check(client, server.URL, `up{instance="test"}`, "", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "value array too short")
	})

	t.Run("invalid value type not string", func(t *testing.T) {
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
							Value:  []interface{}{1234567890.0, 123}, // Value is int, not string
						},
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		scraper := NewPrometheusScraper()
		client := &http.Client{Timeout: 5 * time.Second}

		_, err := scraper.check(client, server.URL, `up{instance="test"}`, "", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "value is not a string")
	})

	t.Run("metric value with special characters", func(t *testing.T) {
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
							Value:  []interface{}{1234567890.0, "0.5"}, // Non 0/1 value
						},
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		scraper := NewPrometheusScraper()
		client := &http.Client{Timeout: 5 * time.Second}

		result, err := scraper.check(client, server.URL, `up{instance="test"}`, "", "")
		require.NoError(t, err)
		assert.Equal(t, 0, result) // Any value != "1" should return 0
	})
}

func TestPrometheusResponse(t *testing.T) {
	t.Run("prometheus response struct", func(t *testing.T) {
		// Test that the struct can be marshaled and unmarshaled correctly
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
						Value:  []interface{}{1234567890.0, "1"},
					},
				},
			},
		}

		// Marshal to JSON
		data, err := json.Marshal(response)
		require.NoError(t, err)

		// Unmarshal back
		var decoded PrometheusResponse
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "success", decoded.Status)
		assert.Equal(t, "vector", decoded.Data.ResultType)
		assert.Len(t, decoded.Data.Result, 1)
		assert.Equal(t, "1", decoded.Data.Result[0].Value[1])
	})
}

func TestPrometheusScraper_EdgeCases(t *testing.T) {
	setupPrometheusTest()

	t.Run("concurrent scraping with rate limiting", func(t *testing.T) {
		requestCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			// Simulate some processing time
			time.Sleep(10 * time.Millisecond)

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
							Value:  []interface{}{1234567890.0, "1"},
						},
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		scraper := NewPrometheusScraper()
		source := config.Source{
			Name: "test-prometheus",
			URL:  server.URL,
			Apps: []config.App{
				{Name: "app1", Location: "loc1", Metric: "up1"},
				{Name: "app2", Location: "loc2", Metric: "up2"},
				{Name: "app3", Location: "loc3", Metric: "up3"},
				{Name: "app4", Location: "loc4", Metric: "up4"},
				{Name: "app5", Location: "loc5", Metric: "up5"},
			},
		}

		start := time.Now()
		statuses, _, err := scraper.Scrape(source, config.ServerSettings{}, 5*time.Second, 2, nil) // max 2 concurrent
		duration := time.Since(start)

		require.NoError(t, err)
		require.Len(t, statuses, 5)

		// With 2 concurrent requests max, 5 requests should take at least 30ms
		// (3 batches * 10ms per request)
		assert.Greater(t, duration, 25*time.Millisecond)
		assert.Equal(t, 5, requestCount)
	})

	t.Run("timeout handling", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate slow response
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		scraper := NewPrometheusScraper()
		source := config.Source{
			Name: "test-prometheus",
			URL:  server.URL,
			Apps: []config.App{
				{
					Name:     "test-app",
					Location: "test-location",
					Metric:   `up{instance="test"}`,
				},
			},
		}

		// Use very short timeout
		statuses, _, err := scraper.Scrape(source, config.ServerSettings{}, 10*time.Millisecond, 1, nil)
		require.NoError(t, err) // Scrape method always returns success
		require.Len(t, statuses, 1)
		assert.Equal(t, "unavailable", statuses[0].Status) // Should be unavailable due to timeout
	})
}
