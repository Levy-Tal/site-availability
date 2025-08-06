package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"site-availability/authentication/middleware"
	"site-availability/authentication/rbac"
	"site-availability/config"
	"site-availability/labels"
	"site-availability/logging"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetLabelsWithAuthz(t *testing.T) {
	// Initialize logger for tests
	err := logging.Init()
	require.NoError(t, err)

	// Set up test data
	labelManager = labels.NewLabelManager()

	// Add some test apps with labels
	testApps := []labels.AppInfo{
		{
			Name:     "app1",
			Location: "location1",
			Status:   "up",
			Source:   "source1",
			Labels: []labels.Label{
				{Key: "env", Value: "production"},
				{Key: "team", Value: "backend"},
			},
		},
		{
			Name:     "app2",
			Location: "location2",
			Status:   "down",
			Source:   "source2",
			Labels: []labels.Label{
				{Key: "env", Value: "staging"},
				{Key: "team", Value: "frontend"},
			},
		},
	}

	// Update label manager with test apps
	labelManager.UpdateAppLabels(testApps)

	// Create test config
	cfg := &config.Config{
		ServerSettings: config.ServerSettings{
			Labels: map[string]string{
				"server_env": "production",
			},
		},
	}

	t.Run("get all labels - admin user", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/labels", nil)
		w := httptest.NewRecorder()

		// Set admin permissions in context
		adminPermissions := rbac.UserPermissions{
			HasFullAccess: true,
		}
		ctx := context.WithValue(req.Context(), middleware.PermissionsContextKey, adminPermissions)
		req = req.WithContext(ctx)

		GetLabelsWithAuthz(w, req, cfg)

		assert.Equal(t, http.StatusOK, w.Code)

		var response LabelsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Should return all label keys
		expectedLabels := []string{"env", "team"}
		assert.ElementsMatch(t, expectedLabels, response.Labels)
	})

	t.Run("get specific label values - restricted user", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/labels?env", nil)
		w := httptest.NewRecorder()

		// Set restricted permissions in context
		restrictedPermissions := rbac.UserPermissions{
			HasFullAccess: false,
			AllowedLabels: map[string]rbac.LabelPermission{
				"env": {
					AllowedValues: []string{"production"},
				},
			},
		}
		ctx := context.WithValue(req.Context(), middleware.PermissionsContextKey, restrictedPermissions)
		req = req.WithContext(ctx)

		GetLabelsWithAuthz(w, req, cfg)

		assert.Equal(t, http.StatusOK, w.Code)

		var response LabelsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Should only return values the user has permission for
		expectedValues := []string{"production"}
		assert.ElementsMatch(t, expectedValues, response.Labels)
	})
}

func TestGetAppsWithAuthz(t *testing.T) {
	// Initialize logger for tests
	err := logging.Init()
	require.NoError(t, err)

	// Set up test data
	testApps := []AppStatus{
		{
			Name:     "app1",
			Location: "location1",
			Status:   "up",
			Source:   "source1",
			Labels: []labels.Label{
				{Key: "env", Value: "production"},
				{Key: "team", Value: "backend"},
			},
		},
		{
			Name:     "app2",
			Location: "location2",
			Status:   "down",
			Source:   "source2",
			Labels: []labels.Label{
				{Key: "env", Value: "staging"},
				{Key: "team", Value: "frontend"},
			},
		},
	}

	// Set app status cache directly
	appStatusCache = make(map[string]map[string]map[string]AppStatus)
	appStatusCache["http://localhost:8080"] = make(map[string]map[string]AppStatus)
	appStatusCache["http://localhost:8080"]["source1"] = make(map[string]AppStatus)
	appStatusCache["http://localhost:8080"]["source1"]["app1"] = testApps[0]
	appStatusCache["http://localhost:8080"]["source2"] = make(map[string]AppStatus)
	appStatusCache["http://localhost:8080"]["source2"]["app2"] = testApps[1]

	// Create test config
	cfg := &config.Config{
		ServerSettings: config.ServerSettings{
			Labels: map[string]string{
				"server_env": "production",
			},
		},
	}

	t.Run("get all apps - admin user", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/apps", nil)
		w := httptest.NewRecorder()

		// Set admin permissions in context
		adminPermissions := rbac.UserPermissions{
			HasFullAccess: true,
		}
		ctx := context.WithValue(req.Context(), middleware.PermissionsContextKey, adminPermissions)
		req = req.WithContext(ctx)

		GetAppsWithAuthz(w, req, cfg)

		assert.Equal(t, http.StatusOK, w.Code)

		var response AppsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Should return all apps
		assert.Len(t, response.Apps, 2)
	})

	t.Run("get apps - restricted user", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/apps", nil)
		w := httptest.NewRecorder()

		// Set restricted permissions in context
		restrictedPermissions := rbac.UserPermissions{
			HasFullAccess: false,
			AllowedLabels: map[string]rbac.LabelPermission{
				"env": {
					AllowedValues: []string{"production"},
				},
			},
		}
		ctx := context.WithValue(req.Context(), middleware.PermissionsContextKey, restrictedPermissions)
		req = req.WithContext(ctx)

		GetAppsWithAuthz(w, req, cfg)

		assert.Equal(t, http.StatusOK, w.Code)

		var response AppsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Should only return apps with production env
		assert.Len(t, response.Apps, 1)
		assert.Equal(t, "app1", response.Apps[0].Name)
	})
}

func TestWriteJSONResponse(t *testing.T) {
	t.Run("successful JSON response", func(t *testing.T) {
		w := httptest.NewRecorder()
		testData := map[string]string{"key": "value"}

		writeJSONResponse(w, testData, "test")

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "value", response["key"])
	})

	t.Run("JSON encoding error", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Create a channel which cannot be JSON encoded
		unencodableData := make(chan int)

		writeJSONResponse(w, unencodableData, "test")

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to encode test")
	})
}
