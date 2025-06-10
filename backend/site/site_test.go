package site

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"site-availability/config"
	"site-availability/handlers"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSiteSync(t *testing.T) {
	// Create test config
	cfg := &config.Config{
		ServerSettings: config.ServerSettings{
			SyncEnable: true,
		},
		Sites: []config.Site{
			{
				Name:          "test-site",
				URL:           "https://test-site:3030",
				Enabled:       true,
				CheckInterval: "1s",
				Timeout:       "5s",
			},
		},
	}

	// Test creating new SiteSync
	sync := NewSiteSync(cfg)
	assert.NotNil(t, sync)
	assert.Equal(t, cfg, sync.config)
	assert.NotNil(t, sync.metrics)

	// Start the sync to populate the sites map
	err := sync.Start()
	require.NoError(t, err)

	assert.NotEmpty(t, sync.sites)
	assert.NotEmpty(t, sync.siteClients)
	assert.NotNil(t, sync.httpClient)

	// Test with disabled site
	cfg.Sites[0].Enabled = false
	sync = NewSiteSync(cfg)
	err = sync.Start()
	require.NoError(t, err)
	assert.Empty(t, sync.siteClients)
}

func TestCreateHTTPClient(t *testing.T) {
	// Create test site without CA
	site := config.Site{
		Name:          "test-site",
		URL:           "https://test-site:3030",
		Enabled:       true,
		CheckInterval: "1s",
		Timeout:       "5s",
	}

	// Test creating basic HTTP client without CA
	sync := &SiteSync{}
	client, err := sync.createHTTPClient(site)
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, 5*time.Second, client.Timeout)

	// Create test CA certificate
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "ca.crt")
	err = os.WriteFile(certPath, []byte("test certificate"), 0644)
	require.NoError(t, err)

	// Test with invalid timeout
	site.Timeout = "invalid"
	_, err = sync.createHTTPClient(site)
	assert.Error(t, err)

	// Test with invalid CA path
	site.Timeout = "5s"
	site.CustomCAPath = "nonexistent.crt"
	_, err = sync.createHTTPClient(site)
	assert.Error(t, err)

	// Test with invalid CA certificate
	site.CustomCAPath = certPath
	_, err = sync.createHTTPClient(site)
	assert.Error(t, err)
}

func TestStartAndStop(t *testing.T) {
	// Create test config
	cfg := &config.Config{
		ServerSettings: config.ServerSettings{
			SyncEnable: true,
		},
		Sites: []config.Site{
			{
				Name:          "test-site",
				URL:           "https://test-site:3030",
				Enabled:       true,
				CheckInterval: "1s",
				Timeout:       "5s",
			},
		},
	}

	// Test starting sync
	sync := NewSiteSync(cfg)
	err := sync.Start()
	require.NoError(t, err)

	// Test stopping sync
	err = sync.Stop()
	require.NoError(t, err)
}

