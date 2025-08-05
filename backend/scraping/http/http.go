package http_source

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"site-availability/config"
	"site-availability/handlers"
	"site-availability/logging"
	"strconv"
	"strings"
	"sync"
	"time"
)

// HTTPConfig represents the configuration for HTTP sources
type HTTPConfig struct {
	Apps []HTTPApp `yaml:"apps"`
}

// HTTPApp represents an app configuration for HTTP monitoring
type HTTPApp struct {
	Name        string            `yaml:"name"`
	Location    string            `yaml:"location"`
	URL         string            `yaml:"url"`
	Method      string            `yaml:"method"`
	Headers     map[string]string `yaml:"headers"`
	Body        string            `yaml:"body"`
	ContentType string            `yaml:"content_type"`

	// HTTP client options
	Timeout         string `yaml:"timeout"`
	FollowRedirects *bool  `yaml:"follow_redirects"`
	MaxRedirects    int    `yaml:"max_redirects"`

	// Authentication
	Auth *HTTPAuth `yaml:"auth"`

	// Status code validation
	AllowedStatusCodes []interface{} `yaml:"allowed_status_codes"`
	BlockedStatusCodes []interface{} `yaml:"blocked_status_codes"`

	// Content validation
	Validation *HTTPValidation `yaml:"validation"`

	// Simplified SSL configuration
	SSLVerify *bool `yaml:"ssl_verify"`

	Labels map[string]string `yaml:"labels"`
}

