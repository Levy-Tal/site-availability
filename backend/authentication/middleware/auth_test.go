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
