package oidc

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/url"

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
}

// UserInfo represents authenticated user information from OIDC
type UserInfo struct {
	Username   string   `json:"username"`
	IsAdmin    bool     `json:"is_admin"`
	Roles      []string `json:"roles"`
	Groups     []string `json:"groups"`
	AuthMethod string   `json:"auth_method"`
}

// OIDCClaims represents the claims we extract from OIDC tokens
type OIDCClaims struct {
	PreferredUsername string   `json:"preferred_username"`
	Name              string   `json:"name"`
	Email             string   `json:"email"`
	Groups            []string `json:"groups"`
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

	// If already initialized, return
	if oa.provider != nil {
		return nil
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
		RedirectURL:  "http://localhost:8080/auth/oidc/callback", // Will be configured dynamically
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email", oa.config.ServerSettings.OIDC.Config.GroupScope},
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

	// Update redirect URL if provided
	if redirectURL != "" {
		oa.oauth2Config.RedirectURL = redirectURL
	}

	authURL := oa.oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
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

	// Extract claims
	var claims OIDCClaims
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to extract claims: %w", err)
	}

	// Get username from configured scope
	username := oa.extractUsername(claims)
	if username == "" {
		return nil, fmt.Errorf("failed to extract username from token")
	}

	// Get user roles based on username and groups
	roles, isAdmin := oa.getUserRoles(username, claims.Groups)

	return &UserInfo{
		Username:   username,
		IsAdmin:    isAdmin,
		Roles:      roles,
		Groups:     claims.Groups,
		AuthMethod: "oidc",
	}, nil
}

// extractUsername gets the username from claims based on configured scope
func (oa *OIDCAuthenticator) extractUsername(claims OIDCClaims) string {
	userNameScope := oa.config.ServerSettings.OIDC.Config.UserNameScope

	switch userNameScope {
	case "name":
		return claims.Name
	case "email":
		return claims.Email
	case "preferred_username":
		return claims.PreferredUsername
	default:
		// Default to preferred_username
		if claims.PreferredUsername != "" {
			return claims.PreferredUsername
		}
		if claims.Name != "" {
			return claims.Name
		}
		return claims.Email
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

	// If no roles assigned, return empty slice
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
