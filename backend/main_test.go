package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"site-availability/config"
	"site-availability/logging"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// Setup test environment
	if err := logging.Init(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestLoadConfig(t *testing.T) {
	// Create temporary test files
	tmpDir := t.TempDir()

	// Create test CA certificates
	serverCAPath := filepath.Join(tmpDir, "ca.crt")
	err := os.WriteFile(serverCAPath, []byte("test server certificate"), 0644)
	require.NoError(t, err)

	siteCAPath := filepath.Join(tmpDir, "site-ca.crt")
	err = os.WriteFile(siteCAPath, []byte("test site certificate"), 0644)
	require.NoError(t, err)

	// Create test config.yaml
	configContent := fmt.Sprintf(`
server_settings:
  port: "8080"
  sync_enable: true
  custom_ca_path: "%s"

scraping:
  interval: "60s"
  timeout: "15s"
  max_parallel: 10

documentation:
  title: "Test Documentation"
  url: "https://test.example.com/docs"

locations:
  - name: "test location"
    latitude: 31.782904
    longitude: 35.214774

prometheus_servers:
  - name: "prom1"
    url: "http://prometheus:9090/"

applications:
  - name: "test app"
    location: "test location"
    metric: 'up{instance="test"}'
    prometheus: "prom1"

sites:
  - name: "Test Site"
    url: "https://test-site:3030"
    enabled: true
    timeout: "5s"
    custom_ca_path: "%s"
`, serverCAPath, siteCAPath)
	configPath := filepath.Join(tmpDir, "config.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Create test credentials.yaml
	credentialsContent := `
server_settings:
  token: "test-server-token"

prometheus_servers:
  - name: "prom1"
    auth: "bearer"
    token: "test-prometheus-token"

sites:
  - name: "Test Site"
    token: "test-site-token"
`
	credentialsPath := filepath.Join(tmpDir, "credentials.yaml")
	err = os.WriteFile(credentialsPath, []byte(credentialsContent), 0644)
	require.NoError(t, err)

	// Set environment variables
	os.Setenv("CONFIG_FILE", configPath)
	os.Setenv("CREDENTIALS_FILE", credentialsPath)
	defer func() {
		os.Unsetenv("CONFIG_FILE")
		os.Unsetenv("CREDENTIALS_FILE")
	}()

	// Test loading configuration
	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Verify server settings
	assert.Equal(t, "8080", cfg.ServerSettings.Port)
	assert.True(t, cfg.ServerSettings.SyncEnable)
	assert.Equal(t, serverCAPath, cfg.ServerSettings.CustomCAPath)
	assert.Equal(t, "test-server-token", cfg.ServerSettings.Token)

	// Verify scraping settings
	assert.Equal(t, "60s", cfg.Scraping.Interval)
	assert.Equal(t, "15s", cfg.Scraping.Timeout)
	assert.Equal(t, 10, cfg.Scraping.MaxParallel)

	// Verify documentation
	assert.Equal(t, "Test Documentation", cfg.Documentation.Title)
	assert.Equal(t, "https://test.example.com/docs", cfg.Documentation.URL)

	// Verify locations
	require.Len(t, cfg.Locations, 1)
	assert.Equal(t, "test location", cfg.Locations[0].Name)
	assert.Equal(t, 31.782904, cfg.Locations[0].Latitude)
	assert.Equal(t, 35.214774, cfg.Locations[0].Longitude)

	// Verify Prometheus servers
	require.Len(t, cfg.PrometheusServers, 1)
	assert.Equal(t, "prom1", cfg.PrometheusServers[0].Name)
	assert.Equal(t, "http://prometheus:9090/", cfg.PrometheusServers[0].URL)
	assert.Equal(t, "bearer", cfg.PrometheusServers[0].Auth)
	assert.Equal(t, "test-prometheus-token", cfg.PrometheusServers[0].Token)

	// Verify applications
	require.Len(t, cfg.Applications, 1)
	assert.Equal(t, "test app", cfg.Applications[0].Name)
	assert.Equal(t, "test location", cfg.Applications[0].Location)
	assert.Equal(t, `up{instance="test"}`, cfg.Applications[0].Metric)
	assert.Equal(t, "prom1", cfg.Applications[0].Prometheus)

	// Verify sites
	require.Len(t, cfg.Sites, 1)
	assert.Equal(t, "Test Site", cfg.Sites[0].Name)
	assert.Equal(t, "https://test-site:3030", cfg.Sites[0].URL)
	assert.True(t, cfg.Sites[0].Enabled)
	assert.Equal(t, "5s", cfg.Sites[0].Timeout)
	assert.Equal(t, siteCAPath, cfg.Sites[0].CustomCAPath)
	assert.Equal(t, "test-site-token", cfg.Sites[0].Token)
}

func TestLoadConfigWithMissingFiles(t *testing.T) {
	// Test with missing config file
	os.Setenv("CONFIG_FILE", "nonexistent.yaml")
	os.Setenv("CREDENTIALS_FILE", "nonexistent-credentials.yaml")
	defer func() {
		os.Unsetenv("CONFIG_FILE")
		os.Unsetenv("CREDENTIALS_FILE")
	}()

	_, err := config.LoadConfig()
	assert.Error(t, err)
}

func TestLoadConfigWithInvalidYAML(t *testing.T) {
	// Create temporary test file with invalid YAML
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")
	err := os.WriteFile(configPath, []byte("invalid: yaml: content:"), 0644)
	require.NoError(t, err)

	os.Setenv("CONFIG_FILE", configPath)
	defer os.Unsetenv("CONFIG_FILE")

	_, err = config.LoadConfig()
	assert.Error(t, err)
}

func TestLoadConfigWithMissingRequiredFields(t *testing.T) {
	// Create temporary test file with missing required fields
	tmpDir := t.TempDir()
	configContent := `
server_settings:
  port: "8080"
`
	configPath := filepath.Join(tmpDir, "incomplete.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	os.Setenv("CONFIG_FILE", configPath)
	defer os.Unsetenv("CONFIG_FILE")

	_, err = config.LoadConfig()
	assert.Error(t, err)
}

func TestLoadConfigWithInvalidLocationCoordinates(t *testing.T) {
	// Create temporary test file with invalid coordinates
	tmpDir := t.TempDir()
	configContent := `
server_settings:
  port: "8080"

locations:
  - name: "invalid location"
    latitude: 200.0  # Invalid latitude
    longitude: 35.214774
`
	configPath := filepath.Join(tmpDir, "invalid-coords.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	os.Setenv("CONFIG_FILE", configPath)
	defer os.Unsetenv("CONFIG_FILE")

	_, err = config.LoadConfig()
	assert.Error(t, err)
}

func TestLoadConfigWithOptionalCredentials(t *testing.T) {
	// Create temporary test files
	tmpDir := t.TempDir()

	// Create test config.yaml with minimal required fields
	configContent := `
server_settings:
  port: "8080"

locations:
  - name: "test location"
    latitude: 31.782904
    longitude: 35.214774

prometheus_servers:
  - name: "prom1"
    url: "http://prometheus:9090/"

sites:
  - name: "Test Site"
    url: "https://test-site:3030"
    enabled: true
`
	configPath := filepath.Join(tmpDir, "minimal-config.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Create empty credentials file
	credentialsPath := filepath.Join(tmpDir, "empty-credentials.yaml")
	err = os.WriteFile(credentialsPath, []byte(""), 0644)
	require.NoError(t, err)

	os.Setenv("CONFIG_FILE", configPath)
	os.Setenv("CREDENTIALS_FILE", credentialsPath)
	defer func() {
		os.Unsetenv("CONFIG_FILE")
		os.Unsetenv("CREDENTIALS_FILE")
	}()

	// Test loading configuration with optional credentials
	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Verify that optional fields are empty
	assert.Empty(t, cfg.ServerSettings.Token)
	assert.Empty(t, cfg.PrometheusServers[0].Auth)
	assert.Empty(t, cfg.PrometheusServers[0].Token)
	assert.Empty(t, cfg.Sites[0].Token)
}

func TestLoadConfigWithEnvironmentVariables(t *testing.T) {
	// Test default values when environment variables are not set
	os.Unsetenv("CONFIG_FILE")
	os.Unsetenv("CREDENTIALS_FILE")

	// Create default config file
	tmpDir := t.TempDir()
	configContent := `
server_settings:
  port: "8080"

locations:
  - name: "test location"
    latitude: 31.782904
    longitude: 35.214774

prometheus_servers:
  - name: "prom1"
    url: "http://prometheus:9090/"

sites:
  - name: "Test Site"
    url: "https://test-site:3030"
    enabled: true
`
	defaultConfigPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(defaultConfigPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(originalDir)
		require.NoError(t, err)
	}()

	// Test loading configuration with default paths
	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "8080", cfg.ServerSettings.Port)
}
