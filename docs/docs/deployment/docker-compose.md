---
sidebar_position: 1
---

# Docker Compose Deployment

Deploy Site Availability Monitoring using Docker Compose for development and small-scale production environments.

## Quick Start

```bash
# Clone repository
git clone https://github.com/Levy-Tal/site-availability.git
cd site-availability

# Start with single server setup
cd helpers/docker-compose/single-server
docker-compose up -d
```

## Available Setups

### Single Server

Best for development and testing:

```bash
cd helpers/docker-compose/single-server
docker-compose up -d
```

### Multiple Servers

Production-like environment:

```bash
cd helpers/docker-compose/multiple-servers
docker-compose up -d
```

### With Nginx

Includes reverse proxy:

```bash
cd helpers/docker-compose/single-server-nginx
docker-compose up -d
```

## Custom Deployment

Create your own `docker-compose.yml`:

```yaml
version: "3.8"

services:
  backend:
    image: site-availability/backend:latest
    ports:
      - "8080:8080"
    environment:
      - CONFIG_FILE=/app/config.yaml
    volumes:
      - ./config.yaml:/app/config.yaml
      - ./certs:/app/certs
    depends_on:
      - prometheus

  frontend:
    image: site-availability/frontend:latest
    ports:
      - "3000:80"
    environment:
      - REACT_APP_API_URL=http://localhost:8080

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
```

## Production Considerations

- Use external volumes for data persistence
- Configure resource limits
- Set up monitoring and logging
- Use secrets management
- Enable TLS/SSL
