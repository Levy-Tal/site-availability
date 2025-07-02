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
SA_LOG_LEVEL=debug ./site-availability
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

2. **Check Prometheus configuration:**

```bash
# Verify Prometheus targets
curl http://prometheus:9090/api/v1/targets

# Check Prometheus logs
docker-compose logs prometheus
```

3. **Authentication issues:**

```bash
# Check HMAC secret
echo $SA_AUTHENTICATION_HMAC_SECRET

# Verify authentication headers
curl -H "Authorization: HMAC-SHA256 <signature>" http://localhost:8080/api/apps
```

### Frontend Issues

#### Frontend Won't Load

**Symptoms:**

- Blank page or error messages
- Build failures
- API connection errors

**Solutions:**

1. **Check Node.js setup:**

```bash
# Verify Node.js version
node --version  # Should be 18.0+

# Clear npm cache
npm cache clean --force

# Reinstall dependencies
rm -rf node_modules package-lock.json
npm install
```

2. **Check API connectivity:**

```bash
# Test backend API
curl http://localhost:8080/health

# Check CORS settings
curl -H "Origin: http://localhost:3000" http://localhost:8080/api/apps
```

3. **Check frontend configuration:**

```javascript
// src/config.js
const config = {
  apiUrl: "http://localhost:8080", // Verify this matches your backend
  updateInterval: 30000,
};
```

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

#### Container Build Failures

**Symptoms:**

- Docker build errors
- Missing dependencies
- Permission denied errors

**Solutions:**

1. **Check Dockerfile:**

```bash
# Build with verbose output
docker build --no-cache -t site-availability-backend backend/

# Check base image
docker pull golang:1.21-alpine

# Verify file permissions
ls -la backend/
```

2. **Disk space issues:**

```bash
# Check available space
df -h

# Clean up Docker
docker system prune -a
docker volume prune
```

#### Container Runtime Issues

**Symptoms:**

- Containers keep restarting
- Out of memory errors
- Network connectivity issues

**Solutions:**

1. **Check container logs:**

```bash
# View logs
docker-compose logs backend
docker-compose logs frontend

# Follow logs in real-time
docker-compose logs -f backend
```

2. **Resource limits:**

```bash
# Check resource usage
docker stats

# Monitor memory usage
docker exec -it <container> free -h
```

3. **Network issues:**

```bash
# Check network configuration
docker network ls
docker network inspect site-availability_default

# Test inter-container connectivity
docker-compose exec frontend ping backend
```

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

2. **Check data mapping:**

```bash
# Verify app-to-location mapping
curl http://localhost:8080/api/apps | jq '.[] | {name, location}'
```

## Debugging Tools

### Log Analysis

```bash
# Backend logs with debug level
SA_LOG_LEVEL=debug ./site-availability

# Follow logs in real-time
tail -f /var/log/site-availability.log

# Search for specific errors
grep -i "error" /var/log/site-availability.log
```

### API Testing

```bash
# Health check
curl http://localhost:8080/health

# Get all applications
curl http://localhost:8080/api/apps | jq

# Get locations
curl http://localhost:8080/api/locations | jq

# Test metrics endpoint
curl http://localhost:8080/metrics
```

### Network Debugging

```bash
# Test connectivity
telnet prometheus 9090

# Check DNS resolution
nslookup prometheus

# Trace network path
traceroute prometheus

# Check firewall rules
iptables -L
```

## Performance Issues

### High Memory Usage

**Solutions:**

1. Increase scrape intervals
2. Reduce number of monitored metrics
3. Implement metric filtering
4. Use Prometheus recording rules

### Slow Response Times

**Solutions:**

1. Optimize Prometheus queries
2. Add caching layers
3. Use Prometheus federation
4. Scale horizontally

### High CPU Usage

**Solutions:**

1. Profile the application
2. Optimize metric processing
3. Reduce scraping frequency
4. Use more efficient queries

## Getting Help

### Before Asking for Help

1. **Check logs** for error messages
2. **Search existing issues** on GitHub
3. **Try minimal configuration** to isolate the problem
4. **Document reproduction steps** clearly

### When Reporting Issues

Include the following information:

```bash
# System information
uname -a
docker --version
go version
node --version

# Application logs
SA_LOG_LEVEL=debug ./site-availability 2>&1

# Configuration (remove sensitive data)
cat config.yaml

# Container status (if using Docker)
docker-compose ps
docker-compose logs
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
