---
sidebar_position: 1
---

# Site Availability

Welcome to **Site Availability Exporter** - an open-source exporter designed to monitor the availability of applications and services across multiple locations with real-time visualization and comprehensive metrics collection.

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
- 🔐 **HMAC authentication** for secure API access
- 🏷️ **Label-based filtering** for organized monitoring

## Architecture

The application consists of three main components:

### Frontend

- **React-based web interface** that displays application statuses on an interactive world map
- **Real-time updates** showing current status of monitored applications
- **Historical data visualization** for trend analysis
- **Responsive design** that works on desktop and mobile devices
- **Status filtering** by application status and labels
- **Collapsible sidebar** for configuration and navigation

### Backend

- **Go-based server** that fetches application statuses from multiple sources
- **REST APIs** for serving status data and configuration
- **Efficient scraping** of Prometheus metrics and HTTP endpoints
- **HMAC authentication** for secure API access
- **Comprehensive logging** for monitoring and debugging
- **Label management** for organizing applications
- **Custom CA certificate support** for secure connections

### Monitoring Stack

- **Prometheus integration** for metrics collection
- **Grafana dashboards** for advanced visualization
- **Alert management** for incident response
- **Custom metrics** for application-specific monitoring
- **Multiple source support** (Prometheus, HTTP, Site monitoring)

## System Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│     Frontend    │    │     Backend     │    │   Data Sources  │
│   (React App)   │───▶│   (Go Server)   │───▶│   (Prometheus)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│      Browser    │    │      API        │    │   Target Apps   │
│   (World Map)   │    │   (REST/JSON)   │    │  (Monitored)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Core Components

### Backend Structure

```
backend/
├── main.go              # Application entry point
├── server/              # HTTP server and routing
├── handlers/            # HTTP request handlers
├── scraping/            # Data collection from sources
│   ├── prometheus/      # Prometheus client
│   ├── http/           # HTTP source monitoring
│   └── site/           # Site monitoring logic
├── config/             # Configuration management
├── logging/            # Structured logging
├── metrics/            # Application metrics
├── labels/             # Label management
└── authentication/     # HMAC authentication
```

### Frontend Structure

```
frontend/src/
├── App.js              # Main application component
├── map.js              # World map visualization
├── Sidebar.js          # Configuration sidebar
├── AppStatusPanel.js   # Status details panel
├── api/                # API client functions
├── utils/              # Utility functions
└── styles/             # CSS styling
```

## Getting Started

Ready to start monitoring your applications? Choose your preferred setup method:

1. **[Quick Start](./usage/quickstart)** - Get up and running in minutes
2. **[Docker Setup](./usage/installation/docker-compose)** - Containerized deployment
3. **[Helm Chart](./usage/installation/helm-chart)** - Kubernetes deployment

## What You'll Need

- **Node.js** (v18.0 or above) for frontend development
- **Go** (v1.19 or above) for backend development
- **Docker** for containerization
- **Helm** for Kubernetes deployment
- **Prometheus** for metrics collection (optional)

## Next Steps

- 📖 Follow the [Quick Start](./usage/quickstart) to get up and running
- 📚 Learn about [Terminology](./usage/terminology) and concepts
- ⚙️ Configure your [Server Settings](./usage/configuration/server)
- 🔌 Set up [Data Sources](./usage/configuration/sources/prometheus)
- 🚀 Explore [Deployment](./usage/installation/docker-compose) strategies
- 🛠️ Check out the [Development Guide](./development/setup) if you want to contribute

## Community & Support

- 🐛 **Report issues**: [GitHub Issues](https://github.com/Levy-Tal/site-availability/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/Levy-Tal/site-availability/discussions)
- 📄 **License**: Available under the Apache 2.0 License

---

_Ready to monitor your applications like never before? Let's get started!_
