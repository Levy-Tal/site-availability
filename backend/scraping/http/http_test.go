package http_source

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"site-availability/config"
	"site-availability/handlers"
	"site-availability/logging"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	_ = logging.Init() // Initialize logging for tests
}

func TestNewHTTPScraper(t *testing.T) {
	scraper := NewHTTPScraper()
	assert.NotNil(t, scraper)
}

func TestGetDefaultHTTPApp(t *testing.T) {
	defaults := getDefaultHTTPApp()

	assert.Equal(t, "GET", defaults.Method)
	assert.Equal(t, true, *defaults.FollowRedirects)
	assert.Equal(t, 10, defaults.MaxRedirects)
	assert.Equal(t, []interface{}{"2XX"}, defaults.AllowedStatusCodes)
	assert.Equal(t, []interface{}{"4XX", "5XX"}, defaults.BlockedStatusCodes)

	assert.NotNil(t, defaults.Validation)
	assert.Empty(t, defaults.Validation.Success)
	assert.Empty(t, defaults.Validation.Failure)

	assert.NotNil(t, defaults.SSLVerify)
	assert.True(t, *defaults.SSLVerify)

	assert.NotNil(t, defaults.Headers)
	assert.NotNil(t, defaults.Labels)
}

func TestBoolPtr(t *testing.T) {
	truePtr := boolPtr(true)
	falsePtr := boolPtr(false)

	assert.True(t, *truePtr)
	assert.False(t, *falsePtr)
}

func TestMergeWithDefaults(t *testing.T) {
	tests := []struct {
		name  string
		input HTTPApp
		check func(t *testing.T, result HTTPApp)
	}{
		{
			name: "empty app gets all defaults",
			input: HTTPApp{
				Name:     "test-app",
				Location: "test-location",
				URL:      "http://example.com",
			},
			check: func(t *testing.T, result HTTPApp) {
				assert.Equal(t, "test-app", result.Name)
				assert.Equal(t, "test-location", result.Location)
				assert.Equal(t, "http://example.com", result.URL)
				assert.Equal(t, "GET", result.Method)
				assert.Equal(t, true, *result.FollowRedirects)
				assert.Equal(t, 10, result.MaxRedirects)
				assert.Equal(t, []interface{}{"2XX"}, result.AllowedStatusCodes)
				assert.Equal(t, []interface{}{"4XX", "5XX"}, result.BlockedStatusCodes)
				assert.True(t, *result.SSLVerify)
			},
		},
		{
			name: "partial app config gets merged",
			input: HTTPApp{
				Name:               "test-app",
				Location:           "test-location",
				URL:                "http://example.com",
				Method:             "POST",
				AllowedStatusCodes: []interface{}{200, 201},
			},
			check: func(t *testing.T, result HTTPApp) {
				assert.Equal(t, "test-app", result.Name)
				assert.Equal(t, "test-location", result.Location)
				assert.Equal(t, "http://example.com", result.URL)
				assert.Equal(t, "POST", result.Method)
				assert.Equal(t, []interface{}{200, 201}, result.AllowedStatusCodes)
				assert.Equal(t, []interface{}{"4XX", "5XX"}, result.BlockedStatusCodes)
			},
		},
		{
			name: "custom ssl config gets merged",
			input: HTTPApp{
				Name:      "test-app",
				Location:  "test-location",
				URL:       "http://example.com",
				SSLVerify: boolPtr(false),
			},
			check: func(t *testing.T, result HTTPApp) {
				assert.Equal(t, "test-app", result.Name)
				assert.Equal(t, "test-location", result.Location)
				assert.Equal(t, "http://example.com", result.URL)
				assert.NotNil(t, result.SSLVerify)
				assert.False(t, *result.SSLVerify)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeWithDefaults(tt.input)
			tt.check(t, result)
		})
	}
}

