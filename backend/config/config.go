package config

import (
	"fmt"
	"os"
	"site-availability/logging"
	"strings"

	"gopkg.in/yaml.v2"
)

// Config structure defines the combined config.yaml and credentials.yaml structure
type Config struct {
	ServerSettings ServerSettings   `yaml:"server_settings"`
	Scraping       ScrapingSettings `yaml:"scraping"`
	Documentation  Documentation    `yaml:"documentation"`
	Locations      []Location       `yaml:"locations"`
	Sources        []Source         `yaml:"sources"`
}

type ServerSettings struct {
	Port         string            `yaml:"port"`
	CustomCAPath string            `yaml:"custom_ca_path"`
	SyncEnable   bool              `yaml:"sync_enable"`
	Token        string            `yaml:"token"`            // Optional token for server authentication
	Labels       map[string]string `yaml:"labels,omitempty"` // Server-level labels
}

type ScrapingSettings struct {
	Interval    string `yaml:"interval"`
	Timeout     string `yaml:"timeout"`
	MaxParallel int    `yaml:"max_parallel"`
}

type Documentation struct {
	Title string `yaml:"title"`
	URL   string `yaml:"url"`
}

type Location struct {
	Name      string  `yaml:"name" json:"name"`
	Latitude  float64 `yaml:"latitude" json:"latitude"`
	Longitude float64 `yaml:"longitude" json:"longitude"`
}

type Source struct {
	Name   string                 `yaml:"name"`
	Type   string                 `yaml:"type"`
	Labels map[string]string      `yaml:"labels,omitempty"` // Source-level labels
	Config map[string]interface{} `yaml:"config"`           // Holds arbitrary nested source-specific config
}

// DecodeConfig is a helper function to decode source-specific config
func DecodeConfig[T any](cfg map[string]interface{}, sourceName string) (T, error) {
	var out T
	bytes, err := yaml.Marshal(cfg)
	if err != nil {
		return out, fmt.Errorf("failed to marshal config for source %s: %w", sourceName, err)
	}

	if err := yaml.Unmarshal(bytes, &out); err != nil {
		return out, fmt.Errorf("failed to unmarshal config for source %s: %w", sourceName, err)
	}

	return out, nil
}

// Legacy structures for backward compatibility during migration
type App struct {
	Name     string            `yaml:"name"`
	Location string            `yaml:"location"`
	Metric   string            `yaml:"metric"`
	Labels   map[string]string `yaml:"labels,omitempty"` // App-level labels
}

// LoadConfig loads both the YAML configuration and credentials files
func LoadConfig() (*Config, error) {
	configFile := GetEnv("CONFIG_FILE", "config.yaml")
	credentialsFile := GetEnv("CREDENTIALS_FILE", "credentials.yaml")

	logging.Logger.WithFields(map[string]interface{}{
		"config_file":      configFile,
		"credentials_file": credentialsFile,
	}).Info("Loading configuration files")

	config := &Config{}

	// Load main config file
	if err := loadYAMLFile(configFile, config); err != nil {
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}

	// Check if credentials file exists and is not empty
	if fileInfo, err := os.Stat(credentialsFile); err == nil && fileInfo.Size() > 0 {
		credentials := &Config{}
		if err := loadYAMLFile(credentialsFile, credentials); err != nil {
			return nil, fmt.Errorf("failed to load credentials file: %w", err)
		}

		// Merge credentials into main config
		mergeCredentials(config, credentials)
	}

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}

// loadYAMLFile loads a YAML file into the provided struct
func loadYAMLFile(filePath string, target interface{}) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("failed to decode file %s: %w", filePath, err)
	}

	return nil
}

