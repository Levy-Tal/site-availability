---
sidebar_position: 3
---

# Site Source Configuration

The **Site** source lets you monitor another Site Availability instance. The server will scrape all app statuses and locations from the remote instance by making a request to its `/sync` endpoint.

- The `/sync` endpoint is protected by HMAC authentication. The server uses the `token` from the `server_settings` of the scraped site. (and you put it in the source config) to generate the HMAC signature. See the [HMAC authentication documentation](../../../authentication/hmac.md) for details.
- There is built-in protection against circular scraping: you can safely scrape a site that scrapes from you.

## How It Works

- The site source makes a GET request to `http://other-site:8080/sync`.
- The request includes a timestamp and HMAC signature in the headers.
- The remote site returns all its app statuses and locations, which are merged into your monitoring view.

## Minimal Example

```yaml
sources:
  - name: "Server A"
    type: site
    config:
      url: http://server-a:8080
```

## Authentication Example (credentials.yaml)

```yaml
server_settings:
  token: "server-b-token"

sources:
  - name: Server A
    config:
      token: "server-a-token"
```

- The `token` under `server_settings` is used for authenticating incoming /sync requests.
- The `token` under the source config (in credentials.yaml) is used for authenticating outgoing /sync requests to the remote site.

## Source Configuration Options

- **name**: Unique name for the source (required)
- **type**: Must be `site` (required)
- **config.url**: Base URL of the remote Site Availability instance (required)
- **config.token**: HMAC token for authenticating to the remote site.
- **labels**: Optional labels for this source

## Important Notes

- The site source will import all apps and locations from the remote instance.
- If the remote site is unavailable or returns an error, no apps will be imported from that source.
- You can safely chain or loop site sources; the system prevents infinite recursion.

## Best Practices

- Use strong, unique tokens for HMAC authentication.
- Store tokens in `credentials.yaml` for security.
- Enable `sync_enable: true` in `server_settings` to allow your instance to be scraped by others.

---

For more details, see the [HMAC authentication documentation](../../../authentication/hmac.md) or the codebase.
