package scraping

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"
	"site-availability/config"
	"site-availability/handlers"
	"site-availability/logging"
	"site-availability/scraping/prometheus"
	"site-availability/scraping/site"
	"strings"
	"time"
)

// Source defines the interface for all data sources (Prometheus, Site, etc.)
type Source interface {
	// ValidateConfig validates the source-specific configuration
	ValidateConfig(source config.Source) error
	// Scrape performs a single scrape operation for a source with the given timeout and max parallel settings.
	// It returns the app statuses, locations, and an error if scraping fails.
	// The serverSettings parameter is passed for label merging purposes.
	Scrape(source config.Source, serverSettings config.ServerSettings, timeout time.Duration, maxParallel int, tlsConfig *tls.Config) ([]handlers.AppStatus, []handlers.Location, error)
}

var (
	Scrapers        = make(map[string]Source)
	globalTLSConfig *tls.Config
)

// createUnavailableStatuses creates empty app statuses when a scraper fails completely
// Since we can't access source-specific config when scraper fails, we return empty slice
// Individual scrapers should handle their own failure cases appropriately
func createUnavailableStatuses(source config.Source) []handlers.AppStatus {
	logging.Logger.WithFields(map[string]interface{}{
		"source": source.Name,
	}).Warn("Scraper failed completely - returning empty app statuses")

	return []handlers.AppStatus{}
}

// InitCertificateFromPath loads CA certificates from the given file path(s).
// Multiple paths can be separated by ":".
func InitCertificateFromPath(caPath string) {
	if caPath == "" {
		logging.Logger.Info("No CA certificate path provided. Using default TLS config.")
		globalTLSConfig = nil
		return
	}

	caCertPool := x509.NewCertPool()
	paths := strings.Split(caPath, ":")

	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}

		certData, err := os.ReadFile(path)
		if err != nil {
			logging.Logger.WithError(err).WithField("path", path).Error("Failed to read CA certificate")
			continue
		}

		if ok := caCertPool.AppendCertsFromPEM(certData); !ok {
			logging.Logger.WithField("path", path).Error("Failed to append CA certificate")
			continue
		}
	}

	globalTLSConfig = &tls.Config{
		RootCAs:            caCertPool,
		InsecureSkipVerify: false, // Only set true if you know what you're doing
	}

	logging.Logger.WithField("ca_path", caPath).Info("Custom CA certificates loaded successfully")
}

// GetHTTPClient creates an HTTP client with the configured timeout and TLS config
func GetHTTPClient(timeout time.Duration) *http.Client {
	client := &http.Client{Timeout: timeout}

	if globalTLSConfig != nil {
		client.Transport = &http.Transport{
			TLSClientConfig: globalTLSConfig,
		}
	}

	return client
}

func InitScrapers(cfg *config.Config) {
	for _, src := range cfg.Sources {
		var scraper Source
		switch src.Type {
		case "prometheus":
			scraper = prometheus.NewPrometheusScraper()
		case "site":
			scraper = site.NewSiteScraper()
		default:
			// Fail immediately on first unsupported type
			logging.Logger.WithFields(map[string]interface{}{
				"source_name":     src.Name,
				"source_type":     src.Type,
				"supported_types": []string{"prometheus", "site"},
			}).Fatal("Unsupported source type encountered. Please fix the source type in your configuration file.")
		}

		// Validate the source configuration
		if err := scraper.ValidateConfig(src); err != nil {
			logging.Logger.WithError(err).WithField("source", src.Name).Fatal("Invalid source configuration")
		}

		Scrapers[src.Name] = scraper
	}
}

func Start(cfg *config.Config) {
	interval, err := time.ParseDuration(cfg.Scraping.Interval)
	if err != nil {
		logging.Logger.WithError(err).Fatal("Invalid scraping interval")
	}

	timeout, err := time.ParseDuration(cfg.Scraping.Timeout)
	if err != nil {
		logging.Logger.WithError(err).Fatal("Invalid scraping timeout")
	}

	for _, source := range cfg.Sources {
		go func(source config.Source) {
			scraper, ok := Scrapers[source.Name]
			if !ok {
				logging.Logger.WithField("source", source.Name).Fatal("Unsupported source name")
			}

			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			// Perform initial scrape immediately
			statuses, locations, err := scraper.Scrape(source, cfg.ServerSettings, timeout, cfg.Scraping.MaxParallel, globalTLSConfig)
			if err != nil {
				logging.Logger.WithError(err).WithField("source", source.Name).Error("Initial scraper failed")
				// Create unavailable statuses for all configured apps when scraper fails completely
				statuses = createUnavailableStatuses(source)
				locations = []handlers.Location{} // No locations when scraper fails
			}

			// Always update caches, even on scraper failure
			handlers.UpdateAppStatus(source.Name, statuses, source, cfg.ServerSettings)
			handlers.UpdateLocationCache(source.Name, locations)
			logging.Logger.WithFields(map[string]interface{}{
				"source":         source.Name,
				"app_count":      len(statuses),
				"location_count": len(locations),
				"scraper_error":  err != nil,
			}).Info("Updated app status and location caches after initial scrape")

			// Continue scraping at intervals
			for range ticker.C {
				statuses, locations, err := scraper.Scrape(source, cfg.ServerSettings, timeout, cfg.Scraping.MaxParallel, globalTLSConfig)
				if err != nil {
					logging.Logger.WithError(err).WithField("source", source.Name).Error("Scraper failed")
					// Create unavailable statuses for all configured apps when scraper fails completely
					statuses = createUnavailableStatuses(source)
					locations = []handlers.Location{} // No locations when scraper fails
				}

				// Always update caches, even on scraper failure
				handlers.UpdateAppStatus(source.Name, statuses, source, cfg.ServerSettings)
				handlers.UpdateLocationCache(source.Name, locations)
				logging.Logger.WithFields(map[string]interface{}{
					"source":         source.Name,
					"app_count":      len(statuses),
					"location_count": len(locations),
					"scraper_error":  err != nil,
				}).Debug("Updated app status and location caches after scrape")
			}
		}(source)
	}
}
