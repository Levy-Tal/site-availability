package metrics

import (
	"net/http"
	"site-availability/handlers"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Per location/app status metrics
	// Note: This will be recreated dynamically in SetupMetricsHandler to support dynamic labels
	siteAvailabilityStatus *prometheus.GaugeVec

	// Per location aggregated metrics
	siteAvailabilityApps = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "site_availability_apps",
			Help: "Total apps monitored in a location",
		},
		[]string{"location", "source"},
	)
	siteAvailabilityAppsUp = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "site_availability_apps_up",
			Help: "Count of apps in up status per location",
		},
		[]string{"location", "source"},
	)
	siteAvailabilityAppsDown = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "site_availability_apps_down",
			Help: "Count of apps in down status per location",
		},
		[]string{"location", "source"},
	)
	siteAvailabilityAppsUnavailable = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "site_availability_apps_unavailable",
			Help: "Count of apps in unavailable status per location",
		},
		[]string{"location", "source"},
	)

	// Global metrics
	siteAvailabilityTotalApps = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "site_availability_total_apps",
			Help: "Total apps monitored in all locations",
		},
	)
	siteAvailabilityTotalAppsUp = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "site_availability_total_apps_up",
			Help: "Total apps in up status across all locations",
		},
	)
	siteAvailabilityTotalAppsDown = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "site_availability_total_apps_down",
			Help: "Total apps in down status across all locations",
		},
	)
	siteAvailabilityTotalAppsUnavailable = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "site_availability_total_apps_unavailable",
			Help: "Total apps in unavailable status across all locations",
		},
	)

	// Site sync metrics
	siteSyncAttempts = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "site_availability_sync_attempts_total",
			Help: "Total number of sync attempts",
		},
	)
	siteSyncFailures = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "site_availability_sync_failures_total",
			Help: "Total number of sync failures",
		},
	)
	siteSyncLatency = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "site_availability_sync_latency_seconds",
			Help:    "Sync operation latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)
	siteSyncLastSuccess = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "site_availability_sync_last_success_timestamp",
			Help: "Timestamp of last successful sync",
		},
	)
	siteSyncStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "site_availability_sync_status",
			Help: "Status of each synced site",
		},
		[]string{"site", "status"},
	)
)

// SetupMetricsHandler returns the handler for /metrics endpoint
func SetupMetricsHandler() http.Handler {
	appStatuses := handlers.GetAppStatusCache()

	// Step 1: Collect all unique label keys across all apps
	labelKeys := collectUniqueLabels(appStatuses)

	// Step 2: Create the dynamic metric with all collected label keys
	createDynamicMetric(labelKeys)

	// Step 3: Reset all metrics
	if siteAvailabilityStatus != nil {
		siteAvailabilityStatus.Reset()
	}
	siteAvailabilityApps.Reset()
	siteAvailabilityAppsUp.Reset()
	siteAvailabilityAppsDown.Reset()
	siteAvailabilityAppsUnavailable.Reset()

	// Step 4: Set values for each app with its specific labels
	for _, appStatus := range appStatuses {
		// Build label values for this app
		labelValues := buildLabelValues(appStatus, labelKeys)

		// Set the status value
		statusValue := 0.0
		if appStatus.Status == "up" {
			statusValue = 1.0
		}

		if siteAvailabilityStatus != nil {
			siteAvailabilityStatus.WithLabelValues(labelValues...).Set(statusValue)
		}
	}

	// Step 5: Update aggregated metrics (unchanged logic)
	updateAggregatedMetrics(appStatuses)

	return promhttp.Handler()
}

