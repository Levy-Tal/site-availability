# HTTP Source Configuration

The HTTP source monitors websites and HTTP endpoints by making HTTP requests and determining if an app is **up** or **down** based on the response status code (and optionally, content validation).

## How It Works

- For each configured app, the HTTP source sends an HTTP request to the specified URL.
- If the response status code matches `allowed_status_codes` (default: 2XX), the app is considered **up**.
- If the status code matches `blocked_status_codes` (default: 4XX, 5XX), or content validation fails, the app is **down**.

## Minimal Example

```yaml
sources:
  - name: web-monitoring
    type: http
    config:
      apps:
        - name: website
          location: "US-East"
          url: "https://example.com"
```

## Source Configuration Options

- **name**: Unique name for the source (required)
- **type**: Must be `http` (required)
- **labels**: Optional labels for all apps in this source
- **config.apps**: List of app configurations (see below)

## App Configuration Options

Each app under `apps` supports the following options:

```yaml
apps:
  - name: "website" # Required: Unique app name
    location: "US-East" # Required: Must match a defined location
    url: "https://example.com" # Required: URL to monitor
    method: "GET" # Optional: HTTP method (GET, POST, etc). Default: GET
    headers: {} # Optional: Custom HTTP headers
    body: "" # Optional: Request body for POST/PUT
    content_type: "" # Optional: Content-Type header
    timeout: "" # Optional: Request timeout (e.g., 10s). Default: global scraping timeout
    follow_redirects: true # Optional: Follow HTTP redirects. Default: true
    max_redirects: 10 # Optional: Max redirects to follow. Default: 10
    ssl_verify: true # Optional: Verify SSL certificates. Default: true
    allowed_status_codes: ["2XX"] # Optional: Status codes considered up. Default: ["2XX"]
    blocked_status_codes: ["4XX", "5XX"] # Optional: Status codes considered down. Default: ["4XX", "5XX"]
    validation: # Optional: Content validation rules
      success: [] # List of success conditions
      failure: [] # List of failure conditions
    auth: # Optional: Authentication settings
      type: "basic" # "basic", "bearer", "digest", "oauth2"
      username: "user" # For basic/digest
      password: "pass" # For basic/digest
      token: "token" # For bearer/oauth2
    labels: {} # Optional: App-specific labels
```

## Content Validation

- Use the `validation` section to define rules for checking response content, status code, or response time.
- Supported types: `body_contains`, `body_not_contains`, `status_code`, `response_time`, `json_path`.

## Authentication

- Configure authentication under `auth` (supports `basic`, `bearer`, `digest`, `oauth2`).
- For sensitive credentials, use `credentials.yaml` with the same structure as your main config.

## Status Code Ranges

- Use exact codes (e.g., `200`) or ranges (e.g., `"2XX"`, `"5XX"`).

## Best Practices

- Set timeouts to avoid hanging requests.
- Keep SSL verification enabled in production.
- Use content validation for critical endpoints.
- Store credentials in `credentials.yaml`.

---

For more details on validation and authentication, see the comments in your configuration files or the codebase.
