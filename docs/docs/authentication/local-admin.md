# Local Admin Authentication

This guide explains how to set up and use the local admin authentication method for Site Availability. Local admin provides a simple username/password authentication system that's useful for small deployments, testing, or as a fallback when OIDC is unavailable.

## Overview

Local admin authentication allows you to:

- Set up simple username/password authentication
- Provide emergency access when OIDC is down
- Quick setup for development and testing
- Small deployments without enterprise identity providers

## Configuration

### Basic Setup

Add the following to your `config.yaml`:

```yaml
server_settings:
  port: "8080"
  host_url: "https://your-domain.com"
  session_timeout: 12h # Optional: defaults to system default

  # Local admin configuration
  local_admin:
    enabled: true
    username: "admin" # Your chosen admin username
```

### Credentials Setup

**Important**: Never put passwords in your main config file. Instead, create a separate `credentials.yaml`:

```yaml
server_settings:
  local_admin:
    password: "your-secure-password-here"
```

### Complete Example

**config.yaml**:

```yaml
server_settings:
  port: "8080"
  host_url: "https://your-domain.com"
  session_timeout: 12h

  # Define application roles (optional, for role-based access)
  roles:
    viewer:
      env: production
    developer:
      team: frontend
      team: backend
    admin: {} # Full access

  # Local admin setup
  local_admin:
    enabled: true
    username: "admin"
```

**credentials.yaml**:

```yaml
server_settings:
  local_admin:
    password: "MySecurePassword123!"
```

## Security Considerations

### Password Requirements

Choose a strong password that includes:

- At least 12 characters
- Mix of uppercase and lowercase letters
- Numbers and special characters
- Avoid common dictionary words

### File Permissions

Secure your credentials file:

```bash
# Set restrictive permissions on credentials file
chmod 600 credentials.yaml

# Ensure only the application user can read it
chown site-availability:site-availability credentials.yaml
```

### Production Security

For production deployments:

1. **Use strong, unique passwords**
2. **Rotate passwords regularly**
3. **Monitor authentication logs**
4. **Consider using local admin only as fallback**
5. **Enable session timeouts**

## Session Management

### Session Timeout

Configure session timeout to automatically log out inactive users:

```yaml
server_settings:
  session_timeout: "8h" # 8 hours
  # session_timeout: "30m"  # 30 minutes
  # session_timeout: "1h30m" # 1 hour 30 minutes
```

Valid formats:

- `1h` - 1 hour
- `30m` - 30 minutes
- `90s` - 90 seconds
- `1h30m` - 1 hour 30 minutes
- `12h` - 12 hours

### Proxy Headers

If running behind a reverse proxy, enable trust for proxy headers:

```yaml
server_settings:
  trust_proxy_headers: true
```

## Usage

### Logging In

1. Navigate to your Site Availability instance
2. If local admin is the only authentication method, you'll see a login form
3. If both local admin and OIDC are enabled, click "Local Admin Login"
4. Enter your configured username and password

### Admin Privileges

Local admin users have full access to:

- All monitoring data regardless of labels
- Administrative functions
- Configuration viewing (where applicable)
- All API endpoints

## Role-Based Access with Local Admin

While local admin provides full access by default, you can combine it with role-based access control:

```yaml
server_settings:
  roles:
    prod-viewer:
      env: production
    dev-access:
      env: development
      env: staging
    admin: {} # Full access

  local_admin:
    enabled: true
    username: "admin"
```

The local admin user will automatically have admin privileges, giving them access to all data.

## Combining with OIDC

Local admin works well as a fallback to OIDC authentication:

```yaml
server_settings:
  # Primary authentication via OIDC
  oidc:
    enabled: true
    config:
      issuer: "https://your-provider.com"
      clientID: "site-availability"
      # ... other OIDC config

  # Fallback authentication
  local_admin:
    enabled: true
    username: "emergency-admin"
```

With this setup:

- Users normally log in via OIDC
- If OIDC is down, administrators can use local admin
- The local admin account provides emergency access

## Testing

### Verify Configuration

1. **Start the application**:

   ```bash
   ./site-availability
   ```

2. **Check authentication config**:

   ```bash
   curl http://localhost:8080/auth/config
   ```

   Expected response should include:

   ```json
   {
     "localAdminEnabled": true,
     "oidcEnabled": false
   }
   ```

3. **Test login**:
   - Navigate to `http://localhost:8080`
   - Enter your configured username and password
   - Verify you can access all monitoring data

### Verification Checklist

- ✅ Can log in with configured credentials
- ✅ Invalid credentials are rejected
- ✅ Session persists across page refreshes
- ✅ Session expires after configured timeout
- ✅ Can access all monitoring data (admin privileges)

## Troubleshooting

### Common Issues

#### 1. "Invalid username or password"

**Symptoms**: Login fails with correct credentials

**Solutions**:

- Verify username in `config.yaml` matches what you're entering
- Check password in `credentials.yaml` (no extra spaces/characters)
- Ensure files are properly formatted YAML
- Check file permissions allow application to read files

#### 2. "Local admin not enabled"

**Symptoms**: No local admin login option available

**Solutions**:

- Verify `local_admin.enabled: true` in config
- Check configuration file is being loaded correctly
- Restart application after configuration changes

#### 3. Session expires immediately

**Symptoms**: Logged out right after login

**Solutions**:

- Check `session_timeout` configuration
- Verify system clock is correct
- Check for conflicting proxy settings

#### 4. Cannot access credentials file

**Symptoms**: Application fails to start with file access errors

**Solutions**:

- Check file permissions: `ls -la credentials.yaml`
- Verify file ownership matches application user
- Ensure file exists in expected location

### Debugging

Enable debug logging to troubleshoot authentication issues:

```yaml
# Add to config.yaml temporarily
logging:
  level: debug
```

Look for log entries related to:

- Configuration file loading
- Authentication attempts
- Session creation/validation

## Migration and Maintenance

### Changing Passwords

1. **Update credentials.yaml** with new password
2. **Restart the application** to load new credentials
3. **Test login** with new password
4. **Inform other administrators** of the change

### Disabling Local Admin

To disable local admin (e.g., after OIDC setup is complete):

```yaml
server_settings:
  local_admin:
    enabled: false
```

### Backup and Recovery

**Important**: Always backup your credentials file securely:

```bash
# Create encrypted backup
gpg -c credentials.yaml

# Store backup in secure location
cp credentials.yaml.gpg /secure/backup/location/
```

## Best Practices

1. **Use local admin as fallback only** - Prefer OIDC for primary authentication
2. **Strong passwords** - Use password managers to generate secure passwords
3. **Regular rotation** - Change passwords periodically
4. **Monitor access** - Review authentication logs regularly
5. **Secure storage** - Protect credentials file with proper permissions
6. **Document access** - Keep track of who has admin credentials
7. **Emergency procedures** - Document password reset procedures
