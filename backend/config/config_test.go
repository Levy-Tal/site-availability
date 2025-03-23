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
	defer os.Remove(tempFile.Name())

	// Write a sample config content
	configContent := `
ScrapeInterval: "60s"
Locations:
  - name: "site a"
    Latitude: 32.43843612145413
    Longitude: 34.899453546836334
Apps:
  - name: "app 1"
    location: "site a"
    Metric: "up{instance=\"app1\"}"
    Prometheus:
      - "prometheus1.app.url"
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
}

func TestLoadConfig_InvalidFile(t *testing.T) {
	// Load from a non-existent file
	_, err := LoadConfig("non_existent_file.yaml")
	assert.NotNil(t, err)
}
