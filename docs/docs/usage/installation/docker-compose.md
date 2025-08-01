---
sidebar_position: 1
---

# Production Deployment with Docker Compose

This guide explains how to deploy the Site Availability application in a production environment using Docker Compose, with Nginx as a secure reverse proxy for SSL termination and authentication.

## 1. Prepare Configuration Files

Before deploying, create the following minimal configuration files in your deployment directory:

### `config.yaml` (minimal example)

```yaml
server_settings:
  port: 8080
  host_url: "https://your-domain.com" # Required: Used for OIDC callback URLs

locations:
  - name: New York
    latitude: 40.712776
    longitude: -74.005974
  - name: San Francisco
    latitude: 37.774929
    longitude: -122.419418
  - name: Chicago
    latitude: 41.878113
    longitude: -87.629799

sources:
  - name: prom1
    type: prometheus
    config:
      url: http://nginx:9090
      apps:
        - name: app1
          location: New York
          metric: up{instance="app:8080", job="app"}
```

### `credentials.yaml` (minimal example)

```yaml
sources:
  - name: prom1
    config:
      auth: bearer
      token: "test-token-123"
```

## 2. Create SSL Certificates

For production, use certificates from a trusted CA (e.g., Let's Encrypt). Place your `fullchain.pem` and `privkey.pem` in a `certs/` directory.

## 3. Create `docker-compose.yml`

Below is a production-ready Docker Compose file with Nginx as a reverse proxy for SSL and optional authentication. It uses a single app container for Site Availability, a fixed image tag, and best practices for security and automatic restarts.

```yaml
services:
  app:
    image: levytal/site-availability:2.4.0
    restart: always
    environment:
      - CONFIG_FILE=/app/config.yaml
      - CREDENTIALS_FILE=/app/credentials.yaml
    volumes:
      - ./config.yaml:/app/config.yaml:ro
      - ./credentials.yaml:/app/credentials.yaml:ro
    ports:
      - "8080:8080"

  nginx:
    image: nginx:latest
    restart: always
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./certs:/etc/nginx/certs:ro
      # Uncomment if using HTTP Basic Auth
      # - ./htpasswd:/etc/nginx/.htpasswd:ro
    ports:
      - "80:80"
      - "443:443"
    depends_on:
      - app

networks:
  default:
    driver: bridge
```

## 4. Example `nginx.conf`

Configure Nginx to:

- Terminate SSL (HTTPS)
- Forward all requests to the single app container
- Optionally, enable HTTP Basic Auth for `/api` (but NOT for `/sync`)
- Ensure `/sync` is SSL-protected and proxied, but not basic-authenticated

```nginx
events {}

http {
  server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$host$request_uri;
  }

  server {
    listen 443 ssl;
    server_name your-domain.com;

    ssl_certificate /etc/nginx/certs/fullchain.pem;
    ssl_certificate_key /etc/nginx/certs/privkey.pem;

    # /sync is NOT protected by HTTP Basic Auth, but is SSL-terminated
    location /sync {
      proxy_pass http://app:8080/sync;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Uncomment to enable HTTP Basic Auth for /api (except /sync)
    # location ^~ /api/ {
    #   auth_basic "Protected";
    #   auth_basic_user_file /etc/nginx/.htpasswd;
    #   proxy_pass http://app:8080/api/;
    #   proxy_set_header Host $host;
    #   proxy_set_header X-Real-IP $remote_addr;
    #   proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    #   proxy_set_header X-Forwarded-Proto $scheme;
    # }

    location /api/ {
      proxy_pass http://app:8080/api/;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_set_header X-Forwarded-Proto $scheme;
    }

    location / {
      proxy_pass http://app:8080/;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_set_header X-Forwarded-Proto $scheme;
    }
  }
}
```

> **Tip:** To generate a `.htpasswd` file for HTTP Basic Auth, use:
> `docker run --rm httpd:alpine htpasswd -nbB admin strongpassword > certs/.htpasswd`

## 5. Best Practices

- **Use `restart: always`** to ensure containers restart after server reboots or failures.
- **Mount configuration and credentials as read-only** (`:ro`) for security.
- **Store SSL certificates securely** and restrict permissions.
- **Keep your images up to date** and use version tags for reproducibility.
- **Monitor logs and set up alerting** for failures.
- **Back up your configuration and credentials regularly.**

## 6. Start the Stack

```bash
docker compose up -d
```

Your Site Availability app will be available at `https://your-domain.com`.

---

For advanced configuration, scaling, or troubleshooting, see the rest of the documentation.
