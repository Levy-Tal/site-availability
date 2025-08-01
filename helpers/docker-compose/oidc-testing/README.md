# OIDC Testing Environment

This Docker Compose setup provides a complete testing environment for Site Availability's OIDC authentication with Keycloak.

## Overview

The setup includes:

- **Keycloak**: Identity provider with pre-configured realm, client, users, and groups
- **PostgreSQL**: Database for Keycloak persistence
- **Site Availability**: The application configured to use OIDC authentication
- **Pre-configured test data**: Sample sources and locations for testing authorization

## Quick Start

1. **Build and start services**:

   ```bash
   # From the project root
   cd helpers/docker-compose/oidc-testing
   docker-compose up --build -d
   ```

2. **Wait for services to be ready** (about 2-3 minutes):

   ```bash
   # Check status
   docker-compose ps

   # Watch logs
   docker-compose logs -f
   ```

3. **Access the applications**:
   - **Site Availability**: http://localhost:8080
   - **Keycloak Admin**: http://localhost:8090 (admin/admin123)

## Test Users and Scenarios

### Pre-configured Test Users

| Username        | Password      | Groups     | Roles                  | Access                       |
| --------------- | ------------- | ---------- | ---------------------- | ---------------------------- |
| `admin`         | `admin123`    | devops     | devops, admin          | Full access (all sources)    |
| `frontend-user` | `frontend123` | developers | frontend, backend      | Frontend and backend sources |
| `backend-user`  | `backend123`  | developers | frontend, backend      | Frontend and backend sources |
| `qa-user`       | `qa123`       | qa-team    | qa-staging, qa-testing | Only staging sources         |
| `eu-user`       | `eu123`       | eu-team    | eu-west, eu-central    | Only EU sources              |

### Test Scenarios

#### Scenario 1: Admin Access

1. Go to http://localhost:8080
2. Click "Login with Keycloak SSO"
3. Login as `admin` / `admin123`
4. **Expected**: See all sources (6 total)
5. **Check**: All API endpoints return full data

#### Scenario 2: Developer Access

1. Login as `frontend-user` / `frontend123`
2. **Expected**: See only frontend and backend sources (4 total)
3. **Check**: `/api/labels` shows only `team` and `env` labels
4. **Check**: `/api/apps` filtered by team permissions

#### Scenario 3: QA Access

1. Login as `qa-user` / `qa123`
2. **Expected**: See only staging sources (1 total)
3. **Check**: Limited access to staging environment only

#### Scenario 4: Regional Access

1. Login as `eu-user` / `eu123`
2. **Expected**: See only EU sources (2 total)
3. **Check**: Regional filtering working correctly

#### Scenario 5: Fallback Authentication

1. Stop Keycloak: `docker-compose stop keycloak`
2. Refresh Site Availability page
3. **Expected**: Warning message about OIDC unavailable
4. **Check**: Local admin login still works (`admin` / `admin123`)

## Testing Authorization

### API Testing with curl

```bash
# Get session cookie first (after OIDC login)
COOKIE=$(curl -c - -b - -s http://localhost:8080/auth/user | grep session_id)

# Test filtered labels
curl -H "Cookie: $COOKIE" http://localhost:8080/api/labels

# Test filtered apps
curl -H "Cookie: $COOKIE" http://localhost:8080/api/apps

# Test filtered locations
curl -H "Cookie: $COOKIE" http://localhost:8080/api/locations
```

### Frontend Testing

1. **Login Flow**: Test both OIDC and local admin login options
2. **User Info**: Click user info button to see user details
3. **Data Filtering**: Verify sources shown match user permissions
4. **Navigation**: Test all UI features with different user roles

## Configuration Details

### Source Labels for Testing

The test configuration includes sources with different labels:

- **Team labels**: `frontend`, `backend`, `devops`
- **Environment labels**: `production`, `staging`
- **Region labels**: `us-east`, `eu-west`, `eu-central`

### Role Mappings

```yaml
# Group -> Role mapping in Site Availability
developers: [frontend, backend]
devops: [devops, admin]
qa-team: [qa-staging, qa-testing]
eu-team: [eu-west, eu-central]
```

### OIDC Configuration

```yaml
# Keycloak settings
issuer: "http://keycloak:8080/realms/site-availability"
clientID: "site-availability-client"
clientSecret: "site-availability-secret-123"
```

## Troubleshooting

### Services Won't Start

```bash
# Check service status
docker-compose ps

# View logs
docker-compose logs keycloak
docker-compose logs site-availability

# Restart services
docker-compose restart
```

### Keycloak Not Ready

```bash
# Wait for Keycloak health check
curl http://localhost:8090/health/ready

# Check realm import
docker-compose logs keycloak | grep "Imported realm"
```

### OIDC Login Issues

1. **Check redirect URI**: Ensure Keycloak client config matches
2. **Check network**: Services must communicate via Docker network
3. **Check logs**: Look for OIDC errors in site-availability logs

### Site Availability Won't Start

```bash
# Check configuration
docker-compose exec site-availability cat /app/config/config.yaml

# Check credentials
docker-compose exec site-availability cat /app/config/credentials.yaml

# Check connectivity to Keycloak
docker-compose exec site-availability curl http://keycloak:8080/health
```

## Customization

### Adding Test Users

1. Access Keycloak admin: http://localhost:8090
2. Go to Users → Add user
3. Set password in Credentials tab
4. Add to groups in Groups tab

### Modifying Roles

Edit `config/config.yaml`:

```yaml
roles:
  custom-role:
    label-key: label-value
```

### Adding Test Sources

Edit `config/config.yaml`:

```yaml
sources:
  - name: custom-service
    url: https://httpbin.org/status/200
    labels:
      custom-label: custom-value
```

## Cleanup

```bash
# Stop and remove containers
docker-compose down

# Remove volumes (data will be lost)
docker-compose down -v

# Remove images
docker-compose down --rmi all
```

## Development Notes

- Keycloak runs on port 8090 (to avoid conflict with Site Availability on 8080)
- Services use Docker internal networking for communication
- Volumes persist Keycloak and PostgreSQL data between restarts
- Configuration is mounted read-only to prevent accidental changes

## Security Notes

⚠️ **This is a testing environment only**:

- Uses default passwords and secrets
- HTTP only (no TLS)
- Permissive CORS settings
- Not suitable for production use

For production deployment, ensure:

- Strong, unique passwords and secrets
- HTTPS with valid certificates
- Proper firewall and network security
- Regular security updates
