---
sidebar_position: 2
---

# API Endpoints

Detailed documentation for all available API endpoints.

## Health Endpoints

### GET /health

System health check endpoint.

**Response:**

```json
{
  "status": "healthy",
  "timestamp": "2023-12-01T10:00:00Z",
  "uptime": "2h30m15s",
  "version": "1.0.0"
}
```

**Status Codes:**

- `200`: Service is healthy
- `503`: Service is unhealthy

---

## Application Endpoints

### GET /api/apps

Get all monitored applications and their current status.

**Response:**

```json
{
  "data": [
    {
      "name": "frontend",
      "location": "New York",
      "status": "up",
      "last_check": "2023-12-01T10:00:00Z",
      "response_time": 0.142,
      "coordinates": {
        "latitude": 40.712776,
        "longitude": -74.005974
      },
      "metric": "up{instance=\"frontend:80\", job=\"frontend\"}",
      "prometheus_url": "http://prometheus:9090/"
    }
  ],
  "meta": {
    "count": 1,
    "timestamp": "2023-12-01T10:00:00Z",
    "last_update": "2023-12-01T09:59:30Z"
  }
}
```

**Status Codes:**

- `200`: Success
- `500`: Internal server error

---

### GET /api/apps/{name}

Get details for a specific application.

**Parameters:**

- `name` (path): Application name

**Response:**

```json
{
  "data": {
    "name": "frontend",
    "location": "New York",
    "status": "up",
    "last_check": "2023-12-01T10:00:00Z",
    "response_time": 0.142,
    "coordinates": {
      "latitude": 40.712776,
      "longitude": -74.005974
    },
    "metric": "up{instance=\"frontend:80\", job=\"frontend\"}",
    "prometheus_url": "http://prometheus:9090/",
    "history": [
      {
        "timestamp": "2023-12-01T09:59:00Z",
        "status": "up",
        "response_time": 0.138
      }
    ]
  }
}
```

**Status Codes:**

- `200`: Success
- `404`: Application not found

---

## Location Endpoints

### GET /api/locations

Get all configured monitoring locations.

**Response:**

```json
{
  "data": [
    {
      "name": "New York",
      "latitude": 40.712776,
      "longitude": -74.005974,
      "apps_count": 3,
      "apps": ["frontend", "backend", "database"]
    },
    {
      "name": "London",
      "latitude": 51.507351,
      "longitude": -0.127758,
      "apps_count": 2,
      "apps": ["api", "cache"]
    }
  ],
  "meta": {
    "count": 2,
    "timestamp": "2023-12-01T10:00:00Z"
  }
}
```

---

### GET /api/locations/{name}

Get details for a specific location.

**Parameters:**

- `name` (path): Location name (URL encoded)

**Response:**

```json
{
  "data": {
    "name": "New York",
    "latitude": 40.712776,
    "longitude": -74.005974,
    "apps_count": 3,
    "apps": [
      {
        "name": "frontend",
        "status": "up",
        "last_check": "2023-12-01T10:00:00Z"
      }
    ],
    "summary": {
      "total_apps": 3,
      "up_count": 2,
      "down_count": 1,
      "availability_percentage": 66.67
    }
  }
}
```

---

## Configuration Endpoints

### GET /api/config

Get current system configuration.

**Response:**

```json
{
  "data": {
    "scrape_interval": "30s",
    "log_level": "info",
    "total_apps": 5,
    "total_locations": 2,
    "version": "1.0.0",
    "uptime": "2h30m15s",
    "prometheus_urls": ["http://prometheus:9090/"]
  },
  "meta": {
    "timestamp": "2023-12-01T10:00:00Z"
  }
}
```

---

### POST /api/scrape-interval

Update the scraping interval.

**Request:**

```json
{
  "interval": "60s"
}
```

**Response:**

```json
{
  "data": {
    "interval": "60s",
    "previous_interval": "30s",
    "updated_at": "2023-12-01T10:00:00Z"
  },
  "meta": {
    "timestamp": "2023-12-01T10:00:00Z"
  }
}
```

**Status Codes:**

- `200`: Successfully updated
- `400`: Invalid interval format
- `403`: Unauthorized (if authentication enabled)

---

## Metrics Endpoints

### GET /metrics

Prometheus metrics endpoint.

**Response:**

```text
# HELP site_availability_scrape_duration_seconds Duration of scrape requests
# TYPE site_availability_scrape_duration_seconds histogram
site_availability_scrape_duration_seconds_bucket{target="frontend",le="0.1"} 45
site_availability_scrape_duration_seconds_bucket{target="frontend",le="0.5"} 50
site_availability_scrape_duration_seconds_sum{target="frontend"} 12.5
site_availability_scrape_duration_seconds_count{target="frontend"} 50

# HELP site_availability_up Application availability status
# TYPE site_availability_up gauge
site_availability_up{app="frontend",location="New York"} 1
site_availability_up{app="backend",location="New York"} 0

# HELP site_availability_scrape_errors_total Total scrape errors
# TYPE site_availability_scrape_errors_total counter
site_availability_scrape_errors_total{target="backend"} 3

# HELP http_requests_total Total HTTP requests
# TYPE http_requests_total counter
http_requests_total{method="GET",path="/api/apps",status="200"} 1245
http_requests_total{method="GET",path="/api/locations",status="200"} 567
```

