---
sidebar_position: 3
---

# Runtime Data Structures

This document provides a comprehensive overview of all runtime data structures in the Site Availability Monitoring application.

## Overview

The application maintains several interconnected data structures in memory to provide fast access to monitoring data, user sessions, and filtering capabilities. All structures are designed for high-performance read operations with thread-safe concurrent access.

## 1. Core Application Cache (handlers package)

### App Status Cache

The primary data structure storing all application status information, organized by origin URL to prevent circular scraping and enable proper deduplication.

```go
appStatusCache = make(map[string]map[string]map[string]AppStatus)
// Structure: [origin_url][source][app_name]AppStatus
```

**Structure:**

```
appStatusCache
├── "https://site-a.com" (Origin URL)
│   ├── "prometheus-source"
│   │   ├── "backend-app" → AppStatus{...}
│   │   └── "frontend-app" → AppStatus{...}
│   └── "http-source"
│       ├── "api-service" → AppStatus{...}
│       └── "web-service" → AppStatus{...}
├── "https://site-b.com" (Origin URL)
│   └── "site-b-source"
│       ├── "site-b-app" → AppStatus{...}
│       └── "scraped-from-c" → AppStatus{...}
└── "https://site-e.com" (Origin URL)
    └── "site-e-source"
        ├── "site-e-app" → AppStatus{...}
        └── "e-scraped-from-f" → AppStatus{...}
```

**AppStatus Object:**

```go
type AppStatus struct {
    Name      string            // Unique app name within source
    Location  string            // Geographic location
    Status    string            // "up", "down", "unavailable"
    Source    string            // Source name (redundant with map key)
    OriginURL string            // Where app originally came from
    Labels    map[string]string // Merged labels (app + source + server)
}
```

**Key Characteristics:**

- **Uniqueness**: `originURL + sourceName + appName` forms the unique identifier
- **Origin URL Grouping**: Apps are grouped by their origin URL for circular prevention
- **Source Isolation**: Within each origin URL, each source has its own namespace
- **Thread Safety**: Protected by `cacheMutex sync.RWMutex`
- **URL Normalization**: Origin URLs are normalized (lowercase, no trailing slashes, no default ports)

### Location Cache

Stores geographic location information for each source with conflict resolution to prevent duplicates.

```go
locationCache = make(map[string][]Location)
```

**Structure:**

```
locationCache
├── "prometheus-source" → [Location{Name: "us-east", Lat: 40.7, Lon: -74.0}, ...]
├── "http-source" → [Location{Name: "us-west", Lat: 37.7, Lon: -122.4}, ...]
└── "site-e-source" → [Location{Name: "eu-west", Lat: 51.5, Lon: -0.1}, ...]
```

**Conflict Resolution:**

- **Universal Rule**: Any source that returns locations will have conflicting locations filtered out
- **Server Precedence**: Server's own configured locations always take precedence over scraped locations
- **No Duplicates**: Prevents duplicate locations when scraped locations match server's configured locations

**Location Object:**

```go
type Location struct {
    Name      string  // Location name
    Latitude  float64 // Geographic latitude
    Longitude float64 // Geographic longitude
    Source    string  // Source that provided this location
    Status    *string // Calculated status based on apps in location
}
```

### Performance Metrics

Tracks operational statistics for monitoring and debugging.

```go
updateMetrics = struct {
    totalUpdates     int64 // Total number of cache updates
    totalAppsAdded   int64 // Total apps added to cache
    totalAppsSkipped int64 // Total apps skipped (filtered)
    totalErrors      int64 // Total errors encountered
    avgDurationMs    int64 // Average update duration in milliseconds
}{}
```

## 2. Label Management System (labels package)

### Label Manager

Provides O(1) filtering capabilities for both system fields and user labels.

```go
type LabelManager struct {
    appsByField map[string]map[string][]string  // field_name -> value -> [app_ids]
    mutex       sync.RWMutex
}
```

**Structure:**

```
labelManager.appsByField
├── "location" (System Field)
│   ├── "us-east" → ["prometheus:backend-app", "http:api-service"]
│   ├── "us-west" → ["prometheus:frontend-app", "http:web-service"]
│   └── "eu-west" → ["site-e:site-e-app"]
├── "status" (System Field)
│   ├── "up" → ["prometheus:backend-app", "site-e:site-e-app"]
│   └── "down" → ["http:api-service", "site-e:e-scraped-from-f"]
├── "labels.env" (User Label)
│   ├── "prod" → ["prometheus:backend-app", "site-e:site-e-app"]
│   └── "staging" → ["http:api-service"]
└── "labels.team" (User Label)
    ├── "backend" → ["prometheus:backend-app"]
    └── "frontend" → ["prometheus:frontend-app", "http:web-service"]
```

**Key Features:**

- **Dual Indexing**: System fields and user labels with "labels." prefix
- **Unique IDs**: App identifiers in format `source:appname`
- **Fast Queries**: O(1) lookup for field-based filtering

