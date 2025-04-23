package metrics_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"site-availability/handlers"
	"site-availability/metrics"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// setupMockAppStatusCache sets mock app status data for testing
func setupMockAppStatusCache(data []handlers.AppStatus) {
	handlers.UpdateAppStatus(data)
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
		{Name: "app1", Location: "us-east", Status: "up"},
		{Name: "app2", Location: "us-east", Status: "down"},
		{Name: "app3", Location: "us-west", Status: "unavailable"},
		{Name: "app4", Location: "us-west", Status: "up"},
		{Name: "app5", Location: "us-west", Status: "unavailable"},
		{Name: "app6", Location: "eu-central", Status: "down"},
		{Name: "app7", Location: "eu-central", Status: "up"},
	}
	setupMockAppStatusCache(mockData)

	handler := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	body, _ := io.ReadAll(rr.Result().Body)
	output := string(body)

	assertContains(t, output, `site_availability_status{app="app1",location="us-east"} 1`)
	assertContains(t, output, `site_availability_status{app="app2",location="us-east"} 0`)
	assertContains(t, output, `site_availability_status{app="app4",location="us-west"} 1`)
	assertContains(t, output, `site_availability_status{app="app6",location="eu-central"} 0`)
	assertContains(t, output, `site_availability_status{app="app7",location="eu-central"} 1`)

	assertContains(t, output, `site_availability_apps{location="us-east"} 2`)
	assertContains(t, output, `site_availability_apps_up{location="us-east"} 1`)
	assertContains(t, output, `site_availability_apps_down{location="us-east"} 1`)
	assertContains(t, output, `site_availability_apps_unavailable{location="us-east"} 0`)

	assertContains(t, output, `site_availability_apps{location="us-west"} 3`)
	assertContains(t, output, `site_availability_apps_up{location="us-west"} 1`)
	assertContains(t, output, `site_availability_apps_down{location="us-west"} 0`)
	assertContains(t, output, `site_availability_apps_unavailable{location="us-west"} 2`)

	assertContains(t, output, `site_availability_apps{location="eu-central"} 2`)
	assertContains(t, output, `site_availability_apps_up{location="eu-central"} 1`)
	assertContains(t, output, `site_availability_apps_down{location="eu-central"} 1`)
	assertContains(t, output, `site_availability_apps_unavailable{location="eu-central"} 0`)

	assertContains(t, output, `site_availability_total_apps 7`)
	assertContains(t, output, `site_availability_total_apps_up 3`)
	assertContains(t, output, `site_availability_total_apps_down 2`)
	assertContains(t, output, `site_availability_total_apps_unavailable 2`)
}

func TestMetricsWithAllUnavailableApps(t *testing.T) {
	mockData := []handlers.AppStatus{
		{Name: "app1", Location: "loc1", Status: "unavailable"},
		{Name: "app2", Location: "loc1", Status: "unavailable"},
		{Name: "app3", Location: "loc2", Status: "unavailable"},
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
		{Name: "app1", Location: "loc1", Status: "down"},
		{Name: "app2", Location: "loc1", Status: "down"},
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
		{Name: "appX", Location: "locX", Status: "weird-status"},
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
