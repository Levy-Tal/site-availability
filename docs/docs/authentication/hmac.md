# HMAC Authentication

Site Availability uses HMAC (Hash-based Message Authentication Code) authentication to secure the `/sync` endpoint, which enables federation between multiple Site Availability instances.

## Overview

HMAC authentication ensures that:

- Requests to the `/sync` endpoint are authenticated using a shared secret token
- Messages cannot be tampered with during transmission
- Replay attacks are prevented using timestamp validation
- Only authorized instances can access federation endpoints

## How HMAC Works

HMAC authentication uses SHA-256 and includes:

1. **Shared Token**: A secret token configured on both the server (providing the endpoint) and client (accessing the endpoint)
2. **Timestamp**: RFC3339 formatted timestamp included in each request
3. **Signature**: HMAC-SHA256 signature calculated from the timestamp and request body
4. **Headers**: Authentication information transmitted via HTTP headers

### Signature Generation

The HMAC signature is generated using:

```
HMAC-SHA256(token, timestamp + request_body)
```

Where:

- `token` is the shared secret
- `timestamp` is the RFC3339 formatted current time
- `request_body` is the HTTP request body (empty for GET requests)

## Configuration

### Server Configuration (Providing /sync endpoint)

Enable the sync endpoint and configure a token in your `config.yaml`:

```yaml
server_settings:
  sync_enable: true
  token: "your-strong-secret-token"
```

### Client Configuration (Site Source)

Configure the site source to authenticate with the remote server:

```yaml
sources:
  - name: "remote-site"
    type: "site"
    config:
      url: "https://remote-site.example.com"
      token: "your-strong-secret-token" # Same token as server
```

## HTTP Headers

HMAC authentication uses two custom headers:

- `X-Site-Sync-Timestamp`: RFC3339 formatted timestamp
- `X-Site-Sync-Signature`: Hex-encoded HMAC-SHA256 signature

## Request Flow

1. **Client** generates RFC3339 timestamp
2. **Client** calculates HMAC signature using token, timestamp, and request body
3. **Client** sends request with timestamp and signature headers
4. **Server** validates the signature matches expected value
5. **Server** validates timestamp is within 5-minute window (prevents replay attacks)
6. **Server** processes request if validation passes

## Example Request

```http
GET /sync HTTP/1.1
Host: remote-site.example.com
X-Site-Sync-Timestamp: 2024-01-15T10:30:00Z
X-Site-Sync-Signature: a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456
```

## Security Considerations

### Token Management

- **Use strong tokens**: Generate cryptographically secure random tokens (recommended: 32+ characters)
- **Store securely**: Keep tokens in secure configuration files with restricted permissions
- **Rotate regularly**: Change tokens periodically and update all instances
- **Don't log tokens**: Ensure tokens are not logged in application logs

### Network Security

- **Use HTTPS**: Always use HTTPS for production deployments
- **Timestamp validation**: 5-minute window prevents replay attacks
- **Rate limiting**: Consider implementing rate limiting on the `/sync` endpoint

### Example Token Generation

```bash
# Generate a secure token
openssl rand -hex 32
```

## Troubleshooting

### Common Issues

1. **401 Unauthorized**

   - Check that tokens match exactly on both server and client
   - Verify timestamp is within 5-minute window
   - Ensure HTTPS is used if required

2. **Clock Skew**

   - Synchronize clocks on all instances using NTP
   - The system allows up to 5 minutes of clock difference

3. **Invalid Signature**
   - Verify token configuration
   - Check for any modifications to request headers or body
   - Ensure proper URL encoding

### Debug Logging

Enable debug logging to troubleshoot HMAC issues:

```yaml
logging:
  level: debug
```

Look for log entries containing:

- "Generated HMAC signature for site sync request"
- "HMAC validation failed"
- "Timestamp validation failed"

## Implementation Details

The HMAC authentication is implemented in:

- `backend/authentication/hmac/hmac.go` - Core HMAC validation logic
- `backend/handlers/handlers.go` - `/sync` endpoint protection
- `backend/scraping/site/site.go` - Client-side signature generation

### Key Functions

- `ValidateRequest()` - Validates both HMAC signature and timestamp
- `GenerateSignature()` - Creates HMAC signature for outbound requests
- `ValidateTimestamp()` - Ensures request is within time window

## See Also

- [Site Source Configuration](../usage/configuration/sources/site.md)
- [Server Configuration](../usage/configuration/server.md)
- [API Endpoints](../api/endpoints.md)
