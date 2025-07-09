package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// Test struct for type validation in DecodeConfig
type TestConfigStruct struct {
	URL     string `yaml:"url"`
	Timeout int    `yaml:"timeout"`
}

func TestDecodeConfig(t *testing.T) {
	tests := []struct {
		name        string
		input       map[string]interface{}
		sourceName  string
		expectError bool
		expected    interface{}
	}{
		{
			name: "valid config decode to map",
			input: map[string]interface{}{
				"url":     "https://example.com",
				"timeout": 30,
			},
			sourceName:  "test-source",
			expectError: false,
			expected: map[string]interface{}{
				"url":     "https://example.com",
				"timeout": 30,
			},
		},
		{
			name: "valid config decode to struct",
			input: map[string]interface{}{
				"url":     "https://example.com",
				"timeout": 30,
			},
			sourceName:  "test-source",
			expectError: false,
			expected: TestConfigStruct{
				URL:     "https://example.com",
				Timeout: 30,
			},
		},
		{
			name: "decode error with type mismatch",
			input: map[string]interface{}{
				"url":     "https://example.com",
				"timeout": "not-a-number",
			},
			sourceName:  "test-source",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch expected := tt.expected.(type) {
			case map[string]interface{}:
				result, err := DecodeConfig[map[string]interface{}](tt.input, tt.sourceName)
				if tt.expectError {
					if err == nil {
						t.Errorf("DecodeConfig() expected error but got none")
					}
					return
				}
				if err != nil {
					t.Errorf("DecodeConfig() unexpected error: %v", err)
					return
				}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("DecodeConfig() = %v, want %v", result, expected)
				}
			case TestConfigStruct:
				result, err := DecodeConfig[TestConfigStruct](tt.input, tt.sourceName)
				if tt.expectError {
					if err == nil {
						t.Errorf("DecodeConfig() expected error but got none")
					}
					return
				}
				if err != nil {
					t.Errorf("DecodeConfig() unexpected error: %v", err)
					return
				}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("DecodeConfig() = %v, want %v", result, expected)
				}
			default:
				// This is the error case - try to decode to struct expecting int timeout
				_, err := DecodeConfig[TestConfigStruct](tt.input, tt.sourceName)
				if tt.expectError {
					if err == nil {
						t.Errorf("DecodeConfig() expected error but got none")
					}
					return
				}
				if err != nil {
					t.Errorf("DecodeConfig() unexpected error: %v", err)
					return
				}
			}
		})
	}
}

