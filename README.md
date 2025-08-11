<div align="center">

<img src="docs/static/img/logo-full.png" alt="Site Availability Monitor Logo" width="400">

[![CI](https://github.com/Levy-Tal/site-availability/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/Levy-Tal/site-availability/actions/workflows/ci.yaml)
[![codecov](https://codecov.io/gh/Levy-Tal/site-availability/branch/main/graph/badge.svg?token=K3PLCUDMX9)](https://codecov.io/gh/Levy-Tal/site-availability)

</div>

## Site Availability Monitoring

Site Availability is an open-source platform that helps SRE teams organize the availability of company resources to ensure they meet the company SLOs. In this app, you define the rules and queries that determine whether your services are up or down. It generates consistent availability metrics, shows a live status of all your services around the world in the UI for fast response, ships an out-of-the-box Grafana dashboard to track availability and SLOs, and gives your team a clear timeline of what changed and when.

This project is designed for SRE and platform teams who want to stop building and maintaining complex, bespoke dashboards for every service and environment. Configure once, standardize everywhere.

### Always open-source, always free

- Licensed under Apache 2.0
- No paid editions or feature gating
- Community-driven roadmap

## The problem it solves

Modern SRE teams juggle many tools and inconsistent definitions of “availability” across hundreds of services:

- Multiple monitoring backends, each with different metrics and labels
- Inconsistent SLI definitions per team/service
- Dashboards that are hard to keep consistent across environments

Site Availability solves this by:

- Unifying data from Prometheus, HTTP checks, and other Site Availability instances
- Normalizing everything into a single, binary availability metric with rich labels
- Providing a ready-to-use Grafana dashboard and time-series views of status changes

## What you get

- Single source of truth for availability across teams, environments, and regions
- Simple up/down signal with consistent labels you control (env, team, app, region, etc.)
- Out-of-the-box Grafana dashboards (see `chart/grafana-dashboards/`)
- Historical timelines of incidents and recoveries via Prometheus time series
- Lightweight, stateless backend and a modern web UI with a world-map visualization
- Production-ready Helm chart and Docker Compose examples
- Optional authentication (local admin or OIDC) and metrics protection

## How it works

- Backend (Go): Scrapes configured sources on an interval, merges and normalizes results, and exposes REST + Prometheus metrics.
- Frontend (React): Displays current status on a world map with filters and details.
- Metrics: Exposes a normalized metric you can alert on and explore over time.

Key metric (normalized availability signal):

```prometheus
# 1 = up, 0 = down; dynamic labels include name, location, source, origin_url, and your custom labels
site_availability_status{name="backend-app",location="eu-west-1",source="prometheus-main",env="prod",team="payments"} 1
```

Additional metrics include scrape duration and request counters. See the documentation for details.

## Quickstart

Pick one of the following deployment options.

### Option A: Docker Compose (local demo)

1. Create `config.yaml` and `prometheus.yml` using the examples in the documentation.

2. Start the stack:

```bash
docker compose up -d
```

3. Open the UI at http://localhost:8080 and Prometheus at http://localhost:9090.

Full guide: Documentation › Usage › Quickstart

### Option B: Helm (Kubernetes)

```bash
helm repo add site-availability https://levytal.github.io/site-availability/
helm install site-availability site-availability/site-availability \
  --set replicaCount=3
```

Full guide: Documentation › Usage › Installation › Helm Chart

## Configuration overview

Configuration is YAML-based and designed to be readable and to enforce consistent labels. At a minimum, set `server_settings.host_url`, `locations`, and one or more `sources`.

Example (Prometheus + HTTP sources):

```yaml
server_settings:
  port: 8080
  host_url: "http://localhost:8080"
  labels:
    env: "production"
    team: "backend"

locations:
  - name: "New York"
    latitude: 40.712776
    longitude: -74.005974
  - name: "London"
    latitude: 51.507351
    longitude: -0.127758

sources:
  - name: prometheus-main
    type: prometheus
    config:
      url: http://prometheus:9090
      apps:
        - name: users-api
          location: London
          metric: 'up{job="users-api"}'
          labels:
            app: users-api
            tier: backend

  - name: basic-http
    type: http
    config:
      apps:
        - name: website
          location: New York
          url: https://example.com
```

Supported sources today: `prometheus`, `http`, and `site` (scrape another Site Availability instance via `/sync` with HMAC).

For full configuration, see Documentation › Usage › Configuration › Server and Sources.

## Dashboards and timelines

- Grafana dashboards are included in the Helm chart (`chart/grafana-dashboards/`).
- Because availability is exposed as time series, you immediately get a clear timeline of what went down, where, and when.
- Recording/alerting rule examples are provided in the documentation.

## Security and access

- Local admin or OIDC authentication for the UI and APIs
- Role-based access using labels to scope what users can see
- Optional authentication for `/metrics` (basic or bearer)
- HMAC-protected `/sync` endpoint for cross-site aggregation

## Documentation

Full documentation is available at the project website: `https://levy-tal.github.io/site-availability/`.

Highlights:

- Why Site Availability (problem and solution)
- Quickstart and production installation (Docker Compose, Helm)
- Configuration reference (server, sources, credentials)
- Metrics and Grafana setup
- Authentication and RBAC

## License and community

- License: Apache 2.0 (see `LICENSE`)
- Issues and discussions: open on the GitHub repository
- Contributions are welcome via pull requests and issues
