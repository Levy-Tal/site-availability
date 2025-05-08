package config

import (
	"fmt"
	"os"
	"site-availability/logging"

	"gopkg.in/yaml.v2"
)

// Config structure defines the config.yaml structure
type Config struct {
	ServerSettings    ServerSettings     `yaml:"server_settings"`
	Scraping          ScrapingSettings   `yaml:"scraping"`
	Documentation     Documentation      `yaml:"documentation"`
	PrometheusServers []PrometheusServer `yaml:"prometheus_servers"`
	Locations         []Location         `yaml:"locations"`
	Applications      []Application      `yaml:"applications"`
}

type ServerSettings struct {
	Port         string `yaml:"port"`
	CustomCAPath string `yaml:"custom_ca_path"`
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
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
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

// LoadConfig loads the YAML configuration from a file
func LoadConfig(filePath string) (*Config, error) {
	logging.Logger.WithField("file", filePath).Info("Loading configuration file")

	config := &Config{}

	// Open the configuration file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	// Decode the YAML file
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	// Validate that at least one location is provided
	if len(config.Locations) == 0 {
		return nil, fmt.Errorf("config validation error: at least one location is required")
	}

	// Validate latitude and longitude for each location
	for _, location := range config.Locations {
		if location.Latitude < -90 || location.Latitude > 90 {
			return nil, fmt.Errorf("location %q has an invalid latitude: %f", location.Name, location.Latitude)
		}
		if location.Longitude < -180 || location.Longitude > 180 {
			return nil, fmt.Errorf("location %q has an invalid longitude: %f", location.Name, location.Longitude)
		}
	}

	// Validate that at least one Prometheus server is provided
	if len(config.PrometheusServers) == 0 {
		return nil, fmt.Errorf("config validation error: at least one Prometheus server is required")
	}

	return config, nil
}

func GetEnv(name, defaultValue string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return defaultValue
}
