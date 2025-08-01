package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"site-availability/logging"
	"site-availability/yaml"

	goyaml "gopkg.in/yaml.v2"
)

type Config struct {
	ServerSettings ServerSettings   `yaml:"server_settings"`
	Scraping       ScrapingSettings `yaml:"scraping"`
	Documentation  Documentation    `yaml:"documentation"`
	Locations      []Location       `yaml:"locations"`
	Sources        []Source         `yaml:"sources"`
}

type ServerSettings struct {
	Port           string                `yaml:"port"`
	CustomCAPath   string                `yaml:"custom_ca_path"`
	SyncEnable     bool                  `yaml:"sync_enable"`
	Token          string                `yaml:"token"`
	Labels         map[string]string     `yaml:"labels,omitempty"`
	SessionTimeout string                `yaml:"session_timeout,omitempty"`
	LocalAdmin     LocalAdminConfig      `yaml:"local_admin,omitempty"`
	Roles          map[string]RoleConfig `yaml:"roles,omitempty"`
	OIDC           OIDCConfig            `yaml:"oidc,omitempty"`
}

type LocalAdminConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}

type RoleConfig struct {
	Labels map[string]string `yaml:",inline"`
}

type OIDCConfig struct {
	Enabled     bool               `yaml:"enabled"`
	Config      OIDCProviderConfig `yaml:"config,omitempty"`
	Permissions OIDCPermissions    `yaml:"permissions,omitempty"`
}

type OIDCProviderConfig struct {
	Name          string `yaml:"name,omitempty"`
	Issuer        string `yaml:"issuer,omitempty"`
	ClientID      string `yaml:"clientID,omitempty"`
	ClientSecret  string `yaml:"clientSecret,omitempty"`
	GroupScope    string `yaml:"groupScope,omitempty"`
	UserNameScope string `yaml:"userNameScope,omitempty"`
}

type OIDCPermissions struct {
	Users  map[string][]string `yaml:"users,omitempty"`
	Groups map[string][]string `yaml:"groups,omitempty"`
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
	Labels map[string]string      `yaml:"labels,omitempty"`
	Config map[string]interface{} `yaml:"config"`
}

func DecodeConfig[T any](cfg map[string]interface{}, sourceName string) (T, error) {
	var out T
	bytes, err := goyaml.Marshal(cfg)
	if err != nil {
		return out, fmt.Errorf("failed to marshal config for source %s: %w", sourceName, err)
	}

	if err := goyaml.Unmarshal(bytes, &out); err != nil {
		return out, fmt.Errorf("failed to unmarshal config for source %s: %w", sourceName, err)
	}

	return out, nil
}

func LoadConfig() (*Config, error) {
	configFile := GetEnv("CONFIG_FILE", "config.yaml")
	credentialsFile := GetEnv("CREDENTIALS_FILE", "credentials.yaml")

	logging.Logger.WithFields(map[string]interface{}{
		"config_file":      configFile,
		"credentials_file": credentialsFile,
	}).Info("Loading configuration files")

	merged, err := yaml.MergeFiles(configFile, credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to merge config files: %w", err)
	}

	var config Config
	bytes, err := goyaml.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal merged config: %w", err)
	}

	if err := goyaml.Unmarshal(bytes, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal merged config: %w", err)
	}

	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	// Apply default values after validation
	applyAuthDefaults(&config.ServerSettings)

	return &config, nil
}

func validateLabels(labels map[string]string, context string) error {
	reservedChars := []string{"&", "=", "?", "#", "/", ":"}

	for key, value := range labels {
		if strings.TrimSpace(key) == "" {
			return fmt.Errorf("label validation error in %s: label key cannot be empty", context)
		}
		if len(key) > 100 {
			return fmt.Errorf("label validation error in %s: label key %q exceeds max length", context, key)
		}
		if len(value) > 500 {
			return fmt.Errorf("label validation error in %s: label value for key %q exceeds max length", context, key)
		}
		for _, reserved := range reservedChars {
			if strings.Contains(key, reserved) {
				return fmt.Errorf("label validation error in %s: label key %q contains reserved char %q", context, key, reserved)
			}
			if strings.Contains(value, reserved) {
				return fmt.Errorf("label validation error in %s: label value %q for key %q contains reserved char %q", context, value, key, reserved)
			}
		}
	}
	return nil
}

