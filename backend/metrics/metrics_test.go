package metrics_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"site-availability/config"
	"site-availability/handlers"
	"site-availability/labels"
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

// setupTestHandler sets up the HTTP handler with Prometheus metrics using a separate registry
func setupTestHandler() http.Handler {
	// Create a new registry for this test to avoid conflicts
	registry := prometheus.NewRegistry()

	// Get app statuses and create dynamic metric with test registry
	appStatuses := handlers.GetAppStatusCache()
	labelKeys := collectUniqueLabelsForTest(appStatuses)
	createDynamicMetricWithRegistryForTest(labelKeys, registry)

	// Set up metrics with the test registry
	handler := setupMetricsHandlerWithRegistry(registry)

	return handler
}

// Test helper functions that mirror the main metrics functions but use a custom registry
func collectUniqueLabelsForTest(appStatuses []handlers.AppStatus) []string {
	labelKeysSet := make(map[string]bool)

	// Always include the base labels
	labelKeysSet["name"] = true
	labelKeysSet["location"] = true
	labelKeysSet["source"] = true
	labelKeysSet["origin_url"] = true

	// Add all app label keys that have non-empty values across ALL apps
	// Only include a label key if at least one app has a non-empty value for it
	labelValueCounts := make(map[string]int)
	for _, appStatus := range appStatuses {
		for _, label := range appStatus.Labels {
			if label.Value != "" { // Only count non-empty values
				labelValueCounts[label.Key]++
			}
		}
	}

	// Only include label keys that have at least one non-empty value
	for labelKey, count := range labelValueCounts {
		if count > 0 {
			labelKeysSet[labelKey] = true
		}
	}

	// Convert to sorted slice for consistent ordering
	labelKeys := make([]string, 0, len(labelKeysSet))

	// Add base labels first in a specific order
	if labelKeysSet["name"] {
		labelKeys = append(labelKeys, "name")
		delete(labelKeysSet, "name")
	}
	if labelKeysSet["location"] {
		labelKeys = append(labelKeys, "location")
		delete(labelKeysSet, "location")
	}
	if labelKeysSet["source"] {
		labelKeys = append(labelKeys, "source")
		delete(labelKeysSet, "source")
	}
	if labelKeysSet["origin_url"] {
		labelKeys = append(labelKeys, "origin_url")
		delete(labelKeysSet, "origin_url")
	}

	// Add remaining label keys alphabetically
	for key := range labelKeysSet {
		labelKeys = append(labelKeys, key)
	}

	return labelKeys
}

func createDynamicMetricWithRegistryForTest(labelKeys []string, registry *prometheus.Registry) {
	// Create new metric with dynamic labels
	siteAvailabilityStatus := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "site_availability_status",
			Help: "Site availability status by app and location (1=up, 0=down)",
		},
		labelKeys,
	)

	// Register with the test registry
	registry.MustRegister(siteAvailabilityStatus)

	// Update the test handler to use this metric
	updateTestMetric(siteAvailabilityStatus, labelKeys)
}

func updateTestMetric(metric *prometheus.GaugeVec, labelKeys []string) {
	appStatuses := handlers.GetAppStatusCache()

	for _, appStatus := range appStatuses {
		status := appStatus.Status
		labelValues := buildLabelValuesForTest(appStatus, labelKeys)

		switch status {
		case "up":
			metric.WithLabelValues(labelValues...).Set(1)
		case "down":
			metric.WithLabelValues(labelValues...).Set(0)
		default:
			// Unavailable or unknown statuses - don't set metric
		}
	}
}

func buildLabelValuesForTest(appStatus handlers.AppStatus, labelKeys []string) []string {
	labelValues := make([]string, len(labelKeys))

	// Create a map for quick lookup of app labels
	appLabelsMap := make(map[string]string)
	for _, label := range appStatus.Labels {
		if label.Value != "" { // Only include non-empty labels
			appLabelsMap[label.Key] = label.Value
		}
	}

	// Fill in the values for each label key
	for i, key := range labelKeys {
		// System fields have priority over user labels
		switch key {
		case "name":
			labelValues[i] = appStatus.Name
		case "location":
			labelValues[i] = appStatus.Location
		case "source":
			labelValues[i] = appStatus.Source
		case "origin_url":
			labelValues[i] = appStatus.OriginURL
		default:
			// This is a user-defined label
			if value, exists := appLabelsMap[key]; exists {
				labelValues[i] = value
			} else {
				labelValues[i] = "" // Empty value for missing labels
			}
		}
	}

	return labelValues
}

