package oidc

import (
	"context"
	"testing"

	"site-availability/config"
	"site-availability/logging"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOIDCAuthenticator(t *testing.T) {
	// Initialize logger for tests
	err := logging.Init()
	require.NoError(t, err)

	t.Run("create new OIDC authenticator with valid config", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				OIDC: config.OIDCConfig{
					Enabled: true,
					Config: config.OIDCProviderConfig{
						Issuer:        "https://example.com",
						ClientID:      "test-client",
						ClientSecret:  "test-secret",
						GroupScope:    "groups",
						UserNameScope: "username",
					},
				},
			},
		}

		authenticator, err := NewOIDCAuthenticator(cfg)
		require.NoError(t, err)
		assert.NotNil(t, authenticator)
		assert.Equal(t, cfg, authenticator.config)
		assert.True(t, authenticator.IsEnabled()) // Should be true when OIDC is enabled in config
	})

	t.Run("create new OIDC authenticator with disabled config", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				OIDC: config.OIDCConfig{
					Enabled: false,
				},
			},
		}

		authenticator, err := NewOIDCAuthenticator(cfg)
		require.NoError(t, err)
		assert.NotNil(t, authenticator)
		assert.False(t, authenticator.IsEnabled())
	})
}

func TestIsEnabled(t *testing.T) {
	// Initialize logger for tests
	err := logging.Init()
	require.NoError(t, err)

	t.Run("OIDC enabled", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				OIDC: config.OIDCConfig{
					Enabled: true,
				},
			},
		}

		authenticator, err := NewOIDCAuthenticator(cfg)
		require.NoError(t, err)
		assert.True(t, authenticator.IsEnabled())
	})

	t.Run("OIDC disabled", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				OIDC: config.OIDCConfig{
					Enabled: false,
				},
			},
		}

		authenticator, err := NewOIDCAuthenticator(cfg)
		require.NoError(t, err)
		assert.False(t, authenticator.IsEnabled())
	})
}

func TestGetProviderName(t *testing.T) {
	// Initialize logger for tests
	err := logging.Init()
	require.NoError(t, err)

	t.Run("with custom provider name", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				OIDC: config.OIDCConfig{
					Enabled: true,
					Config: config.OIDCProviderConfig{
						Name: "Custom Provider",
					},
				},
			},
		}

		authenticator, err := NewOIDCAuthenticator(cfg)
		require.NoError(t, err)
		assert.Equal(t, "Custom Provider", authenticator.GetProviderName())
	})

	t.Run("without custom provider name", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				OIDC: config.OIDCConfig{
					Enabled: true,
					Config:  config.OIDCProviderConfig{},
				},
			},
		}

		authenticator, err := NewOIDCAuthenticator(cfg)
		require.NoError(t, err)
		assert.Equal(t, "OIDC Provider", authenticator.GetProviderName())
	})
}

func TestExtractUsername(t *testing.T) {
	// Initialize logger for tests
	err := logging.Init()
	require.NoError(t, err)

	t.Run("extract username from claims", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				OIDC: config.OIDCConfig{
					Enabled: true,
					Config: config.OIDCProviderConfig{
						UserNameScope: "preferred_username",
					},
				},
			},
		}

		authenticator, err := NewOIDCAuthenticator(cfg)
		require.NoError(t, err)

		claims := map[string]interface{}{
			"preferred_username": "testuser",
			"email":              "test@example.com",
		}

		username := authenticator.extractUsername(claims)
		assert.Equal(t, "testuser", username)
	})

	t.Run("username scope not configured", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				OIDC: config.OIDCConfig{
					Enabled: true,
					Config: config.OIDCProviderConfig{
						UserNameScope: "",
					},
				},
			},
		}

		authenticator, err := NewOIDCAuthenticator(cfg)
		require.NoError(t, err)

		claims := map[string]interface{}{
			"preferred_username": "testuser",
		}

		username := authenticator.extractUsername(claims)
		assert.Equal(t, "", username)
	})

	t.Run("username claim not found", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				OIDC: config.OIDCConfig{
					Enabled: true,
					Config: config.OIDCProviderConfig{
						UserNameScope: "preferred_username",
					},
				},
			},
		}

		authenticator, err := NewOIDCAuthenticator(cfg)
		require.NoError(t, err)

		claims := map[string]interface{}{
			"email": "test@example.com",
		}

		username := authenticator.extractUsername(claims)
		assert.Equal(t, "", username)
	})

	t.Run("username claim is not string", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				OIDC: config.OIDCConfig{
					Enabled: true,
					Config: config.OIDCProviderConfig{
						UserNameScope: "preferred_username",
					},
				},
			},
		}

		authenticator, err := NewOIDCAuthenticator(cfg)
		require.NoError(t, err)

		claims := map[string]interface{}{
			"preferred_username": 123,
		}

		username := authenticator.extractUsername(claims)
		assert.Equal(t, "", username)
	})
}

