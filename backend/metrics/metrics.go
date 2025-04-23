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
		[]string{"app", "location"},
	)

	// Per location aggregated metrics
	siteAvailabilityApps = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "site_availability_apps",
			Help: "Total apps monitored in a location",
		},
		[]string{"location"},
	)
	siteAvailabilityAppsUp = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "site_availability_apps_up",
			Help: "Count of apps in up status per location",
		},
		[]string{"location"},
	)
	siteAvailabilityAppsDown = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "site_availability_apps_down",
			Help: "Count of apps in down status per location",
		},
		[]string{"location"},
	)
	siteAvailabilityAppsUnavailable = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "site_availability_apps_unavailable",
			Help: "Count of apps in unavailable status per location",
		},
		[]string{"location"},
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

	// Track per-location counts
	locationCounts := make(map[string]struct {
		total       int
		up          int
		down        int
		unavailable int
	})

	for _, appStatus := range appStatuses {
		app := appStatus.Name
		location := appStatus.Location
		status := appStatus.Status

		if _, exists := locationCounts[location]; !exists {
			locationCounts[location] = struct {
				total, up, down, unavailable int
			}{}
		}
		counts := locationCounts[location]
		counts.total++
		totalApps++

		switch status {
		case "up":
			siteAvailabilityStatus.WithLabelValues(app, location).Set(1)
			counts.up++
			totalUp++
		case "down":
			siteAvailabilityStatus.WithLabelValues(app, location).Set(0)
			counts.down++
			totalDown++
		default:
			// Unavailable or unknown statuses
			counts.unavailable++
			totalUnavailable++
		}

		locationCounts[location] = counts
	}

	// Update per-location metrics
	for location, counts := range locationCounts {
		siteAvailabilityApps.WithLabelValues(location).Set(float64(counts.total))
		siteAvailabilityAppsUp.WithLabelValues(location).Set(float64(counts.up))
		siteAvailabilityAppsDown.WithLabelValues(location).Set(float64(counts.down))
		siteAvailabilityAppsUnavailable.WithLabelValues(location).Set(float64(counts.unavailable))
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
}
