package config

import (
	"log"
	"os"

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
	config := &Config{}
	yamlFile, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading config file: %v", err)
		return nil, err
	}

	err = yaml.Unmarshal(yamlFile, config)
	if err != nil {
		log.Printf("Error unmarshalling config file: %v", err)
		return nil, err
	}

	return config, nil
}