// collectUniqueLabels collects all unique label keys from all apps
func collectUniqueLabels(appStatuses []handlers.AppStatus) []string {
	labelKeysSet := make(map[string]bool)

	// Always include system labels in a specific order
	labelKeysSet["name"] = true
	labelKeysSet["location"] = true
	labelKeysSet["source"] = true
	labelKeysSet["origin_url"] = true

	// Add all app-specific labels (from all apps)
	for _, appStatus := range appStatuses {
		for _, label := range appStatus.Labels {
			if label.Value != "" { // Only include labels with non-empty values
				labelKeysSet[label.Key] = true
			}
		}
	}

	// Convert to sorted slice for consistent ordering
	labelKeys := make([]string, 0, len(labelKeysSet))

	// Add system labels first in specific order
	labelKeys = append(labelKeys, "name", "location", "source", "origin_url")

	// Add app-specific labels alphabetically
	var appLabels []string
	for key := range labelKeysSet {
		if key != "name" && key != "location" && key != "source" && key != "origin_url" {
			appLabels = append(appLabels, key)
		}
	}

	// Sort app labels alphabetically for consistency
	for i := 0; i < len(appLabels); i++ {
		for j := i + 1; j < len(appLabels); j++ {
			if appLabels[i] > appLabels[j] {
				appLabels[i], appLabels[j] = appLabels[j], appLabels[i]
			}
		}
	}

	labelKeys = append(labelKeys, appLabels...)
	return labelKeys
}

// createDynamicMetric creates and registers the siteAvailabilityStatus metric with dynamic labels
func createDynamicMetric(labelKeys []string) {
	// Unregister existing metric if it exists
	if siteAvailabilityStatus != nil {
		prometheus.DefaultRegisterer.Unregister(siteAvailabilityStatus)
	}

	// Create new metric with dynamic labels
	siteAvailabilityStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "site_availability_status",
			Help: "Site availability status by app and location (1=up, 0=down)",
		},
		labelKeys,
	)

	// Register the new metric
	prometheus.MustRegister(siteAvailabilityStatus)
}

// buildLabelValues builds the label values array for a given app
func buildLabelValues(appStatus handlers.AppStatus, labelKeys []string) []string {
	labelValues := make([]string, len(labelKeys))

	// Create a map for quick lookup of app labels
	appLabelsMap := make(map[string]string)
	for _, label := range appStatus.Labels {
		if label.Value != "" {
			appLabelsMap[label.Key] = label.Value
		}
	}

	// Fill in the values for each label key
	for i, key := range labelKeys {
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
			// For app-specific labels, use the value if it exists, otherwise empty string
			if value, exists := appLabelsMap[key]; exists {
				labelValues[i] = value
			} else {
				labelValues[i] = ""
			}
		}
	}

	return labelValues
}

// updateAggregatedMetrics updates the aggregated metrics (location stats, totals)
func updateAggregatedMetrics(appStatuses []handlers.AppStatus) {
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

// Init registers all Prometheus metrics
func Init() {
	// Note: siteAvailabilityStatus is registered dynamically in SetupMetricsHandler
	prometheus.MustRegister(siteAvailabilityApps)
	prometheus.MustRegister(siteAvailabilityAppsUp)
	prometheus.MustRegister(siteAvailabilityAppsDown)
	prometheus.MustRegister(siteAvailabilityAppsUnavailable)
	prometheus.MustRegister(siteAvailabilityTotalApps)
	prometheus.MustRegister(siteAvailabilityTotalAppsUp)
	prometheus.MustRegister(siteAvailabilityTotalAppsDown)
	prometheus.MustRegister(siteAvailabilityTotalAppsUnavailable)

	// Register site sync metrics
	prometheus.MustRegister(siteSyncAttempts)
	prometheus.MustRegister(siteSyncFailures)
	prometheus.MustRegister(siteSyncLatency)
	prometheus.MustRegister(siteSyncLastSuccess)
	prometheus.MustRegister(siteSyncStatus)
}

// SiteSyncMetrics provides access to site sync metrics
type SiteSyncMetrics struct {
	SyncAttempts prometheus.Counter
	SyncFailures prometheus.Counter
	SyncLatency  prometheus.Histogram
	LastSyncTime prometheus.Gauge
	SiteStatus   *prometheus.GaugeVec
}

// NewSiteSyncMetrics returns a new SiteSyncMetrics instance
func NewSiteSyncMetrics() *SiteSyncMetrics {
	return &SiteSyncMetrics{
		SyncAttempts: siteSyncAttempts,
		SyncFailures: siteSyncFailures,
		SyncLatency:  siteSyncLatency,
		LastSyncTime: siteSyncLastSuccess,
		SiteStatus:   siteSyncStatus,
	}
}
