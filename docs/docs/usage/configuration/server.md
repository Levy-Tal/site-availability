---
sidebar_position: 1
---

# Server Configuration

Site Availability loads its configuration by merging two files: `config.yaml` and `credentials.yaml`. All configuration options can be placed in either file, and the server will merge them at startup. This allows you to keep sensitive data (like tokens or secrets) in `credentials.yaml` and the rest in `config.yaml`.

## Server Settings (`server_settings`)

The `server_settings` section controls core server behavior:

```yaml
server_settings:
  port: 8080 # Port for the HTTP server (string or int)
  sync_enable: false # Enable/disable /sync endpoint (bool)
  custom_ca_path: "" # Path to custom CA certificates (string, optional)
  token: "" # Optional global API token for authentication (string)
  labels: # Optional key-value labels for this server (map)
    env: "production"
    region: "us-east"
```

**Options:**

- **port**: The port the HTTP server listens on (default: 8080).
- **sync_enable**: Enables the `/sync` endpoint for distributed sync (default: false).
- **custom_ca_path**: Path to a custom CA bundle for outbound HTTPS requests (optional).
- **token**: Optional global API token for authenticating requests (optional).
- **labels**: Optional map of key-value labels applied to all apps on this server (used for filtering, grouping, or authorization).

## Locations

The `locations` section defines the monitored geographic locations:

```yaml
locations:
  - name: "New York"
    latitude: 40.712776
    longitude: -74.005974
  - name: "San Francisco"
    latitude: 37.774929
    longitude: -122.419418
```

**Each location must include:**

- **name**: Unique name for the location.
- **latitude**: Latitude in decimal degrees (between -90 and 90).
- **longitude**: Longitude in decimal degrees (between -180 and 180).

At least one location is required. All monitored apps must reference a valid location by name.

## Configuration Merging

- The server merges `config.yaml` and `credentials.yaml` into a single configuration at startup.
- Any option can be placed in either file; values in `credentials.yaml` override those in `config.yaml` if both are present.
- This allows you to keep secrets and sensitive data out of your main config.

## Example Minimal Configuration

```yaml
server_settings:
  port: 8080
  sync_enable: false

locations:
  - name: "New York"
    latitude: 40.712776
    longitude: -74.005974
```

For more details on other configuration sections (sources, etc.), see the related documentation pages.
