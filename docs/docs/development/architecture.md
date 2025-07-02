---
sidebar_position: 2
---

# Architecture Overview

Understanding the architecture and design principles behind Site Availability Monitoring.

## System Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│     Frontend    │    │     Backend     │    │   Prometheus    │
│   (React App)   │───▶│   (Go Server)   │───▶│   (Metrics)     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│      Browser    │    │      API        │    │   Target Apps   │
│   (World Map)   │    │   (REST/JSON)   │    │  (Monitored)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Backend Architecture

### Core Components

```go
main.go
├── server/          // HTTP server and routing
├── handlers/        // HTTP request handlers
├── scraping/        // Prometheus data collection
│   ├── prometheus/  // Prometheus client
│   └── site/       // Site monitoring logic
├── config/         // Configuration management
├── logging/        // Structured logging
├── metrics/        // Application metrics
├── labels/         // Label management
└── authentication/ // HMAC authentication
    └── hmac/       // HMAC implementation
```

### Key Packages

#### Server Package (`server/`)

- HTTP server setup and configuration
- Middleware integration
- Graceful shutdown handling
- Health check endpoints

#### Handlers Package (`handlers/`)

- REST API endpoints
- Request/response processing
- Error handling
- Authentication validation

#### Scraping Package (`scraping/`)

- Prometheus query execution
- Data aggregation and processing
- Site status calculation
- Concurrent scraping management

#### Configuration Package (`config/`)

- YAML configuration parsing
- Environment variable overrides
- Configuration validation
- Hot reload capabilities

### Data Flow

1. **Configuration Loading**:

   ```go
   config.yaml → config.Config → validation → runtime config
   ```

2. **Scraping Process**:

   ```go
   Timer → Scraper → Prometheus Query → Data Processing → Cache Update
   ```

3. **API Request Flow**:
   ```go
   HTTP Request → Authentication → Handler → Data Retrieval → JSON Response
   ```

## Frontend Architecture

### Component Structure

```
src/
├── components/
│   ├── Map.js          // Main world map component
│   ├── Sidebar.js      // Status sidebar
│   └── AppStatusPanel.js // Application status display
├── api/
│   ├── appStatusAPI.js    // Application data API
│   ├── docsAPI.js         // Documentation API
│   └── scrapeIntervalAPI.js // Configuration API
├── styles/
│   └── main.css        // Global styles
├── config.js           // Frontend configuration
├── App.js             // Main application component
└── index.js           // Application entry point
```

### State Management

The frontend uses a simple state management approach:

```javascript
// App.js - Main state container
const [appStatuses, setAppStatuses] = useState([]);
const [locations, setLocations] = useState([]);
const [loading, setLoading] = useState(true);

// Data flow: API → State → Components → Render
```

### Map Integration

Uses D3.js for world map visualization:

```javascript
// Map rendering process
Data Loading → SVG Creation → Country Paths → Marker Placement → Event Binding
```

## Database Design

Currently, the application is stateless and doesn't use a persistent database. Data is:

- **Configuration**: Stored in YAML files
- **Metrics**: Retrieved from Prometheus
- **State**: Maintained in memory with periodic refresh

### Future Database Schema

For persistent storage, consider this schema:

```sql
-- Locations table
CREATE TABLE locations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    latitude DECIMAL(10,8) NOT NULL,
    longitude DECIMAL(11,8) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Applications table
CREATE TABLE applications (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    location_id INTEGER REFERENCES locations(id),
    metric_query TEXT NOT NULL,
    prometheus_url VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Status history table
CREATE TABLE status_history (
    id SERIAL PRIMARY KEY,
    application_id INTEGER REFERENCES applications(id),
    status BOOLEAN NOT NULL,
    response_time DECIMAL(10,4),
    timestamp TIMESTAMP DEFAULT NOW()
);
```

## API Design

### RESTful Endpoints

