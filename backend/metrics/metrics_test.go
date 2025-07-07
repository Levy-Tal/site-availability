package metrics_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"site-availability/config"
	"site-availability/handlers"
	"site-availability/metrics"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// setupMockAppStatusCache sets mock app status data for testing
func setupMockAppStatusCache(data []handlers.AppStatus) {
	// Reset all caches and global state for clean test isolation
	handlers.ResetCacheForTesting()
	updateAppStatusTest("test-source", data)
}

// setupTestHandler sets up the HTTP handler with Prometheus metrics.
func setupTestHandler() http.Handler {
	handler := metrics.SetupMetricsHandler()

	// Wrap with promhttp handler
	return promhttp.InstrumentMetricHandler(
		prometheus.DefaultRegisterer, // Use the default Prometheus registry
		handler,
	)
}

// assertContains is a helper function to assert a substring exists in a string.
func assertContains(t *testing.T, output, metricLine string) {
	if !strings.Contains(output, metricLine) {
		t.Errorf("Expected metric line missing: %q", metricLine)
	}
}

// TestMain is the entry point for setting up the test environment
func TestMain(m *testing.M) {
	// Initialize the metrics globally to avoid duplicate registration
	metrics.Init()

	// Run the tests
	m.Run()
}

func TestSetupMetricsHandler(t *testing.T) {
	mockData := []handlers.AppStatus{
		{Name: "app1", Location: "us-east", Status: "up", Source: "test-source", OriginURL: "http://test-origin.com"},
		{Name: "app2", Location: "us-east", Status: "down", Source: "test-source", OriginURL: "http://test-origin.com"},
		{Name: "app3", Location: "us-west", Status: "unavailable", Source: "test-source", OriginURL: "http://test-origin.com"},
		{Name: "app4", Location: "us-west", Status: "up", Source: "test-source", OriginURL: "http://test-origin.com"},
		{Name: "app5", Location: "us-west", Status: "unavailable", Source: "test-source", OriginURL: "http://test-origin.com"},
		{Name: "app6", Location: "eu-central", Status: "down", Source: "test-source", OriginURL: "http://test-origin.com"},
		{Name: "app7", Location: "eu-central", Status: "up", Source: "test-source", OriginURL: "http://test-origin.com"},
	}
	setupMockAppStatusCache(mockData)

	handler := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	body, _ := io.ReadAll(rr.Result().Body)
	output := string(body)

	assertContains(t, output, `site_availability_status{app="app1",location="us-east",source="test-source"} 1`)
	assertContains(t, output, `site_availability_status{app="app2",location="us-east",source="test-source"} 0`)
	assertContains(t, output, `site_availability_status{app="app4",location="us-west",source="test-source"} 1`)
	assertContains(t, output, `site_availability_status{app="app6",location="eu-central",source="test-source"} 0`)
	assertContains(t, output, `site_availability_status{app="app7",location="eu-central",source="test-source"} 1`)

	assertContains(t, output, `site_availability_apps{location="us-east",source="test-source"} 2`)
	assertContains(t, output, `site_availability_apps_up{location="us-east",source="test-source"} 1`)
	assertContains(t, output, `site_availability_apps_down{location="us-east",source="test-source"} 1`)
	assertContains(t, output, `site_availability_apps_unavailable{location="us-east",source="test-source"} 0`)

	assertContains(t, output, `site_availability_apps{location="us-west",source="test-source"} 3`)
	assertContains(t, output, `site_availability_apps_up{location="us-west",source="test-source"} 1`)
	assertContains(t, output, `site_availability_apps_down{location="us-west",source="test-source"} 0`)
	assertContains(t, output, `site_availability_apps_unavailable{location="us-west",source="test-source"} 2`)

	assertContains(t, output, `site_availability_apps{location="eu-central",source="test-source"} 2`)
	assertContains(t, output, `site_availability_apps_up{location="eu-central",source="test-source"} 1`)
	assertContains(t, output, `site_availability_apps_down{location="eu-central",source="test-source"} 1`)
	assertContains(t, output, `site_availability_apps_unavailable{location="eu-central",source="test-source"} 0`)

	assertContains(t, output, `site_availability_total_apps 7`)
	assertContains(t, output, `site_availability_total_apps_up 3`)
	assertContains(t, output, `site_availability_total_apps_down 2`)
	assertContains(t, output, `site_availability_total_apps_unavailable 2`)
}

func TestMetricsWithAllUnavailableApps(t *testing.T) {
	mockData := []handlers.AppStatus{
		{Name: "app1", Location: "loc1", Status: "unavailable", Source: "test-source", OriginURL: "http://test-origin.com"},
		{Name: "app2", Location: "loc1", Status: "unavailable", Source: "test-source", OriginURL: "http://test-origin.com"},
		{Name: "app3", Location: "loc2", Status: "unavailable", Source: "test-source", OriginURL: "http://test-origin.com"},
	}
	setupMockAppStatusCache(mockData)

	handler := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	body, _ := io.ReadAll(rr.Result().Body)
	output := string(body)

	assertContains(t, output, `site_availability_total_apps 3`)
	assertContains(t, output, `site_availability_total_apps_up 0`)
	assertContains(t, output, `site_availability_total_apps_down 0`)
	assertContains(t, output, `site_availability_total_apps_unavailable 3`)
}