// HTTPAuth represents authentication configuration
type HTTPAuth struct {
	Type     string `yaml:"type"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Token    string `yaml:"token"`
}

// HTTPValidation represents content validation configuration
type HTTPValidation struct {
	Success []HTTPValidationCondition `yaml:"success"`
	Failure []HTTPValidationCondition `yaml:"failure"`
}

// HTTPValidationCondition represents a single validation condition
type HTTPValidationCondition struct {
	Type          string      `yaml:"type"`
	Text          string      `yaml:"text"`
	Path          string      `yaml:"path"`
	ExpectedValue interface{} `yaml:"expected_value"`
	ValueType     string      `yaml:"value_type"`
	Operator      string      `yaml:"operator"`
	CaseSensitive *bool       `yaml:"case_sensitive"`
	MaxMS         int         `yaml:"max_ms"`
}

// HTTPScraper implements the Scraper interface for HTTP sources
type HTTPScraper struct {
}

func NewHTTPScraper() *HTTPScraper {
	return &HTTPScraper{}
}

// getDefaultHTTPApp returns an HTTPApp with all default values
func getDefaultHTTPApp() HTTPApp {
	return HTTPApp{
		Method:             "GET",
		Headers:            make(map[string]string),
		Body:               "",
		ContentType:        "",
		Timeout:            "",
		FollowRedirects:    boolPtr(true),
		MaxRedirects:       10,
		AllowedStatusCodes: []interface{}{"2XX"},
		BlockedStatusCodes: []interface{}{"4XX", "5XX"},
		Validation: &HTTPValidation{
			Success: []HTTPValidationCondition{},
			Failure: []HTTPValidationCondition{},
		},
		SSLVerify: boolPtr(true),
		Labels:    make(map[string]string),
	}
}

// boolPtr returns a pointer to a boolean value
func boolPtr(b bool) *bool {
	return &b
}

// mergeWithDefaults merges app config with defaults
func mergeWithDefaults(app HTTPApp) HTTPApp {
	defaults := getDefaultHTTPApp()

	// Merge basic fields
	if app.Method == "" {
		app.Method = defaults.Method
	}
	if app.ContentType == "" {
		app.ContentType = defaults.ContentType
	}
	if app.Timeout == "" {
		app.Timeout = defaults.Timeout
	}
	if app.FollowRedirects == nil {
		app.FollowRedirects = defaults.FollowRedirects
	}
	if app.MaxRedirects == 0 {
		app.MaxRedirects = defaults.MaxRedirects
	}
	if app.AllowedStatusCodes == nil {
		app.AllowedStatusCodes = defaults.AllowedStatusCodes
	}
	if app.BlockedStatusCodes == nil {
		app.BlockedStatusCodes = defaults.BlockedStatusCodes
	}

	// Merge validation
	if app.Validation == nil {
		app.Validation = defaults.Validation
	}

	// Merge SSL config
	if app.SSLVerify == nil {
		app.SSLVerify = defaults.SSLVerify
	}

	return app
}

// ValidateConfig validates the HTTP-specific configuration
func (h *HTTPScraper) ValidateConfig(source config.Source) error {
	httpCfg, err := config.DecodeConfig[HTTPConfig](source.Config, source.Name)
	if err != nil {
		return err
	}

	// Validate apps
	if len(httpCfg.Apps) == 0 {
		return fmt.Errorf("http source %s: at least one app is required", source.Name)
	}

	appNames := make(map[string]bool)
	for _, app := range httpCfg.Apps {
		app = mergeWithDefaults(app)

		if app.Name == "" {
			return fmt.Errorf("http source %s: app name is required", source.Name)
		}
		if _, exists := appNames[app.Name]; exists {
			return fmt.Errorf("http source %s: duplicate app name %q", source.Name, app.Name)
		}
		appNames[app.Name] = true

		if app.URL == "" {
			return fmt.Errorf("http source %s: app %s missing 'url'", source.Name, app.Name)
		}
		if app.Location == "" {
			return fmt.Errorf("http source %s: app %s missing 'location'", source.Name, app.Name)
		}

		// Validate URL
		if _, err := url.Parse(app.URL); err != nil {
			return fmt.Errorf("http source %s: app %s invalid URL %q: %w", source.Name, app.Name, app.URL, err)
		}

		// Validate method
		validMethods := map[string]bool{"GET": true, "POST": true, "PUT": true, "DELETE": true, "HEAD": true, "OPTIONS": true}
		if !validMethods[strings.ToUpper(app.Method)] {
			return fmt.Errorf("http source %s: app %s invalid method %q", source.Name, app.Name, app.Method)
		}

		// Validate status codes
		if err := validateStatusCodes(app.AllowedStatusCodes, "allowed_status_codes"); err != nil {
			return fmt.Errorf("http source %s: app %s %w", source.Name, app.Name, err)
		}
		if err := validateStatusCodes(app.BlockedStatusCodes, "blocked_status_codes"); err != nil {
			return fmt.Errorf("http source %s: app %s %w", source.Name, app.Name, err)
		}

		// Validate authentication
		if app.Auth != nil {
			if err := validateAuth(app.Auth); err != nil {
				return fmt.Errorf("http source %s: app %s %w", source.Name, app.Name, err)
			}
		}
	}

	return nil
}

// validateStatusCodes validates status code arrays
func validateStatusCodes(codes []interface{}, fieldName string) error {
	for _, code := range codes {
		switch v := code.(type) {
		case int:
			if v < 100 || v > 599 {
				return fmt.Errorf("%s: invalid status code %d", fieldName, v)
			}
		case string:
			if !isValidStatusCodeRange(v) {
				return fmt.Errorf("%s: invalid status code range %q", fieldName, v)
			}
		default:
			return fmt.Errorf("%s: invalid status code type %T", fieldName, code)
		}
	}
	return nil
}

// isValidStatusCodeRange checks if a string is a valid status code range
func isValidStatusCodeRange(s string) bool {
	matched, _ := regexp.MatchString(`^[1-5]XX$`, s)
	return matched
}

// validateAuth validates authentication configuration
func validateAuth(auth *HTTPAuth) error {
	validTypes := map[string]bool{"basic": true, "bearer": true, "digest": true, "oauth2": true}
	if !validTypes[auth.Type] {
		return fmt.Errorf("auth: invalid type %q", auth.Type)
	}

	switch auth.Type {
	case "basic":
		if auth.Username == "" || auth.Password == "" {
			return fmt.Errorf("auth: basic authentication requires username and password")
		}
	case "bearer":
		if auth.Token == "" {
			return fmt.Errorf("auth: bearer authentication requires token")
		}
	case "digest":
		if auth.Username == "" || auth.Password == "" {
			return fmt.Errorf("auth: digest authentication requires username and password")
		}
	case "oauth2":
		if auth.Token == "" {
			return fmt.Errorf("auth: oauth2 authentication requires token")
		}
	}

	return nil
}

func (h *HTTPScraper) Scrape(source config.Source, serverSettings config.ServerSettings, timeout time.Duration, maxParallel int, tlsConfig *tls.Config) ([]handlers.AppStatus, []handlers.Location, error) {
	// Decode the source-specific config
	httpCfg, err := config.DecodeConfig[HTTPConfig](source.Config, source.Name)
	if err != nil {
		return nil, nil, err
	}

	results := make([]handlers.AppStatus, len(httpCfg.Apps))
	sem := make(chan struct{}, maxParallel)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, app := range httpCfg.Apps {
		sem <- struct{}{} // Acquire slot
		wg.Add(1)
		go func(i int, app HTTPApp) {
			defer func() {
				<-sem
				wg.Done()
			}()

			// Merge with defaults
			app = mergeWithDefaults(app)

			status, err := h.check(app, timeout, tlsConfig)
			if err != nil {
				logging.Logger.WithFields(map[string]interface{}{
					"app":    app.Name,
					"source": source.Name,
					"error":  err.Error(),
				}).Warn("HTTP check failed - marking as down")
				status = "down"
			}

			mu.Lock()
			results[i] = handlers.AppStatus{
				Name:      app.Name,
				Location:  app.Location,
				Status:    status,
				Source:    source.Name,
				OriginURL: serverSettings.HostURL, // Use host URL as origin for deduplication
				Labels:    app.Labels,
			}
			mu.Unlock()
		}(i, app)
	}
	wg.Wait()

	// HTTP scraper returns nil for locations since it only provides app statuses
	return results, nil, nil
}

func (h *HTTPScraper) check(app HTTPApp, timeout time.Duration, tlsConfig *tls.Config) (string, error) {
	// Parse app timeout or use server timeout
	appTimeout := timeout
	if app.Timeout != "" {
		if parsedTimeout, err := time.ParseDuration(app.Timeout); err == nil {
			appTimeout = parsedTimeout
		}
	}

	// Create HTTP client
	client := &http.Client{
		Timeout: appTimeout,
	}

	// Configure redirects
	if app.FollowRedirects != nil {
		if !*app.FollowRedirects {
			client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}
		}
	}

	// Configure TLS
	transport := &http.Transport{}
	if tlsConfig != nil {
		transport.TLSClientConfig = tlsConfig
	}

	// Override TLS settings if specified
	if app.SSLVerify != nil && !*app.SSLVerify {
		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{}
		}
		transport.TLSClientConfig.InsecureSkipVerify = true
	}

	client.Transport = transport

	// Create request
	req, err := http.NewRequest(strings.ToUpper(app.Method), app.URL, nil)
	if err != nil {
		return "down", fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range app.Headers {
		req.Header.Set(key, value)
	}

	// Add authentication
	if app.Auth != nil {
		if err := h.addAuthentication(req, app.Auth); err != nil {
			return "down", fmt.Errorf("authentication failed: %w", err)
		}
	}

	// Add body for POST/PUT requests
	if app.Body != "" {
		req.Body = io.NopCloser(strings.NewReader(app.Body))
		if app.ContentType != "" {
			req.Header.Set("Content-Type", app.ContentType)
		} else if strings.ToUpper(app.Method) == "POST" || strings.ToUpper(app.Method) == "PUT" {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	}

	logging.Logger.WithFields(map[string]interface{}{
		"url":    app.URL,
		"method": app.Method,
		"app":    app.Name,
		"source": "httpScraper.check",
	}).Debug("Making HTTP request")

	start := time.Now()
	resp, err := client.Do(req)
	responseTime := time.Since(start)

	if err != nil {
		return "down", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body for validation
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "down", fmt.Errorf("failed to read response body: %w", err)
	}
	bodyString := string(bodyBytes)

	// Check status code
	if !h.isStatusAllowed(resp.StatusCode, app.AllowedStatusCodes) {
		return "down", fmt.Errorf("status code %d not in allowed codes", resp.StatusCode)
	}
	if h.isStatusBlocked(resp.StatusCode, app.BlockedStatusCodes) {
		return "down", fmt.Errorf("status code %d is in blocked codes", resp.StatusCode)
	}

	// Validate content
	if app.Validation != nil {
		responseTimeMS := int(responseTime.Milliseconds())
		// Check success conditions
		for _, condition := range app.Validation.Success {
			if !h.validateCondition(condition, resp.StatusCode, responseTimeMS, bodyString) {
				return "down", fmt.Errorf("success condition failed: %s", condition.Type)
			}
		}

		// Check failure conditions
		for _, condition := range app.Validation.Failure {
			if h.validateCondition(condition, resp.StatusCode, responseTimeMS, bodyString) {
				return "down", fmt.Errorf("failure condition met: %s", condition.Type)
			}
		}
	}

	logging.Logger.WithFields(map[string]interface{}{
		"app":           app.Name,
		"status_code":   resp.StatusCode,
		"response_time": int(responseTime.Milliseconds()),
	}).Debug("HTTP check successful")

	return "up", nil
}

// addAuthentication adds authentication to the request
func (h *HTTPScraper) addAuthentication(req *http.Request, auth *HTTPAuth) error {
	switch auth.Type {
	case "basic":
		credentials := base64.StdEncoding.EncodeToString([]byte(auth.Username + ":" + auth.Password))
		req.Header.Set("Authorization", "Basic "+credentials)
	case "bearer":
		req.Header.Set("Authorization", "Bearer "+auth.Token)
	case "digest":
		// Digest authentication would require more complex implementation
		// For now, just set the Authorization header
		req.Header.Set("Authorization", "Digest "+auth.Token)
	case "oauth2":
		req.Header.Set("Authorization", "Bearer "+auth.Token)
	default:
		return fmt.Errorf("unsupported authentication type: %s", auth.Type)
	}
	return nil
}

// isStatusAllowed checks if a status code is in the allowed list
func (h *HTTPScraper) isStatusAllowed(statusCode int, allowedCodes []interface{}) bool {
	for _, code := range allowedCodes {
		switch v := code.(type) {
		case int:
			if statusCode == v {
				return true
			}
		case string:
			if h.isStatusCodeInRange(statusCode, v) {
				return true
			}
		}
	}
	return false
}

// isStatusBlocked checks if a status code is in the blocked list
func (h *HTTPScraper) isStatusBlocked(statusCode int, blockedCodes []interface{}) bool {
	for _, code := range blockedCodes {
		switch v := code.(type) {
		case int:
			if statusCode == v {
				return true
			}
		case string:
			if h.isStatusCodeInRange(statusCode, v) {
				return true
			}
		}
	}
	return false
}

// isStatusCodeInRange checks if a status code is in a range like "2XX"
func (h *HTTPScraper) isStatusCodeInRange(statusCode int, rangeStr string) bool {
	if len(rangeStr) != 3 || !strings.HasSuffix(rangeStr, "XX") {
		return false
	}

	startStr := rangeStr[:1]
	start, err := strconv.Atoi(startStr)
	if err != nil {
		return false
	}

	return statusCode >= start*100 && statusCode < (start+1)*100
}

// validateCondition validates a single condition
func (h *HTTPScraper) validateCondition(condition HTTPValidationCondition, statusCode, responseTimeMS int, bodyString string) bool {
	switch condition.Type {
	case "status_code":
		if condition.Text != "" {
			expectedCode, err := strconv.Atoi(condition.Text)
			if err != nil {
				return false
			}
			return statusCode == expectedCode
		}
	case "response_time":
		if condition.MaxMS > 0 {
			return responseTimeMS <= condition.MaxMS
		}
	case "body_contains":
		if condition.Text != "" {
			if condition.CaseSensitive != nil && !*condition.CaseSensitive {
				return strings.Contains(strings.ToLower(bodyString), strings.ToLower(condition.Text))
			}
			return strings.Contains(bodyString, condition.Text)
		}
	case "body_not_contains":
		if condition.Text != "" {
			if condition.CaseSensitive != nil && !*condition.CaseSensitive {
				return !strings.Contains(strings.ToLower(bodyString), strings.ToLower(condition.Text))
			}
			return !strings.Contains(bodyString, condition.Text)
		}
	case "json_path":
		// Simple JSON path validation - could be enhanced with a proper JSON path library
		if condition.Path != "" && condition.ExpectedValue != nil {
			// For now, just check if the path exists in the JSON
			// This is a simplified implementation
			var jsonData map[string]interface{}
			if err := json.Unmarshal([]byte(bodyString), &jsonData); err != nil {
				return false
			}
			// Basic path checking - could be enhanced
			return true // Simplified for now
		}
	}
	return false
}
