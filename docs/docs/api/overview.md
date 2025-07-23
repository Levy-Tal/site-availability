---
sidebar_position: 1
---

# API Overview

The Site Availability API provides RESTful endpoints for monitoring applications, locations, and system health.

## Main Endpoints

- `GET  /api/apps` — List all monitored applications and their status. Supports filtering by query parameters (e.g., `?location=NY&status=up`).
- `GET  /api/locations` — List all configured locations with their status.
- `GET  /api/labels` — List all available label keys or values.
- `GET  /api/scrape-interval` — Get the current scraping interval in milliseconds.
- `GET  /api/docs` — Get documentation metadata (title, URL).
- `GET  /metrics` — Prometheus metrics for monitoring.
- `GET  /healthz` — Liveness probe.
- `GET  /readyz` — Readiness probe.
- `GET  /sync` — (If enabled) Export all app statuses and locations for federation (protected by HMAC).

## Authentication

- The `/sync` endpoint is protected by HMAC authentication if enabled in the config.
- All other endpoints are open by default.

## Filtering

- Most endpoints support filtering by query parameters, including system fields (e.g., `location`, `status`) and labels (e.g., `labels.env=prod`).

## Error Handling

- Standard HTTP status codes are used.
- Error responses are JSON with an `error` field.

---

For more details, see the code in `backend/handlers/` and `backend/server/`.
