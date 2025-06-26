package metrics

import (
	"net/http"
	"site-availability/handlers"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Per location/app status metrics
	siteAvailabilityStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "site_availability_status",
			Help: "Site availability status by app and location (1=up, 0=down)",
		},
		[]string{"app", "location", "source"},
	)

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
			Name: "site_sync_attempts_total",
			Help: "Total number of sync attempts",
		},
	)
	siteSyncFailures = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "site_sync_failures_total",
			Help: "Total number of sync failures",
		},
	)
	siteSyncLatency = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "site_sync_latency_seconds",
			Help:    "Sync operation latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)
	siteSyncLastSuccess = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "site_sync_last_success_timestamp",
			Help: "Timestamp of last successful sync",
		},
	)
	siteSyncStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "site_sync_status",
			Help: "Status of each synced site",
		},
		[]string{"site", "status"},
	)
)

// SetupMetricsHandler returns the handler for /metrics endpoint
func SetupMetricsHandler() http.Handler {
	appStatuses := handlers.GetAppStatusCache()

	// Reset all metrics
	siteAvailabilityStatus.Reset()
	siteAvailabilityApps.Reset()
	siteAvailabilityAppsUp.Reset()
	siteAvailabilityAppsDown.Reset()
	siteAvailabilityAppsUnavailable.Reset()

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
		app := appStatus.Name
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
			siteAvailabilityStatus.WithLabelValues(app, location, source).Set(1)
			counts.up++
			totalUp++
		case "down":
			siteAvailabilityStatus.WithLabelValues(app, location, source).Set(0)
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

	return promhttp.Handler()
}

// Init registers all Prometheus metrics and starts any background collectors
func Init() {
	prometheus.MustRegister(siteAvailabilityStatus)
	prometheus.MustRegister(siteAvailabilityApps)
	prometheus.MustRegister(siteAvailabilityAppsUp)
	prometheus.MustRegister(siteAvailabilityAppsDown)
	prometheus.MustRegister(siteAvailabilityAppsUnavailable)
	prometheus.MustRegister(siteAvailabilityTotalApps)
	prometheus.MustRegister(siteAvailabilityTotalAppsUp)
	prometheus.MustRegister(siteAvailabilityTotalAppsDown)
	prometheus.MustRegister(siteAvailabilityTotalAppsUnavailable)
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
