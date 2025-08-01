package oidc

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/url"
	"sync"

	"site-availability/config"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// OIDCAuthenticator handles OIDC authentication
type OIDCAuthenticator struct {
	config       *config.Config
	provider     *oidc.Provider
	oauth2Config *oauth2.Config
	verifier     *oidc.IDTokenVerifier
	initMutex    sync.RWMutex
}

// UserInfo represents authenticated user information from OIDC
type UserInfo struct {
	Username   string   `json:"username"`
	IsAdmin    bool     `json:"is_admin"`
	Roles      []string `json:"roles"`
	Groups     []string `json:"groups"`
	AuthMethod string   `json:"auth_method"`
}

// NewOIDCAuthenticator creates a new OIDC authenticator
func NewOIDCAuthenticator(cfg *config.Config) (*OIDCAuthenticator, error) {
	// Just store the config - don't initialize provider until needed
	return &OIDCAuthenticator{config: cfg}, nil
}

// IsEnabled returns whether OIDC authentication is enabled
func (oa *OIDCAuthenticator) IsEnabled() bool {
	return oa.config.ServerSettings.OIDC.Enabled
}

// initProvider initializes the OIDC provider if not already done
func (oa *OIDCAuthenticator) initProvider() error {
	if !oa.config.ServerSettings.OIDC.Enabled {
		return fmt.Errorf("OIDC is not enabled")
	}

	oa.initMutex.RLock()
	if oa.provider != nil {
		oa.initMutex.RUnlock()
		return nil
	}
	oa.initMutex.RUnlock()

	oa.initMutex.Lock()
	defer oa.initMutex.Unlock()

	// If already initialized, return
	if oa.provider != nil {
		return nil
	}

	// Validate required configuration
	if oa.config.ServerSettings.OIDC.Config.Issuer == "" {
		return fmt.Errorf("OIDC issuer is required")
	}
	if oa.config.ServerSettings.OIDC.Config.ClientID == "" {
		return fmt.Errorf("OIDC clientID is required")
	}
	if oa.config.ServerSettings.OIDC.Config.ClientSecret == "" {
		return fmt.Errorf("OIDC clientSecret is required")
	}
	if oa.config.ServerSettings.OIDC.Config.GroupScope == "" {
		return fmt.Errorf("OIDC groupScope is required")
	}

	ctx := context.Background()

	// Initialize OIDC provider
	provider, err := oidc.NewProvider(ctx, oa.config.ServerSettings.OIDC.Config.Issuer)
	if err != nil {
		return fmt.Errorf("failed to initialize OIDC provider: %w", err)
	}

	// Configure OAuth2
	oauth2Config := &oauth2.Config{
		ClientID:     oa.config.ServerSettings.OIDC.Config.ClientID,
		ClientSecret: oa.config.ServerSettings.OIDC.Config.ClientSecret,
		RedirectURL:  oa.config.ServerSettings.HostURL + "/auth/oidc/callback",
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, oa.config.ServerSettings.OIDC.Config.GroupScope},
	}

	// Configure ID token verifier
	verifier := provider.Verifier(&oidc.Config{
		ClientID: oa.config.ServerSettings.OIDC.Config.ClientID,
	})

	oa.provider = provider
	oa.oauth2Config = oauth2Config
	oa.verifier = verifier

	return nil
}

// GetProviderName returns the configured provider name
func (oa *OIDCAuthenticator) GetProviderName() string {
	if oa.config.ServerSettings.OIDC.Config.Name != "" {
		return oa.config.ServerSettings.OIDC.Config.Name
	}
	return "OIDC Provider"
}

// GenerateAuthURL creates an OAuth2 authorization URL with state
func (oa *OIDCAuthenticator) GenerateAuthURL(redirectURL string) (string, string, error) {
	if !oa.IsEnabled() {
		return "", "", fmt.Errorf("OIDC is not enabled")
	}

	// Initialize provider if needed
	if err := oa.initProvider(); err != nil {
		return "", "", fmt.Errorf("failed to initialize OIDC provider: %w", err)
	}

	// Generate state parameter for CSRF protection
	state, err := generateState()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate state: %w", err)
	}

	// Create a copy of oauth2Config to avoid modifying shared state
	configCopy := *oa.oauth2Config
	if redirectURL != "" {
		configCopy.RedirectURL = redirectURL
	}

	authURL := configCopy.AuthCodeURL(state)
	return authURL, state, nil
}

