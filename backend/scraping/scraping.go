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
	// Scrape performs a single scrape operation for a source with the given timeout and max parallel settings.
	// It returns the app statuses and an error if scraping fails.
	Scrape(source config.Source, timeout time.Duration, maxParallel int, tlsConfig *tls.Config) ([]handlers.AppStatus, error)
}

var (
	Scrapers        = make(map[string]Source)
	globalTLSConfig *tls.Config
)

// InitCertificate loads CA certificates from the file paths listed in the given environment variable name.
// The environment variable value may contain multiple file paths separated by ":".
func InitCertificate(envVarName string) {
	caPath := os.Getenv(envVarName)
	if caPath == "" {
		logging.Logger.WithField("env", envVarName).Info("Env var not set. Using default TLS config.")
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

	logging.Logger.WithField("env", envVarName).Info("Custom CA certificates loaded successfully")
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
		switch src.Type {
		case "prometheus":
			Scrapers[src.Name] = prometheus.NewPrometheusScraper()
		case "site":
			Scrapers[src.Name] = site.NewSiteScraper()
			// Add more types as needed
		}
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
			statuses, err := scraper.Scrape(source, timeout, cfg.Scraping.MaxParallel, globalTLSConfig)
			if err != nil {
				logging.Logger.WithError(err).WithField("source", source.Name).Error("Initial scraper failed")
			} else {
				// Update the app status cache
				handlers.UpdateAppStatus(source.Name, statuses)
				logging.Logger.WithFields(map[string]interface{}{
					"source":    source.Name,
					"app_count": len(statuses),
				}).Info("Updated app status cache after initial scrape")
			}

			// Continue scraping at intervals
			for range ticker.C {
				statuses, err := scraper.Scrape(source, timeout, cfg.Scraping.MaxParallel, globalTLSConfig)
				if err != nil {
					logging.Logger.WithError(err).WithField("source", source.Name).Error("Scraper failed")
				} else {
					// Update the app status cache
					handlers.UpdateAppStatus(source.Name, statuses)
					logging.Logger.WithFields(map[string]interface{}{
						"source":    source.Name,
						"app_count": len(statuses),
					}).Debug("Updated app status cache after scrape")
				}
			}
		}(source)
	}
}
