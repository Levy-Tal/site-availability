---
sidebar_position: 2
---

# Architecture Overview

This page summarizes the core architecture of Site Availability Monitoring.

## Backend Structure

- **main.go**: Entry point
- **server/**: HTTP server and routing
- **handlers/**: API endpoints and request handling
- **scraping/**: Source scrapers (prometheus, http, site)
- **config/**: Configuration loading and validation
- **logging/**: Structured logging
- **metrics/**: Prometheus metrics
- **labels/**: Label management
- **authentication/**: HMAC authentication

## Data Flow

1. **Configuration**: Load and validate config.yaml and credentials.yaml
2. **Scraping**: Periodically scrape sources (Prometheus, HTTP, Site)
3. **Cache Update**: Store latest app statuses and locations in memory
4. **API**: Serve data via REST endpoints

## API Endpoints

- `GET  /health` — Health check
- `GET  /metrics` — Prometheus metrics
- `GET  /api/apps` — Application statuses
- `GET  /api/locations` — Locations
- `GET  /api/config` — Configuration info

## Security

- **HMAC authentication** for protected endpoints
- **Token-based**: Use tokens in config/credentials for secure sync between instances

## Frontend Structure

- **src/App.js**: Main app component
- **src/index.js**: Entry point
- **src/config.js**: Frontend config
- **src/map.js**: Map rendering
- **src/Sidebar.js**: Sidebar and app list
- **src/api/**: API helpers
- **src/utils/**: Utilities
- **src/styles/**: CSS

## Key Principles

- **Stateless backend**: No persistent database, all state in memory
- **Config-driven**: All behavior controlled by YAML config and environment variables
- **Extensible sources**: Add new source types easily
- **Simple REST API**: For integration and UI

---

For more details, see the codebase and config examples.
