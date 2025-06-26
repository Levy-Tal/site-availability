package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getSiteSource is a helper function to find a site source from a list of sources
func getSiteSource(sources []Source) *Source {
	for i := range sources {
		if sources[i].Type == "site" {
			return &sources[i]
		}
	}
	return nil
}

// Test suite for LoadConfig function
func TestLoadConfig(t *testing.T) {
	t.Run("successful config load", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a valid server CA file
		serverCAPath := filepath.Join(tmpDir, "ca.crt")
		err := os.WriteFile(serverCAPath, []byte("test server certificate"), 0644)
		require.NoError(t, err)

		// Create main config file with proper YAML structure
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
    url: "http://prometheus:9090/"
    auth: "bearer"
    token: "test-prometheus-token"
    apps:
      - name: "test app"
        location: "test location"
        metric: 'up{instance="test"}'
  - name: "site-a"
    type: "site"
    url: "https://test-site:3030"
    token: "test-site-token"
`, serverCAPath)

		configPath := filepath.Join(tmpDir, "config.yaml")
		err = os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		// Create credentials file
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

		// Load and test configuration
		cfg, err := LoadConfig()
		require.NoError(t, err)
		assert.NotNil(t, cfg)

		// Test server settings
		assert.Equal(t, "8080", cfg.ServerSettings.Port)
		assert.True(t, cfg.ServerSettings.SyncEnable)
		assert.Equal(t, serverCAPath, cfg.ServerSettings.CustomCAPath)
		assert.Equal(t, "test-server-token", cfg.ServerSettings.Token)

		// Test scraping settings
		assert.Equal(t, "60s", cfg.Scraping.Interval)
		assert.Equal(t, "15s", cfg.Scraping.Timeout)
		assert.Equal(t, 10, cfg.Scraping.MaxParallel)

		// Test documentation
		assert.Equal(t, "Test Documentation", cfg.Documentation.Title)
		assert.Equal(t, "https://test.example.com/docs", cfg.Documentation.URL)

		// Test locations
		require.Len(t, cfg.Locations, 1)
		assert.Equal(t, "test location", cfg.Locations[0].Name)
		assert.Equal(t, 31.782904, cfg.Locations[0].Latitude)
		assert.Equal(t, 35.214774, cfg.Locations[0].Longitude)

		// Test sources
		require.Len(t, cfg.Sources, 2)

		// Test Prometheus source
		prom := cfg.Sources[0]
		assert.Equal(t, "prom1", prom.Name)
		assert.Equal(t, "prometheus", prom.Type)
		assert.Equal(t, "http://prometheus:9090/", prom.URL)
		assert.Equal(t, "bearer", prom.Auth)
		assert.Equal(t, "test-prometheus-token", prom.Token)
		require.Len(t, prom.Apps, 1)
		assert.Equal(t, "test app", prom.Apps[0].Name)
		assert.Equal(t, "test location", prom.Apps[0].Location)
		assert.Equal(t, `up{instance="test"}`, prom.Apps[0].Metric)

		// Test site source
		site := cfg.Sources[1]
		assert.Equal(t, "site-a", site.Name)
		assert.Equal(t, "site", site.Type)
		assert.Equal(t, "https://test-site:3030", site.URL)
		assert.Equal(t, "test-site-token", site.Token)
	})

	t.Run("missing config file", func(t *testing.T) {
		os.Setenv("CONFIG_FILE", "nonexistent.yaml")
		os.Setenv("CREDENTIALS_FILE", "nonexistent-credentials.yaml")
		defer func() {
			os.Unsetenv("CONFIG_FILE")
			os.Unsetenv("CREDENTIALS_FILE")
		}()

		_, err := LoadConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load config file")
	})

	t.Run("invalid YAML format", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "invalid.yaml")
		err := os.WriteFile(configPath, []byte("invalid: yaml: content:"), 0644)
		require.NoError(t, err)

		os.Setenv("CONFIG_FILE", configPath)
		defer os.Unsetenv("CONFIG_FILE")

		_, err = LoadConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode file")
	})

	t.Run("missing required fields", func(t *testing.T) {
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

		_, err = LoadConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least one location is required")
	})

	t.Run("invalid location coordinates", func(t *testing.T) {
		tmpDir := t.TempDir()
		configContent := `
server_settings:
  port: "8080"

locations:
  - name: "invalid location"
    latitude: 200.0
    longitude: 35.214774

sources:
  - name: "prom1"
    type: "prometheus"
    url: "http://prometheus:9090/"
`
		configPath := filepath.Join(tmpDir, "invalid-coords.yaml")
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		os.Setenv("CONFIG_FILE", configPath)
		defer os.Unsetenv("CONFIG_FILE")

		_, err = LoadConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid latitude")
	})

	t.Run("optional credentials handling", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create minimal valid config
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
    url: "http://prometheus:9090/"
  - name: "site-a"
    type: "site"
    url: "https://test-site:3030"
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

		cfg, err := LoadConfig()
		require.NoError(t, err)
		assert.NotNil(t, cfg)

		// Verify optional fields are empty when not provided
		assert.Empty(t, cfg.ServerSettings.Token)
		assert.Empty(t, cfg.Sources[0].Auth)
		assert.Empty(t, cfg.Sources[0].Token)
		assert.Empty(t, cfg.Sources[0].Apps)

		siteSource := getSiteSource(cfg.Sources)
		if siteSource != nil {
			assert.Empty(t, siteSource.Token)
		}
	})

	t.Run("environment variables", func(t *testing.T) {
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
    url: "http://prometheus:9090/"
`
		configPath := filepath.Join(tmpDir, "test-config.yaml")
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		os.Setenv("CONFIG_FILE", configPath)
		defer os.Unsetenv("CONFIG_FILE")

		cfg, err := LoadConfig()
		require.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "8080", cfg.ServerSettings.Port)
	})
}

