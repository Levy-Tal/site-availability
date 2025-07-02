---
sidebar_position: 3
---

# Helm Chart Deployment

Deploy Site Availability Monitoring using the provided Helm chart for Kubernetes.

## Chart Overview

The Helm chart includes:

- Backend deployment and service
- Frontend deployment and service
- ConfigMap for configuration
- ServiceMonitor for Prometheus
- Grafana dashboard ConfigMap
- Ingress configuration
- Optional external secrets

## Installation

### Quick Install

```bash
# From the repository root
helm install site-availability chart/ --namespace monitoring --create-namespace
```

### Custom Installation

Create a `values.yaml` file:

```yaml
# Image configuration
backend:
  image:
    repository: your-registry/site-availability-backend
    tag: "v1.0.0"
    pullPolicy: IfNotPresent

  replicas: 3

  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 512Mi

frontend:
  image:
    repository: your-registry/site-availability-frontend
    tag: "v1.0.0"

  replicas: 2

# Configuration
config:
  scrapeInterval: 30s
  logLevel: info

  locations:
    - name: Production East
      latitude: 40.712776
      longitude: -74.005974
    - name: Production West
      latitude: 37.7749
      longitude: -122.4194

  apps:
    - name: api-east
      location: Production East
      metric: up{instance="api-east:8080", job="api"}
      prometheus: http://prometheus:9090/
    - name: api-west
      location: Production West
      metric: up{instance="api-west:8080", job="api"}
      prometheus: http://prometheus:9090/

# Ingress configuration
ingress:
  enabled: true
  className: nginx
  hostname: monitoring.example.com
  tls:
    enabled: true
    secretName: site-availability-tls

# Monitoring
serviceMonitor:
  enabled: true
  namespace: monitoring

grafanaDashboard:
  enabled: true
  namespace: monitoring

# External secrets (optional)
externalSecrets:
  enabled: false
  secretStore: vault-backend
  secrets:
    hmacSecret:
      key: site-availability/hmac-secret
      property: secret
```

Install with custom values:

```bash
helm install site-availability chart/ -f values.yaml --namespace monitoring
```

## Configuration Options

### Backend Configuration

```yaml
backend:
  image:
    repository: site-availability/backend
    tag: latest
    pullPolicy: IfNotPresent

  replicas: 1

  service:
    type: ClusterIP
    port: 8080

  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 512Mi

  nodeSelector: {}
  tolerations: []
  affinity: {}

  env:
    - name: LOG_LEVEL
      value: info
```

### Frontend Configuration

```yaml
frontend:
  image:
    repository: site-availability/frontend
    tag: latest

  replicas: 1

  service:
    type: ClusterIP
    port: 80

  env:
    - name: REACT_APP_API_URL
      value: /api
```

### Ingress Configuration

```yaml
ingress:
  enabled: true
  className: nginx
  hostname: monitoring.example.com

  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    cert-manager.io/cluster-issuer: letsencrypt-prod

  tls:
    enabled: true
    secretName: site-availability-tls

  paths:
    frontend:
      path: /
      pathType: Prefix
    backend:
      path: /api
      pathType: Prefix
```

## Upgrades

### Upgrade Process

```bash
# Update chart values
helm upgrade site-availability chart/ -f values.yaml --namespace monitoring

# Force recreation of pods
helm upgrade site-availability chart/ --force --namespace monitoring

# Rollback if needed
helm rollback site-availability 1 --namespace monitoring
```

### Version History

```bash
# View release history
helm history site-availability --namespace monitoring

# Get release status
helm status site-availability --namespace monitoring
```

## Monitoring Integration

### Prometheus ServiceMonitor

```yaml
serviceMonitor:
  enabled: true
  namespace: monitoring
  labels:
    release: prometheus
  interval: 30s
  path: /metrics
```

### Grafana Dashboard

```yaml
grafanaDashboard:
  enabled: true
  namespace: monitoring
  labels:
    grafana_dashboard: "1"
  annotations:
    grafana_folder: "Site Availability"
```

## External Secrets

Integration with external secret management:

```yaml
externalSecrets:
  enabled: true
  secretStore: vault-backend
  refreshInterval: 1h

  secrets:
    hmacSecret:
      name: site-availability-hmac
      key: site-availability/hmac-secret
      property: secret

    tlsCert:
      name: site-availability-tls
      key: site-availability/tls
      properties:
        - cert
        - key
```

## Customization

### Custom Values Templates

Create environment-specific values:

```bash
# values-development.yaml
config:
  logLevel: debug
backend:
  replicas: 1
ingress:
  hostname: monitoring-dev.example.com

# values-production.yaml
config:
  logLevel: warn
backend:
  replicas: 3
  resources:
    requests:
      cpu: 200m
      memory: 256Mi
    limits:
      cpu: 1000m
      memory: 1Gi
```

### Helm Hooks

The chart includes pre-install and pre-upgrade hooks for:

- Database migrations
- Configuration validation
- Health checks

## Troubleshooting

### Chart Validation

```bash
# Lint the chart
helm lint chart/

# Dry run installation
helm install site-availability chart/ --dry-run --debug

# Template the chart
helm template site-availability chart/ -f values.yaml
```

### Common Issues

**Image pull errors:**

```bash
# Check image exists
docker pull site-availability/backend:latest

# Verify image registry credentials
kubectl get secrets -n monitoring
```

**Configuration errors:**

```bash
# Check ConfigMap
kubectl get configmap site-availability-config -o yaml

# Validate YAML syntax
helm template chart/ | kubectl apply --dry-run=client -f -
```

**Service connectivity:**

```bash
# Test internal service
kubectl run debug --image=busybox -it --rm -- /bin/sh
wget -qO- http://site-availability-backend:8080/health
```

## Uninstallation

```bash
# Uninstall release
helm uninstall site-availability --namespace monitoring

# Clean up namespace (if desired)
kubectl delete namespace monitoring
```
