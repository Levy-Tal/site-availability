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
	ScrapeTimeout      string     `yaml:"scrape_timeout"`       // New field
	MaxParallelScrapes int        `yaml:"max_parallel_scrapes"` // New field
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
		logging.Logger.WithError(err).Error("Failed to open configuration file")
		return nil, err
	}
	defer file.Close()

	// Decode the YAML file into the Config struct
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		logging.Logger.WithError(err).Error("Failed to parse configuration YAML")
		return nil, err
	}

	// Validate the configuration
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}

// validateConfig checks for missing or invalid fields in the configuration
func validateConfig(config *Config) error {
	// Create a map of valid location names for quick lookup
	validLocations := make(map[string]Location)
	for _, location := range config.Locations {
		// Validate latitude and longitude
		if location.Latitude < -90 || location.Latitude > 90 {
			return fmt.Errorf("location '%s' has an invalid latitude: %f (must be between -90 and 90)", location.Name, location.Latitude)
		}
		if location.Longitude < -180 || location.Longitude > 180 {
			return fmt.Errorf("location '%s' has an invalid longitude: %f (must be between -180 and 180)", location.Name, location.Longitude)
		}
		validLocations[location.Name] = location
	}

	// Validate each app
	for _, app := range config.Apps {
		if app.Name == "" {
			return fmt.Errorf("app is missing a name: %+v", app)
		}
		if app.Metric == "" {
			return fmt.Errorf("app '%s' is missing a metric", app.Name)
		}
		if app.Prometheus == "" {
			return fmt.Errorf("app '%s' is missing a Prometheus URL", app.Name)
		}
		location, exists := validLocations[app.Location]
		if !exists {
			return fmt.Errorf("app '%s' has an unknown location '%s'", app.Name, app.Location)
		}
		// Validate the app's location coordinates (redundant but ensures correctness)
		if location.Latitude < -90 || location.Latitude > 90 || location.Longitude < -180 || location.Longitude > 180 {
			return fmt.Errorf("app '%s' references a location '%s' with invalid coordinates (latitude: %f, longitude: %f)", app.Name, app.Location, location.Latitude, location.Longitude)
		}
	}

	return nil
}
