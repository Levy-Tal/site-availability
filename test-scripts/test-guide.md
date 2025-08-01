# Site Availability Authentication & Authorization Test Guide

This guide walks you through testing all authentication and authorization features we've implemented.

## Test Setup

### Prerequisites

- `curl` command available
- `jq` (optional, for JSON formatting)
- Terminal access

### Test Configurations

We've created several test configurations:

1. **`test-configs/test-no-auth.yaml`** - Authentication disabled
2. **`test-configs/test-with-auth.yaml`** - Authentication enabled with roles
3. **`test-configs/test-credentials.yaml`** - Test credentials file

## Test Scenarios

### Scenario 1: Authentication Disabled

Test that the app works normally when authentication is disabled.

```bash
# Start server with no authentication
cd backend
CONFIG_FILE=../test-configs/test-no-auth.yaml ./site-availability

# Test API endpoints (should work without authentication)
curl http://localhost:8080/auth/config
curl http://localhost:8080/api/labels
curl http://localhost:8080/api/apps
curl http://localhost:8080/api/locations
```

**Expected Results:**

- `/auth/config` returns `{"auth_enabled": false}`
- All API endpoints return data without requiring authentication
- Frontend should work normally without login page

### Scenario 2: Authentication Enabled (Admin User)

Test authentication with admin user having full access.

```bash
# Start server with authentication
cd backend
CONFIG_FILE=../test-configs/test-with-auth.yaml CREDENTIALS_FILE=../test-configs/test-credentials.yaml ./site-availability

# Run the comprehensive test script
cd ../test-scripts
chmod +x test-api.sh
./test-api.sh
```

**Expected Results:**

- `/auth/config` returns `{"auth_enabled": true, "auth_methods": ["local"]}`
- API endpoints require authentication (return 401 without login)
- Login with `admin/testpass123` succeeds
- Admin user sees all labels, apps, and locations
- Logout invalidates session

### Scenario 3: Frontend Authentication Flow

Test the complete frontend authentication experience.

```bash
# Build and serve frontend
cd frontend
npm run build
npx serve -s build -p 3000

# Open browser to http://localhost:3000
```

**Expected Results:**

- **Auth Disabled**: Direct access to dashboard
- **Auth Enabled**: Login page appears
- Login with `admin/testpass123` grants access
- User info modal shows correct information
- Logout returns to login page

## Manual API Testing

### Authentication Endpoints

```bash
# Check auth configuration
curl http://localhost:8080/auth/config

# Login (get session cookie)
curl -c cookies.txt -X POST \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"testpass123"}' \
  http://localhost:8080/auth/login

# Get user info (using session cookie)
curl -b cookies.txt http://localhost:8080/auth/user

# Logout
curl -b cookies.txt -X POST http://localhost:8080/auth/logout
```

### Authorization Testing

```bash
# Login first
curl -c cookies.txt -X POST \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"testpass123"}' \
  http://localhost:8080/auth/login

# Test label access
curl -b cookies.txt http://localhost:8080/api/labels
curl -b cookies.txt http://localhost:8080/api/labels?team
curl -b cookies.txt http://localhost:8080/api/labels?env

# Test app filtering
curl -b cookies.txt http://localhost:8080/api/apps
curl -b cookies.txt "http://localhost:8080/api/apps?labels.team=frontend"
curl -b cookies.txt "http://localhost:8080/api/apps?labels.env=production"

# Test location filtering
curl -b cookies.txt http://localhost:8080/api/locations
curl -b cookies.txt "http://localhost:8080/api/locations?labels.env=production"
```

## Expected Test Results

### Admin User (Full Access)

When logged in as admin, you should see:

**Labels endpoint:**

```json
{
  "labels": ["team", "env", "region", "shared"]
}
```

**Apps endpoint:**

```json
{
  "apps": [
    {
      "name": "frontend-app-prod",
      "labels": {
        "team": "frontend",
        "env": "production",
        "region": "us-east"
      },
      "status": "up"
    }
    // ... all other apps
  ]
}
```

### Role-Based Access (Future OIDC)

Once OIDC is implemented, users with specific roles should see filtered results:

**Frontend Role User:**

- Labels: `["team", "env"]` (only labels they have access to)
- Apps: Only apps with `team=frontend` OR `env=production`

**Backend Role User:**

- Labels: `["team", "shared"]`
- Apps: Only apps with `team=backend` OR `shared=yes`

## Troubleshooting

### Common Issues

1. **401 Unauthorized on API calls**

   - Ensure you're using session cookies (`-b cookies.txt`)
   - Check that login was successful
   - Verify credentials are correct

2. **Authentication config shows disabled**

   - Check configuration file has `local_admin.enabled: true`
   - Verify correct config file is being loaded

3. **Login fails with valid credentials**

   - Check credentials file is loaded correctly
   - Verify password in credentials.yaml matches test password

4. **Empty response from APIs**
   - Wait a few seconds for scraping to populate data
   - Check server logs for scraping errors
   - Verify test URLs (httpbin.org) are accessible

### Debug Tips

1. **Check server logs** for authentication/authorization debug messages
2. **Use browser developer tools** to inspect network requests
3. **Test with curl -v** for verbose HTTP debugging
4. **Check session cookies** are being set and sent correctly

## Performance Testing

### Load Testing Authentication

```bash
# Test multiple concurrent logins
for i in {1..10}; do
  curl -s -X POST \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"testpass123"}' \
    http://localhost:8080/auth/login &
done
wait
```

### Memory Usage

```bash
# Monitor memory usage during authentication tests
top -p $(pgrep site-availability)

# Check session count (if exposed in metrics)
curl http://localhost:8080/metrics | grep session
```

## Test Checklist

### Authentication Features

- [ ] Authentication can be disabled/enabled via config
- [ ] Login with valid credentials succeeds
- [ ] Login with invalid credentials fails (401)
- [ ] Protected endpoints require authentication
- [ ] User info endpoint returns correct data
- [ ] Logout invalidates session
- [ ] Session timeout works correctly

### Authorization Features

- [ ] Admin users have full access to all data
- [ ] API responses are filtered based on user permissions
- [ ] Labels endpoint shows only accessible labels
- [ ] Apps endpoint filters by user's label permissions
- [ ] Locations endpoint filters by accessible apps
- [ ] Multiple roles combine permissions correctly

### Frontend Integration

- [ ] Authentication state detection works
- [ ] Login page appears when auth is enabled
- [ ] User modal shows correct information
- [ ] Logout functionality works
- [ ] Protected routes redirect to login

### Security Features

- [ ] Passwords are never exposed in responses
- [ ] Session cookies are HttpOnly and secure
- [ ] CSRF protection is enabled
- [ ] No sensitive data in logs
- [ ] Rate limiting prevents brute force attacks

---

## Next Steps

After completing these tests:

1. **Phase 3**: Implement OIDC integration for enterprise authentication
2. **Advanced Features**: Add more granular permissions, audit logging
3. **Production Hardening**: SSL/TLS, security headers, monitoring
