package hmac

import (
	"bytes"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestValidator_ValidateRequest(t *testing.T) {
	tests := []struct {
		name      string
		token     string
		timestamp string
		body      string
		signature string
		want      bool
	}{
		{
			name:      "valid request",
			token:     "test-token",
			timestamp: time.Now().Format(time.RFC3339),
			body:      "test body",
			signature: "", // Will be set in test
			want:      true,
		},
		{
			name:      "invalid signature",
			token:     "test-token",
			timestamp: time.Now().Format(time.RFC3339),
			body:      "test body",
			signature: "invalid-signature",
			want:      false,
		},
		{
			name:      "expired timestamp",
			token:     "test-token",
			timestamp: time.Now().Add(-6 * time.Minute).Format(time.RFC3339),
			body:      "test body",
			signature: "", // Will be set in test
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator(tt.token)

			// Create request
			req, err := http.NewRequest("GET", "/sync", strings.NewReader(tt.body))
			if err != nil {
				t.Fatal(err)
			}

			// Set timestamp
			req.Header.Set("X-Site-Sync-Timestamp", tt.timestamp)

			// Set signature
			if tt.signature == "" {
				tt.signature = v.GenerateSignature(tt.timestamp, []byte(tt.body))
			}
			req.Header.Set("X-Site-Sync-Signature", tt.signature)

			// Test validation
			if got := v.ValidateRequest(req); got != tt.want {
				t.Errorf("Validator.ValidateRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidator_GenerateSignature(t *testing.T) {
	v := NewValidator("test-token")
	timestamp := time.Now().Format(time.RFC3339)
	body := []byte("test body")

	signature := v.GenerateSignature(timestamp, body)
	if signature == "" {
		t.Error("GenerateSignature() returned empty signature")
	}

	// Verify the signature is valid
	req, err := http.NewRequest("GET", "/sync", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-Site-Sync-Timestamp", timestamp)
	req.Header.Set("X-Site-Sync-Signature", signature)

	if !v.ValidateRequest(req) {
		t.Error("Generated signature failed validation")
	}
}
