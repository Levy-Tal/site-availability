---
sidebar_position: 4
---

# Backend Sources

This page explains how backend sources work and how to add a new source type to Site Availability.

## Source Interface

All sources must implement the following interface (see `scraping/scraping.go`):

```go
type Source interface {
    // ValidateConfig validates the source-specific configuration
    ValidateConfig(source config.Source) error
    // Scrape performs a single scrape operation for a source
    Scrape(source config.Source, serverSettings config.ServerSettings, timeout time.Duration, maxParallel int, tlsConfig *tls.Config) ([]handlers.AppStatus, []handlers.Location, error)
}
```

## Supported Source Types

- **prometheus**: Scrapes metrics from a Prometheus instance using PromQL.
- **http**: Checks HTTP endpoints for status and content.
- **site**: Scrapes all app statuses from another Site Availability instance via `/sync`.

## Adding a New Source

1. **Create a new package** under `backend/scraping/` (e.g., `mynewscraper`).
2. **Implement the Source interface** with your logic.
3. **Register your source type** in `scraping.go` (add a case in `InitScrapers`).
4. **Add configuration validation and documentation.**

## Minimal Source Config Example

```yaml
sources:
  - name: "prometheus-main"
    type: prometheus
    config:
      url: http://prometheus:9090
      apps:
        - name: myApp
          location: New York
          metric: up{instance="app:8080", job="app"}

  - name: "web-monitoring"
    type: http
    config:
      apps:
        - name: website
          location: US-East
          url: https://example.com

  - name: "Server A"
    type: site
    config:
      url: http://server-a:8080
```

## Best Practices

- Validate all required config fields in `ValidateConfig`.
- Handle errors gracefully in `Scrape`.
- Use structured logging for debugging.

---

For more details, see the code in `backend/scraping/` and the real config examples in `helpers/config/`.
