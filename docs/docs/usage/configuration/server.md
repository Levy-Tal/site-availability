---
sidebar_position: 1
---

# Server Configuration

## Basic Settings

The server configuration includes basic networking and operational settings.

### Port Configuration

```yaml
server_settings:
  port: "8080"
```

### Custom CA Certificates

```yaml
server_settings:
  custom_ca_path: "/path/to/custom/ca.crt"
```

## Authentication

Site Availability Monitor supports authentication to secure access to the monitoring interface.

### Local Admin Authentication

Enable local admin authentication to require login before accessing the dashboard:

```yaml
server_settings:
  session_timeout: "12h" # Session duration (default: 12h)
  local_admin:
    enabled: true # Enable authentication
    username: "admin" # Admin username
```

**Credentials file** (`credentials.yaml`):

```yaml
server_settings:
  local_admin:
    password: "your-secure-password"
```

#### Security Features

- **Secure Sessions**: Cryptographically secure session IDs with configurable timeout
- **Password Security**: Supports both plaintext (development) and bcrypt hashed passwords
- **Protected Endpoints**: All `/api/*` endpoints require authentication when enabled
- **Session Cookies**: HttpOnly cookies with CSRF protection

#### Password Security Best Practices

For production environments, use bcrypt hashed passwords:

1. **Generate a bcrypt hash**:

   ```bash
   # Using htpasswd (if available)
   echo "your-password" | htpasswd -bnBC 12 "" | tr -d ':\n'

   # Or use online bcrypt generators with cost factor 12
   ```

2. **Use the hash in credentials.yaml**:
   ```yaml
   server_settings:
     local_admin:
       password: "$2b$12$rDKx8UXp3F8P7xYV9oGzTeBN6K8aHVWHZxXzGQQJ8E1QXh8l2F9Da"
   ```

## Authorization

### Role-Based Access Control

When authentication is enabled, you can configure role-based authorization to control which labels and apps users can access.

#### Admin Role

The local admin user automatically has the **admin** role with full access to all labels and apps.

#### Custom Roles

Define custom roles in the configuration to control label access:

```yaml
server_settings:
  roles:
    # Frontend team role - can see frontend apps
    frontend:
      team: "frontend"
      env: "production"

    # Backend team role - can see backend and shared apps
    backend:
      team: "backend"
      shared: "yes"

    # DevOps role - can see multiple environments
    devops:
    env: "production"
    region: "us-east"

    # QA role - can see staging environment
    qa:
      env: "staging"
```

#### How Authorization Works

1. **Label Filtering**: Users can only see labels they have permission for
2. **App Filtering**: Users can only see apps that have at least one label they're authorized for
3. **Location Filtering**: Users only see locations that contain authorized apps

#### API Behavior with Authorization

When a user makes API requests:

- **`/api/labels`**: Returns only label keys the user has access to
- **`/api/labels?team`**: Returns only values for the "team" label that the user can see
- **`/api/apps`**: Returns only apps with labels the user is authorized for
- **`/api/locations`**: Returns only locations containing authorized apps

#### Example Scenarios

**Frontend Team User** (role: `frontend`):

- Can see: `team=frontend` and `env=production` labels
- Apps visible: Only apps with `team=frontend` OR `env=production`
- Labels endpoint returns: `["team", "env"]`
- Apps endpoint returns: Apps matching their label permissions

**Admin User**:

- Can see: All labels and apps
- No filtering applied
- Full access to all monitoring data

#### Multi-Role Support

Users can have multiple roles. The system combines permissions from all roles:

```yaml
# If a user has both "frontend" and "qa" roles, they can see:
# - team=frontend (from frontend role)
# - env=production (from frontend role)
# - env=staging (from qa role)
```

## OIDC Authentication

Site Availability supports OpenID Connect (OIDC) authentication for integration with enterprise identity providers like Keycloak, Auth0, Azure AD, and others.

### Configuration

Configure OIDC in your `config.yaml`:

