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
		require.NotNil(t, prom.Config)
		assert.Equal(t, "http://prometheus:9090/", prom.Config["url"])
		assert.Equal(t, "bearer", prom.Config["auth"])
		assert.Equal(t, "test-prometheus-token", prom.Config["token"])

		// Test prometheus apps
		apps, ok := prom.Config["apps"].([]interface{})
		require.True(t, ok)
		require.Len(t, apps, 1)
		app := apps[0].(map[interface{}]interface{})
		assert.Equal(t, "test app", app["name"])
		assert.Equal(t, "test location", app["location"])
		assert.Equal(t, `up{instance="test"}`, app["metric"])

		// Test site source
		site := cfg.Sources[1]
		assert.Equal(t, "site-a", site.Name)
		assert.Equal(t, "site", site.Type)
		require.NotNil(t, site.Config)
		assert.Equal(t, "https://test-site:3030", site.Config["url"])
		assert.Equal(t, "test-site-token", site.Config["token"])
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
    config:
      url: "http://prometheus:9090/"
  - name: "site-a"
    type: "site"
    config:
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
		assert.Equal(t, "http://prometheus:9090/", cfg.Sources[0].Config["url"])
		assert.Nil(t, cfg.Sources[0].Config["auth"])
		assert.Nil(t, cfg.Sources[0].Config["token"])
		assert.Nil(t, cfg.Sources[0].Config["apps"])

		siteSource := getSiteSource(cfg.Sources)
		if siteSource != nil {
			assert.Equal(t, "https://test-site:3030", siteSource.Config["url"])
			assert.Nil(t, siteSource.Config["token"])
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
    config:
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
					Config: map[string]interface{}{
						"url": "http://prometheus:9090/",
					},
				},
				{
					Name: "site-a",
					Type: "site",
					Config: map[string]interface{}{
						"url": "https://test-site:3030",
					},
				},
			},
		}

		credentials := &Config{
			ServerSettings: ServerSettings{
				Token: "test-server-token",
			},
			Sources: []Source{
				{
					Name: "prom1",
					Config: map[string]interface{}{
						"auth":  "bearer",
						"token": "test-prometheus-token",
					},
				},
				{
					Name: "site-a",
					Config: map[string]interface{}{
						"token": "test-site-token",
					},
				},
			},
		}

		mergeCredentials(mainConfig, credentials)

		// Verify server token was merged
		assert.Equal(t, "test-server-token", mainConfig.ServerSettings.Token)

		// Verify Prometheus source credentials were merged
		assert.Equal(t, "bearer", mainConfig.Sources[0].Config["auth"])
		assert.Equal(t, "test-prometheus-token", mainConfig.Sources[0].Config["token"])

		// Verify site source token was merged
		siteSource := getSiteSource(mainConfig.Sources)
		require.NotNil(t, siteSource)
		assert.Equal(t, "test-site-token", siteSource.Config["token"])
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
					Config: map[string]interface{}{
						"url": "http://prometheus:9090/",
					},
				},
				{
					Name: "site-a",
					Type: "site",
					Config: map[string]interface{}{
						"url": "https://test-site:3030",
					},
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
					Config: map[string]interface{}{
						"url": "http://prometheus:9090/",
					},
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
					Config: map[string]interface{}{
						"url": "http://prometheus:9090/",
					},
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
					Config: map[string]interface{}{
						"url": "http://prometheus:9090/",
					},
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
					Config: map[string]interface{}{
						"url": "http://prometheus:9090/",
					},
				},
			},
		}

		err := validateConfig(invalidConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "source name is required")
	})

	t.Run("source missing config", func(t *testing.T) {
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
					// Missing config section
				},
			},
		}

		err := validateConfig(invalidConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "config is required")
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
					Config: map[string]interface{}{
						"url": "http://prometheus:9090/",
					},
				},
				{
					Name: "prom1", // Duplicate name
					Type: "prometheus",
					Config: map[string]interface{}{
						"url": "http://prometheus2:9090/",
					},
				},
			},
		}

		err := validateConfig(invalidConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate source name")
	})
}

