---
sidebar_position: 4
---

# Metrics Integration

Monitor Site Availability with Prometheus metrics and integrate with observability platforms.

## Prometheus Metrics Endpoint

The application exposes metrics at `/metrics` in Prometheus format.

### Default Metrics

#### Application Metrics

```prometheus
# Application availability status
site_availability_up{app="frontend",location="New York"} 1
site_availability_up{app="backend",location="London"} 0

# Scrape duration in seconds
site_availability_scrape_duration_seconds{target="frontend"} 0.142

# Scrape requests total
site_availability_scrape_requests_total{target="frontend",status="success"} 1245
site_availability_scrape_requests_total{target="backend",status="error"} 3

# Last successful scrape timestamp
site_availability_last_scrape_timestamp{target="frontend"} 1638360000
```

#### HTTP Metrics

```prometheus
# HTTP requests total
http_requests_total{method="GET",path="/api/apps",status="200"} 1245
http_requests_total{method="POST",path="/api/scrape-interval",status="200"} 12

# HTTP request duration
http_request_duration_seconds{method="GET",path="/api/apps"} 0.045

# Active HTTP connections
http_connections_active 5
```

#### System Metrics

```prometheus
# Go runtime metrics
go_goroutines 25
go_memstats_alloc_bytes 2.5e+06
go_memstats_gc_duration_seconds 0.001

# Process metrics
process_cpu_seconds_total 12.5
process_resident_memory_bytes 2.5e+07
process_uptime_seconds 3600
```

## Custom Metrics

### Business Metrics

```prometheus
# Overall system availability
site_availability_system_uptime_percentage 98.7

# Applications by status
site_availability_apps_by_status{status="up"} 4
site_availability_apps_by_status{status="down"} 1

# Response time percentiles
site_availability_response_time_p50 0.125
site_availability_response_time_p95 0.245
site_availability_response_time_p99 0.389
```

### Location Metrics

```prometheus
# Applications per location
site_availability_location_apps{location="New York"} 3
site_availability_location_apps{location="London"} 2

# Location availability
site_availability_location_uptime{location="New York"} 0.667
site_availability_location_uptime{location="London"} 1.0
```

## Prometheus Configuration

### Scrape Configuration

Add Site Availability Monitoring to your `prometheus.yml`:

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: "site-availability"
    static_configs:
      - targets: ["site-availability:8080"]
    scrape_interval: 30s
    metrics_path: "/metrics"
    scheme: "http"

    # Optional: Add labels
    relabel_configs:
      - target_label: "service"
        replacement: "site-availability"
      - target_label: "environment"
        replacement: "production"
```

### Service Discovery

For Kubernetes deployments:

```yaml
scrape_configs:
  - job_name: "site-availability-k8s"
    kubernetes_sd_configs:
      - role: pod
        namespaces:
          names: ["monitoring"]

    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_label_app]
        action: keep
        regex: site-availability-backend

      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true

      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
```

## Recording Rules

Create recording rules for common queries:

```yaml
# recording-rules.yml
groups:
  - name: site_availability_rules
    interval: 30s
    rules:
      # Overall system availability
      - record: site_availability:system:uptime_5m
        expr: avg_over_time(site_availability_up[5m])

      # Application availability by location
      - record: site_availability:location:uptime_5m
        expr: avg_over_time(site_availability_up[5m]) by (location)

      # Response time moving average
      - record: site_availability:response_time:avg_5m
        expr: avg_over_time(site_availability_scrape_duration_seconds[5m])

      # Error rate
      - record: site_availability:error_rate_5m
        expr: rate(site_availability_scrape_requests_total{status="error"}[5m])
```

## Alerting Rules

Set up alerts for critical conditions:

```yaml
# alerting-rules.yml
groups:
  - name: site_availability_alerts
    rules:
      # Application down alert
      - alert: ApplicationDown
        expr: site_availability_up == 0
        for: 1m
        labels:
          severity: critical
          service: site-availability
        annotations:
          summary: "Application {{ $labels.app }} is down"
          description: "Application {{ $labels.app }} in {{ $labels.location }} has been down for more than 1 minute"
          runbook_url: "https://docs.example.com/runbooks/app-down"

      # High error rate alert
      - alert: HighErrorRate
        expr: rate(site_availability_scrape_requests_total{status="error"}[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
          service: site-availability
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value | humanizePercentage }} for target {{ $labels.target }}"

      # System availability alert
      - alert: LowSystemAvailability
        expr: site_availability:system:uptime_5m < 0.95
        for: 10m
        labels:
          severity: warning
          service: site-availability
        annotations:
          summary: "System availability below threshold"
          description: "System availability is {{ $value | humanizePercentage }}, below 95% threshold"

      # Scraping issues
      - alert: ScrapingDown
        expr: up{job="site-availability"} == 0
        for: 2m
        labels:
          severity: critical
          service: site-availability
        annotations:
          summary: "Site Availability Monitoring scraping is down"
          description: "Prometheus cannot scrape Site Availability Monitoring metrics"