## 3. Scraping System (scraping package)

### Scrapers Registry

Maintains a registry of all configured scrapers with generic orchestration.

```go
Scrapers = make(map[string]Source)
```

**Structure:**

```
Scrapers
├── "prometheus-source" → PrometheusScraper{}
├── "http-source" → HTTPScraper{}
└── "site-e-source" → SiteScraper{directScrapedSites: ["http://server-b:8080", "http://server-c:8080"]}
```

**Source Interface:**

```go
type Source interface {
    ValidateConfig(source config.Source) error
    Scrape(source config.Source, serverSettings config.ServerSettings,
           timeout time.Duration, maxParallel int, tlsConfig *tls.Config)
           ([]handlers.AppStatus, []handlers.Location, error)
}
```

**Generic Orchestration:**

- **Uniform Treatment**: All scrapers are treated identically by the orchestration layer
- **Encapsulated Logic**: Each scraper handles its own specific logic internally
- **No Source-Type Conditionals**: The `Start` function makes generic calls to `scraper.Scrape()`
- **Configuration at Startup**: Site scrapers are configured with `directScrapedSites` during initialization

### TLS Configuration

Global TLS configuration for secure scraping.

```go
globalTLSConfig *tls.Config
```

## 4. Authentication & Authorization (authentication package)

### Session Management

Manages user sessions with automatic expiration.

```go
type Manager struct {
    sessions map[string]*Session
    mutex    sync.RWMutex
    timeout  time.Duration
}
```

**Structure:**

```
sessionManager.sessions
├── "sess_abc123" → Session{
│   ID: "sess_abc123",
│   Username: "john.doe",
│   IsAdmin: false,
│   Roles: ["developer"],
│   Groups: ["team-a"],
│   AuthMethod: "oidc",
│   CreatedAt: 2024-01-15T10:30:00Z,
│   ExpiresAt: 2024-01-15T18:30:00Z
│ }
├── "sess_def456" → Session{
│   ID: "sess_def456",
│   Username: "admin.user",
│   IsAdmin: true,
│   Roles: ["admin"],
│   Groups: ["admin-team"],
│   AuthMethod: "local",
│   CreatedAt: 2024-01-15T09:00:00Z,
│   ExpiresAt: 2024-01-15T17:00:00Z
│ }
└── "sess_ghi789" → Session{...}
```

**Session Object:**

```go
type Session struct {
    ID         string    // Unique session identifier
    Username   string    // User's username
    IsAdmin    bool      // Admin privileges flag
    Roles      []string  // User's roles
    Groups     []string  // User's groups
    AuthMethod string    // "local" or "oidc"
    CreatedAt  time.Time // Session creation timestamp
    ExpiresAt  time.Time // Session expiration timestamp
}
```

### RBAC Permissions

Role-based access control for label filtering.

```go
type UserPermissions struct {
    IsAdmin       bool
    AllowedLabels map[string]LabelPermission
    HasFullAccess bool
}
```

**Structure:**

```
userPermissions.AllowedLabels
├── "env" → LabelPermission{
│   Key: "env",
│   AllowedValues: ["prod", "staging"],
│   AllowAll: false
│ }
├── "team" → LabelPermission{
│   Key: "team",
│   AllowedValues: ["backend", "frontend"],
│   AllowAll: false
│ }
└── "region" → LabelPermission{
    Key: "region",
    AllowedValues: ["us-east", "us-west"],
    AllowAll: false
}
```

**LabelPermission Object:**

```go
type LabelPermission struct {
    Key           string   // Label key
    AllowedValues []string // Allowed values for this label
    AllowAll      bool     // Admin can access all values
}
```

## 5. Metrics & Monitoring (metrics package)

### Location Source Counts

Tracks metrics per location and source combination.

```go
locationSourceCounts := make(map[locsrc]struct {
    count   int
    sources []string
})
```

**Structure:**

```
locationSourceCounts
├── locsrc{Location: "us-east", Source: "prometheus"} → {count: 5, sources: ["prom1", "prom2"]}
├── locsrc{Location: "us-west", Source: "http"} → {count: 3, sources: ["http1"]}
└── locsrc{Location: "eu-west", Source: "site"} → {count: 2, sources: ["site-e"]}
```

## 6. Data Flow Patterns

### Data Ingestion Flow

```
1. Scrapers (Prometheus/HTTP/Site)
   ↓
2. UpdateAppStatus()
   ↓
3. Group by normalized origin URL
   ↓
4. appStatusCache[originURL][sourceName][appName] = AppStatus
   ↓
5. labelManager.UpdateAppLabels()
   ↓
6. labelManager.appsByField[field][value] = [app_ids]
```

### API Request Flow

