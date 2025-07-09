# HTTP Source Configuration

The HTTP source allows you to monitor websites and HTTP endpoints. It supports authentication, content validation, and customizable HTTP settings.

## Basic Configuration

Here's a minimal configuration example:

```yaml
sources:
  - name: web-monitoring
    type: http
    labels:
      service: "web-monitoring"
    config:
      apps:
        - name: website
          location: "US-East"
          url: "https://example.com"
          method: "GET"
          allowed_status_codes: ["2XX"]
```

## Full Configuration Reference

### Source Level Configuration

```yaml
sources:
  - name: web-monitoring # Required: Unique name for the source
    type: http # Required: Must be "http"
    labels: # Optional: Labels for all apps in this source
      service: "web"
      environment: "prod"
    config:
      apps: # Required: List of apps to monitor
        -  # App configuration (see below)
```

### App Level Configuration

Each app under the `apps` section supports the following options:

```yaml
apps:
  - name: "website" # Required: Unique name for the app
    location: "US-East" # Required: Must match a location defined in your config
    url: "https://example.com" # Required: URL to monitor

    # Basic HTTP Options
    method:
      "GET" # Optional: HTTP method (GET, POST, PUT, DELETE, HEAD, OPTIONS)
      # Default: GET
    headers: {} # Optional: Custom HTTP headers
    body: "" # Optional: Request body for POST/PUT
    content_type: "" # Optional: Content-Type header

    # Connection Options
    timeout:
      "30s" # Optional: Request timeout
      # Default: Uses global scraping timeout
    follow_redirects:
      true # Optional: Whether to follow redirects
      # Default: true
    max_redirects:
      10 # Optional: Maximum number of redirects to follow
      # Default: 10

    # SSL Configuration
    ssl_verify:
      true # Optional: Whether to verify SSL certificates
      # Default: true

    # Status Code Configuration
    allowed_status_codes: # Optional: Status codes considered successful
      - "2XX" # Can use ranges (1XX-5XX) or specific codes (200)
      - 200 # Default: ["2XX"]
      - 201
    blocked_status_codes: # Optional: Status codes considered failures
      - "5XX" # Default: ["4XX", "5XX"]
      - 429

    # Content Validation
    validation: # Optional: Content validation rules
      success: # Conditions that must be true for success
        - type: "body_contains"
          text: "Welcome"
          case_sensitive: false
        - type: "body_not_contains"
          text: "Error"
        - type: "status_code"
          text: "200"
        - type: "response_time"
          max_ms: 3000
        - type: "json_path"
          path: "$.status"
          expected_value: "ok"
      failure: [] # Conditions that indicate failure if true

    # Authentication
    auth: # Optional
      type: "basic" # "basic", "bearer", "digest", "oauth2"
      username: "user" # For basic/digest
      password: "pass" # For basic/digest
      token: "token" # For bearer/oauth2

    labels: {} # Optional: App-specific labels
```

## Authentication

Authentication should be configured in your credentials file, which will be merged with your main configuration. The credentials file must follow the same structure as your main configuration file for proper merging.

### Main Configuration File (config.yaml):

```yaml
sources:
  - name: web-monitoring
    type: http
    config:
      apps:
        - name: secure-api
          location: "US-East"
          url: "https://api.example.com"
          method: "GET"
        - name: oauth-service
          location: "US-West"
          url: "https://oauth.example.com"
          method: "GET"
```

### Credentials Configuration (credentials.yaml):

```yaml
sources:
  - name: web-monitoring # Must match source name in main config
    config:
      apps:
        - name: secure-api # Must match app name in main config
          auth:
            type: "basic"
            username: "user"
            password: "pass"

        - name: oauth-service # Must match app name in main config
          auth:
            type: "bearer"
            token: "your-token"
```

### Important Notes:

1. The structure in credentials.yaml must exactly match your main config
2. The `name` fields must match at both source and app levels
3. The files will be merged based on matching names
4. Credentials take precedence over any auth settings in the main config

### Supported Authentication Types:

- `basic`: Basic HTTP authentication
  ```yaml
  auth:
    type: "basic"
    username: "user"
    password: "pass"
  ```
- `bearer`: Bearer token authentication
  ```yaml
  auth:
    type: "bearer"
    token: "your-token"
  ```
- `digest`: Digest authentication
  ```yaml
  auth:
    type: "digest"
    username: "user"
    password: "pass"
  ```
- `oauth2`: OAuth 2.0 token authentication
  ```yaml
  auth:
    type: "oauth2"
    token: "your-oauth-token"
  ```

## Content Validation Types

The following validation types are supported:

1. `body_contains`:

   - Checks if response body contains specified text
   - Options:
     - `text`: Text to search for
     - `case_sensitive`: Whether to perform case-sensitive search (default: false)

2. `body_not_contains`:

   - Checks if response body does not contain specified text
   - Same options as `body_contains`

3. `status_code`:

   - Validates specific status code
   - Options:
     - `text`: Expected status code as string

4. `response_time`:

   - Validates response time
   - Options:
     - `max_ms`: Maximum allowed response time in milliseconds

5. `json_path`:
   - Validates JSON response using path expression
   - Options:
     - `path`: JSON path expression
     - `expected_value`: Expected value at path
     - `value_type`: Type of value to compare (string, number, boolean)
     - `operator`: Comparison operator (eq, ne, gt, lt, etc.)

## Status Code Ranges

Status codes can be specified as:

- Exact codes: `200`, `201`, `404`
- Ranges: `"2XX"`, `"4XX"`, `"5XX"`
  - `"1XX"`: Informational responses (100-199)
  - `"2XX"`: Successful responses (200-299)
  - `"3XX"`: Redirection messages (300-399)
  - `"4XX"`: Client errors (400-499)
  - `"5XX"`: Server errors (500-599)

## Best Practices

1. **Timeouts**: Always set appropriate timeouts to avoid hanging requests
2. **SSL Verification**: Keep SSL verification enabled in production unless you specifically need to disable it
3. **Content Validation**: Use specific content validation to ensure responses are valid
4. **Labels**: Add meaningful labels for better filtering and organization
5. **Authentication**: Store credentials in a separate credentials file
6. **Status Codes**: Be specific about which status codes are acceptable