// Test GetEnv utility function
func TestGetEnv(t *testing.T) {
	t.Run("environment variable exists", func(t *testing.T) {
		os.Setenv("TEST_VAR", "test_value")
		defer os.Unsetenv("TEST_VAR")

		result := GetEnv("TEST_VAR", "default_value")
		assert.Equal(t, "test_value", result)
	})

	t.Run("environment variable missing", func(t *testing.T) {
		result := GetEnv("NONEXISTENT_VAR", "default_value")
		assert.Equal(t, "default_value", result)
	})
}

// Test loadYAMLFile function
func TestLoadYAMLFile(t *testing.T) {
	t.Run("valid YAML file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configContent := `
server_settings:
  port: "8080"
`
		configPath := filepath.Join(tmpDir, "test.yaml")
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		config := &Config{}
		err = loadYAMLFile(configPath, config)
		require.NoError(t, err)
		assert.Equal(t, "8080", config.ServerSettings.Port)
	})

	t.Run("nonexistent file", func(t *testing.T) {
		config := &Config{}
		err := loadYAMLFile("nonexistent.yaml", config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to open file")
	})

	t.Run("invalid YAML content", func(t *testing.T) {
		tmpDir := t.TempDir()
		invalidPath := filepath.Join(tmpDir, "invalid.yaml")
		err := os.WriteFile(invalidPath, []byte("invalid: yaml: content:"), 0644)
		require.NoError(t, err)

		config := &Config{}
		err = loadYAMLFile(invalidPath, config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode file")
	})
}

// Test mergeCredentials function
func TestMergeCredentials(t *testing.T) {
	t.Run("merge server token and source credentials", func(t *testing.T) {
		mainConfig := &Config{
			ServerSettings: ServerSettings{
				Port: "8080",
			},
			Sources: []Source{
				{
					Name: "prom1",
					Type: "prometheus",
					URL:  "http://prometheus:9090/",
				},
				{
					Name: "site-a",
					Type: "site",
					URL:  "https://test-site:3030",
				},
			},
		}

		credentials := &Config{
			ServerSettings: ServerSettings{
				Token: "test-server-token",
			},
			Sources: []Source{
				{
					Name:  "prom1",
					Auth:  "bearer",
					Token: "test-prometheus-token",
				},
				{
					Name:  "site-a",
					Token: "test-site-token",
				},
			},
		}

		mergeCredentials(mainConfig, credentials)

		// Verify server token was merged
		assert.Equal(t, "test-server-token", mainConfig.ServerSettings.Token)

		// Verify Prometheus source credentials were merged
		assert.Equal(t, "bearer", mainConfig.Sources[0].Auth)
		assert.Equal(t, "test-prometheus-token", mainConfig.Sources[0].Token)

		// Verify site source token was merged
		siteSource := getSiteSource(mainConfig.Sources)
		require.NotNil(t, siteSource)
		assert.Equal(t, "test-site-token", siteSource.Token)
	})
}

// Test validateConfig function
func TestValidateConfig(t *testing.T) {
	t.Run("valid configuration", func(t *testing.T) {
		validConfig := &Config{
			Locations: []Location{
				{
					Name:      "test location",
					Latitude:  31.782904,
					Longitude: 35.214774,
				},
			},
			Sources: []Source{
				{
					Name: "prom1",
					Type: "prometheus",
					URL:  "http://prometheus:9090/",
				},
				{
					Name: "site-a",
					Type: "site",
					URL:  "https://test-site:3030",
				},
			},
		}

		err := validateConfig(validConfig)
		assert.NoError(t, err)
	})

	t.Run("missing locations", func(t *testing.T) {
		invalidConfig := &Config{
			Sources: []Source{
				{
					Name: "prom1",
					Type: "prometheus",
					URL:  "http://prometheus:9090/",
				},
			},
		}

		err := validateConfig(invalidConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least one location is required")
	})

	t.Run("invalid latitude", func(t *testing.T) {
		invalidConfig := &Config{
			Locations: []Location{
				{
					Name:      "invalid location",
					Latitude:  200.0, // Invalid latitude
					Longitude: 35.214774,
				},
			},
			Sources: []Source{
				{
					Name: "prom1",
					Type: "prometheus",
					URL:  "http://prometheus:9090/",
				},
			},
		}

		err := validateConfig(invalidConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid latitude")
	})

	t.Run("invalid longitude", func(t *testing.T) {
		invalidConfig := &Config{
			Locations: []Location{
				{
					Name:      "invalid location",
					Latitude:  31.782904,
					Longitude: 200.0, // Invalid longitude
				},
			},
			Sources: []Source{
				{
					Name: "prom1",
					Type: "prometheus",
					URL:  "http://prometheus:9090/",
				},
			},
		}

		err := validateConfig(invalidConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid longitude")
	})

	t.Run("source missing name", func(t *testing.T) {
		invalidConfig := &Config{
			Locations: []Location{
				{
					Name:      "test location",
					Latitude:  31.782904,
					Longitude: 35.214774,
				},
			},
			Sources: []Source{
				{
					// Missing name
					Type: "prometheus",
					URL:  "http://prometheus:9090/",
				},
			},
		}

		err := validateConfig(invalidConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "source name is required")
	})

	t.Run("source missing URL", func(t *testing.T) {
		invalidConfig := &Config{
			Locations: []Location{
				{
					Name:      "test location",
					Latitude:  31.782904,
					Longitude: 35.214774,
				},
			},
			Sources: []Source{
				{
					Name: "prom1",
					Type: "prometheus",
					// Missing URL
				},
			},
		}

		err := validateConfig(invalidConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "URL is required")
	})

	t.Run("duplicate source names", func(t *testing.T) {
		invalidConfig := &Config{
			Locations: []Location{
				{
					Name:      "test location",
					Latitude:  31.782904,
					Longitude: 35.214774,
				},
			},
			Sources: []Source{
				{
					Name: "prom1",
					Type: "prometheus",
					URL:  "http://prometheus:9090/",
				},
				{
					Name: "prom1", // Duplicate name
					Type: "prometheus",
					URL:  "http://prometheus2:9090/",
				},
			},
		}

		err := validateConfig(invalidConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate source name")
	})

	t.Run("duplicate app names in source", func(t *testing.T) {
		invalidConfig := &Config{
			Locations: []Location{
				{
					Name:      "test location",
					Latitude:  31.782904,
					Longitude: 35.214774,
				},
			},
			Sources: []Source{
				{
					Name: "prom1",
					Type: "prometheus",
					URL:  "http://prometheus:9090/",
					Apps: []App{
						{
							Name:     "app1",
							Location: "test location",
							Metric:   "up",
						},
						{
							Name:     "app1", // Duplicate app name
							Location: "test location",
							Metric:   "up2",
						},
					},
				},
			},
		}

		err := validateConfig(invalidConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate app name")
	})
}
