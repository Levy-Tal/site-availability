---
sidebar_position: 4
---

# Backend Sources Development

This guide covers how to develop and extend the backend sources for Site Availability.

## Overview

The backend sources are responsible for collecting monitoring data from various providers. Each source implements a common interface and can be easily extended to support new monitoring systems.

## Source Interface

All sources implement the following interface:

```go
type Source interface {
    Name() string
    Collect(ctx context.Context) ([]Status, error)
    Validate() error
}
```

## Creating a New Source

### 1. Define the Source Structure

```go
type MySource struct {
    name     string
    url      string
    interval time.Duration
    timeout  time.Duration
}

func NewMySource(config Config) (*MySource, error) {
    return &MySource{
        name:     config.Name,
        url:      config.URL,
        interval: config.Interval,
        timeout:  config.Timeout,
    }, nil
}
```

### 2. Implement the Interface Methods

```go
func (s *MySource) Name() string {
    return s.name
}

func (s *MySource) Validate() error {
    if s.name == "" {
        return errors.New("name is required")
    }
    if s.url == "" {
        return errors.New("url is required")
    }
    return nil
}

func (s *MySource) Collect(ctx context.Context) ([]Status, error) {
    // Implementation here
    return statuses, nil
}
```

### 3. Register the Source

Add your source to the source registry:

```go
func init() {
    sources.Register("my-source", NewMySource)
}
```

## Example: HTTP Source

Here's a complete example of a simple HTTP source:

```go
package http

import (
    "context"
    "errors"
    "net/http"
    "time"
)

type HTTPSource struct {
    name     string
    url      string
    method   string
    headers  map[string]string
    timeout  time.Duration
    client   *http.Client
}

func NewHTTPSource(config Config) (*HTTPSource, error) {
    client := &http.Client{
        Timeout: config.Timeout,
    }

    return &HTTPSource{
        name:    config.Name,
        url:     config.URL,
        method:  config.Method,
        headers: config.Headers,
        timeout: config.Timeout,
        client:  client,
    }, nil
}

func (s *HTTPSource) Name() string {
    return s.name
}

func (s *HTTPSource) Validate() error {
    if s.name == "" {
        return errors.New("name is required")
    }
    if s.url == "" {
        return errors.New("url is required")
    }
    return nil
}

func (s *HTTPSource) Collect(ctx context.Context) ([]Status, error) {
    req, err := http.NewRequestWithContext(ctx, s.method, s.url, nil)
    if err != nil {
        return nil, err
    }

    // Add headers
    for key, value := range s.headers {
        req.Header.Set(key, value)
    }

    resp, err := s.client.Do(req)
    if err != nil {
        return []Status{{
            Name:   s.name,
            Status: "down",
            Error:  err.Error(),
        }}, nil
    }
    defer resp.Body.Close()

    status := "up"
    if resp.StatusCode >= 400 {
        status = "down"
    }

    return []Status{{
        Name:   s.name,
        Status: status,
        Code:   resp.StatusCode,
    }}, nil
}
```

## Configuration

Sources should support configuration through YAML:

```yaml
sources:
  - name: "my-service"
    type: "my-source"
    url: "https://api.example.com/health"
    method: "GET"
    headers:
      Authorization: "Bearer token"
    timeout: 10s
    interval: 30s
```

## Error Handling

Sources should handle errors gracefully:

1. **Network errors**: Return status as "down" with error message
2. **Configuration errors**: Log and skip the source
3. **Partial failures**: Return available data with warnings

## Testing

Create comprehensive tests for your source:

```go
func TestHTTPSource(t *testing.T) {
    config := Config{
        Name:    "test",
        URL:     "https://httpbin.org/status/200",
        Method:  "GET",
        Timeout: 10 * time.Second,
    }

    source, err := NewHTTPSource(config)
    require.NoError(t, err)

    statuses, err := source.Collect(context.Background())
    require.NoError(t, err)
    require.Len(t, statuses, 1)
    require.Equal(t, "up", statuses[0].Status)
}
```

## Best Practices

1. **Use context**: Always respect context cancellation
2. **Handle timeouts**: Set appropriate timeouts for external calls
3. **Validate configuration**: Check required fields and valid values
4. **Log appropriately**: Use structured logging for debugging
5. **Return meaningful errors**: Provide clear error messages
6. **Test thoroughly**: Include unit and integration tests

## Integration

To integrate your source with the main application:

1. Add configuration validation
2. Register the source in the source manager
3. Add configuration examples to documentation
4. Update the configuration schema
5. Add tests to the test suite

## Common Patterns

### Retry Logic

```go
func (s *MySource) collectWithRetry(ctx context.Context) ([]Status, error) {
    var lastErr error
    for i := 0; i < s.retries; i++ {
        statuses, err := s.collect(ctx)
        if err == nil {
            return statuses, nil
        }
        lastErr = err
        time.Sleep(s.retryDelay)
    }
    return nil, lastErr
}
```

### Rate Limiting

```go
func (s *MySource) Collect(ctx context.Context) ([]Status, error) {
    s.rateLimiter.Wait(ctx)
    return s.collect(ctx)
}
```

### Caching

```go
func (s *MySource) Collect(ctx context.Context) ([]Status, error) {
    if cached := s.cache.Get(s.name); cached != nil {
        return cached.([]Status), nil
    }

    statuses, err := s.collect(ctx)
    if err == nil {
        s.cache.Set(s.name, statuses, s.cacheTTL)
    }
    return statuses, err
}
```
