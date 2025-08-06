package rbac

import (
	"testing"
	"time"

	"site-availability/authentication/session"
	"site-availability/config"
	"site-availability/labels"
)

func TestNewAuthorizer(t *testing.T) {
	cfg := &config.Config{}
	authorizer := NewAuthorizer(cfg)

	if authorizer == nil {
		t.Fatal("NewAuthorizer returned nil")
	}

	if authorizer.config != cfg {
		t.Errorf("Expected config to be %v, got %v", cfg, authorizer.config)
	}
}

func TestGetUserPermissions_AdminUser(t *testing.T) {
	cfg := &config.Config{}
	authorizer := NewAuthorizer(cfg)

	userSession := &session.Session{
		ID:         "test-session",
		Username:   "admin",
		IsAdmin:    true,
		Roles:      []string{"admin"},
		Groups:     []string{},
		AuthMethod: "local",
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(time.Hour),
	}

	permissions := authorizer.GetUserPermissions(userSession)

	if !permissions.IsAdmin {
		t.Error("Expected IsAdmin to be true for admin user")
	}

	if !permissions.HasFullAccess {
		t.Error("Expected HasFullAccess to be true for admin user")
	}

	if len(permissions.AllowedLabels) != 0 {
		t.Errorf("Expected empty AllowedLabels for admin user, got %d", len(permissions.AllowedLabels))
	}
}

func TestGetUserPermissions_RegularUser(t *testing.T) {
	cfg := &config.Config{
		ServerSettings: config.ServerSettings{
			Roles: map[string]config.RoleConfig{
				"developer": {
					Labels: map[string]string{
						"environment": "dev",
						"team":        "backend",
					},
				},
				"tester": {
					Labels: map[string]string{
						"environment": "test",
						"team":        "qa",
					},
				},
			},
		},
	}
	authorizer := NewAuthorizer(cfg)

	userSession := &session.Session{
		ID:         "test-session",
		Username:   "user1",
		IsAdmin:    false,
		Roles:      []string{"developer", "tester"},
		Groups:     []string{},
		AuthMethod: "oidc",
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(time.Hour),
	}

	permissions := authorizer.GetUserPermissions(userSession)

	if permissions.IsAdmin {
		t.Error("Expected IsAdmin to be false for regular user")
	}

	if permissions.HasFullAccess {
		t.Error("Expected HasFullAccess to be false for regular user")
	}

	if len(permissions.AllowedLabels) != 2 {
		t.Errorf("Expected 2 labels, got %d", len(permissions.AllowedLabels))
	}

	// Check environment label
	if envPerm, exists := permissions.AllowedLabels["environment"]; exists {
		if len(envPerm.AllowedValues) != 2 {
			t.Errorf("Expected 2 environment values, got %d", len(envPerm.AllowedValues))
		}
		expectedValues := map[string]bool{"dev": true, "test": true}
		for _, value := range envPerm.AllowedValues {
			if !expectedValues[value] {
				t.Errorf("Unexpected environment value: %s", value)
			}
		}
	} else {
		t.Error("Expected environment label permission")
	}

	// Check team label
	if teamPerm, exists := permissions.AllowedLabels["team"]; exists {
		if len(teamPerm.AllowedValues) != 2 {
			t.Errorf("Expected 2 team values, got %d", len(teamPerm.AllowedValues))
		}
		expectedValues := map[string]bool{"backend": true, "qa": true}
		for _, value := range teamPerm.AllowedValues {
			if !expectedValues[value] {
				t.Errorf("Unexpected team value: %s", value)
			}
		}
	} else {
		t.Error("Expected team label permission")
	}
}

func TestGetUserPermissions_UserWithNoRoles(t *testing.T) {
	cfg := &config.Config{
		ServerSettings: config.ServerSettings{
			Roles: map[string]config.RoleConfig{
				"developer": {
					Labels: map[string]string{
						"environment": "dev",
					},
				},
			},
		},
	}
	authorizer := NewAuthorizer(cfg)

	userSession := &session.Session{
		ID:         "test-session",
		Username:   "user1",
		IsAdmin:    false,
		Roles:      []string{},
		Groups:     []string{},
		AuthMethod: "local",
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(time.Hour),
	}

	permissions := authorizer.GetUserPermissions(userSession)

	if permissions.IsAdmin {
		t.Error("Expected IsAdmin to be false")
	}

	if permissions.HasFullAccess {
		t.Error("Expected HasFullAccess to be false")
	}

	if len(permissions.AllowedLabels) != 0 {
		t.Errorf("Expected no labels, got %d", len(permissions.AllowedLabels))
	}
}

