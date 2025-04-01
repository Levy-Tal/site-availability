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

	// Write a sample config content (fixed YAML structure)
	configContent := `scrape_interval: "60s"
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
	_, err = tempFile.WriteString(configContent)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	// Close the file
	err = tempFile.Close()
	if err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Load the config from the temporary file
	cfg, err := LoadConfig(tempFile.Name())
	assert.Nil(t, err)
	assert.Equal(t, "60s", cfg.ScrapeInterval)
	assert.Len(t, cfg.Locations, 1)
	assert.Len(t, cfg.Apps, 1)
	assert.Equal(t, "site a", cfg.Locations[0].Name)
	assert.Equal(t, "app 1", cfg.Apps[0].Name)
	assert.Equal(t, "prometheus1.app.url", cfg.Apps[0].Prometheus)
}

func TestLoadConfig_InvalidFile(t *testing.T) {
	// Load from a non-existent file
	_, err := LoadConfig("non_existent_file.yaml")
	assert.NotNil(t, err)
}