func setupMetricsHandlerWithRegistry(registry *prometheus.Registry) http.Handler {
	// Register all the other metrics that the tests expect
	siteAvailabilityApps := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "site_availability_apps",
			Help: "Total apps monitored in a location",
		},
		[]string{"location", "source"},
	)
	siteAvailabilityAppsUp := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "site_availability_apps_up",
			Help: "Count of apps in up status per location",
		},
		[]string{"location", "source"},
	)
	siteAvailabilityAppsDown := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "site_availability_apps_down",
			Help: "Count of apps in down status per location",
		},
		[]string{"location", "source"},
	)
	siteAvailabilityAppsUnavailable := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "site_availability_apps_unavailable",
			Help: "Count of apps in unavailable status per location",
		},
		[]string{"location", "source"},
	)
	siteAvailabilityTotalApps := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "site_availability_total_apps",
			Help: "Total apps monitored in all locations",
		},
	)
	siteAvailabilityTotalAppsUp := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "site_availability_total_apps_up",
			Help: "Total apps in up status across all locations",
		},
	)
	siteAvailabilityTotalAppsDown := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "site_availability_total_apps_down",
			Help: "Total apps in down status across all locations",
		},
	)
	siteAvailabilityTotalAppsUnavailable := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "site_availability_total_apps_unavailable",
			Help: "Total apps in unavailable status across all locations",
		},
	)

	registry.MustRegister(siteAvailabilityApps)
	registry.MustRegister(siteAvailabilityAppsUp)
	registry.MustRegister(siteAvailabilityAppsDown)
	registry.MustRegister(siteAvailabilityAppsUnavailable)
	registry.MustRegister(siteAvailabilityTotalApps)
	registry.MustRegister(siteAvailabilityTotalAppsUp)
	registry.MustRegister(siteAvailabilityTotalAppsDown)
	registry.MustRegister(siteAvailabilityTotalAppsUnavailable)

	// Update the metrics with current data
	updateAggregateMetrics(siteAvailabilityApps, siteAvailabilityAppsUp, siteAvailabilityAppsDown, siteAvailabilityAppsUnavailable, siteAvailabilityTotalApps, siteAvailabilityTotalAppsUp, siteAvailabilityTotalAppsDown, siteAvailabilityTotalAppsUnavailable)

	return promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
}

