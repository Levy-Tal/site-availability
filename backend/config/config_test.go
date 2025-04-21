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
	validConfigContent := `scrape_interval: "60s"
scrape_timeout: "15s"
locations:
  - name: "site a"
    latitude: 32.43843612145413
    longitude: 34.899453546836334
apps:
  - name: "app 1"
    location: "site a"
    metric: "up{instance=\"app1\"}"
    prometheus: "prometheus1.app.url"
`
	_, err = tempFile.WriteString(validConfigContent)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	// Load the valid config
	cfg, err := LoadConfig(tempFile.Name())
	assert.Nil(t, err)
	assert.Equal(t, "60s", cfg.ScrapeInterval)
	assert.Equal(t, "15s", cfg.ScrapeTimeout)
}

func TestLoadConfig_InvalidLatitude(t *testing.T) {
	// Create a temporary config file with an invalid latitude
	tempFile, err := os.CreateTemp("", "config_invalid_latitude.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name()) // Ensure cleanup

	// Write a sample config with an invalid latitude
	invalidLatitudeConfig := `scrape_interval: "60s"
scrape_timeout: "15s"
locations:
  - name: "site a"
    latitude: 95.0
    longitude: 34.899453546836334
apps:
  - name: "app 1"
    location: "site a"
    metric: "up{instance=\"app1\"}"
    prometheus: "prometheus1.app.url"
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
	noLocationsConfig := `scrape_interval: "60s"
scrape_timeout: "15s"
locations: []
apps:
  - name: "app 1"
    location: "site a"
    metric: "up{instance=\"app1\"}"
    prometheus: "prometheus1.app.url"
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
