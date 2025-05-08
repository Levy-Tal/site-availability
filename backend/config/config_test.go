package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file for testing
	tempFile, err := os.CreateTemp("", "config.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name()) // Ensure cleanup

	// Write a sample valid config content
	validConfigContent := `server_settings:
  port: 8080
  custom_ca_path: /app/ca.crt

scraping:
  interval: "60s"
  timeout: "15s"
  max_parallel: 10

documentation:
  title: "DR documentation"
  url: "https://google.com"

prometheus_servers:
  - name: prom1
    url: http://prometheus1-operated:9090/

locations:
  - name: "site a"
    latitude: 32.43843612145413
    longitude: 34.899453546836334

applications:
  - name: "app 1"
    location: "site a"
    metric: "up{instance=\"app1\"}"
    prometheus: "prom1"
`
	_, err = tempFile.WriteString(validConfigContent)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	// Load the valid config
	cfg, err := LoadConfig(tempFile.Name())
	assert.Nil(t, err)
	assert.Equal(t, "60s", cfg.Scraping.Interval)
	assert.Equal(t, "15s", cfg.Scraping.Timeout)
	assert.Equal(t, 10, cfg.Scraping.MaxParallel)
	assert.Equal(t, 8080, cfg.ServerSettings.Port)
	assert.Equal(t, "/app/ca.crt", cfg.ServerSettings.CustomCAPath)
	assert.Equal(t, "DR documentation", cfg.Documentation.Title)
	assert.Equal(t, "https://google.com", cfg.Documentation.URL)
}

func TestLoadConfig_InvalidLatitude(t *testing.T) {
	// Create a temporary config file with an invalid latitude
	tempFile, err := os.CreateTemp("", "config_invalid_latitude.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name()) // Ensure cleanup

	// Write a sample config with an invalid latitude
	invalidLatitudeConfig := `server_settings:
  port: 8080
  custom_ca_path: /app/ca.crt

scraping:
  interval: "60s"
  timeout: "15s"
  max_parallel: 10

documentation:
  title: "DR documentation"
  url: "https://google.com"

prometheus_servers:
  - name: prom1
    url: http://prometheus1-operated:9090/

locations:
  - name: "site a"
    latitude: 95.0
    longitude: 34.899453546836334

applications:
  - name: "app 1"
    location: "site a"
    metric: "up{instance=\"app1\"}"
    prometheus: "prom1"
`
	_, err = tempFile.WriteString(invalidLatitudeConfig)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	// Load the invalid config and check for error
	_, err = LoadConfig(tempFile.Name())
	assert.NotNil(t, err, "Expected an error for invalid latitude but got nil")
	assert.Contains(t, err.Error(), "has an invalid latitude", "Error message does not contain expected text")
}

func TestLoadConfig_InvalidFile(t *testing.T) {
	// Load from a non-existent file
	_, err := LoadConfig("non_existent_file.yaml")
	assert.NotNil(t, err)
}

func TestLoadConfig_NoLocations(t *testing.T) {
	// Create a temporary config file with no locations
	tempFile, err := os.CreateTemp("", "config_no_locations.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name()) // Ensure cleanup

	// Write a sample config with no locations
	noLocationsConfig := `server_settings:
  port: 8080
  custom_ca_path: /app/ca.crt

scraping:
  interval: "60s"
  timeout: "15s"
  max_parallel: 10

documentation:
  title: "DR documentation"
  url: "https://google.com"

prometheus_servers:
  - name: prom1
    url: http://prometheus1-operated:9090/

locations: []

applications:
  - name: "app 1"
    location: "site a"
    metric: "up{instance=\"app1\"}"
    prometheus: "prom1"
`
	_, err = tempFile.WriteString(noLocationsConfig)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	// Load the config and check for error
	_, err = LoadConfig(tempFile.Name())
	assert.NotNil(t, err, "Expected an error for no locations but got nil")
	assert.Contains(t, err.Error(), "at least one location is required", "Error message does not contain expected text")
}