```
GET  /health                 // Health check
GET  /metrics               // Prometheus metrics
GET  /api/apps              // Application statuses
GET  /api/locations         // Geographic locations
GET  /api/config            // Configuration info
POST /api/scrape-interval   // Update scrape interval
```

### Response Format

Standard JSON response structure:

```json
{
  "data": [...],
  "meta": {
    "count": 10,
    "last_updated": "2023-12-01T10:00:00Z"
  },
  "error": null
}
```

## Security Architecture

### Authentication Flow

```
Client Request → HMAC Validation → Request Processing → Response
```

### HMAC Implementation

```go
// HMAC signature calculation
func CalculateHMAC(data, secret string) string {
    h := hmac.New(sha256.New, []byte(secret))
    h.Write([]byte(data))
    return hex.EncodeToString(h.Sum(nil))
}
```

## Monitoring and Observability

### Application Metrics

The backend exposes Prometheus metrics:

```go
// Custom metrics
var (
    scrapeRequests = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "scrape_requests_total",
            Help: "Total number of scrape requests",
        },
        []string{"status"},
    )

    scrapeLatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "scrape_duration_seconds",
            Help: "Scrape request latency",
        },
        []string{"target"},
    )
)
```

### Logging Strategy

Structured logging with different levels:

```go
logger.Info("Starting scrape",
    zap.String("target", target),
    zap.Duration("interval", interval))

logger.Error("Scrape failed",
    zap.Error(err),
    zap.String("target", target))
```

## Performance Considerations

### Backend Optimization

1. **Concurrent Scraping**: Parallel Prometheus queries
2. **Caching**: In-memory cache for frequent requests
3. **Connection Pooling**: Reuse HTTP connections
4. **Graceful Degradation**: Fallback mechanisms

### Frontend Optimization

1. **Lazy Loading**: Load map data on demand
2. **Debouncing**: Limit API call frequency
3. **Memoization**: Cache expensive calculations
4. **Code Splitting**: Bundle optimization

## Scalability Design

### Horizontal Scaling

- **Stateless Backend**: Multiple backend instances
- **Load Balancing**: Distribute requests
- **Prometheus Federation**: Multi-prometheus setup
- **CDN**: Static asset distribution

### Vertical Scaling

- **Resource Limits**: CPU and memory optimization
- **Database Indexing**: Query optimization
- **Caching Layers**: Redis/Memcached integration

## Development Patterns

### Error Handling

```go
// Consistent error handling pattern
func (s *Service) ProcessData() error {
    if err := s.validate(); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }

    if err := s.process(); err != nil {
        return fmt.Errorf("processing failed: %w", err)
    }

    return nil
}
```

### Testing Strategy

- **Unit Tests**: Individual component testing
- **Integration Tests**: API endpoint testing
- **E2E Tests**: Full workflow testing
- **Performance Tests**: Load and stress testing

## Configuration Management

### Hierarchical Configuration

```yaml
# Default configuration
defaults:
  scrape_interval: 30s
  log_level: info

# Environment overrides
development:
  log_level: debug
  scrape_interval: 10s

production:
  log_level: warn
  authentication:
    enabled: true
```

### Environment Variables

```bash
# Override any config value
SA_SCRAPE_INTERVAL=60s
SA_LOG_LEVEL=debug
SA_AUTHENTICATION_HMAC_SECRET=secret
```

## Future Architecture Enhancements

### Microservices Evolution

```
Current: Monolith → Future: Microservices
├── Config Service    // Configuration management
├── Scraper Service   // Metrics collection
├── API Gateway       // Request routing
├── Auth Service      // Authentication
└── UI Service        // Frontend hosting
```

### Event-Driven Architecture

```
Events: Config Change → Scraper Update → Status Change → UI Update
```

### Data Pipeline

```
Prometheus → Kafka → Stream Processing → Time Series DB → API
```

This architecture provides a solid foundation for monitoring applications while remaining simple enough for easy development and deployment.