func validateConfig(config *Config) error {
	if err := validateLabels(config.ServerSettings.Labels, "server settings"); err != nil {
		return err
	}

	// Validate authentication configuration
	if err := validateAuthConfig(&config.ServerSettings); err != nil {
		return err
	}

	if len(config.Locations) == 0 {
		return fmt.Errorf("config validation error: at least one location is required")
	}
	for _, location := range config.Locations {
		if location.Latitude < -90 || location.Latitude > 90 {
			return fmt.Errorf("location %q has invalid latitude: %f", location.Name, location.Latitude)
		}
		if location.Longitude < -180 || location.Longitude > 180 {
			return fmt.Errorf("location %q has invalid longitude: %f", location.Name, location.Longitude)
		}
	}

	sourceNames := make(map[string]bool)
	for _, source := range config.Sources {
		if source.Name == "" {
			return fmt.Errorf("source config error: source name is required")
		}
		if _, exists := sourceNames[source.Name]; exists {
			return fmt.Errorf("source config error: duplicate source name %q", source.Name)
		}
		sourceNames[source.Name] = true

		if source.Type == "" {
			return fmt.Errorf("source config error: type is required for source %s", source.Name)
		}
		// Note: source.Config validation is deferred to source initialization
		// This allows sources to be skipped if their config is invalid rather than failing the entire application

		if err := validateLabels(source.Labels, fmt.Sprintf("source %s", source.Name)); err != nil {
			return err
		}
	}

	return nil
}

func validateAuthConfig(serverSettings *ServerSettings) error {
	// If local admin is enabled, validate configuration
	if serverSettings.LocalAdmin.Enabled {
		if strings.TrimSpace(serverSettings.LocalAdmin.Username) == "" {
			return fmt.Errorf("auth config error: local admin username is required when local admin is enabled")
		}
		if strings.TrimSpace(serverSettings.LocalAdmin.Password) == "" {
			return fmt.Errorf("auth config error: local admin password is required when local admin is enabled")
		}
	}

	// Validate session timeout format if provided
	if serverSettings.SessionTimeout != "" {
		if _, err := time.ParseDuration(serverSettings.SessionTimeout); err != nil {
			return fmt.Errorf("auth config error: invalid session_timeout format %q: %w (valid examples: '12h', '30m', '90s', '1h30m')",
				serverSettings.SessionTimeout, err)
		}
	}

	// If OIDC is enabled, validate configuration
	if serverSettings.OIDC.Enabled {
		if strings.TrimSpace(serverSettings.OIDC.Config.Issuer) == "" {
			return fmt.Errorf("auth config error: OIDC issuer is required when OIDC is enabled")
		}
		if strings.TrimSpace(serverSettings.OIDC.Config.ClientID) == "" {
			return fmt.Errorf("auth config error: OIDC clientID is required when OIDC is enabled")
		}
		if strings.TrimSpace(serverSettings.OIDC.Config.ClientSecret) == "" {
			return fmt.Errorf("auth config error: OIDC clientSecret is required when OIDC is enabled")
		}
	}

	return nil
}

// applyAuthDefaults sets default values for authentication configuration
// This should be called after validation, during config loading
func applyAuthDefaults(serverSettings *ServerSettings) {
	if serverSettings.OIDC.Enabled {
		if strings.TrimSpace(serverSettings.OIDC.Config.GroupScope) == "" {
			serverSettings.OIDC.Config.GroupScope = "groups" // Default value
		}
		if strings.TrimSpace(serverSettings.OIDC.Config.UserNameScope) == "" {
			serverSettings.OIDC.Config.UserNameScope = "preferred_username" // Default value
		}
	}
}

func GetEnv(name, defaultValue string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return defaultValue
}
