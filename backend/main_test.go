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

func getSourceByType(sources []config.Source, typ string) *config.Source {
	for i := range sources {
		if sources[i].Type == typ {
			return &sources[i]
		}
	}
	return nil
}

func getSourceByName(sources []config.Source, name string) *config.Source {
	for i := range sources {
		if sources[i].Name == name {
			return &sources[i]
		}
	}
	return nil
}

// setupMainTest sets up a clean environment for each test
func setupMainTest() {
	// Unset environment variables to ensure clean state
	os.Unsetenv("CONFIG_FILE")
	os.Unsetenv("CREDENTIALS_FILE")
}

func TestMain(m *testing.M) {
	// Setup test environment
	if err := logging.Init(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestConfigurationLoading(t *testing.T) {
	setupMainTest()

	t.Run("load complete valid configuration", func(t *testing.T) {
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

sources:
  - name: "prom1"
    type: "prometheus"
    config:
      url: "http://prometheus:9090/"
      auth: "bearer"
      token: "test-prometheus-token"
      apps:
        - name: "test app"
          location: "test location"
          metric: 'up{instance="test"}'
  - name: "site-a"
    type: "site"
    config:
      url: "https://test-site:3030"
      token: "test-site-token"
`, serverCAPath)
		configPath := filepath.Join(tmpDir, "config.yaml")
		err = os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		// Create test credentials.yaml
		credentialsContent := `
server_settings:
  token: "test-server-token"
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

		// Verify sources
		require.Len(t, cfg.Sources, 2)

		// Verify prometheus source
		prom := getSourceByType(cfg.Sources, "prometheus")
		require.NotNil(t, prom)
		assert.Equal(t, "prom1", prom.Name)
		assert.Equal(t, "prometheus", prom.Type)
		require.NotNil(t, prom.Config)
		assert.Equal(t, "http://prometheus:9090/", prom.Config["url"])
		assert.Equal(t, "bearer", prom.Config["auth"])
		assert.Equal(t, "test-prometheus-token", prom.Config["token"])

		// Verify prometheus apps
		apps, ok := prom.Config["apps"].([]interface{})
		require.True(t, ok)
		require.Len(t, apps, 1)
		app := apps[0].(map[interface{}]interface{})
		assert.Equal(t, "test app", app["name"])
		assert.Equal(t, "test location", app["location"])
		assert.Equal(t, `up{instance="test"}`, app["metric"])

		// Verify site source
		site := getSourceByType(cfg.Sources, "site")
		require.NotNil(t, site)
		assert.Equal(t, "site-a", site.Name)
		assert.Equal(t, "site", site.Type)
		require.NotNil(t, site.Config)
		assert.Equal(t, "https://test-site:3030", site.Config["url"])
		assert.Equal(t, "test-site-token", site.Config["token"])
	})

	t.Run("load configuration with optional credentials", func(t *testing.T) {
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

sources:
  - name: "prom1"
    type: "prometheus"
    config:
      url: "http://prometheus:9090/"
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
		prom := getSourceByType(cfg.Sources, "prometheus")
		require.NotNil(t, prom)
		// Config should have url but no auth/token since they're not in the minimal config
		assert.Equal(t, "http://prometheus:9090/", prom.Config["url"])
		assert.Nil(t, prom.Config["auth"])
		assert.Nil(t, prom.Config["token"])
	})

	t.Run("load configuration with environment variables", func(t *testing.T) {
		// Test default values when environment variables are not set
		setupMainTest()

		// Create default config file
		tmpDir := t.TempDir()
		configContent := `
server_settings:
  port: "8080"

locations:
  - name: "test location"
    latitude: 31.782904
    longitude: 35.214774

sources:
  - name: "prom1"
    type: "prometheus"
    config:
      url: "http://prometheus:9090/"
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
	})
}

func TestConfigurationErrors(t *testing.T) {
	setupMainTest()

	t.Run("missing config file", func(t *testing.T) {
		// Test with missing config file
		os.Setenv("CONFIG_FILE", "nonexistent.yaml")
		defer os.Unsetenv("CONFIG_FILE")

		_, err := config.LoadConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read base file")
	})

	t.Run("invalid YAML syntax", func(t *testing.T) {
		// Create temporary test file with invalid YAML
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "invalid.yaml")
		err := os.WriteFile(configPath, []byte("invalid: yaml: content:"), 0644)
		require.NoError(t, err)

		os.Setenv("CONFIG_FILE", configPath)
		defer os.Unsetenv("CONFIG_FILE")

		_, err = config.LoadConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse base YAML")
	})

	t.Run("missing required fields", func(t *testing.T) {
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
		assert.Contains(t, err.Error(), "at least one location is required")
	})

	t.Run("invalid location coordinates", func(t *testing.T) {
		// Create temporary test file with invalid coordinates
		tmpDir := t.TempDir()
		configContent := `
server_settings:
  port: "8080"

locations:
  - name: "invalid location"
    latitude: 200.0  # Invalid latitude
    longitude: 35.214774

sources:
  - name: "test-source"
    type: "prometheus"
    config:
      url: "http://prometheus:9090"
`
		configPath := filepath.Join(tmpDir, "invalid-coords.yaml")
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		os.Setenv("CONFIG_FILE", configPath)
		defer os.Unsetenv("CONFIG_FILE")

		_, err = config.LoadConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "has invalid latitude")
	})

	t.Run("invalid source configuration", func(t *testing.T) {
		// Note: Source config validation is now deferred to source initialization
		// This test case is no longer valid as invalid source configs don't fail config loading
		// Instead, they are logged as errors and skipped during source initialization
		t.Skip("Source config validation is now deferred to source initialization")
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("getSourceByType", func(t *testing.T) {
		sources := []config.Source{
			{Name: "prom1", Type: "prometheus", Config: map[string]interface{}{"url": "http://prom:9090"}},
			{Name: "site1", Type: "site", Config: map[string]interface{}{"url": "http://site:3030"}},
		}

		// Test finding existing source
		prom := getSourceByType(sources, "prometheus")
		require.NotNil(t, prom)
		assert.Equal(t, "prom1", prom.Name)

		// Test finding non-existent source
		nonExistent := getSourceByType(sources, "nonexistent")
		assert.Nil(t, nonExistent)
	})

	t.Run("getSourceByName", func(t *testing.T) {
		sources := []config.Source{
			{Name: "prom1", Type: "prometheus", Config: map[string]interface{}{"url": "http://prom:9090"}},
			{Name: "site1", Type: "site", Config: map[string]interface{}{"url": "http://site:3030"}},
		}

		// Test finding existing source
		site := getSourceByName(sources, "site1")
		require.NotNil(t, site)
		assert.Equal(t, "site", site.Type)

		// Test finding non-existent source
		nonExistent := getSourceByName(sources, "nonexistent")
		assert.Nil(t, nonExistent)
	})
}
