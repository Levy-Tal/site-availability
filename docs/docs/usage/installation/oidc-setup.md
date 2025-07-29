# OIDC Authentication Setup

This guide walks you through setting up OpenID Connect (OIDC) authentication for Site Availability with popular identity providers.

## Overview

Site Availability supports OIDC authentication, which allows you to:

- Integrate with enterprise identity providers
- Implement role-based access control using existing groups
- Provide single sign-on (SSO) experience
- Maintain security with industry-standard protocols

## Prerequisites

- Identity provider that supports OpenID Connect
- Admin access to configure the identity provider
- Understanding of your organization's group structure
- SSL/TLS enabled for production deployments

## Step 1: Identity Provider Configuration

### Keycloak Setup

1. **Create a new client**:

   - Client ID: `site-availability`
   - Client Protocol: `openid-connect`
   - Access Type: `confidential`

2. **Configure redirect URIs**:

   ```
   http://localhost:8080/auth/oidc/callback  # Development
   https://your-domain.com/auth/oidc/callback  # Production
   ```

3. **Enable required scopes**:

   - Go to Client Scopes
   - Add `groups` scope to default client scopes
   - Ensure `profile` and `email` scopes are included

4. **Configure group mapper**:
   - Create a new mapper
   - Type: `Group Membership`
   - Token Claim Name: `groups`
   - Full group path: `false`

### Azure AD Setup

1. **Register application**:

   - Go to Azure AD > App registrations
   - New registration
   - Name: `Site Availability`
   - Redirect URI: `https://your-domain.com/auth/oidc/callback`

2. **Configure authentication**:

   - Platform: Web
   - Implicit grant tokens: ID tokens
   - Allow public client flows: No

3. **API permissions**:

   - Add `openid`, `profile`, `email`
   - Add `GroupMember.Read.All` (for group claims)

4. **Create client secret**:
   - Go to Certificates & secrets
   - New client secret
   - Copy the secret value

### Auth0 Setup

1. **Create application**:

   - Type: Regular Web Application
   - Name: Site Availability

2. **Configure settings**:

   - Allowed Callback URLs: `https://your-domain.com/auth/oidc/callback`
   - Allowed Web Origins: `https://your-domain.com`

3. **Enable groups claim**:
   - Go to Rules or Actions
   - Add groups to ID token:
   ```javascript
   function (user, context, callback) {
     context.idToken.groups = user.groups;
     callback(null, user, context);
   }
   ```

## Step 2: Application Configuration

### Configuration File

Create or update your `config.yaml`:

```yaml
server_settings:
  port: "8080"
  session_timeout: 12h

  # Define application roles
  roles:
    viewer:
      env: production
    developer:
      team: frontend
      team: backend
    ops:
      team: devops
    admin: {} # Full access

  # OIDC Configuration
  oidc:
    enabled: true
    config:
      name: "Your Company SSO"
      issuer: "https://your-provider.com/realms/master"  # Keycloak
      # issuer: "https://login.microsoftonline.com/{tenant}/v2.0"  # Azure AD
      # issuer: "https://your-domain.auth0.com/"  # Auth0
      clientID: "site-availability"
      groupScope: "groups"
      userNameScope: "preferred_username"  # or "email" for Azure AD
    permissions:
      users:
        # Direct user mappings
        admin@company.com:
          - admin
      groups:
        # Group-based mappings
        developers:
          - developer
        operations:
          - ops
        executives:
          - viewer
```

### Credentials File

Create `credentials.yaml`:

```yaml
server_settings:
  oidc:
    config:
      clientSecret: "your-client-secret-here"

  # Optional: Local admin fallback
  local_admin:
    password: "secure-fallback-password"
```

## Step 3: Role Design

Design your roles based on your labeling strategy:

### Example 1: Team-Based Access

```yaml
roles:
  frontend-team:
    team: frontend
  backend-team:
    team: backend
  devops-team:
    team: devops
  qa-team:
    env: staging
    env: testing
```

### Example 2: Environment-Based Access

```yaml
roles:
  prod-access:
    env: production
  staging-access:
    env: staging
  dev-access:
    env: development
```

### Example 3: Regional Access

```yaml
roles:
  us-team:
    region: us-east
    region: us-west
  eu-team:
    region: eu-west
    region: eu-central
```

## Step 4: Testing

### Test Configuration

1. **Start the application**:

   ```bash
   ./site-availability
   ```

2. **Check authentication status**:

   ```bash
   curl http://localhost:8080/auth/config
   ```

3. **Test OIDC login**:
   - Navigate to `http://localhost:8080`
   - Click "Login with Your Company SSO"
   - Complete authentication flow

### Verification

Verify that:

- ✅ Users can log in via OIDC
- ✅ Groups are correctly mapped to roles
- ✅ API responses are filtered based on user roles
- ✅ Fallback to local admin works when OIDC is down

## Troubleshooting

### Common Issues

#### 1. "OIDC provider unavailable"

**Symptoms**: Warning message on login page, OIDC button disabled

**Solutions**:

- Check issuer URL is accessible
- Verify network connectivity
- Check firewall rules
- Use local admin fallback

#### 2. "Invalid redirect URI"

**Symptoms**: Error after clicking OIDC login button

**Solutions**:

- Verify redirect URI in provider configuration
- Check for HTTP vs HTTPS mismatch
- Ensure trailing slashes match

#### 3. "Failed to extract username"

**Symptoms**: Login fails after successful OIDC authentication

**Solutions**:

- Check `userNameScope` configuration
- Verify claim is present in ID token
- Try alternative scopes: `email`, `name`, `sub`

#### 4. "No groups found"

**Symptoms**: User authenticated but has no permissions

**Solutions**:

- Verify `groupScope` configuration
- Check group claim in ID token
- Ensure group membership in identity provider
- Add user-specific role mapping

### Debugging

Enable detailed logging:

```yaml
# Add to config.yaml for debugging
logging:
  level: debug
```

Check logs for:

- OIDC provider initialization
- Token validation errors
- Group/role mapping results

### Security Checklist

- [ ] Use HTTPS in production
- [ ] Secure client secret storage
- [ ] Regular secret rotation
- [ ] Validate redirect URIs
- [ ] Monitor authentication logs
- [ ] Test fallback mechanisms
- [ ] Review role permissions regularly

## Migration from Local Auth

If migrating from local authentication:

1. **Keep local admin enabled during transition**:

   ```yaml
   server_settings:
     local_admin:
       enabled: true
     oidc:
       enabled: true
   ```

2. **Test OIDC with limited users first**

3. **Gradually migrate user permissions**

4. **Disable local admin once OIDC is stable**:
   ```yaml
   server_settings:
     local_admin:
       enabled: false # Disable after successful migration
   ```

## Production Considerations

### High Availability

- Configure multiple OIDC providers if supported
- Monitor provider health
- Keep local admin as emergency access
- Implement proper backup procedures

### Security

- Use short session timeouts for sensitive environments
- Implement proper HTTPS with valid certificates
- Regular security audits of role assignments
- Monitor for unauthorized access attempts

### Performance

- OIDC adds minimal overhead to authentication
- Authorization filtering is efficient
- Monitor session creation/cleanup
- Consider caching for large deployments
