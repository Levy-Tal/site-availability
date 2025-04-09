package config

import (
	"os"
	"site-availability/logging"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// Config structure defines the config.yaml structure
type Config struct {
	ScrapeInterval string     `yaml:"scrape_interval"`
	Locations      []Location `yaml:"locations"`
	Apps           []App      `yaml:"apps"`
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
	yamlFile, err := os.ReadFile(filePath)
	if err != nil {
		logging.Logger.WithError(err).WithField("file", filePath).Error("Failed to read configuration file")
		return nil, err
	}

	err = yaml.Unmarshal(yamlFile, config)
	if err != nil {
		logging.Logger.WithError(err).WithField("file", filePath).Error("Failed to parse configuration YAML")
		return nil, err
	}

	logging.Logger.WithFields(logrus.Fields{
		"scrape_interval": config.ScrapeInterval,
		"locations":       len(config.Locations),
		"apps":            len(config.Apps),
	}).Debug("Configuration loaded successfully")

	return config, nil
}
