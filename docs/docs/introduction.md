---
sidebar_position: 1
title: "Welcome to Site Availability"
description: "Monitor the availability of applications and services across multiple locations with real-time visualization and comprehensive metrics collection."
keywords: [monitoring, availability, prometheus, grafana, health checks]
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

## Why Site Availability Matters for SRE Teams

### The SRE Challenge: Measuring What Matters

One of the most fundamental responsibilities of Site Reliability Engineering is **measuring service availability** and ensuring services meet their Service Level Objectives (SLOs). However, tracking availability across a large company's infrastructure presents significant challenges that can overwhelm even experienced SRE teams.

### The Complexity Problem

#### Scale and Fragmentation

Modern enterprises operate **hundreds or thousands of microservices**, each with different:

- **Technology stacks** (Java, Python, Go, Node.js)
- **Deployment patterns** (containers, VMs, serverless)
- **Monitoring approaches** (logs, metrics, traces)
- **Team ownership** and access requirements

#### Inconsistent Availability Definition

**"Available" isn't binary.** A service might return HTTP 200 but deliver:

- Wrong data
- Unacceptable latency
- Degraded functionality

Different components require different Service Level Indicators (SLIs):

- **Web APIs**: Success rate + response time
- **Databases**: Query error rate + connection health
- **Message queues**: Processing rate + backlog size
- **External dependencies**: Third-party SLA compliance

#### Organizational Silos

- **Different teams** own different services
- **Varying permissions** and access patterns
- **Inconsistent labeling** and metrics standards
- **Fragmented dashboards** across multiple tools

### How Site Availability Solves These Challenges

#### 🎯 **Unified Service Discovery**

- **Multi-source aggregation**: Pull availability data from Prometheus, HTTP endpoints, and external APIs
- **Automatic labeling**: Consistent metadata across all services regardless of source
- **Team-based filtering**: Show only the services your team owns or cares about

#### 📊 **Consistent Availability Metrics**

- **Standardized SLIs**: HTTP success rates, response times, and custom business metrics
- **Flexible definitions**: Configure what "available" means for each service type
- **Historical tracking**: Trend analysis and SLO compliance reporting

#### 🔍 **Intelligent Aggregation**

- **Geographic grouping**: View availability by region, datacenter, or environment
- **Service hierarchies**: Roll up component availability to business-critical services
- **Smart filtering**: Focus on production services, critical dependencies, or failing components

#### 🚨 **Actionable Alerting**

- **SLO-based alerts**: Get notified when error budgets are at risk
- **Noise reduction**: Alert on patterns, not individual blips
- **Context-rich notifications**: Know which team to contact and what might be affected

#### 🏢 **Enterprise-Ready**

- **RBAC integration**: Respect existing access controls and permissions
- **Multi-tenant**: Support multiple teams and environments in one deployment
- **Audit trail**: Track who accessed what and when for compliance

### The Bottom Line

**Site Availability transforms availability monitoring from a reactive burden into a proactive advantage.**

Instead of spending hours manually correlating metrics across disparate tools, SRE teams get:

- ✅ **Single source of truth** for service availability
- ✅ **Consistent SLO tracking** across all services
- ✅ **Faster incident response** with centralized visibility
- ✅ **Data-driven capacity planning** with historical trends
- ✅ **Improved stakeholder communication** with clear availability reports

## Community & Support

- 🐛 **Report issues**: [GitHub Issues](https://github.com/Levy-Tal/site-availability/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/Levy-Tal/site-availability/discussions)
- 📄 **License**: Available under the Apache 2.0 License

---

_Ready to monitor your applications like never before? Let's get started!_
