# Local Kubernetes Testing with Helm

This directory contains Helm values files for testing the Site Availability application locally on Kubernetes with multiple instances that can scrape each other.

## Prerequisites

1. Local Kubernetes cluster (minikube, kind, or Docker Desktop)
2. Helm 3.x installed
3. NGINX Ingress Controller (if using ingress)
4. Prometheus Operator (optional, for ServiceMonitor support)

## Quick Start

### Deploy Server A

```bash
helm install server-a ./chart -f helpers/helm/values-server-a.yaml
```

### Deploy Server B

```bash
helm install server-b ./chart -f helpers/helm/values-server-b.yaml
```

### Access the Applications

If using ingress, add these entries to your `/etc/hosts`:

```
127.0.0.1 server-a.local
127.0.0.1 server-b.local
```

Then access:

- Server A: http://server-a.local
- Server B: http://server-b.local

### Port Forwarding (Alternative)

If not using ingress:

```bash
kubectl port-forward svc/site-availability-server-a 8080:8080
kubectl port-forward svc/site-availability-server-b 8081:8080
```

Access:

- Server A: http://localhost:8080
- Server B: http://localhost:8081

## Configuration

The configurations mirror the docker-compose multiple-servers setup:

- **Server A**: Production environment with comprehensive monitoring
- **Server B**: Staging environment that also monitors Server A

## Features

- ✅ Multiple server instances
- ✅ Cross-server monitoring (Server B monitors Server A)
- ✅ Prometheus integration
- ✅ ServiceMonitor for Prometheus Operator
- ✅ Ingress for local access
- ✅ Resource limits and security contexts
- ✅ Grafana dashboard and Prometheus rules

## Cleanup

```bash
helm uninstall server-a
helm uninstall server-b
```
