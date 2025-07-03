---
sidebar_position: 1
---

# Site Availability

Welcome to **Site Availability** - an open-source application designed to monitor the availability of applications and services across multiple locations with real-time visualization and comprehensive metrics collection.

## Overview

Site Availability Monitoring provides a complete solution for tracking application uptime and performance across geographically distributed locations. The system combines real-time monitoring with beautiful visualizations and robust alerting capabilities.

### Key Features

- 🌍 **Real-time site status monitoring** across multiple geographic locations
- 🗺️ **Interactive world map visualization** showing application status at a glance
- 📊 **Historical data tracking** with comprehensive metrics collection
- 🔔 **Alert notifications** for immediate incident response
- 📈 **Prometheus integration** for advanced metrics and monitoring
- 🎨 **Modern React frontend** with responsive design
- 🚀 **Go-based backend** for high performance and reliability
- ☸️ **Kubernetes-ready** with Helm chart deployment
- 🐳 **Docker support** for easy containerization
- 📊 **Grafana dashboards** for advanced analytics

## Architecture

The application consists of three main components:

### Frontend

- **React-based web interface** that displays application statuses on an interactive world map
- **Real-time updates** showing current status of monitored applications
- **Historical data visualization** for trend analysis
- **Responsive design** that works on desktop and mobile devices

### Backend

- **Go-based server** that fetches application statuses from Prometheus
- **REST APIs** for serving status data and configuration
- **Efficient scraping** of Prometheus metrics
- **HMAC authentication** for secure API access
- **Comprehensive logging** for monitoring and debugging

### Monitoring Stack

- **Prometheus integration** for metrics collection
- **Grafana dashboards** for advanced visualization
- **Alert management** for incident response
- **Custom metrics** for application-specific monitoring

## Getting Started

Ready to start monitoring your applications? Choose your preferred setup method:

1. **[Quick Start](./getting-started/quick-start)** - Get up and running in minutes
2. **[Docker Setup](./getting-started/docker)** - Containerized deployment
3. **[Installation Guide](./getting-started/installation)** - Detailed setup instructions

## What You'll Need

- **Node.js** (v18.0 or above) for frontend development
- **Go** (v1.19 or above) for backend development
- **Docker** for containerization
- **Helm** for Kubernetes deployment
- **Prometheus** for metrics collection

## Next Steps

- 📖 Follow the [Installation Guide](./getting-started/installation) to set up your environment
- ⚙️ Learn about [Configuration](./configuration/overview) options
- 🚀 Explore [Deployment](./deployment/docker-compose) strategies
- 🛠️ Check out the [Development Guide](./development/setup) if you want to contribute

## Community & Support

- 🐛 **Report issues**: [GitHub Issues](https://github.com/Levy-Tal/site-availability/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/Levy-Tal/site-availability/discussions)
- 📄 **License**: Available under the Apache 2.0 License

---

_Ready to monitor your applications like never before? Let's get started!_