func TestGetUserPermissions_UserWithNonExistentRole(t *testing.T) {
	cfg := &config.Config{
		ServerSettings: config.ServerSettings{
			Roles: map[string]config.RoleConfig{
				"developer": {
					Labels: map[string]string{
						"environment": "dev",
					},
				},
			},
		},
	}
	authorizer := NewAuthorizer(cfg)

	userSession := &session.Session{
		ID:         "test-session",
		Username:   "user1",
		IsAdmin:    false,
		Roles:      []string{"nonexistent"},
		Groups:     []string{},
		AuthMethod: "oidc",
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(time.Hour),
	}

	permissions := authorizer.GetUserPermissions(userSession)

	if permissions.IsAdmin {
		t.Error("Expected IsAdmin to be false")
	}

	if permissions.HasFullAccess {
		t.Error("Expected HasFullAccess to be false")
	}

	if len(permissions.AllowedLabels) != 0 {
		t.Errorf("Expected no labels, got %d", len(permissions.AllowedLabels))
	}
}

func TestCanAccessLabel_AdminUser(t *testing.T) {
	cfg := &config.Config{}
	authorizer := NewAuthorizer(cfg)

	permissions := UserPermissions{
		IsAdmin:       true,
		AllowedLabels: make(map[string]LabelPermission),
		HasFullAccess: true,
	}

	// Admin should have access to any label
	if !authorizer.CanAccessLabel(permissions, "environment", "prod") {
		t.Error("Admin should have access to any label")
	}

	if !authorizer.CanAccessLabel(permissions, "team", "frontend") {
		t.Error("Admin should have access to any label")
	}
}

func TestCanAccessLabel_RegularUser_Allowed(t *testing.T) {
	cfg := &config.Config{}
	authorizer := NewAuthorizer(cfg)

	permissions := UserPermissions{
		IsAdmin:       false,
		HasFullAccess: false,
		AllowedLabels: map[string]LabelPermission{
			"environment": {
				Key:           "environment",
				AllowedValues: []string{"dev", "test"},
				AllowAll:      false,
			},
			"team": {
				Key:           "team",
				AllowedValues: []string{"backend"},
				AllowAll:      false,
			},
		},
	}

	// User should have access to allowed values
	if !authorizer.CanAccessLabel(permissions, "environment", "dev") {
		t.Error("User should have access to 'dev' environment")
	}

	if !authorizer.CanAccessLabel(permissions, "environment", "test") {
		t.Error("User should have access to 'test' environment")
	}

	if !authorizer.CanAccessLabel(permissions, "team", "backend") {
		t.Error("User should have access to 'backend' team")
	}
}

func TestCanAccessLabel_RegularUser_Denied(t *testing.T) {
	cfg := &config.Config{}
	authorizer := NewAuthorizer(cfg)

	permissions := UserPermissions{
		IsAdmin:       false,
		HasFullAccess: false,
		AllowedLabels: map[string]LabelPermission{
			"environment": {
				Key:           "environment",
				AllowedValues: []string{"dev", "test"},
				AllowAll:      false,
			},
		},
	}

	// User should not have access to non-allowed values
	if authorizer.CanAccessLabel(permissions, "environment", "prod") {
		t.Error("User should not have access to 'prod' environment")
	}

	if authorizer.CanAccessLabel(permissions, "team", "frontend") {
		t.Error("User should not have access to 'team' label")
	}
}

func TestCanAccessLabel_UserWithNoPermissions(t *testing.T) {
	cfg := &config.Config{}
	authorizer := NewAuthorizer(cfg)

	permissions := UserPermissions{
		IsAdmin:       false,
		AllowedLabels: make(map[string]LabelPermission),
		HasFullAccess: false,
	}

	// User with no permissions should be denied
	if authorizer.CanAccessLabel(permissions, "environment", "dev") {
		t.Error("User with no permissions should be denied")
	}
}