func TestSyncSite(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		assert.NotEmpty(t, r.Header.Get("X-Site-Sync-Timestamp"))
		if r.Header.Get("X-Site-Sync-Signature") != "" {
			assert.NotEmpty(t, r.Header.Get("X-Site-Sync-Signature"))
		}

		// Return test response
		response := map[string]*SyncedAppStatus{
			"app1": {
				AppStatus: handlers.AppStatus{
					Name:     "app1",
					Location: "loc1",
					Status:   "up",
				},
				LastSynced:  time.Now(),
				LastSuccess: time.Now(),
				SyncEnabled: true,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer server.Close()

	// Create test config
	cfg := &config.Config{
		ServerSettings: config.ServerSettings{
			SyncEnable: true,
		},
		Sites: []config.Site{
			{
				Name:          "test-site",
				URL:           server.URL,
				Enabled:       true,
				CheckInterval: "1s",
				Timeout:       "5s",
				Token:         "test-token",
			},
		},
	}

	// Test syncing site
	sync := NewSiteSync(cfg)
	err := sync.syncSite(&cfg.Sites[0])
	require.NoError(t, err)

	// Verify status was updated
	statuses := handlers.GetAppStatusCache()
	require.Len(t, statuses, 1)
	assert.Equal(t, "app1", statuses[0].Name)
	assert.Equal(t, "loc1", statuses[0].Location)
	assert.Equal(t, "up", statuses[0].Status)
}

func TestPrepareSyncRequest(t *testing.T) {
	// Create test site
	site := config.Site{
		Name:          "test-site",
		URL:           "https://test-site:3030",
		Enabled:       true,
		CheckInterval: "1s",
		Timeout:       "5s",
		Token:         "test-token",
	}

	// Test preparing request
	sync := &SiteSync{}
	req, err := sync.prepareSyncRequest(&site)
	require.NoError(t, err)
	assert.NotNil(t, req)
	assert.Equal(t, "GET", req.Method)
	assert.Equal(t, "https://test-site:3030/sync", req.URL.String())
	assert.NotEmpty(t, req.Header.Get("X-Site-Sync-Timestamp"))
	assert.NotEmpty(t, req.Header.Get("X-Site-Sync-Signature"))

	// Test without token
	site.Token = ""
	req, err = sync.prepareSyncRequest(&site)
	require.NoError(t, err)
	assert.Empty(t, req.Header.Get("X-Site-Sync-Signature"))
}

func TestMergeStatuses(t *testing.T) {
	// Create test data
	sourceSite := "test-site"
	remoteStatuses := map[string]*SyncedAppStatus{
		"app1": {
			AppStatus: handlers.AppStatus{
				Name:     "app1",
				Location: "loc1",
				Status:   "up",
			},
			LastSynced:  time.Now(),
			LastSuccess: time.Now(),
			SyncEnabled: true,
		},
		"app2": {
			AppStatus: handlers.AppStatus{
				Name:     "app2",
				Location: "loc2",
				Status:   "down",
			},
			LastSynced:  time.Now(),
			LastSuccess: time.Now(),
			SyncEnabled: true,
		},
	}

	// Test merging statuses
	sync := &SiteSync{}
	sync.mergeStatuses(sourceSite, remoteStatuses)

	// Verify statuses were merged
	statuses := handlers.GetAppStatusCache()
	require.Len(t, statuses, 2)
	assert.Equal(t, "app1", statuses[0].Name)
	assert.Equal(t, "up", statuses[0].Status)
	assert.Equal(t, "app2", statuses[1].Name)
	assert.Equal(t, "down", statuses[1].Status)
}

func TestUpdateSiteStatus(t *testing.T) {
	// Create test config
	cfg := &config.Config{
		ServerSettings: config.ServerSettings{
			SyncEnable: true,
		},
		Sites: []config.Site{
			{
				Name:    "test-site",
				Enabled: true,
			},
		},
	}

	// Test updating site status
	sync := NewSiteSync(cfg)
	err := sync.Start()
	require.NoError(t, err)

	sync.updateSiteStatus("test-site", true, "")
	site := sync.sites["test-site"]
	assert.NotZero(t, site.LastSuccess)
	assert.Zero(t, site.ErrorCount)
	assert.Empty(t, site.LastError)

	// Test updating with error
	sync.updateSiteStatus("test-site", false, "test error")
	site = sync.sites["test-site"]
	assert.Equal(t, 1, site.ErrorCount)
	assert.Equal(t, "test error", site.LastError)

	// Test updating non-existent site
	sync.updateSiteStatus("nonexistent", true, "")
}

func TestStartSiteSync(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]*SyncedAppStatus{
			"app1": {
				AppStatus: handlers.AppStatus{
					Name:     "app1",
					Location: "loc1",
					Status:   "up",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer server.Close()

	// Create test config
	cfg := &config.Config{
		ServerSettings: config.ServerSettings{
			SyncEnable: true,
		},
		Sites: []config.Site{
			{
				Name:          "test-site",
				URL:           server.URL,
				Enabled:       true,
				CheckInterval: "100ms",
				Timeout:       "5s",
			},
		},
	}

	// Test starting site sync
	sync := NewSiteSync(cfg)
	err := sync.Start()
	require.NoError(t, err)

	// Wait for sync to occur
	time.Sleep(200 * time.Millisecond)

	// Stop sync
	err = sync.Stop()
	require.NoError(t, err)

	// Verify status was updated
	statuses := handlers.GetAppStatusCache()
	require.Len(t, statuses, 1)
	assert.Equal(t, "app1", statuses[0].Name)
	assert.Equal(t, "up", statuses[0].Status)
}

func TestSyncSiteWithError(t *testing.T) {
	// Create test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer server.Close()

	// Create test config
	cfg := &config.Config{
		ServerSettings: config.ServerSettings{
			SyncEnable: true,
		},
		Sites: []config.Site{
			{
				Name:          "test-site",
				URL:           server.URL,
				Enabled:       true,
				CheckInterval: "1s",
				Timeout:       "5s",
			},
		},
	}

	// Test syncing site with error
	sync := NewSiteSync(cfg)
	err := sync.Start()
	require.NoError(t, err)

	err = sync.syncSite(&cfg.Sites[0])
	assert.Error(t, err)

	// Verify site status was updated
	site := sync.sites["test-site"]
	assert.Equal(t, 1, site.ErrorCount)
	assert.NotEmpty(t, site.LastError)

	// Clean up
	err = sync.Stop()
	require.NoError(t, err)
}
