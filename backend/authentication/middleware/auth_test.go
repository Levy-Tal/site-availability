package middleware

import (
	"crypto/tls"
	"net/http"
	"testing"

	"site-availability/config"
)

func TestCreateSessionCookie_SecureFlag(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		expectedSecure bool
		description    string
	}{
		{
			name:           "http_request",
			request:        &http.Request{TLS: nil},
			expectedSecure: false,
			description:    "HTTP request should have Secure=false",
		},
		{
			name:           "https_request",
			request:        &http.Request{TLS: &tls.ConnectionState{}},
			expectedSecure: true,
			description:    "HTTPS request should have Secure=true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cookie := CreateSessionCookie("test-session-id", 3600, tt.request, false)

			if cookie.Secure != tt.expectedSecure {
				t.Errorf("Expected Secure=%v for %s, but got Secure=%v",
					tt.expectedSecure, tt.description, cookie.Secure)
			}

			// Verify other cookie properties are set correctly
			if cookie.Name != "session_id" {
				t.Errorf("Expected cookie name 'session_id', got '%s'", cookie.Name)
			}

			if cookie.Value != "test-session-id" {
				t.Errorf("Expected cookie value 'test-session-id', got '%s'", cookie.Value)
			}

			if cookie.MaxAge != 3600 {
				t.Errorf("Expected MaxAge 3600, got %d", cookie.MaxAge)
			}

			if !cookie.HttpOnly {
				t.Error("Expected HttpOnly=true for security")
			}

			if cookie.SameSite != http.SameSiteLaxMode {
				t.Errorf("Expected SameSite=%v, got %v", http.SameSiteLaxMode, cookie.SameSite)
			}
		})
	}
}

func TestDeleteSessionCookie_SecureFlag(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		expectedSecure bool
		description    string
	}{
		{
			name:           "http_request",
			request:        &http.Request{TLS: nil},
			expectedSecure: false,
			description:    "HTTP request should have Secure=false",
		},
		{
			name:           "https_request",
			request:        &http.Request{TLS: &tls.ConnectionState{}},
			expectedSecure: true,
			description:    "HTTPS request should have Secure=true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cookie := DeleteSessionCookie(tt.request, false)

			if cookie.Secure != tt.expectedSecure {
				t.Errorf("Expected Secure=%v for %s, but got Secure=%v",
					tt.expectedSecure, tt.description, cookie.Secure)
			}

			// Verify other cookie properties are set correctly for deletion
			if cookie.Name != "session_id" {
				t.Errorf("Expected cookie name 'session_id', got '%s'", cookie.Name)
			}

			if cookie.Value != "" {
				t.Errorf("Expected empty cookie value for deletion, got '%s'", cookie.Value)
			}

			if cookie.MaxAge != -1 {
				t.Errorf("Expected MaxAge -1 for deletion, got %d", cookie.MaxAge)
			}

			if !cookie.HttpOnly {
				t.Error("Expected HttpOnly=true for security")
			}

			if cookie.SameSite != http.SameSiteLaxMode {
				t.Errorf("Expected SameSite=%v, got %v", http.SameSiteLaxMode, cookie.SameSite)
			}
		})
	}
}

func TestIsExcludedPath_OIDCAuthenticationPaths(t *testing.T) {
	cfg := &config.Config{
		ServerSettings: config.ServerSettings{
			OIDC: config.OIDCConfig{
				Enabled: true,
			},
		},
	}

	am := NewAuthMiddleware(cfg, nil)

	tests := []struct {
		name          string
		path          string
		shouldExclude bool
		description   string
	}{
		{
			name:          "oidc_login_path",
			path:          "/auth/oidc/login",
			shouldExclude: true,
			description:   "OIDC login path should be excluded",
		},
		{
			name:          "oidc_callback_path",
			path:          "/auth/oidc/callback",
			shouldExclude: true,
			description:   "OIDC callback path should be excluded",
		},
		{
			name:          "local_login_path",
			path:          "/auth/login",
			shouldExclude: true,
			description:   "Local login path should be excluded",
		},
		{
			name:          "root_path",
			path:          "/",
			shouldExclude: true,
			description:   "Root path should be excluded",
		},
		{
			name:          "api_locations_path",
			path:          "/api/locations",
			shouldExclude: false,
			description:   "API locations path should require authentication",
		},
		{
			name:          "api_apps_path",
			path:          "/api/apps",
			shouldExclude: false,
			description:   "API apps path should require authentication",
		},
		{
			name:          "auth_user_path",
			path:          "/auth/user",
			shouldExclude: false,
			description:   "Auth user path should require authentication",
		},
		{
			name:          "healthz_path",
			path:          "/healthz",
			shouldExclude: true,
			description:   "Health check path should be excluded",
		},
		{
			name:          "metrics_path",
			path:          "/metrics",
			shouldExclude: true,
			description:   "Metrics path should be excluded",
		},
		{
			name:          "static_file_path",
			path:          "/static/app.js",
			shouldExclude: true,
			description:   "Static file path should be excluded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isExcluded := am.isExcludedPath(tt.path)

			if tt.shouldExclude && !isExcluded {
				t.Errorf("Expected %s to be excluded for %s, but it wasn't", tt.path, tt.description)
			}

			if !tt.shouldExclude && isExcluded {
				t.Errorf("Expected %s to require authentication for %s, but it was excluded", tt.path, tt.description)
			}
		})
	}
}

func TestIsAuthRequired_OIDCOnly(t *testing.T) {
	tests := []struct {
		name              string
		localAdminEnabled bool
		oidcEnabled       bool
		shouldRequire     bool
		description       string
	}{
		{
			name:              "both_disabled",
			localAdminEnabled: false,
			oidcEnabled:       false,
			shouldRequire:     false,
			description:       "No authentication should be required when both are disabled",
		},
		{
			name:              "local_admin_only",
			localAdminEnabled: true,
			oidcEnabled:       false,
			shouldRequire:     true,
			description:       "Authentication should be required when only local admin is enabled",
		},
		{
			name:              "oidc_only",
			localAdminEnabled: false,
			oidcEnabled:       true,
			shouldRequire:     true,
			description:       "Authentication should be required when only OIDC is enabled",
		},
		{
			name:              "both_enabled",
			localAdminEnabled: true,
			oidcEnabled:       true,
			shouldRequire:     true,
			description:       "Authentication should be required when both are enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				ServerSettings: config.ServerSettings{
					LocalAdmin: config.LocalAdminConfig{
						Enabled: tt.localAdminEnabled,
					},
					OIDC: config.OIDCConfig{
						Enabled: tt.oidcEnabled,
					},
				},
			}

			am := NewAuthMiddleware(cfg, nil)
			isRequired := am.isAuthRequired()

			if tt.shouldRequire && !isRequired {
				t.Errorf("Expected authentication to be required for %s, but it wasn't", tt.description)
			}

			if !tt.shouldRequire && isRequired {
				t.Errorf("Expected authentication to not be required for %s, but it was", tt.description)
			}
		})
	}
}