// Test suite for label functionality
func TestLabelsSupport(t *testing.T) {
	t.Run("successful config load with labels", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create main config file with labels at all levels
		configContent := `
server_settings:
  port: "8080"
  sync_enable: true
  labels:
    server: "serverA"
    environment: "production"

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
    labels:
      instance: "prom1"
      arc: "x86"
    config:
      url: "http://prometheus:9090/"
      apps:
        - name: "test app"
          location: "test location"
          metric: 'up{instance="test"}'
          labels:
            cluster: "prod"
            network: "pvc0213"
  - name: "site-a"
    type: "site"
    labels:
      region: "us-west"
    config:
      url: "https://test-site:3030"
`
		configPath := filepath.Join(tmpDir, "config.yaml")
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		// Create credentials file with additional labels
		credentialsContent := `
server_settings:
  token: "test-server-token"
  labels:
    secrets: "encrypted"

sources:
  - name: "prom1"
    config:
      token: "test-prometheus-token"
    labels:
      auth: "bearer"
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

		// Test server labels (merged from main config and credentials)
		require.NotNil(t, cfg.ServerSettings.Labels)
		assert.Equal(t, "serverA", cfg.ServerSettings.Labels["server"])
		assert.Equal(t, "production", cfg.ServerSettings.Labels["environment"])
		assert.Equal(t, "encrypted", cfg.ServerSettings.Labels["secrets"]) // From credentials

		// Test source labels
		prom := cfg.Sources[0]
		require.NotNil(t, prom.Labels)
		assert.Equal(t, "prom1", prom.Labels["instance"])
		assert.Equal(t, "x86", prom.Labels["arc"])
		assert.Equal(t, "bearer", prom.Labels["auth"]) // From credentials

		// Test app labels (accessed from config map)
		apps, ok := prom.Config["apps"].([]interface{})
		require.True(t, ok)
		require.Len(t, apps, 1)

		appMap, ok := apps[0].(map[interface{}]interface{})
		require.True(t, ok)

		appLabels, ok := appMap["labels"].(map[interface{}]interface{})
		require.True(t, ok)
		assert.Equal(t, "prod", appLabels["cluster"])
		assert.Equal(t, "pvc0213", appLabels["network"])

		// Test site source labels
		site := cfg.Sources[1]
		require.NotNil(t, site.Labels)
		assert.Equal(t, "us-west", site.Labels["region"])
	})

	t.Run("backward compatibility - config without labels", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create config file without any labels (existing format)
		configContent := `
server_settings:
  port: "8080"
  sync_enable: true

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
      apps:
        - name: "test app"
          location: "test location"
          metric: 'up{instance="test"}'
`
		configPath := filepath.Join(tmpDir, "config.yaml")
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		os.Setenv("CONFIG_FILE", configPath)
		defer os.Unsetenv("CONFIG_FILE")

		// Should load successfully with nil/empty labels
		cfg, err := LoadConfig()
		require.NoError(t, err)
		assert.NotNil(t, cfg)

		// Labels should be nil (not causing issues)
		assert.Nil(t, cfg.ServerSettings.Labels)
		assert.Nil(t, cfg.Sources[0].Labels)

		// App labels should also be nil (accessed from config map)
		apps, ok := cfg.Sources[0].Config["apps"].([]interface{})
		require.True(t, ok)
		require.Len(t, apps, 1)

		appMap, ok := apps[0].(map[interface{}]interface{})
		require.True(t, ok)

		// Labels should be nil in app
		assert.Nil(t, appMap["labels"])
	})
}

func TestValidateLabels(t *testing.T) {
	t.Run("valid labels", func(t *testing.T) {
		validLabels := map[string]string{
			"environment": "production",
			"cluster":     "prod",
			"team":        "platformA",
		}

		err := validateLabels(validLabels, "test context")
		assert.NoError(t, err)
	})

	t.Run("empty labels map", func(t *testing.T) {
		err := validateLabels(nil, "test context")
		assert.NoError(t, err)

		err = validateLabels(map[string]string{}, "test context")
		assert.NoError(t, err)
	})

	t.Run("empty label key", func(t *testing.T) {
		invalidLabels := map[string]string{
			"":     "value",
			"test": "valid",
		}

		err := validateLabels(invalidLabels, "test context")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "label key cannot be empty")
	})

	t.Run("whitespace only label key", func(t *testing.T) {
		invalidLabels := map[string]string{
			"   ": "value",
		}

		err := validateLabels(invalidLabels, "test context")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "label key cannot be empty")
	})

	t.Run("label key too long", func(t *testing.T) {
		longKey := make([]byte, 101) // 101 characters, over the 100 limit
		for i := range longKey {
			longKey[i] = 'a'
		}

		invalidLabels := map[string]string{
			string(longKey): "value",
		}

		err := validateLabels(invalidLabels, "test context")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exceeds maximum length of 100 characters")
	})

	t.Run("label value too long", func(t *testing.T) {
		longValue := make([]byte, 501) // 501 characters, over the 500 limit
		for i := range longValue {
			longValue[i] = 'a'
		}

		invalidLabels := map[string]string{
			"test": string(longValue),
		}

		err := validateLabels(invalidLabels, "test context")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exceeds maximum length of 500 characters")
	})

	// Test reserved characters in keys
	reservedChars := []string{"&", "=", "?", "#", "/", ":"}
	for _, char := range reservedChars {
		t.Run(fmt.Sprintf("reserved character %q in key", char), func(t *testing.T) {
			invalidLabels := map[string]string{
				"test" + char + "key": "value",
			}

			err := validateLabels(invalidLabels, "test context")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), fmt.Sprintf("contains reserved character %q", char))
		})
	}

	// Test reserved characters in values
	for _, char := range reservedChars {
		t.Run(fmt.Sprintf("reserved character %q in value", char), func(t *testing.T) {
			invalidLabels := map[string]string{
				"testkey": "value" + char + "test",
			}

			err := validateLabels(invalidLabels, "test context")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), fmt.Sprintf("contains reserved character %q", char))
		})
	}
}

func TestMergeCredentialsWithLabels(t *testing.T) {
	t.Run("merge server labels", func(t *testing.T) {
		mainConfig := &Config{
			ServerSettings: ServerSettings{
				Port: "8080",
				Labels: map[string]string{
					"environment": "production",
					"server":      "main",
				},
			},
		}

		credentials := &Config{
			ServerSettings: ServerSettings{
				Token: "test-server-token",
				Labels: map[string]string{
					"secrets": "encrypted",
					"server":  "override", // Should override main config
				},
			},
		}

		mergeCredentials(mainConfig, credentials)

		// Verify server token was merged
		assert.Equal(t, "test-server-token", mainConfig.ServerSettings.Token)

		// Verify labels were merged correctly
		require.NotNil(t, mainConfig.ServerSettings.Labels)
		assert.Equal(t, "production", mainConfig.ServerSettings.Labels["environment"]) // From main
		assert.Equal(t, "override", mainConfig.ServerSettings.Labels["server"])        // Overridden by credentials
		assert.Equal(t, "encrypted", mainConfig.ServerSettings.Labels["secrets"])      // From credentials
	})

	t.Run("merge source labels", func(t *testing.T) {
		mainConfig := &Config{
			Sources: []Source{
				{
					Name: "prom1",
					Type: "prometheus",
					Config: map[string]interface{}{
						"url": "http://prometheus:9090/",
					},
					Labels: map[string]string{
						"instance": "prom1",
						"type":     "main",
					},
				},
			},
		}

		credentials := &Config{
			Sources: []Source{
				{
					Name: "prom1",
					Config: map[string]interface{}{
						"auth":  "bearer",
						"token": "test-prometheus-token",
					},
					Labels: map[string]string{
						"auth": "bearer",
						"type": "override", // Should override main config
					},
				},
			},
		}

		mergeCredentials(mainConfig, credentials)

		// Verify source credentials were merged
		assert.Equal(t, "bearer", mainConfig.Sources[0].Config["auth"])
		assert.Equal(t, "test-prometheus-token", mainConfig.Sources[0].Config["token"])

		// Verify labels were merged correctly
		require.NotNil(t, mainConfig.Sources[0].Labels)
		assert.Equal(t, "prom1", mainConfig.Sources[0].Labels["instance"]) // From main
		assert.Equal(t, "override", mainConfig.Sources[0].Labels["type"])  // Overridden by credentials
		assert.Equal(t, "bearer", mainConfig.Sources[0].Labels["auth"])    // From credentials
	})

	t.Run("initialize nil labels maps", func(t *testing.T) {
		mainConfig := &Config{
			ServerSettings: ServerSettings{
				Port: "8080",
				// No labels initially
			},
			Sources: []Source{
				{
					Name: "prom1",
					Type: "prometheus",
					Config: map[string]interface{}{
						"url": "http://prometheus:9090/",
					},
					// No labels initially
				},
			},
		}

		credentials := &Config{
			ServerSettings: ServerSettings{
				Labels: map[string]string{
					"from": "credentials",
				},
			},
			Sources: []Source{
				{
					Name: "prom1",
					Labels: map[string]string{
						"auth": "bearer",
					},
				},
			},
		}

		mergeCredentials(mainConfig, credentials)

		// Labels should be initialized and populated
		require.NotNil(t, mainConfig.ServerSettings.Labels)
		assert.Equal(t, "credentials", mainConfig.ServerSettings.Labels["from"])

		require.NotNil(t, mainConfig.Sources[0].Labels)
		assert.Equal(t, "bearer", mainConfig.Sources[0].Labels["auth"])
	})
}

func TestValidateConfigWithLabels(t *testing.T) {
	t.Run("valid config with labels", func(t *testing.T) {
		validConfig := &Config{
			ServerSettings: ServerSettings{
				Labels: map[string]string{
					"environment": "production",
					"server":      "main",
				},
			},
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
					Config: map[string]interface{}{
						"url": "http://prometheus:9090/",
						"apps": []map[string]interface{}{
							{
								"name":     "test app",
								"location": "test location",
								"metric":   "up",
								"labels": map[string]string{
									"cluster": "prod",
									"team":    "platform",
								},
							},
						},
					},
					Labels: map[string]string{
						"instance": "prom1",
						"type":     "prometheus",
					},
				},
			},
		}

		err := validateConfig(validConfig)
		assert.NoError(t, err)
	})

	t.Run("invalid server labels", func(t *testing.T) {
		invalidConfig := &Config{
			ServerSettings: ServerSettings{
				Labels: map[string]string{
					"test&key": "value", // Contains reserved character
				},
			},
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
					Config: map[string]interface{}{
						"url": "http://prometheus:9090/",
					},
				},
			},
		}

		err := validateConfig(invalidConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "server settings")
		assert.Contains(t, err.Error(), "reserved character")
	})

	t.Run("invalid source labels", func(t *testing.T) {
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
					Config: map[string]interface{}{
						"url": "http://prometheus:9090/",
					},
					Labels: map[string]string{
						"valid": "label",
						"":      "empty key", // Invalid empty key
					},
				},
			},
		}

		err := validateConfig(invalidConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "source prom1")
		assert.Contains(t, err.Error(), "empty")
	})
}
