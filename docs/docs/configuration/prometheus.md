---
sidebar_position: 4
---

# Prometheus Configuration

Configure Prometheus integration for metrics collection and monitoring.

## Basic Prometheus Setup

Example `prometheus.yml` configuration:

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "site_availability_rules.yml"

scrape_configs:
  - job_name: "site-availability"
    static_configs:
      - targets: ["site-availability:8080"]

  - job_name: "monitored-apps"
    static_configs:
      - targets:
          - "app1:8080"
          - "app2:8080"
          - "app3:9090"
    scrape_interval: 30s
    metrics_path: "/metrics"

alerting:
  alertmanagers:
    - static_configs:
        - targets:
            - alertmanager:9093
```

## Metrics Configuration

Configure which metrics to scrape from your applications:

```yaml
# config.yaml
apps:
  - name: web-server
    location: New York
    metric: up{instance="web-server:8080", job="web"}
    prometheus: http://prometheus:9090/

  - name: api-server
    location: London
    metric: up{instance="api-server:8080", job="api"}
    prometheus: http://prometheus:9090/

  - name: database
    location: Tokyo
    metric: pg_up{instance="postgres:5432", job="postgres"}
    prometheus: http://prometheus:9090/
```

## Custom Metrics

Define custom metrics for your applications:

```yaml
custom_metrics:
  - name: response_time
    query: 'http_request_duration_seconds{job="web"}'
    threshold: 0.5

  - name: error_rate
    query: 'rate(http_requests_total{status=~"5.."}[5m])'
    threshold: 0.05
```

## Recording Rules

Create recording rules for complex queries:

```yaml
# site_availability_rules.yml
groups:
  - name: site_availability
    rules:
      - record: site:availability:rate5m
        expr: avg_over_time(up[5m])

      - record: site:response_time:avg5m
        expr: avg_over_time(http_request_duration_seconds[5m])
```

## Alerting Rules

Define alerts for monitoring:

```yaml
groups:
  - name: site_availability_alerts
    rules:
      - alert: SiteDown
        expr: up == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Site {{ $labels.instance }} is down"

      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate on {{ $labels.instance }}"
```

## Federation

For multi-prometheus setups:

```yaml
# prometheus-federation.yml
scrape_configs:
  - job_name: "federate"
    scrape_interval: 15s
    honor_labels: true
    metrics_path: "/federate"
    params:
      "match[]":
        - '{job=~"site-availability.*"}'
        - '{__name__=~"up|site:.*"}'
    static_configs:
      - targets:
          - "prometheus-dc1:9090"
          - "prometheus-dc2:9090"
```
