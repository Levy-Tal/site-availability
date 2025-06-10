package config

import (
	"fmt"
	"os"
	"site-availability/logging"
	"time"

	"gopkg.in/yaml.v2"
)

// Config structure defines the combined config.yaml and credentials.yaml structure
type Config struct {
	ServerSettings    ServerSettings     `yaml:"server_settings"`
	Scraping          ScrapingSettings   `yaml:"scraping"`
	Documentation     Documentation      `yaml:"documentation"`
	PrometheusServers []PrometheusServer `yaml:"prometheus_servers"`
	Locations         []Location         `yaml:"locations"`
	Applications      []Application      `yaml:"applications"`
	Sites             []Site             `yaml:"sites"`
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

type PrometheusServer struct {
	Name  string `yaml:"name"`
	URL   string `yaml:"url"`
	Auth  string `yaml:"auth"`  // Optional authentication type
	Token string `yaml:"token"` // Optional token for authentication
}

type Location struct {
	Name      string  `yaml:"name" json:"name"`
	Latitude  float64 `yaml:"latitude" json:"latitude"`
	Longitude float64 `yaml:"longitude" json:"longitude"`
}

type Application struct {
	Name       string `yaml:"name"`
	Location   string `yaml:"location"`
	Metric     string `yaml:"metric"`
	Prometheus string `yaml:"prometheus"`
}

type Site struct {
	Name          string    `yaml:"name"`
	URL           string    `yaml:"url"`
	CheckInterval string    `yaml:"check_interval"`
	Timeout       string    `yaml:"timeout"`
	Enabled       bool      `yaml:"enabled"`
	CustomCAPath  string    `yaml:"custom_ca_path"`
	Token         string    `yaml:"token"` // Optional token for site authentication
	LastSuccess   time.Time `yaml:"-"`     // Last successful sync
	ErrorCount    int       `yaml:"-"`     // Number of consecutive errors
	LastError     string    `yaml:"-"`     // Last error message
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

	// Merge Prometheus server credentials
	for i, promServer := range config.PrometheusServers {
		for _, credPromServer := range credentials.PrometheusServers {
			if promServer.Name == credPromServer.Name {
				config.PrometheusServers[i].Auth = credPromServer.Auth
				config.PrometheusServers[i].Token = credPromServer.Token
				break
			}
		}
	}

	// Merge site tokens
	for i, site := range config.Sites {
		for _, credSite := range credentials.Sites {
			if site.Name == credSite.Name {
				config.Sites[i].Token = credSite.Token
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

	// Validate that at least one Prometheus server is provided
	if len(config.PrometheusServers) == 0 {
		return fmt.Errorf("config validation error: at least one Prometheus server is required")
	}

	// Validate sites configuration
	for _, site := range config.Sites {
		if site.Name == "" {
			return fmt.Errorf("site configuration error: site name is required")
		}
		if site.URL == "" {
			return fmt.Errorf("site configuration error: URL is required for site %s", site.Name)
		}
		if site.CheckInterval == "" {
			site.CheckInterval = "1m" // Default check interval
		}
		if site.Timeout == "" {
			site.Timeout = "5s" // Default timeout
		}
		// If CustomCAPath is set, validate it exists
		if site.CustomCAPath != "" {
			if _, err := os.Stat(site.CustomCAPath); err != nil {
				return fmt.Errorf("site configuration error: CA certificate not found for site %s: %w", site.Name, err)
			}
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
