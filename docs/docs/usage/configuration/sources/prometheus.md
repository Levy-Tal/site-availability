---
sidebar_position: 4
---

# Prometheus Source Configuration

The Prometheus source monitors applications by making a PromQL query to a Prometheus instance. If the query result is **1**, the app is considered **up**; if **0**, it is **down**.

## How It Works

- For each app, the source sends a PromQL query to the configured Prometheus URL.
- If the query result is `1`, the app is **up**. If `0`, the app is **down**.
- You can manipulate your PromQL to return 1 or 0 as needed (e.g., using `up{...}` or custom expressions).
- To flip the result (e.g., if `up` is 0 when you want up), use `1 - (promql)`.

## Minimal Example

```yaml
sources:
  - name: prometheus-main
    type: prometheus
    config:
      url: http://prometheus:9090
      apps:
        - name: myApp
          location: New York
          metric: up{instance="app:8080", job="app"}
```

## Source Configuration Options

- **name**: Unique name for the source (required)
- **type**: Must be `prometheus` (required)
- **config.url**: Prometheus base URL (required)
- **config.token**: Optional authentication token (for bearer/basic auth)
- **config.auth**: Optional authentication type (`bearer` or `basic`)
- **config.apps**: List of app configurations (see below)

## App Configuration Options

Each app under `apps` supports the following options:

```yaml
apps:
  - name: "myApp" # Required: Unique app name
    location: "New York" # Required: Must match a defined location
    metric: 'up{instance="app:8080", job="app"}' # Required: PromQL query
    labels: {} # Optional: App-specific labels
```

- **name**: Unique name for the app (required)
- **location**: Must match a defined location (required)
- **metric**: PromQL query string (required). Should return 1 for up, 0 for down. You can use expressions like `1 - up{...}` to flip the result.
- **labels**: Optional key-value labels for this app

## Authentication

- Use `token` and `auth` in the source config for bearer/basic authentication if your Prometheus requires it.
- For sensitive credentials, use `credentials.yaml` with the same structure as your main config.

### Example: Bearer Authentication

```yaml
sources:
  - name: prometheus-main
    type: prometheus
    config:
      url: http://prometheus:9090
      auth: bearer
      token: "your-bearer-token"
      apps:
        - name: myApp
          location: New York
          metric: up{instance="app:8080", job="app"}
```

### Example: Basic Authentication

```yaml
sources:
  - name: prometheus-main
    type: prometheus
    config:
      url: http://prometheus:9090
      auth: basic
      token: "username:password"
      apps:
        - name: myApp
          location: New York
          metric: up{instance="app:8080", job="app"}
```

### Using credentials.yaml for Sensitive Data

```yaml
# config.yaml
sources:
  - name: prometheus-main
    type: prometheus
    config:
      url: http://prometheus:9090
      auth: bearer
      apps:
        - name: myApp
          location: New York
          metric: up{instance="app:8080", job="app"}

# credentials.yaml
sources:
  - name: prometheus-main
    config:
      token: "your-bearer-token"
```

## Best Practices

- Write PromQL queries that return 1 for up and 0 for down.
- Use `1 - (promql)` to invert results if needed.
- Store sensitive tokens in `credentials.yaml`.

---

For more details, see the code or your configuration files.
