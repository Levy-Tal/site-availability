package local

import (
	"testing"

	"site-availability/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLocalAuthenticator(t *testing.T) {
	t.Run("create new local authenticator", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				LocalAdmin: config.LocalAdminConfig{
					Enabled:  true,
					Username: "admin",
					Password: "password",
				},
			},
		}

		auth := NewLocalAuthenticator(cfg)
		require.NotNil(t, auth)
		assert.Equal(t, cfg, auth.config)
	})
}

func TestIsEnabled(t *testing.T) {
	t.Run("local admin enabled", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				LocalAdmin: config.LocalAdminConfig{
					Enabled:  true,
					Username: "admin",
					Password: "password",
				},
			},
		}

		auth := NewLocalAuthenticator(cfg)
		assert.True(t, auth.IsEnabled())
	})

	t.Run("local admin disabled", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				LocalAdmin: config.LocalAdminConfig{
					Enabled:  false,
					Username: "admin",
					Password: "password",
				},
			},
		}

		auth := NewLocalAuthenticator(cfg)
		assert.False(t, auth.IsEnabled())
	})
}

func TestGetUsername(t *testing.T) {
	t.Run("get configured username", func(t *testing.T) {
		expectedUsername := "testadmin"
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				LocalAdmin: config.LocalAdminConfig{
					Enabled:  true,
					Username: expectedUsername,
					Password: "password",
				},
			},
		}

		auth := NewLocalAuthenticator(cfg)
		assert.Equal(t, expectedUsername, auth.GetUsername())
	})
}

func TestAuthenticate(t *testing.T) {
	t.Run("successful authentication", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				LocalAdmin: config.LocalAdminConfig{
					Enabled:  true,
					Username: "admin",
					Password: "password123",
				},
			},
		}

		auth := NewLocalAuthenticator(cfg)
		err := auth.Authenticate("admin", "password123")
		assert.NoError(t, err)
	})

	t.Run("authentication disabled", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				LocalAdmin: config.LocalAdminConfig{
					Enabled:  false,
					Username: "admin",
					Password: "password",
				},
			},
		}

		auth := NewLocalAuthenticator(cfg)
		err := auth.Authenticate("admin", "password")
		assert.Error(t, err)
		assert.Equal(t, "local admin authentication is disabled", err.Error())
	})

	t.Run("empty username", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				LocalAdmin: config.LocalAdminConfig{
					Enabled:  true,
					Username: "admin",
					Password: "password",
				},
			},
		}

		auth := NewLocalAuthenticator(cfg)
		err := auth.Authenticate("", "password")
		assert.Error(t, err)
		assert.Equal(t, "username is required", err.Error())
	})

	t.Run("empty password", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				LocalAdmin: config.LocalAdminConfig{
					Enabled:  true,
					Username: "admin",
					Password: "password",
				},
			},
		}

		auth := NewLocalAuthenticator(cfg)
		err := auth.Authenticate("admin", "")
		assert.Error(t, err)
		assert.Equal(t, "password is required", err.Error())
	})

	t.Run("wrong username", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				LocalAdmin: config.LocalAdminConfig{
					Enabled:  true,
					Username: "admin",
					Password: "password",
				},
			},
		}

		auth := NewLocalAuthenticator(cfg)
		err := auth.Authenticate("wronguser", "password")
		assert.Error(t, err)
		assert.Equal(t, "invalid credentials", err.Error())
	})

	t.Run("wrong password", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				LocalAdmin: config.LocalAdminConfig{
					Enabled:  true,
					Username: "admin",
					Password: "password",
				},
			},
		}

		auth := NewLocalAuthenticator(cfg)
		err := auth.Authenticate("admin", "wrongpassword")
		assert.Error(t, err)
		assert.Equal(t, "invalid credentials", err.Error())
	})
}

func TestGetUserInfo(t *testing.T) {
	t.Run("get user info for valid username", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				LocalAdmin: config.LocalAdminConfig{
					Enabled:  true,
					Username: "admin",
					Password: "password",
				},
			},
		}

		auth := NewLocalAuthenticator(cfg)
		userInfo := auth.GetUserInfo("admin")

		expectedUserInfo := UserInfo{
			Username:   "admin",
			IsAdmin:    true,
			Roles:      []string{"admin"},
			Groups:     []string{},
			AuthMethod: "local",
		}

		assert.Equal(t, expectedUserInfo, userInfo)
	})

	t.Run("get user info for different username", func(t *testing.T) {
		cfg := &config.Config{
			ServerSettings: config.ServerSettings{
				LocalAdmin: config.LocalAdminConfig{
					Enabled:  true,
					Username: "admin",
					Password: "password",
				},
			},
		}

		auth := NewLocalAuthenticator(cfg)
		userInfo := auth.GetUserInfo("differentuser")

		expectedUserInfo := UserInfo{
			Username:   "differentuser",
			IsAdmin:    true,
			Roles:      []string{"admin"},
			Groups:     []string{},
			AuthMethod: "local",
		}

		assert.Equal(t, expectedUserInfo, userInfo)
	})
}

func TestUserInfoStruct(t *testing.T) {
	t.Run("user info struct fields", func(t *testing.T) {
		userInfo := UserInfo{
			Username:   "testuser",
			IsAdmin:    true,
			Roles:      []string{"admin", "user"},
			Groups:     []string{"admins", "developers"},
			AuthMethod: "local",
		}

		assert.Equal(t, "testuser", userInfo.Username)
		assert.True(t, userInfo.IsAdmin)
		assert.Equal(t, []string{"admin", "user"}, userInfo.Roles)
		assert.Equal(t, []string{"admins", "developers"}, userInfo.Groups)
		assert.Equal(t, "local", userInfo.AuthMethod)
	})
}