func TestFilterLabels_AdminUser(t *testing.T) {
	cfg := &config.Config{}
	authorizer := NewAuthorizer(cfg)

	permissions := UserPermissions{
		IsAdmin:       true,
		AllowedLabels: make(map[string]LabelPermission),
		HasFullAccess: true,
	}

	allLabels := map[string][]string{
		"environment": {"dev", "test", "prod"},
		"team":        {"backend", "frontend", "qa"},
		"region":      {"us-east", "us-west"},
	}

	filtered := authorizer.FilterLabels(permissions, allLabels)

	// Admin should see all labels
	if len(filtered) != len(allLabels) {
		t.Errorf("Expected %d labels, got %d", len(allLabels), len(filtered))
	}

	for key, values := range allLabels {
		if filteredValues, exists := filtered[key]; exists {
			if len(filteredValues) != len(values) {
				t.Errorf("Expected %d values for %s, got %d", len(values), key, len(filteredValues))
			}
		} else {
			t.Errorf("Expected label %s to be present", key)
		}
	}
}

func TestFilterLabels_RegularUser(t *testing.T) {
	cfg := &config.Config{}
	authorizer := NewAuthorizer(cfg)

	permissions := UserPermissions{
		IsAdmin:       false,
		HasFullAccess: false,
		AllowedLabels: map[string]LabelPermission{
			"environment": {
				Key:           "environment",
				AllowedValues: []string{"dev", "test"},
				AllowAll:      false,
			},
			"team": {
				Key:           "team",
				AllowedValues: []string{"backend"},
				AllowAll:      false,
			},
		},
	}

	allLabels := map[string][]string{
		"environment": {"dev", "test", "prod"},
		"team":        {"backend", "frontend", "qa"},
		"region":      {"us-east", "us-west"},
	}

	filtered := authorizer.FilterLabels(permissions, allLabels)

	// User should only see labels they have permission for
	if len(filtered) != 2 {
		t.Errorf("Expected 2 labels, got %d", len(filtered))
	}

	// Check environment label
	if envValues, exists := filtered["environment"]; exists {
		if len(envValues) != 2 {
			t.Errorf("Expected 2 environment values, got %d", len(envValues))
		}
		expectedValues := map[string]bool{"dev": true, "test": true}
		for _, value := range envValues {
			if !expectedValues[value] {
				t.Errorf("Unexpected environment value: %s", value)
			}
		}
	} else {
		t.Error("Expected environment label to be present")
	}

	// Check team label
	if teamValues, exists := filtered["team"]; exists {
		if len(teamValues) != 1 {
			t.Errorf("Expected 1 team value, got %d", len(teamValues))
		}
		if teamValues[0] != "backend" {
			t.Errorf("Expected team value 'backend', got %s", teamValues[0])
		}
	} else {
		t.Error("Expected team label to be present")
	}

	// Region should not be present
	if _, exists := filtered["region"]; exists {
		t.Error("Region label should not be present")
	}
}

func TestFilterLabels_UserWithNoPermissions(t *testing.T) {
	cfg := &config.Config{}
	authorizer := NewAuthorizer(cfg)

	permissions := UserPermissions{
		IsAdmin:       false,
		AllowedLabels: make(map[string]LabelPermission),
		HasFullAccess: false,
	}

	allLabels := map[string][]string{
		"environment": {"dev", "test", "prod"},
		"team":        {"backend", "frontend"},
	}

	filtered := authorizer.FilterLabels(permissions, allLabels)

	if len(filtered) != 0 {
		t.Errorf("Expected 0 labels, got %d", len(filtered))
	}
}

func TestBuildLabelFilters_AdminUser(t *testing.T) {
	cfg := &config.Config{}
	authorizer := NewAuthorizer(cfg)

	permissions := UserPermissions{
		IsAdmin:       true,
		AllowedLabels: make(map[string]LabelPermission),
		HasFullAccess: true,
	}

	filters := authorizer.BuildLabelFilters(permissions)

	// Admin should not need filters
	if filters != nil {
		t.Errorf("Expected nil filters for admin, got %v", filters)
	}
}

func TestBuildLabelFilters_RegularUser_SingleValue(t *testing.T) {
	cfg := &config.Config{}
	authorizer := NewAuthorizer(cfg)

	permissions := UserPermissions{
		IsAdmin:       false,
		HasFullAccess: false,
		AllowedLabels: map[string]LabelPermission{
			"environment": {
				Key:           "environment",
				AllowedValues: []string{"dev"},
				AllowAll:      false,
			},
			"team": {
				Key:           "team",
				AllowedValues: []string{"backend"},
				AllowAll:      false,
			},
		},
	}

	filters := authorizer.BuildLabelFilters(permissions)

	if len(filters) != 2 {
		t.Errorf("Expected 2 filters, got %d", len(filters))
	}

	if filters["environment"] != "dev" {
		t.Errorf("Expected environment filter to be 'dev', got %s", filters["environment"])
	}

	if filters["team"] != "backend" {
		t.Errorf("Expected team filter to be 'backend', got %s", filters["team"])
	}
}

