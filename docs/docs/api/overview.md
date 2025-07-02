---
sidebar_position: 1
---

# API Overview

The Site Availability Monitoring API provides RESTful endpoints for accessing application status, configuration, and metrics data.

## Base URL

```
http://localhost:8080/api
```

## API Design Principles

### RESTful Architecture

- Resource-based URLs
- HTTP methods for actions
- Stateless requests
- JSON responses

### Response Format

All API responses follow a consistent structure:

```json
{
  "data": {...},
  "meta": {
    "timestamp": "2023-12-01T10:00:00Z",
    "count": 10,
    "version": "v1"
  },
  "error": null
}
```

### Error Handling

Error responses include:

```json
{
  "data": null,
  "meta": {
    "timestamp": "2023-12-01T10:00:00Z"
  },
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request parameters",
    "details": ["scrape_interval must be positive"]
  }
}
```

## Available Endpoints

### Health and Status

```http
GET /health
```

System health check endpoint.

**Response:**

```json
{
  "status": "healthy",
  "timestamp": "2023-12-01T10:00:00Z",
  "uptime": "2h30m15s"
}
```

### Applications

```http
GET /api/apps
```

Get all monitored applications and their status.

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
      }
    }
  ],
  "meta": {
    "count": 1,
    "timestamp": "2023-12-01T10:00:00Z"
  }
}
```

### Locations

```http
GET /api/locations
```

Get all configured monitoring locations.

**Response:**

```json
{
  "data": [
    {
      "name": "New York",
      "latitude": 40.712776,
      "longitude": -74.005974,
      "apps_count": 3
    }
  ],
  "meta": {
    "count": 1,
    "timestamp": "2023-12-01T10:00:00Z"
  }
}
```

### Configuration

```http
GET /api/config
```

Get current configuration information.

**Response:**

```json
{
  "data": {
    "scrape_interval": "30s",
    "total_apps": 5,
    "total_locations": 2,
    "version": "1.0.0"
  }
}
```

```http
POST /api/scrape-interval
```

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
    "updated_at": "2023-12-01T10:00:00Z"
  }
}
```

## Authentication

The API supports HMAC-SHA256 authentication for secure access.

### HMAC Authentication

Include the HMAC signature in the request header:

```http
Authorization: HMAC-SHA256 <signature>
X-Timestamp: 1638360000
```

### Signature Generation

```javascript
// JavaScript example
const crypto = require("crypto");

function generateSignature(method, path, body, timestamp, secret) {
  const message = `${method}\n${path}\n${body}\n${timestamp}`;
  return crypto.createHmac("sha256", secret).update(message).digest("hex");
}
```

```go
// Go example
func GenerateSignature(method, path, body, timestamp, secret string) string {
    message := fmt.Sprintf("%s\n%s\n%s\n%s", method, path, body, timestamp)
    h := hmac.New(sha256.New, []byte(secret))
    h.Write([]byte(message))
    return hex.EncodeToString(h.Sum(nil))
}
```

## Rate Limiting

The API implements rate limiting to prevent abuse:

- **Rate**: 100 requests per minute per IP
- **Burst**: 10 requests
- **Headers**: Rate limit information in response headers

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1638360060
```

## Versioning

The API uses URL versioning:

- Current version: `v1`
- Full URL: `http://localhost:8080/api/v1/apps`
- Default: If no version specified, defaults to latest

## CORS Support

Cross-Origin Resource Sharing (CORS) is enabled for frontend applications:

```http
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization, X-Timestamp
```

## WebSocket Support (Future)

Real-time updates via WebSocket:

```javascript
const ws = new WebSocket("ws://localhost:8080/api/ws");

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  // Handle real-time updates
};
```

## SDK and Client Libraries

### JavaScript Client

```javascript
import { SiteAvailabilityClient } from "site-availability-client";

const client = new SiteAvailabilityClient({
  baseUrl: "http://localhost:8080",
  apiKey: "your-api-key",
});

// Get applications
const apps = await client.getApplications();

// Update scrape interval
await client.updateScrapeInterval("60s");
```

### Go Client

```go
import "github.com/levy-tal/site-availability/client"

client := client.New(&client.Config{
    BaseURL: "http://localhost:8080",
    APIKey:  "your-api-key",
})

apps, err := client.GetApplications(context.Background())
```

## Monitoring and Metrics

The API exposes Prometheus metrics:

```http
GET /metrics
```

Available metrics:

- `api_requests_total`: Total API requests
- `api_request_duration_seconds`: Request duration
- `api_errors_total`: Total API errors

## Development Tools

### OpenAPI Specification

The API is documented using OpenAPI 3.0:

```yaml
openapi: 3.0.0
info:
  title: Site Availability Monitoring API
  version: 1.0.0
  description: API for monitoring application availability
servers:
  - url: http://localhost:8080/api/v1
paths:
  /apps:
    get:
      summary: Get all applications
      responses:
        "200":
          description: List of applications
```

### Postman Collection

Import the Postman collection for easy API testing:

```json
{
  "info": {
    "name": "Site Availability Monitoring API",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/"
  },
  "item": [
    {
      "name": "Get Applications",
      "request": {
        "method": "GET",
        "url": "{{baseUrl}}/api/apps"
      }
    }
  ]
}
```

## Error Codes

| Code | Description           |
| ---- | --------------------- |
| 200  | Success               |
| 400  | Bad Request           |
| 401  | Unauthorized          |
| 403  | Forbidden             |
| 404  | Not Found             |
| 429  | Rate Limited          |
| 500  | Internal Server Error |
| 503  | Service Unavailable   |

## Next Steps

- Learn about specific [API Endpoints](./endpoints)
- Set up [Authentication](./authentication)
- Explore [Metrics Integration](./metrics)
- Check out [Client Libraries](./clients)
