---
sidebar_position: 99
---

# Troubleshooting

This guide covers common issues and their solutions when working with Site Availability Monitoring.

## Common Issues

### Backend Issues

#### Backend Won't Start

**Symptoms:**

- Application exits immediately
- Port binding errors
- Configuration errors

**Solutions:**

1. **Check port availability:**

```bash
# Check if port 8080 is in use
lsof -i :8080
netstat -tulpn | grep 8080

# Kill process using the port
kill -9 <PID>
```

2. **Validate configuration:**

```bash
# Check YAML syntax
yamllint config.yaml

# Run with debug logging
LOG_LEVEL=debug ./site-availability
```

3. **Check permissions:**

```bash
# Ensure config file is readable
chmod 644 config.yaml

# Check directory permissions
ls -la config.yaml
```

#### Prometheus Connection Issues

**Symptoms:**

- No data showing in frontend
- Prometheus timeout errors
- Connection refused errors

**Solutions:**

1. **Verify Prometheus connectivity:**

```bash
# Test Prometheus URL
curl http://prometheus:9090/api/v1/query?query=up

# Check network connectivity
ping prometheus

# Test from inside container
docker-compose exec backend wget -qO- http://prometheus:9090/metrics
```

### Frontend Issues

#### Frontend Won't Load

**Symptoms:**

- Blank page or error messages
- Build failures
- API connection errors

**Solutions:**

1. **Check API connectivity:**

```bash
# Test backend API
curl http://localhost:8080/api/apps
```

````

3. **Check browser console:**

```javascript
// Open browser developer tools (F12)
// Look for JavaScript errors in Console tab
// Check Network tab for failed API requests
````

#### Map Not Displaying

**Symptoms:**

- Empty map area
- JavaScript errors in console
- Data not loading

**Solutions:**

1. **Check browser console:**

```javascript
// Open browser developer tools (F12)
// Look for JavaScript errors in Console tab
// Check Network tab for failed API requests
```

2. **Verify data format:**

```bash
# Check API response format
curl http://localhost:8080/api/apps | jq

# Verify locations data
curl http://localhost:8080/api/locations | jq
```

3. **Browser compatibility:**

```bash
# Test in different browsers
# Clear browser cache and cookies
# Disable browser extensions temporarily
```

### Docker Issues

#### Container Runtime Issues

**Symptoms:**

- Containers keep restarting
- Out of memory errors
- Network connectivity issues

**Solutions:**

1. **Check container logs:**

```bash
# View logs
docker ps
docker logs -f <container-name>
```

````

3. **Restart containers:**

```bash

# Restart all services
docker compose down && docker-compose up -d
````

### Kubernetes/Helm Issues

#### Deployment Failures

**Symptoms:**

- Pods not starting
- Image pull errors
- Configuration errors

**Solutions:**

1. **Check pod status:**

```bash
# View pods
kubectl get pods -n site-availability

# Describe failing pod
kubectl describe pod <pod-name> -n site-availability

# Check logs
kubectl logs <pod-name> -n site-availability
```

2. **Configuration issues:**

```bash
# Validate Helm chart
helm lint chart/

# Check values
helm template site-availability chart/ --values chart/values.yaml

# Debug installation
helm install site-availability chart/ --dry-run --debug
```

3. **Resource constraints:**

```bash
# Check node resources
kubectl top nodes

# Check resource quotas
kubectl describe quota -n site-availability
```

### Data Issues

#### No Metrics Data

**Symptoms:**

- Empty dashboards
- Zero values everywhere
- Missing applications

**Solutions:**

1. **Verify metrics endpoints:**

```bash
# Check if applications expose metrics
curl http://your-app:8080/metrics

# Verify Prometheus scraping
curl http://prometheus:9090/api/v1/query?query=up
```

2. **Check scraping configuration:**

```yaml
# Prometheus configuration
scrape_configs:
  - job_name: "your-app"
    static_configs:
      - targets: ["your-app:8080"]
```

3. **Validate metric queries:**

```bash
# Test metric queries directly in Prometheus
# Go to http://prometheus:9090
# Run query: up{instance="your-app:8080"}
```

#### Incorrect Location Display

**Symptoms:**

- Applications in wrong locations
- Missing location markers
- Incorrect coordinates

**Solutions:**

1. **Verify location configuration:**

```yaml
locations:
  - name: "New York"
    latitude: 40.712776 # Check these coordinates
    longitude: -74.005974
```

2. **Check coordinate format:**

```yaml
# Ensure coordinates are decimal degrees
# Latitude: -90 to 90
# Longitude: -180 to 180
```

## Debugging Tools

### Log Analysis

```bash
# Backend logs with debug level
LOG_LEVEL=debug

```

### API Testing

```bash
# Health check
curl http://localhost:8080/healthz

# Get all applications
curl http://localhost:8080/api/apps | jq

# Get locations
curl http://localhost:8080/api/locations | jq

# Test metrics endpoint
curl http://localhost:8080/metrics
```

## Getting Help

### Before Asking for Help

1. **Check logs** for error messages
2. **Search existing issues** on GitHub
3. **Try minimal configuration** to isolate the problem
4. **Document reproduction steps** clearly

### When Reporting Issues

Include the following information:

```bash

# Application logs in debug mode
LOG_LEVEL=debug

# Configuration (remove sensitive data)
cat config.yaml

```

### Community Resources

- 🐛 **GitHub Issues**: [Report bugs](https://github.com/Levy-Tal/site-availability/issues)
- 💬 **Discussions**: [Ask questions](https://github.com/Levy-Tal/site-availability/discussions)
- 📖 **Documentation**: You're reading it!

## Prevention

### Monitoring Your Monitoring

Set up monitoring for the Site Availability Monitoring system itself:

1. **Health checks** for backend services
2. **Alerting** on service failures
3. **Log monitoring** for errors
4. **Resource monitoring** for containers

### Best Practices

1. **Use configuration validation** in CI/CD
2. **Test in staging** environments first
3. **Monitor resource usage** regularly
4. **Keep documentation** up to date
5. **Regular backups** of configuration and data

---

Still having issues? [Open an issue](https://github.com/Levy-Tal/site-availability/issues) on GitHub with detailed information about your problem.
