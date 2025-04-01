package handlers

import (
	"net/http"
	"net/http/httptest"
	"site-availability/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAppStatus(t *testing.T) {
	// Initialize the cache with mock data
	UpdateAppStatus([]AppStatus{
		{
			Name:     "app 1",
			Location: "site a",
			Status:   "up",
		},
	})

	// Set up the HTTP request
	req, err := http.NewRequest("GET", "/api/status", nil)
	assert.NoError(t, err)

	// Mock config (for locations)
	cfg := &config.Config{
		Locations: []config.Location{
			{Name: "site a", Latitude: 0.0, Longitude: 0.0},
		},
	}

	// Record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		GetAppStatus(w, r, cfg)
	})
	handler.ServeHTTP(rr, req)

	// Check the response status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Expected JSON response
	expected := `{"locations":[{"name":"site a","latitude":0.0,"longitude":0.0}],"apps":[{"name":"app 1","location":"site a","status":"up"}]}`

	// Compare expected and actual JSON response
	assert.JSONEq(t, expected, rr.Body.String(), "Response JSON does not match expected output")
}