func TestExtractGroups(t *testing.T) {
	// Initialize logger for tests
	err := logging.Init()
	require.NoError(t, err)

	t.Run("extract groups from string slice", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				OIDC: config.OIDCConfig{
					Enabled: true,
					Config: config.OIDCProviderConfig{
						GroupScope: "groups",
					},
				},
			},
		}

		authenticator, err := NewOIDCAuthenticator(cfg)
		require.NoError(t, err)

		claims := map[string]interface{}{
			"groups": []string{"group1", "group2", "group3"},
		}

		groups := authenticator.extractGroups(claims)
		assert.Equal(t, []string{"group1", "group2", "group3"}, groups)
	})

	t.Run("extract groups from interface slice", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				OIDC: config.OIDCConfig{
					Enabled: true,
					Config: config.OIDCProviderConfig{
						GroupScope: "groups",
					},
				},
			},
		}

		authenticator, err := NewOIDCAuthenticator(cfg)
		require.NoError(t, err)

		claims := map[string]interface{}{
			"groups": []interface{}{"group1", "group2", "group3"},
		}

		groups := authenticator.extractGroups(claims)
		assert.Equal(t, []string{"group1", "group2", "group3"}, groups)
	})

	t.Run("group scope not configured", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				OIDC: config.OIDCConfig{
					Enabled: true,
					Config: config.OIDCProviderConfig{
						GroupScope: "",
					},
				},
			},
		}

		authenticator, err := NewOIDCAuthenticator(cfg)
		require.NoError(t, err)

		claims := map[string]interface{}{
			"groups": []string{"group1", "group2"},
		}

		groups := authenticator.extractGroups(claims)
		assert.Equal(t, []string{}, groups)
	})

	t.Run("groups claim not found", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				OIDC: config.OIDCConfig{
					Enabled: true,
					Config: config.OIDCProviderConfig{
						GroupScope: "groups",
					},
				},
			},
		}

		authenticator, err := NewOIDCAuthenticator(cfg)
		require.NoError(t, err)

		claims := map[string]interface{}{
			"email": "test@example.com",
		}

		groups := authenticator.extractGroups(claims)
		assert.Equal(t, []string{}, groups)
	})

	t.Run("groups claim is not slice", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				OIDC: config.OIDCConfig{
					Enabled: true,
					Config: config.OIDCProviderConfig{
						GroupScope: "groups",
					},
				},
			},
		}

		authenticator, err := NewOIDCAuthenticator(cfg)
		require.NoError(t, err)

		claims := map[string]interface{}{
			"groups": "not-a-slice",
		}

		groups := authenticator.extractGroups(claims)
		assert.Equal(t, []string{}, groups)
	})
}

func TestGetUserRoles(t *testing.T) {
	// Initialize logger for tests
	err := logging.Init()
	require.NoError(t, err)

	t.Run("user with specific role mapping", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				OIDC: config.OIDCConfig{
					Enabled: true,
					Permissions: config.OIDCPermissions{
						Users: map[string][]string{
							"testuser": {"admin", "user"},
						},
					},
				},
			},
		}

		authenticator, err := NewOIDCAuthenticator(cfg)
		require.NoError(t, err)

		roles, isAdmin := authenticator.getUserRoles("testuser", []string{})
		// Since the implementation uses a map, order is not guaranteed
		// Check that both roles are present and admin role is included
		assert.Len(t, roles, 2)
		assert.Contains(t, roles, "admin")
		assert.Contains(t, roles, "user")
		assert.True(t, isAdmin)
	})

	t.Run("user with group-based role mapping", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				OIDC: config.OIDCConfig{
					Enabled: true,
					Permissions: config.OIDCPermissions{
						Groups: map[string][]string{
							"admins": {"admin"},
							"users":  {"user"},
						},
					},
				},
			},
		}

		authenticator, err := NewOIDCAuthenticator(cfg)
		require.NoError(t, err)

		roles, isAdmin := authenticator.getUserRoles("testuser", []string{"admins", "users"})
		// Since the implementation uses a map, order is not guaranteed
		// Check that both roles are present and admin role is included
		assert.Len(t, roles, 2)
		assert.Contains(t, roles, "admin")
		assert.Contains(t, roles, "user")
		assert.True(t, isAdmin)
	})

	t.Run("user with both user and group role mappings", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				OIDC: config.OIDCConfig{
					Enabled: true,
					Permissions: config.OIDCPermissions{
						Users: map[string][]string{
							"testuser": {"user"},
						},
						Groups: map[string][]string{
							"admins": {"admin"},
						},
					},
				},
			},
		}

		authenticator, err := NewOIDCAuthenticator(cfg)
		require.NoError(t, err)

		roles, isAdmin := authenticator.getUserRoles("testuser", []string{"admins"})
		// Since the implementation uses a map, order is not guaranteed
		// Check that both roles are present and admin role is included
		assert.Len(t, roles, 2)
		assert.Contains(t, roles, "admin")
		assert.Contains(t, roles, "user")
		assert.True(t, isAdmin)
	})

	t.Run("user with no role mappings", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				OIDC: config.OIDCConfig{
					Enabled: true,
					Permissions: config.OIDCPermissions{
						Users:  map[string][]string{},
						Groups: map[string][]string{},
					},
				},
			},
		}

		authenticator, err := NewOIDCAuthenticator(cfg)
		require.NoError(t, err)

		roles, isAdmin := authenticator.getUserRoles("testuser", []string{})
		assert.Equal(t, []string{}, roles)
		assert.False(t, isAdmin)
	})

	t.Run("user with non-admin roles", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				OIDC: config.OIDCConfig{
					Enabled: true,
					Permissions: config.OIDCPermissions{
						Users: map[string][]string{
							"testuser": {"user", "viewer"},
						},
					},
				},
			},
		}

		authenticator, err := NewOIDCAuthenticator(cfg)
		require.NoError(t, err)

		roles, isAdmin := authenticator.getUserRoles("testuser", []string{})
		// Since the implementation uses a map, order is not guaranteed
		// Check that both roles are present
		assert.Len(t, roles, 2)
		assert.Contains(t, roles, "user")
		assert.Contains(t, roles, "viewer")
		assert.False(t, isAdmin)
	})
}