func TestValidateLabels(t *testing.T) {
	tests := []struct {
		name    string
		labels  map[string]string
		context string
		wantErr bool
	}{
		{
			name: "valid labels",
			labels: map[string]string{
				"env":     "production",
				"version": "1.0.0",
			},
			context: "test",
			wantErr: false,
		},
		{
			name: "empty label key",
			labels: map[string]string{
				"": "value",
			},
			context: "test",
			wantErr: true,
		},
		{
			name: "whitespace only label key",
			labels: map[string]string{
				"   ": "value",
			},
			context: "test",
			wantErr: true,
		},
		{
			name: "label key too long",
			labels: map[string]string{
				string(make([]byte, 101)): "value",
			},
			context: "test",
			wantErr: true,
		},
		{
			name: "label value too long",
			labels: map[string]string{
				"key": string(make([]byte, 501)),
			},
			context: "test",
			wantErr: true,
		},
		{
			name: "reserved character in key",
			labels: map[string]string{
				"key&": "value",
			},
			context: "test",
			wantErr: true,
		},
		{
			name: "reserved character in value",
			labels: map[string]string{
				"key": "val&ue",
			},
			context: "test",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLabels(tt.labels, tt.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateLabels() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				ServerSettings: ServerSettings{
					Port: "8080",
					Labels: map[string]string{
						"env": "test",
					},
				},
				Locations: []Location{
					{
						Name:      "Test Location",
						Latitude:  40.7128,
						Longitude: -74.0060,
					},
				},
				Sources: []Source{
					{
						Name: "test-source",
						Type: "http",
						Config: map[string]interface{}{
							"url": "https://example.com",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "no locations",
			config: &Config{
				ServerSettings: ServerSettings{
					Port: "8080",
				},
				Locations: []Location{},
				Sources: []Source{
					{
						Name: "test-source",
						Type: "http",
						Config: map[string]interface{}{
							"url": "https://example.com",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid latitude",
			config: &Config{
				ServerSettings: ServerSettings{
					Port: "8080",
				},
				Locations: []Location{
					{
						Name:      "Invalid Location",
						Latitude:  91.0,
						Longitude: -74.0060,
					},
				},
				Sources: []Source{
					{
						Name: "test-source",
						Type: "http",
						Config: map[string]interface{}{
							"url": "https://example.com",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid longitude",
			config: &Config{
				ServerSettings: ServerSettings{
					Port: "8080",
				},
				Locations: []Location{
					{
						Name:      "Invalid Location",
						Latitude:  40.7128,
						Longitude: 181.0,
					},
				},
				Sources: []Source{
					{
						Name: "test-source",
						Type: "http",
						Config: map[string]interface{}{
							"url": "https://example.com",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty source name",
			config: &Config{
				ServerSettings: ServerSettings{
					Port: "8080",
				},
				Locations: []Location{
					{
						Name:      "Test Location",
						Latitude:  40.7128,
						Longitude: -74.0060,
					},
				},
				Sources: []Source{
					{
						Name: "",
						Type: "http",
						Config: map[string]interface{}{
							"url": "https://example.com",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "duplicate source names",
			config: &Config{
				ServerSettings: ServerSettings{
					Port: "8080",
				},
				Locations: []Location{
					{
						Name:      "Test Location",
						Latitude:  40.7128,
						Longitude: -74.0060,
					},
				},
				Sources: []Source{
					{
						Name: "duplicate",
						Type: "http",
						Config: map[string]interface{}{
							"url": "https://example.com",
						},
					},
					{
						Name: "duplicate",
						Type: "tcp",
						Config: map[string]interface{}{
							"host": "example.com",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty source type",
			config: &Config{
				ServerSettings: ServerSettings{
					Port: "8080",
				},
				Locations: []Location{
					{
						Name:      "Test Location",
						Latitude:  40.7128,
						Longitude: -74.0060,
					},
				},
				Sources: []Source{
					{
						Name: "test-source",
						Type: "",
						Config: map[string]interface{}{
							"url": "https://example.com",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid server labels",
			config: &Config{
				ServerSettings: ServerSettings{
					Port: "8080",
					Labels: map[string]string{
						"key&": "value",
					},
				},
				Locations: []Location{
					{
						Name:      "Test Location",
						Latitude:  40.7128,
						Longitude: -74.0060,
					},
				},
				Sources: []Source{
					{
						Name: "test-source",
						Type: "http",
						Config: map[string]interface{}{
							"url": "https://example.com",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid source labels",
			config: &Config{
				ServerSettings: ServerSettings{
					Port: "8080",
				},
				Locations: []Location{
					{
						Name:      "Test Location",
						Latitude:  40.7128,
						Longitude: -74.0060,
					},
				},
				Sources: []Source{
					{
						Name: "test-source",
						Type: "http",
						Labels: map[string]string{
							"key&": "value",
						},
						Config: map[string]interface{}{
							"url": "https://example.com",
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue string
		expected     string
	}{
		{
			name:         "environment variable exists",
			envKey:       "TEST_VAR",
			envValue:     "test_value",
			defaultValue: "default",
			expected:     "test_value",
		},
		{
			name:         "environment variable does not exist",
			envKey:       "NON_EXISTENT_VAR",
			envValue:     "",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "empty environment variable",
			envKey:       "EMPTY_VAR",
			envValue:     "",
			defaultValue: "default",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment before test
			os.Unsetenv(tt.envKey)

			if tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			}

			result := GetEnv(tt.envKey, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("GetEnv() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	// Create temporary directory for test files
	tempDir := t.TempDir()

	// Create test config file
	configFile := filepath.Join(tempDir, "config.yaml")
	configContent := `
server_settings:
  port: "8080"
  labels:
    env: "test"
scraping:
  interval: "30s"
  timeout: "10s"
  max_parallel: 5
documentation:
  title: "Test API"
  url: "https://example.com/docs"
locations:
  - name: "Test Location"
    latitude: 40.7128
    longitude: -74.0060
sources:
  - name: "test-source"
    type: "http"
    config:
      url: "https://example.com"
`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Test successful config loading
	t.Run("load config successfully", func(t *testing.T) {
		// Set environment variables for the test
		os.Setenv("CONFIG_FILE", configFile)
		os.Setenv("CREDENTIALS_FILE", "non-existent.yaml")
		defer func() {
			os.Unsetenv("CONFIG_FILE")
			os.Unsetenv("CREDENTIALS_FILE")
		}()

		config, err := LoadConfig()
		if err != nil {
			t.Errorf("LoadConfig() unexpected error: %v", err)
			return
		}

		if config.ServerSettings.Port != "8080" {
			t.Errorf("Expected port 8080, got %s", config.ServerSettings.Port)
		}

		if len(config.Locations) != 1 {
			t.Errorf("Expected 1 location, got %d", len(config.Locations))
		}

		if len(config.Sources) != 1 {
			t.Errorf("Expected 1 source, got %d", len(config.Sources))
		}

		if config.Sources[0].Name != "test-source" {
			t.Errorf("Expected source name 'test-source', got %s", config.Sources[0].Name)
		}
	})

	// Test config loading with invalid file
	t.Run("load config with invalid file", func(t *testing.T) {
		os.Setenv("CONFIG_FILE", "non-existent.yaml")
		defer os.Unsetenv("CONFIG_FILE")

		_, err := LoadConfig()
		if err == nil {
			t.Error("Expected error for non-existent config file")
		}
	})

	// Test config validation failure
	t.Run("load config with validation failure", func(t *testing.T) {
		invalidConfigFile := filepath.Join(tempDir, "invalid-config.yaml")
		invalidConfigContent := `
server_settings:
  port: "8080"
sources:
  - name: "test-source"
    type: "http"
    config:
      url: "https://example.com"
`
		err := os.WriteFile(invalidConfigFile, []byte(invalidConfigContent), 0644)
		if err != nil {
			t.Fatal(err)
		}

		os.Setenv("CONFIG_FILE", invalidConfigFile)
		defer os.Unsetenv("CONFIG_FILE")

		_, err = LoadConfig()
		if err == nil {
			t.Error("Expected validation error for config without locations")
		}
	})
}