---

### GET /api/metrics/summary

Get metrics summary in JSON format.

**Response:**

```json
{
  "data": {
    "total_requests": 1812,
    "total_errors": 23,
    "average_response_time": 0.145,
    "uptime_percentage": 98.7,
    "apps_summary": [
      {
        "name": "frontend",
        "status": "up",
        "uptime_24h": 99.2,
        "avg_response_time": 0.142
      }
    ],
    "locations_summary": [
      {
        "name": "New York",
        "apps_up": 2,
        "apps_down": 1,
        "availability": 66.67
      }
    ]
  },
  "meta": {
    "timestamp": "2023-12-01T10:00:00Z",
    "period": "24h"
  }
}
```

---

## Status Endpoints

### GET /api/status

Get overall system status.

**Response:**

```json
{
  "data": {
    "system_status": "healthy",
    "overall_availability": 98.7,
    "total_applications": 5,
    "applications_up": 4,
    "applications_down": 1,
    "total_locations": 2,
    "last_update": "2023-12-01T10:00:00Z",
    "prometheus_connectivity": "connected"
  }
}
```

---

### GET /api/status/history

Get historical status data.

**Query Parameters:**

- `period` (optional): Time period (`1h`, `24h`, `7d`, `30d`) - default: `24h`
- `resolution` (optional): Data resolution (`1m`, `5m`, `1h`) - default: `5m`

**Response:**

```json
{
  "data": {
    "period": "24h",
    "resolution": "5m",
    "data_points": [
      {
        "timestamp": "2023-12-01T09:55:00Z",
        "overall_availability": 98.5,
        "apps_up": 4,
        "apps_down": 1
      },
      {
        "timestamp": "2023-12-01T10:00:00Z",
        "overall_availability": 98.7,
        "apps_up": 4,
        "apps_down": 1
      }
    ]
  },
  "meta": {
    "count": 288,
    "timestamp": "2023-12-01T10:00:00Z"
  }
}
```

---

## WebSocket Endpoints (Future)

### WS /api/ws

Real-time updates via WebSocket.

**Connection:**

```javascript
const ws = new WebSocket("ws://localhost:8080/api/ws");
```

**Message Format:**

```json
{
  "type": "status_update",
  "timestamp": "2023-12-01T10:00:00Z",
  "data": {
    "app": "frontend",
    "status": "up",
    "location": "New York"
  }
}
```

**Message Types:**

- `status_update`: Application status changed
- `config_update`: Configuration changed
- `system_event`: System-level events

---

## Error Response Format

All error responses follow this format:

```json
{
  "data": null,
  "meta": {
    "timestamp": "2023-12-01T10:00:00Z"
  },
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message",
    "details": ["Additional error details"],
    "request_id": "req_123456789"
  }
}
```

### Common Error Codes

| Code                  | Description                     |
| --------------------- | ------------------------------- |
| `VALIDATION_ERROR`    | Request validation failed       |
| `NOT_FOUND`           | Resource not found              |
| `UNAUTHORIZED`        | Authentication required         |
| `FORBIDDEN`           | Access denied                   |
| `RATE_LIMITED`        | Rate limit exceeded             |
| `INTERNAL_ERROR`      | Internal server error           |
| `SERVICE_UNAVAILABLE` | Service temporarily unavailable |

---

## Request/Response Examples

### cURL Examples

```bash
# Get all applications
curl -X GET http://localhost:8080/api/apps

# Get specific application
curl -X GET http://localhost:8080/api/apps/frontend

# Update scrape interval with authentication
curl -X POST http://localhost:8080/api/scrape-interval \
  -H "Content-Type: application/json" \
  -H "Authorization: HMAC-SHA256 <signature>" \
  -H "X-Timestamp: 1638360000" \
  -d '{"interval": "60s"}'

# Get metrics
curl -X GET http://localhost:8080/metrics
```

### JavaScript Examples

```javascript
// Fetch applications
const response = await fetch("/api/apps");
const apps = await response.json();

// Update scrape interval
const updateInterval = async (interval) => {
  const response = await fetch("/api/scrape-interval", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ interval }),
  });

  if (!response.ok) {
    throw new Error("Failed to update interval");
  }

  return response.json();
};
```

This completes the API endpoints documentation. Each endpoint includes request/response examples, status codes, and error handling information.