func TestMetricsWithAllDownApps(t *testing.T) {
	mockData := []handlers.AppStatus{
		{Name: "app1", Location: "loc1", Status: "down", Source: "test-source", OriginURL: "http://test-origin.com"},
		{Name: "app2", Location: "loc1", Status: "down", Source: "test-source", OriginURL: "http://test-origin.com"},
	}
	setupMockAppStatusCache(mockData)

	handler := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	body, _ := io.ReadAll(rr.Result().Body)
	output := string(body)

	assertContains(t, output, `site_availability_total_apps 2`)
	assertContains(t, output, `site_availability_total_apps_up 0`)
	assertContains(t, output, `site_availability_total_apps_down 2`)
	assertContains(t, output, `site_availability_total_apps_unavailable 0`)
}

func TestMetricsWithEmptyAppStatus(t *testing.T) {
	setupMockAppStatusCache([]handlers.AppStatus{})

	handler := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	body, _ := io.ReadAll(rr.Result().Body)
	output := string(body)

	assertContains(t, output, `site_availability_total_apps 0`)
	assertContains(t, output, `site_availability_total_apps_up 0`)
	assertContains(t, output, `site_availability_total_apps_down 0`)
	assertContains(t, output, `site_availability_total_apps_unavailable 0`)
}

func TestMetricsWithUnknownStatus(t *testing.T) {
	mockData := []handlers.AppStatus{
		{Name: "appX", Location: "locX", Status: "weird-status", Source: "test-source", OriginURL: "http://test-origin.com"},
	}
	setupMockAppStatusCache(mockData)

	handler := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	body, _ := io.ReadAll(rr.Result().Body)
	output := string(body)

	assertContains(t, output, `site_availability_total_apps_unavailable 1`)
}

func TestSiteSyncMetrics(t *testing.T) {
	data := []handlers.AppStatus{
		{Name: "app1", Status: "up", Source: "test-source", OriginURL: "http://test-origin.com"},
	}
	updateAppStatusTest("test-source", data)

	// Create a sync metrics instance and test it
	syncMetrics := metrics.NewSiteSyncMetrics()

	// Test incrementing sync attempts
	syncMetrics.SyncAttempts.Inc()
	syncMetrics.SyncAttempts.Inc()

	// Test incrementing sync failures
	syncMetrics.SyncFailures.Inc()

	// Test setting sync latency
	syncMetrics.SyncLatency.Observe(0.5) // 500ms
	syncMetrics.SyncLatency.Observe(1.2) // 1.2s

	// Test setting last success time
	syncMetrics.LastSyncTime.SetToCurrentTime()

	// Test setting site status
	syncMetrics.SiteStatus.WithLabelValues("test-site", "up").Set(1)
	syncMetrics.SiteStatus.WithLabelValues("test-site", "down").Set(0)

	// Create a test handler to expose metrics
	handler := promhttp.Handler()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	body, _ := io.ReadAll(rr.Result().Body)
	output := string(body)

	// Verify the sync metrics are present with correct names
	assertContains(t, output, "site_availability_sync_attempts_total 2")
	assertContains(t, output, "site_availability_sync_failures_total 1")
	assertContains(t, output, "site_availability_sync_latency_seconds_bucket")
	assertContains(t, output, "site_availability_sync_latency_seconds_sum")
	assertContains(t, output, "site_availability_sync_latency_seconds_count")
	assertContains(t, output, "site_availability_sync_last_success_timestamp")
	assertContains(t, output, `site_availability_sync_status{site="test-site",status="up"} 1`)
	assertContains(t, output, `site_availability_sync_status{site="test-site",status="down"} 0`)
}

// Helper function to create mock source and server settings for tests
func getMockSourceAndSettings(sourceName string) (config.Source, config.ServerSettings) {
	source := config.Source{
		Name: sourceName,
		Type: "test",
		Config: map[string]interface{}{
			"url": "http://test.example.com",
		},
		Labels: map[string]string{"source_env": "test", "source_type": "mock"},
	}

	serverSettings := config.ServerSettings{
		Labels: map[string]string{"server_env": "test", "server_region": "us-west"},
	}

	return source, serverSettings
}

// Helper function to call UpdateAppStatus with mock parameters
func updateAppStatusTest(sourceName string, statuses []handlers.AppStatus) {
	source, serverSettings := getMockSourceAndSettings(sourceName)
	handlers.UpdateAppStatus(sourceName, statuses, source, serverSettings)
}