// mergeCredentials merges credentials into the main config
func mergeCredentials(config, credentials *Config) {
	// Merge server settings token if present
	if credentials.ServerSettings.Token != "" {
		config.ServerSettings.Token = credentials.ServerSettings.Token
	}

	// Merge server settings labels if present
	if credentials.ServerSettings.Labels != nil {
		if config.ServerSettings.Labels == nil {
			config.ServerSettings.Labels = make(map[string]string)
		}
		for key, value := range credentials.ServerSettings.Labels {
			config.ServerSettings.Labels[key] = value
		}
	}

	// Merge source credentials into config map and labels
	for i, source := range config.Sources {
		for _, credSource := range credentials.Sources {
			if source.Name == credSource.Name {
				// Merge the credential source's config into the main source's config
				if config.Sources[i].Config == nil {
					config.Sources[i].Config = make(map[string]interface{})
				}
				if credSource.Config != nil {
					for key, value := range credSource.Config {
						config.Sources[i].Config[key] = value
					}
				}

				// Merge source labels if present
				if credSource.Labels != nil {
					if config.Sources[i].Labels == nil {
						config.Sources[i].Labels = make(map[string]string)
					}
					for key, value := range credSource.Labels {
						config.Sources[i].Labels[key] = value
					}
				}
				break
			}
		}
	}
}

// validateLabels validates that labels don't contain reserved URL characters
// and meet basic requirements (no empty keys, reasonable lengths)
func validateLabels(labels map[string]string, context string) error {
	// Reserved characters that could interfere with URL query parameters
	reservedChars := []string{"&", "=", "?", "#", "/", ":"}

	for key, value := range labels {
		// Check for empty keys
		if strings.TrimSpace(key) == "" {
			return fmt.Errorf("label validation error in %s: label key cannot be empty", context)
		}

		// Check key length (reasonable limit)
		if len(key) > 100 {
			return fmt.Errorf("label validation error in %s: label key %q exceeds maximum length of 100 characters", context, key)
		}

		// Check value length (reasonable limit)
		if len(value) > 500 {
			return fmt.Errorf("label validation error in %s: label value for key %q exceeds maximum length of 500 characters", context, key)
		}

		// Check for reserved characters in keys
		for _, reserved := range reservedChars {
			if strings.Contains(key, reserved) {
				return fmt.Errorf("label validation error in %s: label key %q contains reserved character %q", context, key, reserved)
			}
			if strings.Contains(value, reserved) {
				return fmt.Errorf("label validation error in %s: label value %q for key %q contains reserved character %q", context, value, key, reserved)
			}
		}
	}

	return nil
}

// validateConfig performs validation on the combined configuration
func validateConfig(config *Config) error {
	// Validate server labels
	if err := validateLabels(config.ServerSettings.Labels, "server settings"); err != nil {
		return err
	}

	// Validate that at least one location is provided
	if len(config.Locations) == 0 {
		return fmt.Errorf("config validation error: at least one location is required")
	}

	// Validate latitude and longitude for each location
	for _, location := range config.Locations {
		if location.Latitude < -90 || location.Latitude > 90 {
			return fmt.Errorf("location %q has an invalid latitude: %f", location.Name, location.Latitude)
		}
		if location.Longitude < -180 || location.Longitude > 180 {
			return fmt.Errorf("location %q has an invalid longitude: %f", location.Name, location.Longitude)
		}
	}

	// Validate sources
	sourceNames := make(map[string]bool)
	for _, source := range config.Sources {
		if source.Name == "" {
			return fmt.Errorf("source configuration error: source name is required")
		}
		if _, exists := sourceNames[source.Name]; exists {
			return fmt.Errorf("source configuration error: duplicate source name %q", source.Name)
		}
		sourceNames[source.Name] = true

		if source.Type == "" {
			return fmt.Errorf("source configuration error: type is required for source %s", source.Name)
		}

		if source.Config == nil {
			return fmt.Errorf("source configuration error: config is required for source %s", source.Name)
		}

		// Validate source labels
		if err := validateLabels(source.Labels, fmt.Sprintf("source %s", source.Name)); err != nil {
			return err
		}

		// Source-specific validation will be handled by each scraper's ValidateConfig method
	}

	return nil
}

func GetEnv(name, defaultValue string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return defaultValue
}
