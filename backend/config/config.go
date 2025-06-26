package config

import (
	"fmt"
	"os"
	"site-availability/logging"

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
	Port         string `yaml:"port"`
	CustomCAPath string `yaml:"custom_ca_path"`
	SyncEnable   bool   `yaml:"sync_enable"`
	Token        string `yaml:"token"` // Optional token for server authentication
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
	Name  string `yaml:"name"`
	Type  string `yaml:"type"`
	URL   string `yaml:"url"`
	Token string `yaml:"token"`
	Auth  string `yaml:"auth"`
	Apps  []App  `yaml:"apps"`
}

type App struct {
	Name     string `yaml:"name"`
	Location string `yaml:"location"`
	Metric   string `yaml:"metric"`
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

	// Merge source credentials
	for i, source := range config.Sources {
		for _, credSource := range credentials.Sources {
			if source.Name == credSource.Name {
				config.Sources[i].Auth = credSource.Auth
				config.Sources[i].Token = credSource.Token
				break
			}
		}
	}
}

// validateConfig performs validation on the combined configuration
func validateConfig(config *Config) error {
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

		if source.URL == "" {
			return fmt.Errorf("source configuration error: URL is required for source %s", source.Name)
		}

		appNames := make(map[string]bool)
		for _, app := range source.Apps {
			if app.Name == "" {
				return fmt.Errorf("app configuration error: app name is required for source %s", source.Name)
			}
			if _, exists := appNames[app.Name]; exists {
				return fmt.Errorf("app configuration error: duplicate app name %q in source %s", app.Name, source.Name)
			}
			appNames[app.Name] = true
		}
	}

	return nil
}

func GetEnv(name, defaultValue string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return defaultValue
}