```yaml
server_settings:
  session_timeout: 12h

  # Define roles and their label permissions
  roles:
    frontend:
      team: frontend
      env: production
    backend:
      team: backend
    admin: {} # Empty means full access

  # OIDC Configuration
  oidc:
    enabled: true
    config:
      name: "Keycloak" # Display name for the provider
      issuer: "https://keycloak.example.com/realms/master"
      clientID: "site-availability"
      groupScope: "groups" # Claim name for user groups
      userNameScope: "preferred_username" # Claim name for username
    permissions:
      users:
        # Map specific users to roles
        john.doe:
          - frontend
          - backend
        admin.user:
          - admin
      groups:
        # Map OIDC groups to roles
        developers:
          - frontend
          - backend
        devops:
          - admin
        qa-team:
          - qa
```

Add the client secret in your `credentials.yaml`:

```yaml
server_settings:
  oidc:
    config:
      clientSecret: "your-oidc-client-secret"
```

### OIDC Provider Setup

#### Keycloak Example

1. Create a new client in Keycloak
2. Set client type to "OpenID Connect"
3. Configure redirect URI: `http://your-domain/auth/oidc/callback`
4. Enable "Client authentication"
5. Add groups to the client scope

#### Generic OIDC Provider

Ensure your OIDC provider includes:

- `groups` claim in ID tokens (or configure `groupScope`)
- `preferred_username` or similar claim (configure `userNameScope`)
- Standard OpenID Connect discovery endpoint

### User and Group Mapping

#### User-Based Mapping

Map individual users to roles:

```yaml
permissions:
  users:
    alice@example.com:
      - frontend
      - qa
    bob@example.com:
      - admin
```

#### Group-Based Mapping

Map OIDC groups to application roles:

```yaml
permissions:
  groups:
    frontend-team:
      - frontend
    backend-team:
      - backend
    devops-team:
      - admin
```

### Fallback Mechanism

If the OIDC provider is unavailable, you can configure local admin as a fallback:

```yaml
server_settings:
  # OIDC configuration (primary)
  oidc:
    enabled: true
    # ... oidc config ...

  # Local admin fallback
  local_admin:
    enabled: true
    username: admin
```

When OIDC provider is down:

- Users will see a warning message
- Local admin authentication remains available
- Server continues to operate normally

### OIDC Security Features

- **State Parameter**: CSRF protection using cryptographically secure state
- **Token Validation**: Full JWT signature verification
- **Timeout Handling**: 10-second timeout for provider initialization
- **Secure Cookies**: HttpOnly, Secure, and SameSite cookie attributes
- **Graceful Degradation**: Automatic fallback to local auth when provider unavailable

### Troubleshooting

#### Common Issues

1. **Provider Unreachable**: Check issuer URL and network connectivity
2. **Invalid Client**: Verify clientID and clientSecret
3. **Missing Claims**: Ensure groups and username scopes are configured
4. **Redirect Mismatch**: Verify redirect URI in provider matches application

#### Error Messages

- `OIDC provider unavailable`: Provider cannot be reached, local auth available
- `Invalid state parameter`: Possible CSRF attack or session timeout
- `Failed to extract username`: Username scope not configured properly

### Security Considerations

- **Principle of Least Privilege**: Only grant access to labels users need
- **Label Design**: Design your labeling strategy with authorization in mind
- **Admin Access**: Admin users bypass all authorization checks
- **Performance**: Authorization filtering is applied efficiently at the API level

#### Excluded Endpoints

The following endpoints are **not** protected by authentication:

- `/` - Login page and static files
- `/sync` - B2B endpoint (protected by HMAC authentication)
- `/healthz` - Health check
- `/readyz` - Readiness check
- `/metrics` - Metrics endpoint

## Sync Configuration

Configure server-to-server synchronization:

```yaml
server_settings:
  sync_enable: true
  token: "your-hmac-token"
```

## Labels

Add custom labels to identify this server instance:

```yaml
server_settings:
  labels:
    environment: "production"
    region: "us-east-1"
    cluster: "main"
```

## Complete Example

```yaml
server_settings:
  port: "8080"
  session_timeout: "12h"
  local_admin:
    enabled: true
    username: "admin"
  roles:
    frontend:
      team: "frontend"
      env: "production"
    backend:
      team: "backend"
      shared: "yes"
  sync_enable: true
  token: "secure-sync-token"
  labels:
    environment: "production"
    region: "us-east-1"
```
