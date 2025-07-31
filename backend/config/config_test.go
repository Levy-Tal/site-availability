package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
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

func TestValidateSessionTimeout(t *testing.T) {
	tests := []struct {
		name           string
		sessionTimeout string
		shouldFail     bool
		description    string
	}{
		// Valid durations
		{"valid_hours", "12h", false, "standard hours format"},
		{"valid_minutes", "30m", false, "standard minutes format"},
		{"valid_seconds", "60s", false, "seconds should be accepted"},
		{"valid_mixed", "1h30m", false, "mixed hours and minutes"},
		{"valid_decimal", "1.5h", false, "decimal hours"},
		{"valid_milliseconds", "500ms", false, "milliseconds"},
		{"valid_nanoseconds", "1000ns", false, "nanoseconds"},
		{"valid_microseconds", "1000μs", false, "microseconds"},
		{"valid_complex", "2h30m45s", false, "complex duration"},

		// Invalid durations
		{"invalid_malformed_h", "abch", true, "malformed with h suffix"},
		{"invalid_malformed_m", "xyz123m", true, "malformed with m suffix"},
		{"invalid_no_unit", "123", true, "number without unit"},
		{"invalid_empty", "", false, "empty string should be allowed"},
		{"invalid_spaces", " 12h ", true, "duration with spaces"},
		{"valid_negative", "-5h", false, "negative duration is valid in Go"},
		{"invalid_text", "twelve hours", true, "text instead of number"},
		{"invalid_mixed_bad", "1h2x", true, "mixed with invalid unit"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serverSettings := &ServerSettings{
				SessionTimeout: tt.sessionTimeout,
			}

			err := validateAuthConfig(serverSettings)

			if tt.shouldFail && err == nil {
				t.Errorf("Expected validation to fail for %s (%s), but it passed",
					tt.sessionTimeout, tt.description)
			}

			if !tt.shouldFail && err != nil {
				t.Errorf("Expected validation to pass for %s (%s), but got error: %v",
					tt.sessionTimeout, tt.description, err)
			}

			// For failing cases, ensure the error message is helpful
			if tt.shouldFail && err != nil {
				errorMsg := err.Error()
				if !strings.Contains(errorMsg, "session_timeout") {
					t.Errorf("Error message should mention 'session_timeout': %s", errorMsg)
				}
				if !strings.Contains(errorMsg, tt.sessionTimeout) {
					t.Errorf("Error message should include the invalid value '%s': %s",
						tt.sessionTimeout, errorMsg)
				}
			}
		})
	}
}
