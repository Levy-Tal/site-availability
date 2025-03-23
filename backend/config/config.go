package config

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

// Config structure defines the config.yaml structure
type Config struct {
	ScrapeInterval string     `yaml:"ScrapeInterval"`
	Locations      []Location `yaml:"Locations"`
	Apps           []App      `yaml:"Apps"`
}

type Location struct {
	Name      string  `yaml:"name"`
	Latitude  float64 `yaml:"Latitude"`
	Longitude float64 `yaml:"Longitude"`
}

type App struct {
	Name       string   `yaml:"name"`
	Location   string   `yaml:"location"`
	Metric     string   `yaml:"Metric"`
	Prometheus []string `yaml:"Prometheus"`
}

// LoadConfig loads the YAML configuration from a file
func LoadConfig(filePath string) (*Config, error) {
	config := &Config{}
	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
		return nil, err
	}

	err = yaml.Unmarshal(yamlFile, config)
	if err != nil {
		log.Fatalf("Error unmarshalling config file: %v", err)
		return nil, err
	}

	return config, nil
}
