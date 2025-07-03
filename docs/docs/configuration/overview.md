---
sidebar_position: 1
---

# Configuration Overview

Site Availability is highly configurable to suit different environments and monitoring needs. This section covers all configuration options and best practices.

## Configuration Architecture

The application uses a hierarchical configuration system:

1. **YAML Configuration Files** - Primary configuration method
2. **Environment Variables** - Override specific settings
3. **Command Line Arguments** - Runtime overrides
4. **Default Values** - Fallback configuration

## Configuration File Structure

The main configuration file follows this structure:

```yaml
# Global settings
scrape_interval: 30s
log_level: info
port: 8080

# Geographic locations for monitoring
locations:
  - name: "Location Name"
    latitude: 40.712776
    longitude: -74.005974

# Applications to monitor
apps:
  - name: "app-name"
    location: "Location Name"
    metric: 'up{instance="app:8080", job="app"}'
    prometheus: "http://prometheus:9090/"

# Authentication settings
authentication:
  hmac:
    enabled: true
    secret: "your-secret-key"

# Custom CA certificates
certificates:
  custom_ca_path: "/path/to/ca-certificates"
```

## Configuration Locations

Configuration files are loaded in the following order (later files override earlier ones):

1. `/etc/site-availability/config.yaml` (system-wide)
2. `~/.site-availability/config.yaml` (user-specific)
3. `./config.yaml` (current directory)
4. File specified by `CONFIG_FILE` environment variable
5. File specified by `--config` command line argument

## Environment Variables

All configuration options can be overridden using environment variables with the `SA_` prefix:

```bash
# Basic settings
SA_PORT=8080
SA_LOG_LEVEL=debug
SA_SCRAPE_INTERVAL=60s

# Authentication
SA_AUTHENTICATION_HMAC_ENABLED=true
SA_AUTHENTICATION_HMAC_SECRET=your-secret

# Custom CA certificates
SA_CERTIFICATES_CUSTOM_CA_PATH=/app/certs
```

## Configuration Validation

The application validates configuration on startup and will fail to start with invalid configuration. Common validation rules:

- **Locations**: Must have valid latitude/longitude coordinates
- **Apps**: Must reference existing locations
- **Prometheus URLs**: Must be valid HTTP/HTTPS URLs
- **Scrape Interval**: Must be a valid duration (e.g., "30s", "1m", "5m")

## Configuration Examples

### Minimal Configuration

```yaml
scrape_interval: 30s

locations:
  - name: Primary DC
    latitude: 40.712776
    longitude: -74.005974

apps:
  - name: web-server
    location: Primary DC
    metric: up{instance="web:8080"}
    prometheus: http://prometheus:9090/
```

### Multi-Location Configuration

```yaml
scrape_interval: 15s
log_level: info

locations:
  - name: New York
    latitude: 40.712776
    longitude: -74.005974
  - name: London
    latitude: 51.507351
    longitude: -0.127758
  - name: Tokyo
    latitude: 35.676676
    longitude: 139.650027

apps:
  - name: api-server-ny
    location: New York
    metric: up{instance="api-ny:8080", job="api"}
    prometheus: http://prometheus-ny:9090/

  - name: api-server-london
    location: London
    metric: up{instance="api-london:8080", job="api"}
    prometheus: http://prometheus-london:9090/

  - name: api-server-tokyo
    location: Tokyo
    metric: up{instance="api-tokyo:8080", job="api"}
    prometheus: http://prometheus-tokyo:9090/
```

### Production Configuration with Authentication

```yaml
scrape_interval: 30s
log_level: warn
port: 8080

# Security settings
authentication:
  hmac:
    enabled: true
    secret: ${SA_HMAC_SECRET}

# Custom certificates
certificates:
  custom_ca_path: /app/certs

locations:
  - name: Production East
    latitude: 39.0458
    longitude: -76.6413
  - name: Production West
    latitude: 37.7749
    longitude: -122.4194

apps:
  - name: frontend
    location: Production East
    metric: up{instance="frontend:443", job="frontend"}
    prometheus: https://prometheus-east.internal:9090/

  - name: backend-api
    location: Production East
    metric: up{instance="backend:8080", job="backend"}
    prometheus: https://prometheus-east.internal:9090/

  - name: database
    location: Production West
    metric: up{instance="postgres:5432", job="postgres"}
    prometheus: https://prometheus-west.internal:9090/
```

## Best Practices

### Security

- Use environment variables for secrets
- Enable HMAC authentication in production
- Use HTTPS for Prometheus connections
- Regularly rotate authentication secrets

### Performance

- Set appropriate scrape intervals (15-60 seconds)
- Monitor only essential applications
- Use Prometheus federation for large deployments
- Configure appropriate timeouts

### Monitoring

- Start with a few critical applications
- Add geographic diversity gradually
- Use descriptive names for locations and apps
- Document your monitoring strategy

### Maintenance

- Version control your configuration files
- Use configuration templates for environments
- Validate configuration in CI/CD pipelines
- Monitor configuration changes

## Configuration Sections

Learn more about specific configuration areas:

- **[Backend Configuration](./backend)** - Server settings and behavior
- **[Frontend Configuration](./frontend)** - UI customization and features
- **[Prometheus Configuration](./prometheus)** - Metrics and scraping setup

## Troubleshooting Configuration

### Configuration Not Loading

```bash
# Check file permissions
ls -la config.yaml

# Validate YAML syntax
yaml-lint config.yaml

# Check environment variables
env | grep SA_
```

### Invalid Configuration Errors

```bash
# Run with debug logging
SA_LOG_LEVEL=debug ./site-availability

# Validate configuration manually
./site-availability --validate-config
```

For more troubleshooting help, see our [Troubleshooting Guide](../troubleshooting).

## Next Steps

- Configure the [Backend](./backend) settings
- Set up [Frontend](./frontend) customization
- Configure [Prometheus](./prometheus) integration
