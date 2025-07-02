---
sidebar_position: 4
---

# Production Deployment

Best practices and considerations for deploying Site Availability Monitoring in production environments.

## Production Checklist

### Security

- [ ] Enable HMAC authentication
- [ ] Use HTTPS/TLS everywhere
- [ ] Configure network policies
- [ ] Set up secrets management
- [ ] Enable audit logging
- [ ] Regular security updates

### Reliability

- [ ] Multi-replica deployments
- [ ] Health checks configured
- [ ] Resource limits set
- [ ] Persistent storage
- [ ] Backup strategy
- [ ] Monitoring and alerting

### Performance

- [ ] Resource sizing
- [ ] Horizontal autoscaling
- [ ] Load balancing
- [ ] CDN for static assets
- [ ] Database optimization
- [ ] Caching strategy

## Architecture Overview

```
Internet → Load Balancer → Ingress Controller → Frontend/Backend
                                              ↓
                                           Prometheus ← Apps
                                              ↓
                                            Grafana
```

## High Availability Setup

### Load Balancer Configuration

```yaml
# AWS ALB example
apiVersion: v1
kind: Service
metadata:
  name: site-availability-frontend
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-type: "alb"
    service.beta.kubernetes.io/aws-load-balancer-backend-protocol: "http"
    service.beta.kubernetes.io/aws-load-balancer-ssl-cert: "arn:aws:acm:..."
spec:
  type: LoadBalancer
  ports:
    - port: 443
      targetPort: 80
  selector:
    app: site-availability-frontend
```

### Multi-Zone Deployment

```yaml
backend:
  replicas: 3
  affinity:
    podAntiAffinity:
      preferredDuringSchedulingIgnoredDuringExecution:
        - weight: 100
          podAffinityTerm:
            labelSelector:
              matchExpressions:
                - key: app
                  operator: In
                  values:
                    - site-availability-backend
            topologyKey: kubernetes.io/hostname
```

## Monitoring and Alerting

### Prometheus Rules

```yaml
groups:
  - name: site-availability
    rules:
      - alert: SiteAvailabilityDown
        expr: up{job="site-availability-backend"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Site Availability service is down"

      - alert: HighMemoryUsage
        expr: container_memory_usage_bytes{pod=~"site-availability-.*"} / container_spec_memory_limit_bytes > 0.8
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage detected"
```

### Grafana Dashboard

Import the dashboard from `chart/grafana-dashboards/dashboard.json` for comprehensive monitoring.

## Security Configuration

### TLS/SSL Setup

```yaml
# cert-manager certificate
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: site-availability-tls
  namespace: monitoring
spec:
  secretName: site-availability-tls
  issuerRef:
    name: letsencrypt-prod
    kind: ClusterIssuer
  dnsNames:
    - monitoring.example.com
```

### RBAC Configuration

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: monitoring
  name: site-availability
rules:
  - apiGroups: [""]
    resources: ["configmaps", "secrets"]
    verbs: ["get", "list"]
```

### Network Policies

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: site-availability-netpol
  namespace: monitoring
spec:
  podSelector:
    matchLabels:
      app: site-availability-backend
  policyTypes:
    - Ingress
    - Egress
  ingress:
    - from:
        - podSelector:
            matchLabels:
              app: site-availability-frontend
      ports:
        - protocol: TCP
          port: 8080
```

## Backup and Recovery

### Configuration Backup

```bash
# Backup Kubernetes resources
kubectl get configmap site-availability-config -o yaml > config-backup.yaml
kubectl get secret site-availability-secrets -o yaml > secrets-backup.yaml

# Backup Prometheus data
kubectl exec prometheus-pod -- tar czf /tmp/prometheus-backup.tar.gz /prometheus
kubectl cp prometheus-pod:/tmp/prometheus-backup.tar.gz ./prometheus-backup.tar.gz
```

### Disaster Recovery Plan

1. **Configuration Recovery**: Restore from version control
2. **Data Recovery**: Restore Prometheus data from backups
3. **Service Recovery**: Redeploy using Helm charts
4. **Verification**: Run health checks and validate monitoring

## Performance Optimization

### Resource Sizing

```yaml
backend:
  resources:
    requests:
      cpu: 200m
      memory: 256Mi
    limits:
      cpu: 1000m
      memory: 1Gi

frontend:
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 256Mi
```

### Horizontal Pod Autoscaling

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: site-availability-backend-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: site-availability-backend
  minReplicas: 3
  maxReplicas: 10
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 80
```

## Operational Procedures

### Deployment Process

1. **Pre-deployment**:

   - Review changes in staging
   - Update documentation
   - Notify stakeholders

2. **Deployment**:

   ```bash
   helm upgrade site-availability chart/ -f values-prod.yaml
   ```

3. **Post-deployment**:
   - Verify health checks
   - Monitor metrics
   - Test functionality

### Maintenance Windows

Schedule regular maintenance for:

- Security updates
- Dependency upgrades
- Performance tuning
- Backup verification

### Incident Response

1. **Detection**: Automated alerts
2. **Assessment**: Check dashboards and logs
3. **Response**: Follow runbooks
4. **Resolution**: Apply fixes
5. **Post-mortem**: Document lessons learned

## Cost Optimization

### Resource Management

```yaml
# Use resource quotas
apiVersion: v1
kind: ResourceQuota
metadata:
  name: site-availability-quota
  namespace: monitoring
spec:
  hard:
    requests.cpu: "4"
    requests.memory: 8Gi
    limits.cpu: "8"
    limits.memory: 16Gi
```

### Auto-scaling Policies

- Scale down during off-hours
- Use spot instances for non-critical workloads
- Implement resource-based scaling
- Monitor and adjust regularly

## Compliance and Governance

### Audit Logging

```yaml
# Enable audit logging
apiVersion: v1
kind: ConfigMap
metadata:
  name: audit-policy
data:
  audit-policy.yaml: |
    apiVersion: audit.k8s.io/v1
    kind: Policy
    rules:
    - level: Metadata
      resources:
      - group: ""
        resources: ["configmaps", "secrets"]
```

### Documentation Requirements

- Architecture diagrams
- Runbooks and procedures
- Security policies
- Change management process
- Disaster recovery plans

## Migration Guide

### From Development to Production

1. **Configuration Changes**:

   - Update resource limits
   - Enable authentication
   - Configure persistent storage
   - Set up monitoring

2. **Data Migration**:

   - Export configuration
   - Migrate historical data
   - Validate data integrity

3. **Testing**:
   - Functional testing
   - Performance testing
   - Security scanning
   - Load testing

### Rollback Procedures

```bash
# Quick rollback
helm rollback site-availability 1

# Gradual rollback with traffic splitting
kubectl patch deployment site-availability-backend -p '{"spec":{"replicas":1}}'
# Monitor and scale back up gradually
```