// HandleCallback processes the OAuth2 callback and exchanges code for tokens
func (oa *OIDCAuthenticator) HandleCallback(ctx context.Context, code string) (*UserInfo, error) {
	if !oa.IsEnabled() {
		return nil, fmt.Errorf("OIDC is not enabled")
	}

	// Initialize provider if needed
	if err := oa.initProvider(); err != nil {
		return nil, fmt.Errorf("failed to initialize OIDC provider: %w", err)
	}

	// Exchange authorization code for tokens
	token, err := oa.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Extract ID token
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("no id_token found in OAuth2 token")
	}

	// Verify ID token
	idToken, err := oa.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID token: %w", err)
	}

	// Extract claims as a map to allow dynamic access based on config
	var allClaims map[string]interface{}
	if err := idToken.Claims(&allClaims); err != nil {
		return nil, fmt.Errorf("failed to extract claims: %w", err)
	}

	// Get username from configured claim field
	username := oa.extractUsername(allClaims)
	if username == "" {
		return nil, fmt.Errorf("failed to extract username from token: userNameScope='%s', claim value is empty or missing",
			oa.config.ServerSettings.OIDC.Config.UserNameScope)
	}

	// Extract groups from configured claim field
	groups := oa.extractGroups(allClaims)

	// Get user roles based on username and groups
	roles, isAdmin := oa.getUserRoles(username, groups)

	return &UserInfo{
		Username:   username,
		IsAdmin:    isAdmin,
		Roles:      roles,
		Groups:     groups,
		AuthMethod: "oidc",
	}, nil
}

// extractUsername extracts the username from claims based on configuration
func (oa *OIDCAuthenticator) extractUsername(claims map[string]interface{}) string {
	userNameScope := oa.config.ServerSettings.OIDC.Config.UserNameScope
	if userNameScope == "" {
		return "" // No username scope configured
	}

	// Get the claim value based on the configured userNameScope
	claimValue, exists := claims[userNameScope]
	if !exists {
		return "" // Configured claim doesn't exist
	}

	// Ensure the claim value is a string
	username, ok := claimValue.(string)
	if !ok {
		return "" // Claim value is not a string
	}

	return username
}

// extractGroups extracts the groups from claims based on the configured groupScope
func (oa *OIDCAuthenticator) extractGroups(claims map[string]interface{}) []string {
	groupScope := oa.config.ServerSettings.OIDC.Config.GroupScope
	if groupScope == "" {
		return []string{} // No group scope configured
	}

	// Get the claim value based on the configured groupScope
	claimValue, exists := claims[groupScope]
	if !exists {
		return []string{} // Configured claim doesn't exist
	}

	// Handle different possible types for the groups claim
	switch v := claimValue.(type) {
	case []string:
		return v
	case []interface{}:
		// Convert []interface{} to []string
		var groups []string
		for _, item := range v {
			if str, ok := item.(string); ok {
				groups = append(groups, str)
			}
		}
		return groups
	default:
		return []string{} // Claim value is not a compatible type
	}
}

// getUserRoles maps username and groups to roles based on configuration
func (oa *OIDCAuthenticator) getUserRoles(username string, groups []string) ([]string, bool) {
	roleSet := make(map[string]bool)
	isAdmin := false

	// Check user-specific role mappings
	if userRoles, exists := oa.config.ServerSettings.OIDC.Permissions.Users[username]; exists {
		for _, role := range userRoles {
			roleSet[role] = true
			if role == "admin" {
				isAdmin = true
			}
		}
	}

	// Check group-based role mappings
	for _, group := range groups {
		if groupRoles, exists := oa.config.ServerSettings.OIDC.Permissions.Groups[group]; exists {
			for _, role := range groupRoles {
				roleSet[role] = true
				if role == "admin" {
					isAdmin = true
				}
			}
		}
	}

	// Convert role set to slice
	var roles []string
	for role := range roleSet {
		roles = append(roles, role)
	}

	// If no roles assigned, return empty slice (user will have no permissions)
	// This is expected behavior for users without explicit role assignments
	if len(roles) == 0 {
		return []string{}, false
	}

	return roles, isAdmin
}

// generateState creates a random state parameter for CSRF protection
func generateState() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// ValidateState validates the state parameter (this would typically be stored in session/cache)
func ValidateState(receivedState, expectedState string) bool {
	return receivedState != "" && receivedState == expectedState
}

// ParseRedirectURL safely parses and validates redirect URLs
func ParseRedirectURL(rawURL string) (string, error) {
	if rawURL == "" {
		return "", fmt.Errorf("redirect URL cannot be empty")
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid redirect URL: %w", err)
	}

	// Basic security check - only allow http/https
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", fmt.Errorf("invalid redirect URL scheme: %s", parsedURL.Scheme)
	}

	return parsedURL.String(), nil
}
