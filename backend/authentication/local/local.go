package local

import (
	"fmt"
	"strings"

	"site-availability/config"
)

// LocalAuthenticator handles local admin authentication
type LocalAuthenticator struct {
	config *config.Config
}

// NewLocalAuthenticator creates a new local authenticator
func NewLocalAuthenticator(cfg *config.Config) *LocalAuthenticator {
	return &LocalAuthenticator{
		config: cfg,
	}
}

// IsEnabled returns whether local admin authentication is enabled
func (la *LocalAuthenticator) IsEnabled() bool {
	return la.config.ServerSettings.LocalAdmin.Enabled
}

// GetUsername returns the configured admin username
func (la *LocalAuthenticator) GetUsername() string {
	return la.config.ServerSettings.LocalAdmin.Username
}

// Authenticate validates username and password for local admin
func (la *LocalAuthenticator) Authenticate(username, password string) error {
	// Check if local admin is enabled
	if !la.IsEnabled() {
		return fmt.Errorf("local admin authentication is disabled")
	}

	// Validate username
	if strings.TrimSpace(username) == "" {
		return fmt.Errorf("username is required")
	}

	if strings.TrimSpace(password) == "" {
		return fmt.Errorf("password is required")
	}

	// Check username match
	configUsername := la.config.ServerSettings.LocalAdmin.Username
	if username != configUsername {
		return fmt.Errorf("invalid credentials")
	}

	// Verify password with simple string comparison
	// Note: The password is stored in the config file
	// as the password (or hash) is already in memory. Security should come from:
	// 1. File system permissions on the config file
	// 2. Using proper secrets management (Kubernetes secrets, HashiCorp Vault, etc.)
	configPassword := la.config.ServerSettings.LocalAdmin.Password
	if password != configPassword {
		return fmt.Errorf("invalid credentials")
	}

	return nil
}

// GetUserInfo returns user information for the local admin
func (la *LocalAuthenticator) GetUserInfo(username string) UserInfo {
	return UserInfo{
		Username:   username,
		IsAdmin:    true,
		Roles:      []string{"admin"},
		Groups:     []string{}, // Local admin has no groups
		AuthMethod: "local",
	}
}

// UserInfo represents user information
type UserInfo struct {
	Username   string   `json:"username"`
	IsAdmin    bool     `json:"is_admin"`
	Roles      []string `json:"roles"`
	Groups     []string `json:"groups"`
	AuthMethod string   `json:"auth_method"`
}
