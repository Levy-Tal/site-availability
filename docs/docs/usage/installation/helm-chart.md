---
sidebar_position: 4
---

# Production Deployment with Helm

This guide shows how to deploy Site Availability Monitoring in production using the official Helm chart, with real configuration examples.

## Minimal Production Install

```bash
helm repo add site-availability https://levytal.github.io/site-availability/
helm install site-availability site-availability/site-availability \
  --set image.tag=2.4.0 \
  --set replicaCount=3
```

## Example values.yaml

Below is a minimal, production-ready values.yaml using real defaults from the chart:

```yaml
replicaCount: 3

image:
  repository: levytal/site-availability
  tag: "2.4.0"
  pullPolicy: Always

configFile:
  enabled: true
  content:
    server_settings:
      port: "8080"
      sync_enable: false
    scraping:
      interval: "10s"
      timeout: "5s"
      max_parallel: 10
    documentation:
      title: "DR documentation"
      url: "https://google.com"
    locations:
      - name: "Hadera"
        latitude: 32.446235
        longitude: 34.903852
      - name: "Jerusalem"
        latitude: 31.782904
        longitude: 35.214774
      - name: "Beer Sheva"
        latitude: 31.245381
        longitude: 34.788745
    sources:
      - name: "prometheus-main"
        type: "prometheus"
        url: "http://prometheus-operated:9090"
        apps:
          - name: "app1"
            location: "Hadera"
            metric: 'up{job="site-availability"}'
          - name: "app2"
            location: "Jerusalem"
            metric: 'up{container="alertmanager"}'
          - name: "app3"
            location: "Beer Sheva"
            metric: 'up{container="alertmanager"}'
```

## Scaling and High Availability

- **replicaCount**: Set to 2+ for high availability.
- **resources**: Adjust CPU/memory as needed for your workload.
- **service.type**: Use `ClusterIP` for internal, or `LoadBalancer` for external access.

## Security Best Practices

- **Use HTTPS** via Ingress.
- **Set resource limits** to prevent noisy neighbor issues.
- **Store credentials in Kubernetes secrets** and reference them via `credentials.credentialsSecretName`.

## Monitoring

- **ServiceMonitor** and **PrometheusRules** are enabled by default for Prometheus integration.
- **Grafana dashboards** are included in the chart.

## Upgrade and Rollback

Upgrade:

```bash
helm upgrade site-availability site-availability/site-availability -f values.yaml
```

Rollback:

```bash
helm rollback site-availability 1
```

---

For more advanced configuration, see the comments in `chart/values.yaml` and the Helm chart documentation.
