---
sidebar_position: 2
---

# Backend Configuration

Configure the Go backend server settings and behavior.

## Server Settings

```yaml
# Server configuration
port: 8080
host: 0.0.0.0
log_level: info
read_timeout: 30s
write_timeout: 30s
```

## Scraping Configuration

```yaml
scrape_interval: 30s
timeout: 10s
max_retries: 3
```

## Prometheus Settings

```yaml
prometheus:
  default_url: http://prometheus:9090
  query_timeout: 10s
  max_concurrent_queries: 10
```

## Authentication

```yaml
authentication:
  hmac:
    enabled: true
    secret: ${SA_HMAC_SECRET}
    algorithm: sha256
```

## Logging

Configure logging levels and output:

```yaml
logging:
  level: info
  format: json
  output: stdout
  file: /var/log/site-availability.log
```

Available levels: `debug`, `info`, `warn`, `error`

## TLS/SSL Configuration

```yaml
tls:
  enabled: false
  cert_file: /path/to/cert.pem
  key_file: /path/to/key.pem
  ca_file: /path/to/ca.pem
```

## Environment Variables

Override any configuration with environment variables:

```bash
SA_PORT=8080
SA_LOG_LEVEL=debug
SA_SCRAPE_INTERVAL=60s
SA_AUTHENTICATION_HMAC_SECRET=your-secret
```