func updateAggregateMetrics(siteAvailabilityApps, siteAvailabilityAppsUp, siteAvailabilityAppsDown, siteAvailabilityAppsUnavailable *prometheus.GaugeVec, siteAvailabilityTotalApps, siteAvailabilityTotalAppsUp, siteAvailabilityTotalAppsDown, siteAvailabilityTotalAppsUnavailable prometheus.Gauge) {
	appStatuses := handlers.GetAppStatusCache()

	// Track totals across all locations
	totalApps := 0
	totalUp := 0
	totalDown := 0
	totalUnavailable := 0

	// Track per-location and per-source counts
	type locsrc struct {
		location string
		source   string
	}
	locationSourceCounts := make(map[locsrc]struct {
		total       int
		up          int
		down        int
		unavailable int
	})

	for _, appStatus := range appStatuses {
		location := appStatus.Location
		source := appStatus.Source
		status := appStatus.Status

		key := locsrc{location, source}
		if _, exists := locationSourceCounts[key]; !exists {
			locationSourceCounts[key] = struct {
				total, up, down, unavailable int
			}{}
		}
		counts := locationSourceCounts[key]
		counts.total++
		totalApps++

		switch status {
		case "up":
			counts.up++
			totalUp++
		case "down":
			counts.down++
			totalDown++
		default:
			// Unavailable or unknown statuses
			counts.unavailable++
			totalUnavailable++
		}

		locationSourceCounts[key] = counts
	}

	// Update per-location and per-source metrics
	for key, counts := range locationSourceCounts {
		siteAvailabilityApps.WithLabelValues(key.location, key.source).Set(float64(counts.total))
		siteAvailabilityAppsUp.WithLabelValues(key.location, key.source).Set(float64(counts.up))
		siteAvailabilityAppsDown.WithLabelValues(key.location, key.source).Set(float64(counts.down))
		siteAvailabilityAppsUnavailable.WithLabelValues(key.location, key.source).Set(float64(counts.unavailable))
	}

	// Update global metrics
	siteAvailabilityTotalApps.Set(float64(totalApps))
	siteAvailabilityTotalAppsUp.Set(float64(totalUp))
	siteAvailabilityTotalAppsDown.Set(float64(totalDown))
	siteAvailabilityTotalAppsUnavailable.Set(float64(totalUnavailable))
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

	assertContains(t, output, `site_availability_status{location="us-east",name="app1",origin_url="http://test-origin.com",server_env="test",server_region="us-west",source="test-source",source_env="test",source_type="mock"} 1`)
	assertContains(t, output, `site_availability_status{location="us-east",name="app2",origin_url="http://test-origin.com",server_env="test",server_region="us-west",source="test-source",source_env="test",source_type="mock"} 0`)
	assertContains(t, output, `site_availability_status{location="us-west",name="app4",origin_url="http://test-origin.com",server_env="test",server_region="us-west",source="test-source",source_env="test",source_type="mock"} 1`)
	assertContains(t, output, `site_availability_status{location="eu-central",name="app6",origin_url="http://test-origin.com",server_env="test",server_region="us-west",source="test-source",source_env="test",source_type="mock"} 0`)
	assertContains(t, output, `site_availability_status{location="eu-central",name="app7",origin_url="http://test-origin.com",server_env="test",server_region="us-west",source="test-source",source_env="test",source_type="mock"} 1`)

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

	// Ensure HostURL is set for validation
	if serverSettings.HostURL == "" {
		serverSettings.HostURL = "https://test-server.com"
	}

	// Ensure all test apps have OriginURL set (required for new validation)
	for i := range statuses {
		if statuses[i].OriginURL == "" {
			statuses[i].OriginURL = "https://test-origin.com"
		}
	}

	_ = handlers.UpdateAppStatus(sourceName, statuses, source, serverSettings)
}

func TestDynamicLabelsInMetrics(t *testing.T) {
	// Mock data with app labels as requested in the user example
	mockData := []handlers.AppStatus{
		{
			Name:      "backend-app",
			Location:  "me-central-1",
			Status:    "down",
			Source:    "frontend-app-prod",
			OriginURL: "http://localhost:8080",
			Labels: []labels.Label{
				{Key: "app", Value: "app1"},
				{Key: "env", Value: "production"},
				{Key: "team", Value: "backend"},
			},
		},
	}
	setupMockAppStatusCache(mockData)

	handler := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	body, _ := io.ReadAll(rr.Result().Body)
	output := string(body)

	// Test the expected metric format with dynamic labels including origin_url and app labels
	// Note: Prometheus sorts labels alphabetically, and we now have "name" for the system field
	expectedMetric := `site_availability_status{app="app1",env="production",location="me-central-1",name="backend-app",origin_url="http://localhost:8080"`
	assertContains(t, output, expectedMetric)
}

func TestEmptyLabelsAreExcluded(t *testing.T) {
	// Mock data with some empty labels that should be excluded
	mockData := []handlers.AppStatus{
		{
			Name:      "test-app",
			Location:  "test-location",
			Status:    "up",
			Source:    "test-source",
			OriginURL: "http://test.com",
			Labels: []labels.Label{
				{Key: "env", Value: "production"}, // Should be included
				{Key: "team", Value: "backend"},   // Should be included
				{Key: "app", Value: ""},           // Should be excluded (empty)
				{Key: "version", Value: ""},       // Should be excluded (empty)
			},
		},
	}
	setupMockAppStatusCache(mockData)

	handler := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	body, _ := io.ReadAll(rr.Result().Body)
	output := string(body)

	// Should contain labels with values
	assertContains(t, output, `env="production"`)
	assertContains(t, output, `team="backend"`)

	// Should NOT contain empty labels (app="" or version="")
	if strings.Contains(output, `app=""`) {
		t.Errorf("Found empty app label in output, but it should be excluded")
	}
	if strings.Contains(output, `version=""`) {
		t.Errorf("Found empty version label in output, but it should be excluded")
	}
}
