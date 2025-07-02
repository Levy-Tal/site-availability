---
sidebar_position: 2
---

# Kubernetes Deployment

Deploy Site Availability Monitoring on Kubernetes for scalable, production environments.

## Prerequisites

- Kubernetes cluster (v1.20+)
- kubectl configured
- Helm 3.0+ (recommended)

## Helm Deployment (Recommended)

### Quick Start

```bash
# Add the repository (if available)
helm repo add site-availability https://levy-tal.github.io/site-availability

# Or use local chart
cd chart/
helm install site-availability . --namespace monitoring --create-namespace
```

### Custom Values

Create `values.yaml`:

```yaml
backend:
  image:
    repository: site-availability/backend
    tag: latest
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
    repository: site-availability/frontend
    tag: latest
  replicas: 2

prometheus:
  enabled: true
  server:
    persistentVolume:
      size: 10Gi

ingress:
  enabled: true
  hostname: monitoring.example.com
  tls: true
```

Deploy with custom values:

```bash
helm install site-availability . -f values.yaml
```

## Manual Kubernetes Deployment

### Namespace

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: site-availability
```

### ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: site-availability-config
  namespace: site-availability
data:
  config.yaml: |
    scrape_interval: 30s
    locations:
      - name: Production
        latitude: 40.712776
        longitude: -74.005974
    apps:
      - name: api
        location: Production
        metric: up{instance="api:8080"}
        prometheus: http://prometheus:9090/
```

### Backend Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: site-availability-backend
  namespace: site-availability
spec:
  replicas: 3
  selector:
    matchLabels:
      app: site-availability-backend
  template:
    metadata:
      labels:
        app: site-availability-backend
    spec:
      containers:
        - name: backend
          image: site-availability/backend:latest
          ports:
            - containerPort: 8080
          env:
            - name: CONFIG_FILE
              value: /app/config.yaml
          volumeMounts:
            - name: config
              mountPath: /app/config.yaml
              subPath: config.yaml
          resources:
            requests:
              cpu: 100m
              memory: 128Mi
            limits:
              cpu: 500m
              memory: 512Mi
      volumes:
        - name: config
          configMap:
            name: site-availability-config
```

### Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: site-availability-backend
  namespace: site-availability
spec:
  selector:
    app: site-availability-backend
  ports:
    - port: 8080
      targetPort: 8080
  type: ClusterIP
```

### Ingress

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: site-availability
  namespace: site-availability
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
    - hosts:
        - monitoring.example.com
      secretName: site-availability-tls
  rules:
    - host: monitoring.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: site-availability-frontend
                port:
                  number: 80
          - path: /api
            pathType: Prefix
            backend:
              service:
                name: site-availability-backend
                port:
                  number: 8080
```

## Monitoring and Observability

### ServiceMonitor (Prometheus Operator)

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: site-availability
  namespace: site-availability
spec:
  selector:
    matchLabels:
      app: site-availability-backend
  endpoints:
    - port: http
      path: /metrics
```

### Grafana Dashboard

Import the provided dashboard from `chart/grafana-dashboards/dashboard.json`.

## Scaling

### Horizontal Pod Autoscaler

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: site-availability-backend-hpa
  namespace: site-availability
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: site-availability-backend
  minReplicas: 2
  maxReplicas: 10
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
```

## Security

### Network Policies

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: site-availability-netpol
  namespace: site-availability
spec:
  podSelector:
    matchLabels:
      app: site-availability-backend
  policyTypes:
    - Ingress
    - Egress
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              name: ingress-nginx
      ports:
        - protocol: TCP
          port: 8080
  egress:
    - to:
        - namespaceSelector:
            matchLabels:
              name: monitoring
      ports:
        - protocol: TCP
          port: 9090
```

## Troubleshooting

```bash
# Check pod status
kubectl get pods -n site-availability

# View pod logs
kubectl logs -f deployment/site-availability-backend -n site-availability

# Describe pod for events
kubectl describe pod <pod-name> -n site-availability

# Check service endpoints
kubectl get endpoints -n site-availability
```
