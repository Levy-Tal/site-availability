---
sidebar_position: 3
---

# Authentication

Secure your Site Availability Monitoring API with HMAC-SHA256 authentication.

## Overview

The API supports HMAC (Hash-based Message Authentication Code) authentication using SHA-256 for secure access. This ensures that requests are authentic and haven't been tampered with.

## Enabling Authentication

### Configuration

Enable HMAC authentication in your configuration:

```yaml
# config.yaml
authentication:
  hmac:
    enabled: true
    secret: ${SA_HMAC_SECRET} # Use environment variable
    algorithm: sha256
    clock_skew_tolerance: 300 # 5 minutes
```

### Environment Variable

Set the HMAC secret via environment variable:

```bash
export SA_HMAC_SECRET="your-very-secure-secret-key-here"
```

## HMAC Signature Process

### 1. Create the Canonical Request

Construct a canonical string from the request:

```
METHOD\nPATH\nBODY\nTIMESTAMP
```

Example:

```
GET\n/api/apps\n\n1638360000
```

For POST requests with body:

```
POST\n/api/scrape-interval\n{"interval":"60s"}\n1638360000
```

### 2. Generate the Signature

Create HMAC-SHA256 signature:

```javascript
const crypto = require("crypto");

function generateSignature(method, path, body, timestamp, secret) {
  const canonicalRequest = `${method}\n${path}\n${body || ""}\n${timestamp}`;
  const signature = crypto
    .createHmac("sha256", secret)
    .update(canonicalRequest)
    .digest("hex");
  return signature;
}
```

### 3. Add Headers to Request

Include the signature and timestamp in request headers:

```http
Authorization: HMAC-SHA256 <signature>
X-Timestamp: <unix_timestamp>
Content-Type: application/json
```

## Implementation Examples

### JavaScript/Node.js

```javascript
class SiteAvailabilityClient {
  constructor(baseUrl, secret) {
    this.baseUrl = baseUrl;
    this.secret = secret;
  }

  generateSignature(method, path, body, timestamp) {
    const crypto = require("crypto");
    const canonicalRequest = `${method}\n${path}\n${body || ""}\n${timestamp}`;
    return crypto
      .createHmac("sha256", this.secret)
      .update(canonicalRequest)
      .digest("hex");
  }

  async makeRequest(method, path, body = null) {
    const timestamp = Math.floor(Date.now() / 1000);
    const bodyString = body ? JSON.stringify(body) : "";
    const signature = this.generateSignature(
      method,
      path,
      bodyString,
      timestamp,
    );

    const headers = {
      Authorization: `HMAC-SHA256 ${signature}`,
      "X-Timestamp": timestamp.toString(),
      "Content-Type": "application/json",
    };

    const response = await fetch(`${this.baseUrl}${path}`, {
      method,
      headers,
      body: bodyString || undefined,
    });

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }

    return response.json();
  }

  // Usage examples
  async getApplications() {
    return this.makeRequest("GET", "/api/apps");
  }

  async updateScrapeInterval(interval) {
    return this.makeRequest("POST", "/api/scrape-interval", { interval });
  }
}

// Usage
const client = new SiteAvailabilityClient(
  "http://localhost:8080",
  "your-secret-key",
);

const apps = await client.getApplications();
```

### Go