func TestValidateConfig(t *testing.T) {
	scraper := NewHTTPScraper()

	tests := []struct {
		name      string
		source    config.Source
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid config",
			source: config.Source{
				Name: "test-http",
				Type: "http",
				Config: map[string]interface{}{
					"apps": []map[string]interface{}{
						{
							"name":     "test-app",
							"location": "test-location",
							"url":      "http://example.com",
						},
					},
				},
			},
			expectErr: false,
		},
		{
			name: "no apps",
			source: config.Source{
				Name: "test-http",
				Type: "http",
				Config: map[string]interface{}{
					"apps": []map[string]interface{}{},
				},
			},
			expectErr: true,
			errMsg:    "at least one app is required",
		},
		{
			name: "missing app name",
			source: config.Source{
				Name: "test-http",
				Type: "http",
				Config: map[string]interface{}{
					"apps": []map[string]interface{}{
						{
							"location": "test-location",
							"url":      "http://example.com",
						},
					},
				},
			},
			expectErr: true,
			errMsg:    "app name is required",
		},
		{
			name: "missing url",
			source: config.Source{
				Name: "test-http",
				Type: "http",
				Config: map[string]interface{}{
					"apps": []map[string]interface{}{
						{
							"name":     "test-app",
							"location": "test-location",
						},
					},
				},
			},
			expectErr: true,
			errMsg:    "missing 'url'",
		},
		{
			name: "missing location",
			source: config.Source{
				Name: "test-http",
				Type: "http",
				Config: map[string]interface{}{
					"apps": []map[string]interface{}{
						{
							"name": "test-app",
							"url":  "http://example.com",
						},
					},
				},
			},
			expectErr: true,
			errMsg:    "missing 'location'",
		},
		{
			name: "invalid url",
			source: config.Source{
				Name: "test-http",
				Type: "http",
				Config: map[string]interface{}{
					"apps": []map[string]interface{}{
						{
							"name":     "test-app",
							"location": "test-location",
							"url":      "://invalid-url",
						},
					},
				},
			},
			expectErr: true,
			errMsg:    "invalid URL",
		},
		{
			name: "invalid method",
			source: config.Source{
				Name: "test-http",
				Type: "http",
				Config: map[string]interface{}{
					"apps": []map[string]interface{}{
						{
							"name":     "test-app",
							"location": "test-location",
							"url":      "http://example.com",
							"method":   "INVALID",
						},
					},
				},
			},
			expectErr: true,
			errMsg:    "invalid method",
		},
		{
			name: "duplicate app names",
			source: config.Source{
				Name: "test-http",
				Type: "http",
				Config: map[string]interface{}{
					"apps": []map[string]interface{}{
						{
							"name":     "test-app",
							"location": "test-location",
							"url":      "http://example.com",
						},
						{
							"name":     "test-app",
							"location": "test-location2",
							"url":      "http://example2.com",
						},
					},
				},
			},
			expectErr: true,
			errMsg:    "duplicate app name",
		},
		{
			name: "invalid status codes",
			source: config.Source{
				Name: "test-http",
				Type: "http",
				Config: map[string]interface{}{
					"apps": []map[string]interface{}{
						{
							"name":                 "test-app",
							"location":             "test-location",
							"url":                  "http://example.com",
							"allowed_status_codes": []interface{}{700}, // Invalid status code
						},
					},
				},
			},
			expectErr: true,
			errMsg:    "invalid status code",
		},
		{
			name: "invalid auth config",
			source: config.Source{
				Name: "test-http",
				Type: "http",
				Config: map[string]interface{}{
					"apps": []map[string]interface{}{
						{
							"name":     "test-app",
							"location": "test-location",
							"url":      "http://example.com",
							"auth": map[string]interface{}{
								"type": "invalid-type",
							},
						},
					},
				},
			},
			expectErr: true,
			errMsg:    "invalid type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := scraper.ValidateConfig(tt.source)

			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateStatusCodes(t *testing.T) {
	tests := []struct {
		name      string
		codes     []interface{}
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid individual codes",
			codes:     []interface{}{200, 201, 204},
			expectErr: false,
		},
		{
			name:      "valid range codes",
			codes:     []interface{}{"2XX", "3XX"},
			expectErr: false,
		},
		{
			name:      "mixed valid codes",
			codes:     []interface{}{200, "3XX", 404},
			expectErr: false,
		},
		{
			name:      "invalid low status code",
			codes:     []interface{}{99},
			expectErr: true,
			errMsg:    "invalid status code 99",
		},
		{
			name:      "invalid high status code",
			codes:     []interface{}{600},
			expectErr: true,
			errMsg:    "invalid status code 600",
		},
		{
			name:      "invalid range format",
			codes:     []interface{}{"6XX"},
			expectErr: true,
			errMsg:    "invalid status code range",
		},
		{
			name:      "invalid type",
			codes:     []interface{}{12.5},
			expectErr: true,
			errMsg:    "invalid status code type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStatusCodes(tt.codes, "test_field")

			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsValidStatusCodeRange(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"1XX", true},
		{"2XX", true},
		{"3XX", true},
		{"4XX", true},
		{"5XX", true},
		{"6XX", false},
		{"0XX", false},
		{"2xx", false},
		{"XX", false},
		{"200", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isValidStatusCodeRange(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateAuth(t *testing.T) {
	tests := []struct {
		name      string
		auth      *HTTPAuth
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid basic auth",
			auth: &HTTPAuth{
				Type:     "basic",
				Username: "user",
				Password: "pass",
			},
			expectErr: false,
		},
		{
			name: "valid bearer auth",
			auth: &HTTPAuth{
				Type:  "bearer",
				Token: "token123",
			},
			expectErr: false,
		},
		{
			name: "valid digest auth",
			auth: &HTTPAuth{
				Type:     "digest",
				Username: "user",
				Password: "pass",
			},
			expectErr: false,
		},
		{
			name: "valid oauth2 auth",
			auth: &HTTPAuth{
				Type:  "oauth2",
				Token: "token123",
			},
			expectErr: false,
		},
		{
			name: "invalid auth type",
			auth: &HTTPAuth{
				Type: "invalid",
			},
			expectErr: true,
			errMsg:    "invalid type",
		},
		{
			name: "basic auth missing username",
			auth: &HTTPAuth{
				Type:     "basic",
				Password: "pass",
			},
			expectErr: true,
			errMsg:    "requires username and password",
		},
		{
			name: "basic auth missing password",
			auth: &HTTPAuth{
				Type:     "basic",
				Username: "user",
			},
			expectErr: true,
			errMsg:    "requires username and password",
		},
		{
			name: "bearer auth missing token",
			auth: &HTTPAuth{
				Type: "bearer",
			},
			expectErr: true,
			errMsg:    "requires token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAuth(tt.auth)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestScrape(t *testing.T) {
	scraper := NewHTTPScraper()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/success":
			w.WriteHeader(200)
			_, _ = w.Write([]byte("OK"))
		case "/error":
			w.WriteHeader(500)
			_, _ = w.Write([]byte("Internal Server Error"))
		case "/auth":
			auth := r.Header.Get("Authorization")
			if auth == "Basic dGVzdDpwYXNz" { // test:pass in base64
				w.WriteHeader(200)
				_, _ = w.Write([]byte("Authenticated"))
			} else {
				w.WriteHeader(401)
				_, _ = w.Write([]byte("Unauthorized"))
			}
		default:
			w.WriteHeader(404)
			_, _ = w.Write([]byte("Not Found"))
		}
	}))
	defer server.Close()

	tests := []struct {
		name           string
		source         config.Source
		expectedStatus []string
		expectErr      bool
	}{
		{
			name: "successful scrape",
			source: config.Source{
				Name: "test-http",
				Type: "http",
				Config: map[string]interface{}{
					"apps": []map[string]interface{}{
						{
							"name":     "success-app",
							"location": "test-location",
							"url":      server.URL + "/success",
						},
					},
				},
			},
			expectedStatus: []string{"up"},
			expectErr:      false,
		},
		{
			name: "failed scrape",
			source: config.Source{
				Name: "test-http",
				Type: "http",
				Config: map[string]interface{}{
					"apps": []map[string]interface{}{
						{
							"name":     "error-app",
							"location": "test-location",
							"url":      server.URL + "/error",
						},
					},
				},
			},
			expectedStatus: []string{"down"},
			expectErr:      false,
		},
		{
			name: "multiple apps",
			source: config.Source{
				Name: "test-http",
				Type: "http",
				Config: map[string]interface{}{
					"apps": []map[string]interface{}{
						{
							"name":     "success-app",
							"location": "test-location",
							"url":      server.URL + "/success",
						},
						{
							"name":     "error-app",
							"location": "test-location",
							"url":      server.URL + "/error",
						},
					},
				},
			},
			expectedStatus: []string{"up", "down"},
			expectErr:      false,
		},
		{
			name: "authenticated request",
			source: config.Source{
				Name: "test-http",
				Type: "http",
				Config: map[string]interface{}{
					"apps": []map[string]interface{}{
						{
							"name":     "auth-app",
							"location": "test-location",
							"url":      server.URL + "/auth",
							"auth": map[string]interface{}{
								"type":     "basic",
								"username": "test",
								"password": "pass",
							},
						},
					},
				},
			},
			expectedStatus: []string{"up"},
			expectErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statuses, locations, err := scraper.Scrape(
				tt.source,
				config.ServerSettings{},
				10*time.Second,
				10,
				nil,
			)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Nil(t, locations) // HTTP scraper doesn't return locations
				assert.Len(t, statuses, len(tt.expectedStatus))

				for i, expectedStatus := range tt.expectedStatus {
					assert.Equal(t, expectedStatus, statuses[i].Status)
					assert.Equal(t, tt.source.Name, statuses[i].Source)
				}
			}
		})
	}
}

func TestAddAuthentication(t *testing.T) {
	scraper := NewHTTPScraper()

	tests := []struct {
		name           string
		auth           *HTTPAuth
		expectedHeader string
		expectErr      bool
	}{
		{
			name: "basic auth",
			auth: &HTTPAuth{
				Type:     "basic",
				Username: "user",
				Password: "pass",
			},
			expectedHeader: "Basic dXNlcjpwYXNz", // user:pass in base64
			expectErr:      false,
		},
		{
			name: "bearer auth",
			auth: &HTTPAuth{
				Type:  "bearer",
				Token: "token123",
			},
			expectedHeader: "Bearer token123",
			expectErr:      false,
		},
		{
			name: "oauth2 auth",
			auth: &HTTPAuth{
				Type:  "oauth2",
				Token: "token123",
			},
			expectedHeader: "Bearer token123",
			expectErr:      false,
		},
		{
			name: "unsupported auth type",
			auth: &HTTPAuth{
				Type: "unsupported",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "http://example.com", nil)

			err := scraper.addAuthentication(req, tt.auth)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedHeader, req.Header.Get("Authorization"))
			}
		})
	}
}

