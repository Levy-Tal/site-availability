---
sidebar_position: 3
---

# Docker Setup

Learn how to run Site Availability Monitoring using Docker and Docker Compose for easy deployment and development.

## Docker Compose Setups

The project includes several pre-configured Docker Compose setups for different scenarios.

### Single Server Setup

Perfect for testing and small deployments:

```bash
cd helpers/docker-compose/single-server
docker-compose up -d
```

This setup includes:

- Site Availability Monitoring backend
- React frontend
- Prometheus for metrics
- Sample application to monitor

**Services:**

- Frontend: http://localhost:3000
- Backend: http://localhost:8080
- Prometheus: http://localhost:9090

### Multiple Servers Setup

For production-like environments with multiple applications:

```bash
cd helpers/docker-compose/multiple-servers
docker-compose up -d
```

This setup includes:

- Multiple backend instances
- Load balancer
- Prometheus federation
- Multiple sample applications

**Services:**

- Frontend: http://localhost:3000
- Backend Cluster: http://localhost:8080
- Prometheus: http://localhost:9090

## Environment Variables

Configure the application using environment variables:

### Backend Environment Variables

```bash
# Configuration
CONFIG_FILE=/app/config.yaml
LOG_LEVEL=info
PORT=8080

# Custom CA certificates
CUSTOM_CA_PATH=/app/certs

# Prometheus settings
PROMETHEUS_URL=http://prometheus:9090
SCRAPE_INTERVAL=30s
```

## Next Steps

- 🚀 Deploy to production with [Kubernetes](../deployment/kubernetes)
- ⚙️ Configure monitoring with [Prometheus](../configuration/prometheus)
- 🛡️ Set up [Authentication](../api/authentication)

## Troubleshooting

For troubleshooting help, see our [Troubleshooting Guide](../troubleshooting).