```go
package main

import (
    "bytes"
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "strconv"
    "time"
)

type Client struct {
    BaseURL string
    Secret  string
    HTTP    *http.Client
}

func NewClient(baseURL, secret string) *Client {
    return &Client{
        BaseURL: baseURL,
        Secret:  secret,
        HTTP:    &http.Client{Timeout: 30 * time.Second},
    }
}

func (c *Client) generateSignature(method, path, body string, timestamp int64) string {
    canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%d", method, path, body, timestamp)
    h := hmac.New(sha256.New, []byte(c.Secret))
    h.Write([]byte(canonicalRequest))
    return hex.EncodeToString(h.Sum(nil))
}

func (c *Client) makeRequest(method, path string, body interface{}) (*http.Response, error) {
    var bodyBytes []byte
    var bodyString string

    if body != nil {
        var err error
        bodyBytes, err = json.Marshal(body)
        if err != nil {
            return nil, err
        }
        bodyString = string(bodyBytes)
    }

    timestamp := time.Now().Unix()
    signature := c.generateSignature(method, path, bodyString, timestamp)

    req, err := http.NewRequest(method, c.BaseURL+path, bytes.NewBuffer(bodyBytes))
    if err != nil {
        return nil, err
    }

    req.Header.Set("Authorization", fmt.Sprintf("HMAC-SHA256 %s", signature))
    req.Header.Set("X-Timestamp", strconv.FormatInt(timestamp, 10))
    req.Header.Set("Content-Type", "application/json")

    return c.HTTP.Do(req)
}

// Usage examples
func (c *Client) GetApplications() ([]Application, error) {
    resp, err := c.makeRequest("GET", "/api/apps", nil)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result struct {
        Data []Application `json:"data"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return result.Data, nil
}

func (c *Client) UpdateScrapeInterval(interval string) error {
    body := map[string]string{"interval": interval}
    resp, err := c.makeRequest("POST", "/api/scrape-interval", body)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("unexpected status: %d", resp.StatusCode)
    }

    return nil
}
```

### Python

```python
import hashlib
import hmac
import json
import time
import requests

class SiteAvailabilityClient:
    def __init__(self, base_url, secret):
        self.base_url = base_url
        self.secret = secret.encode('utf-8')

    def generate_signature(self, method, path, body, timestamp):
        canonical_request = f"{method}\n{path}\n{body or ''}\n{timestamp}"
        signature = hmac.new(
            self.secret,
            canonical_request.encode('utf-8'),
            hashlib.sha256
        ).hexdigest()
        return signature

    def make_request(self, method, path, body=None):
        timestamp = int(time.time())
        body_string = json.dumps(body) if body else ''
        signature = self.generate_signature(method, path, body_string, timestamp)

        headers = {
            'Authorization': f'HMAC-SHA256 {signature}',
            'X-Timestamp': str(timestamp),
            'Content-Type': 'application/json'
        }

        url = f"{self.base_url}{path}"
        response = requests.request(
            method,
            url,
            headers=headers,
            data=body_string if body_string else None
        )

        response.raise_for_status()
        return response.json()

    def get_applications(self):
        return self.make_request('GET', '/api/apps')

    def update_scrape_interval(self, interval):
        return self.make_request('POST', '/api/scrape-interval', {'interval': interval})

# Usage
client = SiteAvailabilityClient('http://localhost:8080', 'your-secret-key')
apps = client.get_applications()
```

### cURL

```bash
#!/bin/bash

SECRET="your-secret-key"
BASE_URL="http://localhost:8080"
METHOD="GET"
PATH="/api/apps"
BODY=""
TIMESTAMP=$(date +%s)

# Create canonical request
CANONICAL_REQUEST="${METHOD}\n${PATH}\n${BODY}\n${TIMESTAMP}"

# Generate signature
SIGNATURE=$(echo -ne "$CANONICAL_REQUEST" | openssl dgst -sha256 -hmac "$SECRET" -hex | cut -d' ' -f2)

# Make request
curl -X "$METHOD" \
  -H "Authorization: HMAC-SHA256 $SIGNATURE" \
  -H "X-Timestamp: $TIMESTAMP" \
  -H "Content-Type: application/json" \
  "$BASE_URL$PATH"
```

## Security Considerations

### Clock Skew Tolerance

The server accepts requests within a configurable time window to account for clock differences:

```yaml
authentication:
  hmac:
    clock_skew_tolerance: 300 # 5 minutes