func TestBuildLabelFilters_RegularUser_MultipleValues(t *testing.T) {
	cfg := &config.Config{}
	authorizer := NewAuthorizer(cfg)

	permissions := UserPermissions{
		IsAdmin:       false,
		HasFullAccess: false,
		AllowedLabels: map[string]LabelPermission{
			"environment": {
				Key:           "environment",
				AllowedValues: []string{"dev", "test"},
				AllowAll:      false,
			},
			"team": {
				Key:           "team",
				AllowedValues: []string{"backend"},
				AllowAll:      false,
			},
		},
	}

	filters := authorizer.BuildLabelFilters(permissions)

	// Should only filter by single-value labels
	if len(filters) != 1 {
		t.Errorf("Expected 1 filter, got %d", len(filters))
	}

	if filters["team"] != "backend" {
		t.Errorf("Expected team filter to be 'backend', got %s", filters["team"])
	}

	// Environment should not be in filters since it has multiple values
	if _, exists := filters["environment"]; exists {
		t.Error("Environment should not be in filters")
	}
}

func TestBuildLabelFilters_UserWithNoPermissions(t *testing.T) {
	cfg := &config.Config{}
	authorizer := NewAuthorizer(cfg)

	permissions := UserPermissions{
		IsAdmin:       false,
		AllowedLabels: make(map[string]LabelPermission),
		HasFullAccess: false,
	}

	filters := authorizer.BuildLabelFilters(permissions)

	if len(filters) != 0 {
		t.Errorf("Expected 0 filters, got %d", len(filters))
	}
}

func TestCanAccessApp_AdminUser(t *testing.T) {
	cfg := &config.Config{}
	authorizer := NewAuthorizer(cfg)

	permissions := UserPermissions{
		IsAdmin:       true,
		AllowedLabels: make(map[string]LabelPermission),
		HasFullAccess: true,
	}

	appLabels := []labels.Label{
		{Key: "environment", Value: "prod"},
		{Key: "team", Value: "frontend"},
	}

	if !authorizer.CanAccessApp(permissions, appLabels) {
		t.Error("Admin should have access to any app")
	}
}

func TestCanAccessApp_RegularUser_Allowed(t *testing.T) {
	cfg := &config.Config{}
	authorizer := NewAuthorizer(cfg)

	permissions := UserPermissions{
		IsAdmin:       false,
		HasFullAccess: false,
		AllowedLabels: map[string]LabelPermission{
			"environment": {
				Key:           "environment",
				AllowedValues: []string{"dev", "test"},
				AllowAll:      false,
			},
			"team": {
				Key:           "team",
				AllowedValues: []string{"backend"},
				AllowAll:      false,
			},
		},
	}

	// App with matching environment label
	appLabels1 := []labels.Label{
		{Key: "environment", Value: "dev"},
		{Key: "team", Value: "frontend"},
	}

	if !authorizer.CanAccessApp(permissions, appLabels1) {
		t.Error("User should have access to app with matching environment label")
	}

	// App with matching team label
	appLabels2 := []labels.Label{
		{Key: "environment", Value: "prod"},
		{Key: "team", Value: "backend"},
	}

	if !authorizer.CanAccessApp(permissions, appLabels2) {
		t.Error("User should have access to app with matching team label")
	}
}

func TestCanAccessApp_RegularUser_Denied(t *testing.T) {
	cfg := &config.Config{}
	authorizer := NewAuthorizer(cfg)

	permissions := UserPermissions{
		IsAdmin:       false,
		HasFullAccess: false,
		AllowedLabels: map[string]LabelPermission{
			"environment": {
				Key:           "environment",
				AllowedValues: []string{"dev", "test"},
				AllowAll:      false,
			},
		},
	}

	// App with no matching labels
	appLabels := []labels.Label{
		{Key: "environment", Value: "prod"},
		{Key: "team", Value: "frontend"},
	}

	if authorizer.CanAccessApp(permissions, appLabels) {
		t.Error("User should not have access to app with no matching labels")
	}
}

