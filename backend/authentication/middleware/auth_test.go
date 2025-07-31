package middleware

import (
	"crypto/tls"
	"net/http"
	"testing"
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
			cookie := CreateSessionCookie("test-session-id", 3600, tt.request)

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
			cookie := DeleteSessionCookie(tt.request)

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
