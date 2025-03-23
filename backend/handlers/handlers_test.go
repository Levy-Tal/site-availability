package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock the status cache
func TestGetAppStatus(t *testing.T) {
	// Initialize the cache with mock data
	UpdateAppStatusCache("app 1", AppStatus{
		Name:     "app 1",
		Location: "site a",
		Status:   "up",
	})

	// Set up the HTTP request
	req, err := http.NewRequest("GET", "/api/status", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetAppStatus)
	handler.ServeHTTP(rr, req)

	// Check the response status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check if the response contains the expected app status
	expected := `{"app 1":{"name":"app 1","location":"site a","status":"up"}}`
	assert.JSONEq(t, expected, rr.Body.String())
}