func TestCanAccessApp_UserWithNoPermissions(t *testing.T) {
	cfg := &config.Config{}
	authorizer := NewAuthorizer(cfg)

	permissions := UserPermissions{
		IsAdmin:       false,
		AllowedLabels: make(map[string]LabelPermission),
		HasFullAccess: false,
	}

	appLabels := []labels.Label{
		{Key: "environment", Value: "dev"},
		{Key: "team", Value: "backend"},
	}

	if authorizer.CanAccessApp(permissions, appLabels) {
		t.Error("User with no permissions should not have access to any app")
	}
}

func TestCanAccessApp_AppWithNoLabels(t *testing.T) {
	cfg := &config.Config{}
	authorizer := NewAuthorizer(cfg)

	permissions := UserPermissions{
		IsAdmin:       false,
		HasFullAccess: false,
		AllowedLabels: map[string]LabelPermission{
			"environment": {
				Key:           "environment",
				AllowedValues: []string{"dev"},
				AllowAll:      false,
			},
		},
	}

	appLabels := []labels.Label{}

	if authorizer.CanAccessApp(permissions, appLabels) {
		t.Error("User should not have access to app with no labels")
	}
}

func TestCanAccessApp_AdminUser_AppWithNoLabels(t *testing.T) {
	cfg := &config.Config{}
	authorizer := NewAuthorizer(cfg)

	permissions := UserPermissions{
		IsAdmin:       true,
		AllowedLabels: make(map[string]LabelPermission),
		HasFullAccess: true,
	}

	appLabels := []labels.Label{}

	if !authorizer.CanAccessApp(permissions, appLabels) {
		t.Error("Admin should have access to app with no labels")
	}
}

// Benchmark tests
func BenchmarkGetUserPermissions_Admin(b *testing.B) {
	cfg := &config.Config{}
	authorizer := NewAuthorizer(cfg)

	userSession := &session.Session{
		ID:         "test-session",
		Username:   "admin",
		IsAdmin:    true,
		Roles:      []string{"admin"},
		Groups:     []string{},
		AuthMethod: "local",
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(time.Hour),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		authorizer.GetUserPermissions(userSession)
	}
}

func BenchmarkGetUserPermissions_RegularUser(b *testing.B) {
	cfg := &config.Config{
		ServerSettings: config.ServerSettings{
			Roles: map[string]config.RoleConfig{
				"developer": {
					Labels: map[string]string{
						"environment": "dev",
						"team":        "backend",
					},
				},
				"tester": {
					Labels: map[string]string{
						"environment": "test",
						"team":        "qa",
					},
				},
			},
		},
	}
	authorizer := NewAuthorizer(cfg)

	userSession := &session.Session{
		ID:         "test-session",
		Username:   "user1",
		IsAdmin:    false,
		Roles:      []string{"developer", "tester"},
		Groups:     []string{},
		AuthMethod: "oidc",
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(time.Hour),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		authorizer.GetUserPermissions(userSession)
	}
}

func BenchmarkCanAccessLabel(b *testing.B) {
	cfg := &config.Config{}
	authorizer := NewAuthorizer(cfg)

	permissions := UserPermissions{
		IsAdmin:       false,
		HasFullAccess: false,
		AllowedLabels: map[string]LabelPermission{
			"environment": {
				Key:           "environment",
				AllowedValues: []string{"dev", "test", "staging"},
				AllowAll:      false,
			},
			"team": {
				Key:           "team",
				AllowedValues: []string{"backend", "frontend"},
				AllowAll:      false,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		authorizer.CanAccessLabel(permissions, "environment", "dev")
	}
}

func BenchmarkFilterLabels(b *testing.B) {
	cfg := &config.Config{}
	authorizer := NewAuthorizer(cfg)

	permissions := UserPermissions{
		IsAdmin:       false,
		HasFullAccess: false,
		AllowedLabels: map[string]LabelPermission{
			"environment": {
				Key:           "environment",
				AllowedValues: []string{"dev", "test"},
				AllowAll:      false,
			},
			"team": {
				Key:           "team",
				AllowedValues: []string{"backend"},
				AllowAll:      false,
			},
		},
	}

	allLabels := map[string][]string{
		"environment": {"dev", "test", "prod", "staging"},
		"team":        {"backend", "frontend", "qa", "devops"},
		"region":      {"us-east", "us-west", "eu-west"},
		"zone":        {"zone-a", "zone-b", "zone-c"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		authorizer.FilterLabels(permissions, allLabels)
	}
}
