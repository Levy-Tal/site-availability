package local

import (
	"fmt"
	"strings"

	"site-availability/config"

	"golang.org/x/crypto/bcrypt"
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

	// Verify password
	configPassword := la.config.ServerSettings.LocalAdmin.Password
	if err := verifyPassword(password, configPassword); err != nil {
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

// verifyPassword checks if the provided password matches the stored password
// It handles both plaintext (for development) and bcrypt hashed passwords
func verifyPassword(providedPassword, storedPassword string) error {
	// Check if stored password is bcrypt hashed (starts with $2a$, $2b$, or $2y$)
	if strings.HasPrefix(storedPassword, "$2a$") ||
		strings.HasPrefix(storedPassword, "$2b$") ||
		strings.HasPrefix(storedPassword, "$2y$") {
		// Use bcrypt verification
		return bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(providedPassword))
	}

	// For backwards compatibility and development, support plaintext comparison
	// In production, passwords should always be hashed
	if providedPassword == storedPassword {
		return nil
	}

	return fmt.Errorf("password verification failed")
}

// HashPassword creates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	// Use bcrypt with cost 12 (good balance of security and performance)
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedBytes), nil
}

// ValidatePasswordStrength performs basic password validation
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	// You can add more validation rules here if needed
	// For now, keeping it simple as requested

	return nil
}