```

## Grafana Integration

### Dashboard Configuration

Import the provided dashboard from `chart/grafana-dashboards/dashboard.json` or create custom dashboards:

#### Overview Dashboard

```json
{
  "dashboard": {
    "title": "Site Availability Monitoring",
    "panels": [
      {
        "title": "System Overview",
        "type": "stat",
        "targets": [
          {
            "expr": "site_availability:system:uptime_5m * 100",
            "legendFormat": "Availability %"
          }
        ]
      },
      {
        "title": "Applications Status",
        "type": "piechart",
        "targets": [
          {
            "expr": "count by (status) (site_availability_up)",
            "legendFormat": "{{ status }}"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "site_availability_scrape_duration_seconds",
            "legendFormat": "{{ target }}"
          }
        ]
      }
    ]
  }
}
```

#### Application Details Dashboard

```json
{
  "dashboard": {
    "title": "Application Details",
    "templating": {
      "list": [
        {
          "name": "app",
          "type": "query",
          "query": "label_values(site_availability_up, app)"
        }
      ]
    },
    "panels": [
      {
        "title": "Availability - $app",
        "type": "graph",
        "targets": [
          {
            "expr": "site_availability_up{app=\"$app\"}",
            "legendFormat": "{{ location }}"
          }
        ]
      }
    ]
  }
}
```

### Grafana Provisioning

Automatically provision dashboards:

```yaml
# grafana/provisioning/dashboards/site-availability.yml
apiVersion: 1
providers:
  - name: "site-availability"
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    allowUiUpdates: true
    options:
      path: /var/lib/grafana/dashboards/site-availability
```

## Custom Metrics Implementation

### Adding New Metrics

```go
// metrics/custom.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // Custom business metric
    ApplicationResponseTime = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "site_availability_app_response_time_seconds",
            Help: "Application response time in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"app", "location", "status_code"},
    )

    // Configuration change metric
    ConfigurationChanges = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "site_availability_config_changes_total",
            Help: "Total configuration changes",
        },
        []string{"type", "user"},
    )

    // Cache hit ratio
    CacheHitRatio = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "site_availability_cache_hit_ratio",
            Help: "Cache hit ratio",
        },
        []string{"cache_type"},
    )
)

// Usage in application code
func recordMetrics(app, location string, responseTime float64, statusCode int) {
    ApplicationResponseTime.WithLabelValues(
        app,
        location,
        fmt.Sprintf("%d", statusCode),
    ).Observe(responseTime)
}
```

### Instrumenting Code

```go
// Example: Instrument HTTP handlers
func instrumentHandler(handler http.HandlerFunc) http.HandlerFunc {
    return promhttp.InstrumentHandlerDuration(
        prometheus.NewHistogramVec(
            prometheus.HistogramOpts{
                Name: "http_request_duration_seconds",
                Help: "HTTP request duration",
            },
            []string{"method", "path", "status"},
        ),
        handler,
    )
}

// Example: Instrument scraping operations
func (s *Scraper) instrumentedScrape(target string) error {
    timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
        ScrapeDuration.WithLabelValues(target).Observe(v)
    }))
    defer timer.ObserveDuration()

    err := s.scrape(target)

    status := "success"
    if err != nil {
        status = "error"
    }

    ScrapeRequests.WithLabelValues(target, status).Inc()

    return err
}
```

## Observability Best Practices

### Metric Naming

Follow Prometheus naming conventions:

- Use `_total` suffix for counters
- Use `_seconds` suffix for time durations
- Use descriptive names with units
- Group related metrics with common prefixes

### Label Best Practices

- Keep cardinality low (< 1000 unique combinations)
- Use meaningful label names
- Avoid high-cardinality labels (user IDs, request IDs)
- Be consistent across metrics

### Performance Considerations

```go
// Use label values caching
var httpRequestsCounter = prometheus.NewCounterVec(...)

// Pre-create metric instances for known label combinations
func init() {
    for _, method := range []string{"GET", "POST", "PUT", "DELETE"} {
        for _, path := range knownPaths {
            httpRequestsCounter.WithLabelValues(method, path)
        }
    }
}
```

## Monitoring the Monitor

Monitor Site Availability Monitoring itself:

### Self-Monitoring Metrics

```prometheus
# Monitor scraping health
up{job="site-availability"} 1

# Monitor response times
prometheus_rule_evaluation_duration_seconds{rule_group="site_availability_rules"}

# Monitor disk usage
prometheus_tsdb_symbol_table_size_bytes
```

### Health Checks

```yaml
# Kubernetes liveness probe
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10

# Kubernetes readiness probe
readinessProbe:
  httpGet:
    path: /metrics
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
```

This comprehensive metrics integration ensures full observability of your Site Availability Monitoring system and the applications it monitors.
