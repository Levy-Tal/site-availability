package hmac

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"time"
)

// Validator handles HMAC validation for site sync requests
type Validator struct {
	token string
}

// NewValidator creates a new HMAC validator with the given token
func NewValidator(token string) *Validator {
	return &Validator{
		token: token,
	}
}

// ValidateRequest checks if the request has a valid HMAC signature and timestamp
func (v *Validator) ValidateRequest(r *http.Request) bool {
	return v.ValidateHMAC(r) && v.ValidateTimestamp(r)
}

// ValidateHMAC checks if the request has a valid HMAC signature
func (v *Validator) ValidateHMAC(r *http.Request) bool {
	signature := r.Header.Get("X-Site-Sync-Signature")
	timestamp := r.Header.Get("X-Site-Sync-Timestamp")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return false
	}
	// Restore body for later use
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	// Create HMAC
	h := hmac.New(sha256.New, []byte(v.token))
	h.Write([]byte(timestamp))
	h.Write(body)
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// ValidateTimestamp checks if the request timestamp is within the allowed window
func (v *Validator) ValidateTimestamp(r *http.Request) bool {
	timestamp := r.Header.Get("X-Site-Sync-Timestamp")
	ts, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return false
	}

	// Allow 5-minute window for clock skew
	now := time.Now()
	return ts.After(now.Add(-5*time.Minute)) && ts.Before(now.Add(5*time.Minute))
}

// GenerateSignature creates an HMAC signature for the given timestamp and body
func (v *Validator) GenerateSignature(timestamp string, body []byte) string {
	h := hmac.New(sha256.New, []byte(v.token))
	h.Write([]byte(timestamp))
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}
