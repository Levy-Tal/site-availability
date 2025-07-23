---
sidebar_position: 3
---

# API Endpoints

This page lists the main API endpoints for Site Availability Monitoring.

## Endpoints

- `GET  /api/apps` — List all monitored applications and their status. Supports filtering by query parameters (e.g., `?location=NY&status=up`).
- `GET  /api/locations` — List all configured locations with their status.
- `GET  /api/labels` — List all available label keys or values.
- `GET  /api/scrape-interval` — Get the current scraping interval in milliseconds.
- `GET  /api/docs` — Get documentation metadata (title, URL).
- `GET  /metrics` — Prometheus metrics for monitoring.
- `GET  /healthz` — Liveness probe.
- `GET  /readyz` — Readiness probe.
- `GET  /sync` — (If enabled) Export all app statuses and locations for federation (protected by HMAC).

## /sync Endpoint Example

### Request

```http
GET /sync HTTP/1.1
Host: site-a.example.com
X-Site-Sync-Timestamp: 2024-06-01T12:00:00Z
X-Site-Sync-Signature: <signature>
```

### Response

```json
{
  "locations": [
    {
      "name": "New York",
      "latitude": 40.712776,
      "longitude": -74.005974,
      "source": "site-a",
      "status": "up"
    }
  ],
  "apps": [
    {
      "name": "frontend",
      "location": "New York",
      "status": "up",
      "source": "site-a",
      "origin_url": "http://site-a:8080",
      "labels": { "env": "prod" }
    }
  ]
}
```

- The `X-Site-Sync-Signature` header contains the HMAC signature (see HMAC docs for details).
- The `X-Site-Sync-Timestamp` header contains the request timestamp (RFC3339 format).

---

For more details, see the code in `backend/handlers/` and `backend/authentication/hmac/`.