```
1. HTTP Request → /api/apps?location=us-east&labels.env=prod
   ↓
2. parseFilters() → filters map
   ↓
3. labelManager.FindAppsByLabels(filters) → [app_ids]
   ↓
4. filterApps() → filtered AppStatus slice
   ↓
5. JSON Response
```

## 7. Concurrency & Thread Safety

### Mutex Protection

All shared data structures are protected by appropriate mutexes:

```go
// handlers package
cacheMutex sync.RWMutex  // Protects appStatusCache, locationCache

// labels package
labelManager.mutex sync.RWMutex  // Protects appsByField

// session package
sessionManager.mutex sync.RWMutex  // Protects sessions map

// scraping package
// No global mutex - each scraper is independent
```

**Locking Strategy:**

- **Read-Write Mutexes**: Allow concurrent reads, exclusive writes
- **Fine-grained Locking**: Each major structure has its own mutex
- **Minimal Lock Time**: Locks are held only for the duration of operations

## 8. Memory Usage Patterns

### Memory Allocation Estimates

```
appStatusCache: ~3-15 MB (depending on app count and origin URL diversity)
locationCache: ~1-5 MB (depending on location count)
labelManager: ~5-20 MB (indexing overhead)
sessionManager: ~1-5 MB (depending on active sessions)
Scrapers: ~1-2 MB (scraper instances)
Total: ~11-47 MB typical runtime
```

### Growth Patterns

- **appStatusCache**: Grows with number of origin URLs × sources × apps per source
- **labelManager**: Grows with number of unique field values
- **sessionManager**: Grows with active user sessions
- **Scrapers**: Fixed size based on configured sources

## 9. Data Relationships

```
AppStatus ←→ Location (via location field)
AppStatus ←→ Labels (via labels map)
AppStatus ←→ Source (via source field)
AppStatus ←→ OriginURL (via origin_url field, used for cache grouping)
Session ←→ UserPermissions (via roles)
UserPermissions ←→ LabelPermission (via allowed labels)
Scrapers ←→ AppStatus (via scraping results)
SiteScraper ←→ directScrapedSites (via internal state for circular prevention)
```

## 10. Performance Characteristics

### Access Patterns

- **App Lookup**: O(1) via `appStatusCache[originURL][source][app]`
- **Label Filtering**: O(1) via `labelManager.appsByField[field][value]`
- **Session Validation**: O(1) via `sessions[sessionID]`
- **Permission Check**: O(n) where n = number of user roles

### Update Patterns

- **App Status Update**: O(1) for cache update + O(n) for label reindexing + O(1) for URL normalization
- **Session Creation**: O(1)
- **Label Manager Update**: O(n×m) where n = apps, m = fields per app
- **Location Conflict Resolution**: O(n) where n = number of scraped locations

### Optimization Strategies

- **Pre-allocated Maps**: Use `make(map[string]string, estimatedSize)` for known sizes
- **Index Reuse**: Label manager maintains persistent indexes
- **Batch Operations**: Multiple app updates are batched where possible
- **Memory Pooling**: Reuse objects where appropriate

## 11. Circular Scraping Prevention

### Deduplication Rules

The application implements sophisticated rules to prevent circular scraping:

1. **Origin URL Validation**: Apps with empty `OriginURL` are rejected (mandatory field)
2. **URL Normalization**: Origin URLs are normalized for consistent grouping
3. **Origin URL Grouping**: Apps are grouped by normalized origin URL in cache
4. **Circular Prevention**: Site scrapers drop apps from directly scraped sites
5. **Origin URL Preservation**: Site scrapers preserve original origin URLs from scraped apps

### Example Scenario

```
Site A (host_url: https://site-a.com) scrapes:
├── Prometheus source → OriginURL = https://site-a.com (kept - not a site scraper)
├── Site B (https://site-b.com) → OriginURL = https://site-b.com (dropped - directly scraped)
└── Site E (https://site-e.com) → OriginURL preserved from Site E (kept - not directly scraped)
```

### Cache Structure Benefits

- **Prevents Duplicates**: Apps with same origin URL are grouped together
- **Enables Deduplication**: Easy to replace entire source data for an origin URL
- **Maintains Uniqueness**: `originURL + sourceName + appName` ensures uniqueness
- **Supports Circular Prevention**: Origin URL grouping enables efficient filtering

## 12. Error Handling & Resilience

### Error States

- **Invalid Data**: Apps with invalid status/location are corrected or skipped
- **Scraper Failures**: Failed scrapers return empty results, don't crash system
- **Memory Pressure**: Large datasets are handled gracefully
- **Concurrent Access**: Race conditions are prevented by proper locking

### Recovery Mechanisms

- **Graceful Degradation**: System continues operating with partial data
- **Automatic Cleanup**: Expired sessions and invalid data are automatically removed
- **Error Logging**: Comprehensive logging for debugging and monitoring

---

This data structure design ensures high performance, thread safety, and scalability while maintaining data integrity and providing rich filtering capabilities.
