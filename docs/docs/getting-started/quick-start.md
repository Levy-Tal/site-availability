---
sidebar_position: 2
---

# Quick Start

Get Site Availability Monitoring up and running in under 5 minutes!

## Option 1: Docker Compose (Recommended)

The fastest way to get started is using Docker Compose, which includes the complete stack with Prometheus.

### 1. Clone and Start

```bash
git clone https://github.com/Levy-Tal/site-availability.git
cd site-availability
```

### 2. Choose Your Setup

#### Single Server Setup

```bash
cd helpers/docker-compose/single-server
docker-compose up -d
```

#### Multiple Servers Setup

```bash
cd helpers/docker-compose/multiple-servers
docker-compose up -d
```

### 3. Access the Application

- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3001 (admin/admin)

That's it! 🎉 Your Site Availability Monitoring is now running.

## Option 2: Local Development

For development or customization:

### 1. Prerequisites Check

```bash
node --version  # Should be v18.0+
go version      # Should be v1.19+
```

### 2. Start Backend

```bash
cd backend
go run main.go
```

### 3. Start Frontend (New Terminal)

```bash
cd frontend
npm install
npm start
```

## What's Next?

### 🎯 Configure Monitoring Targets

Edit the configuration to monitor your applications:

```yaml
# backend/config.yaml
apps:
  - name: my-app
    location: New York City
    metric: up{instance="my-app:8080", job="my-app"}
    prometheus: http://prometheus:9090/
```

### 📊 View Your Data

1. **World Map**: Check the interactive map at http://localhost:3000
2. **Metrics**: View raw metrics at http://localhost:8080/metrics
3. **Grafana**: Create custom dashboards at http://localhost:3001

### 🔧 Customize Configuration

- **Backend**: Edit `backend/config.yaml`
- **Frontend**: Modify `frontend/src/config.js`
- **Prometheus**: Update `helpers/prometheus/prometheus.yml`

## Example Configuration

Here's a minimal working configuration to get you started:

```yaml
# backend/config.yaml
scrape_interval: 30s

locations:
  - name: New York City
    latitude: 40.712776
    longitude: -74.005974
  - name: London
    latitude: 51.507351
    longitude: -0.127758

apps:
  - name: example-app
    location: New York City
    metric: up{instance="example-app:8080"}
    prometheus: http://prometheus:9090/
```

## Quick Verification

Test that everything is working:

```bash
# Check backend health
curl http://localhost:8080/health

# Check if apps are being scraped
curl http://localhost:8080/api/apps

# Verify Prometheus connectivity
curl http://localhost:8080/metrics
```

## Troubleshooting

### Services Not Starting?

```bash
# Check Docker containers
docker-compose ps

# View logs
docker-compose logs backend
docker-compose logs frontend
```

### Can't Access Frontend?

- Ensure port 3000 is not in use: `lsof -i :3000`
- Check firewall settings
- Try accessing via `127.0.0.1:3000` instead of `localhost:3000`

### Prometheus Data Not Showing?

1. Verify Prometheus is running: http://localhost:9090
2. Check targets are up: http://localhost:9090/targets
3. Ensure your applications expose metrics on `/metrics`

## Next Steps

- 📖 Learn about [Docker deployment](./docker) options
- ⚙️ Deep dive into [Configuration](../configuration/overview)
- 🚀 Explore [Production Deployment](../deployment/production)
- 🛠️ Set up [Development Environment](../development/setup)

---

**Need help?** Check our [Troubleshooting Guide](../troubleshooting) or [open an issue](https://github.com/Levy-Tal/site-availability/issues).