func TestGenerateState(t *testing.T) {
	t.Run("generate state parameter", func(t *testing.T) {
		state, err := generateState()
		require.NoError(t, err)
		assert.NotEmpty(t, state)
		assert.Len(t, state, 44) // base64.URLEncoding.EncodeToString of 32 bytes produces 44 characters
	})

	t.Run("generate multiple states are different", func(t *testing.T) {
		state1, err := generateState()
		require.NoError(t, err)
		state2, err := generateState()
		require.NoError(t, err)
		assert.NotEqual(t, state1, state2)
	})
}

func TestValidateState(t *testing.T) {
	t.Run("valid state validation", func(t *testing.T) {
		state := "test-state"
		assert.True(t, ValidateState(state, state))
	})

	t.Run("invalid state validation", func(t *testing.T) {
		assert.False(t, ValidateState("state1", "state2"))
	})

	t.Run("empty state validation", func(t *testing.T) {
		assert.False(t, ValidateState("", "state"))
		assert.False(t, ValidateState("state", ""))
		assert.False(t, ValidateState("", ""))
	})
}

func TestGenerateAuthURL(t *testing.T) {
	// Initialize logger for tests
	err := logging.Init()
	require.NoError(t, err)

	t.Run("OIDC disabled", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				OIDC: config.OIDCConfig{
					Enabled: false,
				},
			},
		}

		authenticator, err := NewOIDCAuthenticator(cfg)
		require.NoError(t, err)

		authURL, state, err := authenticator.GenerateAuthURL("")
		assert.Error(t, err)
		assert.Empty(t, authURL)
		assert.Empty(t, state)
		assert.Contains(t, err.Error(), "OIDC is not enabled")
	})

	t.Run("OIDC enabled but missing required config", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				HostURL: "http://localhost:8080",
				OIDC: config.OIDCConfig{
					Enabled: true,
					Config:  config.OIDCProviderConfig{
						// Missing required fields
					},
				},
			},
		}

		authenticator, err := NewOIDCAuthenticator(cfg)
		require.NoError(t, err)

		authURL, state, err := authenticator.GenerateAuthURL("")
		assert.Error(t, err)
		assert.Empty(t, authURL)
		assert.Empty(t, state)
		assert.Contains(t, err.Error(), "OIDC issuer is required")
	})
}

func TestHandleCallback(t *testing.T) {
	// Initialize logger for tests
	err := logging.Init()
	require.NoError(t, err)

	t.Run("OIDC disabled", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				OIDC: config.OIDCConfig{
					Enabled: false,
				},
			},
		}

		authenticator, err := NewOIDCAuthenticator(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		userInfo, err := authenticator.HandleCallback(ctx, "test-code")
		assert.Error(t, err)
		assert.Nil(t, userInfo)
		assert.Contains(t, err.Error(), "OIDC is not enabled")
	})

	t.Run("OIDC enabled but missing required config", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				OIDC: config.OIDCConfig{
					Enabled: true,
					Config:  config.OIDCProviderConfig{
						// Missing required fields
					},
				},
			},
		}

		authenticator, err := NewOIDCAuthenticator(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		userInfo, err := authenticator.HandleCallback(ctx, "test-code")
		assert.Error(t, err)
		assert.Nil(t, userInfo)
		assert.Contains(t, err.Error(), "OIDC issuer is required")
	})
}
