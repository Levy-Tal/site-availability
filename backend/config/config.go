package config

import (
	"fmt"
	"os"
	"site-availability/logging"

	"gopkg.in/yaml.v2"
)

// Config structure defines the config.yaml structure
type Config struct {
	ScrapeInterval     string     `yaml:"scrape_interval"`
	ScrapeTimeout      string     `yaml:"scrape_timeout"`
	MaxParallelScrapes int        `yaml:"max_parallel_scrapes"`
	DocsTitle          string     `yaml:"docs_title"`
	DocsURL            string     `yaml:"docs_url"`
	Locations          []Location `yaml:"locations"`
	Apps               []App      `yaml:"apps"`
}

type Location struct {
	Name      string  `yaml:"name" json:"name"`
	Latitude  float64 `yaml:"latitude" json:"latitude"`
	Longitude float64 `yaml:"longitude" json:"longitude"`
}

type App struct {
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

	return config, nil
}