```

### Secret Management

**Best Practices:**

- Use strong, randomly generated secrets (minimum 32 characters)
- Store secrets in environment variables or secure key management systems
- Rotate secrets regularly
- Never commit secrets to version control

**Secret Generation:**

```bash
# Generate a secure secret
openssl rand -hex 32
```

### Replay Attack Prevention

The timestamp in the signature helps prevent replay attacks:

- Each request must include a current timestamp
- Server rejects requests outside the tolerance window
- Consider implementing nonce tracking for additional security

## Authentication Errors

### Invalid Signature

```json
{
  "error": {
    "code": "INVALID_SIGNATURE",
    "message": "HMAC signature verification failed",
    "details": ["Check your secret key and signature generation"]
  }
}
```

### Timestamp Issues

```json
{
  "error": {
    "code": "TIMESTAMP_ERROR",
    "message": "Request timestamp outside acceptable range",
    "details": [
      "Current server time: 1638360300",
      "Request timestamp: 1638359000"
    ]
  }
}
```

### Missing Headers

```json
{
  "error": {
    "code": "MISSING_AUTH_HEADERS",
    "message": "Required authentication headers missing",
    "details": ["Authorization and X-Timestamp headers required"]
  }
}
```

## Testing Authentication

### Verification Script

```bash
#!/bin/bash
# test-auth.sh

SECRET="your-secret-key"
BASE_URL="http://localhost:8080"

echo "Testing authentication..."

# Test GET request
METHOD="GET"
PATH="/api/apps"
BODY=""
TIMESTAMP=$(date +%s)
CANONICAL_REQUEST="${METHOD}\n${PATH}\n${BODY}\n${TIMESTAMP}"
SIGNATURE=$(echo -ne "$CANONICAL_REQUEST" | openssl dgst -sha256 -hmac "$SECRET" -hex | cut -d' ' -f2)

echo "GET /api/apps"
curl -s -w "Status: %{http_code}\n" \
  -H "Authorization: HMAC-SHA256 $SIGNATURE" \
  -H "X-Timestamp: $TIMESTAMP" \
  "$BASE_URL$PATH"

# Test POST request
METHOD="POST"
PATH="/api/scrape-interval"
BODY='{"interval":"60s"}'
TIMESTAMP=$(date +%s)
CANONICAL_REQUEST="${METHOD}\n${PATH}\n${BODY}\n${TIMESTAMP}"
SIGNATURE=$(echo -ne "$CANONICAL_REQUEST" | openssl dgst -sha256 -hmac "$SECRET" -hex | cut -d' ' -f2)

echo -e "\nPOST /api/scrape-interval"
curl -s -w "Status: %{http_code}\n" \
  -X POST \
  -H "Authorization: HMAC-SHA256 $SIGNATURE" \
  -H "X-Timestamp: $TIMESTAMP" \
  -H "Content-Type: application/json" \
  -d "$BODY" \
  "$BASE_URL$PATH"
```

### Unit Tests

```go
func TestHMACAuthentication(t *testing.T) {
    secret := "test-secret"
    auth := NewHMACAuth(secret)

    // Test valid signature
    timestamp := time.Now().Unix()
    signature := generateTestSignature("GET", "/api/apps", "", timestamp, secret)

    req := httptest.NewRequest("GET", "/api/apps", nil)
    req.Header.Set("Authorization", fmt.Sprintf("HMAC-SHA256 %s", signature))
    req.Header.Set("X-Timestamp", strconv.FormatInt(timestamp, 10))

    valid, err := auth.ValidateRequest(req)
    assert.NoError(t, err)
    assert.True(t, valid)

    // Test invalid signature
    req.Header.Set("Authorization", "HMAC-SHA256 invalid-signature")
    valid, err = auth.ValidateRequest(req)
    assert.Error(t, err)
    assert.False(t, valid)
}
```

## Troubleshooting

### Common Issues

1. **Clock Synchronization**: Ensure client and server clocks are synchronized
2. **Encoding**: Use UTF-8 encoding for all string operations
3. **Newlines**: Use `\n` (LF) for canonical request formatting
4. **Empty Body**: Use empty string, not null, for requests without body
5. **URL Encoding**: Use the exact path from the URL, not URL-encoded version

### Debug Mode

Enable debug logging to troubleshoot authentication issues:

```yaml
authentication:
  hmac:
    debug: true # Only for development
```

This will log canonical requests and signatures (without revealing the secret).

Remember: HMAC authentication provides strong security when implemented correctly. Always use HTTPS in production to protect against man-in-the-middle attacks.