func TestIsStatusAllowed(t *testing.T) {
	scraper := NewHTTPScraper()

	tests := []struct {
		name         string
		statusCode   int
		allowedCodes []interface{}
		expected     bool
	}{
		{
			name:         "exact match",
			statusCode:   200,
			allowedCodes: []interface{}{200, 201},
			expected:     true,
		},
		{
			name:         "range match",
			statusCode:   201,
			allowedCodes: []interface{}{"2XX"},
			expected:     true,
		},
		{
			name:         "mixed codes match",
			statusCode:   404,
			allowedCodes: []interface{}{200, "4XX"},
			expected:     true,
		},
		{
			name:         "no match",
			statusCode:   500,
			allowedCodes: []interface{}{200, "2XX"},
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scraper.isStatusAllowed(tt.statusCode, tt.allowedCodes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsStatusBlocked(t *testing.T) {
	scraper := NewHTTPScraper()

	tests := []struct {
		name         string
		statusCode   int
		blockedCodes []interface{}
		expected     bool
	}{
		{
			name:         "exact block match",
			statusCode:   500,
			blockedCodes: []interface{}{500, 502},
			expected:     true,
		},
		{
			name:         "range block match",
			statusCode:   404,
			blockedCodes: []interface{}{"4XX"},
			expected:     true,
		},
		{
			name:         "not blocked",
			statusCode:   200,
			blockedCodes: []interface{}{500, "4XX"},
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scraper.isStatusBlocked(tt.statusCode, tt.blockedCodes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsStatusCodeInRange(t *testing.T) {
	scraper := NewHTTPScraper()

	tests := []struct {
		name       string
		statusCode int
		rangeStr   string
		expected   bool
	}{
		{
			name:       "in 2XX range",
			statusCode: 200,
			rangeStr:   "2XX",
			expected:   true,
		},
		{
			name:       "in 2XX range boundary",
			statusCode: 299,
			rangeStr:   "2XX",
			expected:   true,
		},
		{
			name:       "not in 2XX range",
			statusCode: 300,
			rangeStr:   "2XX",
			expected:   false,
		},
		{
			name:       "invalid range format",
			statusCode: 200,
			rangeStr:   "2X",
			expected:   false,
		},
		{
			name:       "invalid range format long",
			statusCode: 200,
			rangeStr:   "200X",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scraper.isStatusCodeInRange(tt.statusCode, tt.rangeStr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateCondition(t *testing.T) {
	scraper := NewHTTPScraper()

	tests := []struct {
		name           string
		condition      HTTPValidationCondition
		statusCode     int
		responseTimeMS int
		bodyString     string
		expected       bool
	}{
		{
			name: "status code condition match",
			condition: HTTPValidationCondition{
				Type: "status_code",
				Text: "200",
			},
			statusCode: 200,
			expected:   true,
		},
		{
			name: "status code condition no match",
			condition: HTTPValidationCondition{
				Type: "status_code",
				Text: "404",
			},
			statusCode: 200,
			expected:   false,
		},
		{
			name: "response time condition within limit",
			condition: HTTPValidationCondition{
				Type:  "response_time",
				MaxMS: 1000,
			},
			responseTimeMS: 500,
			expected:       true,
		},
		{
			name: "response time condition exceeds limit",
			condition: HTTPValidationCondition{
				Type:  "response_time",
				MaxMS: 1000,
			},
			responseTimeMS: 1500,
			expected:       false,
		},
		{
			name: "body contains case sensitive",
			condition: HTTPValidationCondition{
				Type: "body_contains",
				Text: "SUCCESS",
			},
			bodyString: "Operation was SUCCESS",
			expected:   true,
		},
		{
			name: "body contains case insensitive",
			condition: HTTPValidationCondition{
				Type:          "body_contains",
				Text:          "success",
				CaseSensitive: boolPtr(false),
			},
			bodyString: "Operation was SUCCESS",
			expected:   true,
		},
		{
			name: "body not contains",
			condition: HTTPValidationCondition{
				Type: "body_not_contains",
				Text: "ERROR",
			},
			bodyString: "Operation was SUCCESS",
			expected:   true,
		},
		{
			name: "json path condition (simplified)",
			condition: HTTPValidationCondition{
				Type:          "json_path",
				Path:          "$.status",
				ExpectedValue: "ok",
			},
			bodyString: `{"status": "ok"}`,
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scraper.validateCondition(tt.condition, tt.statusCode, tt.responseTimeMS, tt.bodyString)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCheck(t *testing.T) {
	scraper := NewHTTPScraper()

	// Create test server with various endpoints
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/success":
			w.WriteHeader(200)
			_, _ = w.Write([]byte("OK"))
		case "/slow":
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(200)
			_, _ = w.Write([]byte("OK"))
		case "/error":
			w.WriteHeader(500)
			_, _ = w.Write([]byte("Internal Server Error"))
		case "/json":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		case "/post":
			if r.Method != "POST" {
				w.WriteHeader(405)
				return
			}
			w.WriteHeader(201)
			_, _ = w.Write([]byte("Created"))
		default:
			w.WriteHeader(404)
			_, _ = w.Write([]byte("Not Found"))
		}
	}))
	defer server.Close()

	tests := []struct {
		name         string
		app          HTTPApp
		expectStatus string
		expectErr    bool
	}{
		{
			name: "successful check",
			app: HTTPApp{
				Name:     "test-app",
				Location: "test-location",
				URL:      server.URL + "/success",
				Method:   "GET",
			},
			expectStatus: "up",
			expectErr:    false,
		},
		{
			name: "error status code",
			app: HTTPApp{
				Name:     "test-app",
				Location: "test-location",
				URL:      server.URL + "/error",
				Method:   "GET",
			},
			expectStatus: "down",
			expectErr:    true,
		},
		{
			name: "post request",
			app: HTTPApp{
				Name:               "test-app",
				Location:           "test-location",
				URL:                server.URL + "/post",
				Method:             "POST",
				Body:               "test=data",
				AllowedStatusCodes: []interface{}{201},
			},
			expectStatus: "up",
			expectErr:    false,
		},
		{
			name: "timeout check",
			app: HTTPApp{
				Name:     "test-app",
				Location: "test-location",
				URL:      server.URL + "/slow",
				Method:   "GET",
				Timeout:  "50ms",
			},
			expectStatus: "down",
			expectErr:    true,
		},
		{
			name: "invalid URL",
			app: HTTPApp{
				Name:     "test-app",
				Location: "test-location",
				URL:      "://invalid-url",
				Method:   "GET",
			},
			expectStatus: "down",
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Merge with defaults to ensure all required fields are set
			app := mergeWithDefaults(tt.app)

			status, err := scraper.check(app, 5*time.Second, nil)

			assert.Equal(t, tt.expectStatus, status)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckWithValidation(t *testing.T) {
	scraper := NewHTTPScraper()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("SUCCESS: Operation completed"))
	}))
	defer server.Close()

	tests := []struct {
		name         string
		validation   *HTTPValidation
		expectStatus string
		expectErr    bool
	}{
		{
			name: "success condition met",
			validation: &HTTPValidation{
				Success: []HTTPValidationCondition{
					{
						Type: "body_contains",
						Text: "SUCCESS",
					},
				},
			},
			expectStatus: "up",
			expectErr:    false,
		},
		{
			name: "success condition not met",
			validation: &HTTPValidation{
				Success: []HTTPValidationCondition{
					{
						Type: "body_contains",
						Text: "FAILURE",
					},
				},
			},
			expectStatus: "down",
			expectErr:    true,
		},
		{
			name: "failure condition met",
			validation: &HTTPValidation{
				Failure: []HTTPValidationCondition{
					{
						Type: "body_contains",
						Text: "SUCCESS",
					},
				},
			},
			expectStatus: "down",
			expectErr:    true,
		},
		{
			name: "failure condition not met",
			validation: &HTTPValidation{
				Failure: []HTTPValidationCondition{
					{
						Type: "body_contains",
						Text: "ERROR",
					},
				},
			},
			expectStatus: "up",
			expectErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := mergeWithDefaults(HTTPApp{
				Name:       "test-app",
				Location:   "test-location",
				URL:        server.URL,
				Method:     "GET",
				Validation: tt.validation,
			})

			status, err := scraper.check(app, 5*time.Second, nil)

			assert.Equal(t, tt.expectStatus, status)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckWithTLS(t *testing.T) {
	scraper := NewHTTPScraper()

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("OK"))
	}))
	defer server.Close()

	tests := []struct {
		name         string
		sslVerify    bool
		tlsConfig    *tls.Config
		expectStatus string
		expectErr    bool
	}{
		{
			name:         "insecure SSL allowed",
			sslVerify:    false,
			expectStatus: "up",
			expectErr:    false,
		},
		{
			name:      "secure SSL with server's cert",
			sslVerify: true,
			tlsConfig: &tls.Config{
				InsecureSkipVerify: true, // For testing with self-signed cert
			},
			expectStatus: "up",
			expectErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := mergeWithDefaults(HTTPApp{
				Name:      "test-app",
				Location:  "test-location",
				URL:       server.URL,
				Method:    "GET",
				SSLVerify: &tt.sslVerify,
			})

			status, err := scraper.check(app, 5*time.Second, tt.tlsConfig)

			assert.Equal(t, tt.expectStatus, status)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIntegrationScrapeWithRealScenarios(t *testing.T) {
	scraper := NewHTTPScraper()

	// Create test server with complex scenarios
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/health":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "healthy",
				"uptime": 12345,
			})
		case "/slow-api":
			time.Sleep(50 * time.Millisecond)
			w.WriteHeader(200)
			_, _ = w.Write([]byte("OK"))
		case "/auth-required":
			auth := r.Header.Get("Authorization")
			if auth == "Bearer secret-token" {
				w.WriteHeader(200)
				_, _ = w.Write([]byte("Authenticated"))
			} else {
				w.WriteHeader(401)
				_, _ = w.Write([]byte("Unauthorized"))
			}
		default:
			w.WriteHeader(404)
			_, _ = w.Write([]byte("Not Found"))
		}
	}))
	defer server.Close()

	source := config.Source{
		Name: "integration-test",
		Type: "http",
		Config: map[string]interface{}{
			"apps": []map[string]interface{}{
				{
					"name":     "health-check",
					"location": "datacenter-1",
					"url":      server.URL + "/api/health",
					"method":   "GET",
					"validation": map[string]interface{}{
						"success": []map[string]interface{}{
							{
								"type": "status_code",
								"text": "200",
							},
							{
								"type": "body_contains",
								"text": "healthy",
							},
						},
					},
					"labels": map[string]string{
						"service": "health-api",
						"env":     "production",
					},
				},
				{
					"name":     "fast-endpoint",
					"location": "datacenter-1",
					"url":      server.URL + "/slow-api",
					"validation": map[string]interface{}{
						"success": []map[string]interface{}{
							{
								"type":   "response_time",
								"max_ms": 100,
							},
						},
					},
				},
				{
					"name":     "authenticated-api",
					"location": "datacenter-1",
					"url":      server.URL + "/auth-required",
					"auth": map[string]interface{}{
						"type":  "bearer",
						"token": "secret-token",
					},
				},
			},
		},
	}

	statuses, locations, err := scraper.Scrape(
		source,
		config.ServerSettings{},
		5*time.Second,
		5,
		nil,
	)

	require.NoError(t, err)
	assert.Nil(t, locations)
	assert.Len(t, statuses, 3)

	// Verify health check
	healthStatus := findAppStatus(statuses, "health-check")
	require.NotNil(t, healthStatus)
	assert.Equal(t, "up", healthStatus.Status)
	assert.Equal(t, "datacenter-1", healthStatus.Location)
	assert.Equal(t, "integration-test", healthStatus.Source)
	// Find service label
	var serviceLabel string
	for _, label := range healthStatus.Labels {
		if label.Key == "service" {
			serviceLabel = label.Value
			break
		}
	}
	assert.Equal(t, "health-api", serviceLabel)

	// Find env label
	var envLabel string
	for _, label := range healthStatus.Labels {
		if label.Key == "env" {
			envLabel = label.Value
			break
		}
	}
	assert.Equal(t, "production", envLabel)

	// Verify fast endpoint
	fastStatus := findAppStatus(statuses, "fast-endpoint")
	require.NotNil(t, fastStatus)
	assert.Equal(t, "up", fastStatus.Status)

	// Verify authenticated API
	authStatus := findAppStatus(statuses, "authenticated-api")
	require.NotNil(t, authStatus)
	assert.Equal(t, "up", authStatus.Status)
}

// Helper function to find an app status by name
func findAppStatus(statuses []handlers.AppStatus, name string) *handlers.AppStatus {
	for _, status := range statuses {
		if status.Name == name {
			return &status
		}
	}
	return nil
}
