---
sidebar_position: 1
---

# Development Setup

Set up your local development environment for contributing to Site Availability Monitoring.

## Prerequisites

### Required Software

- **Go** (v1.24): [Download](https://golang.org/dl/)
- **Node.js** (v18): [Download](https://nodejs.org/)
- **Git**: [Installation guide](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)

### Recommended Tools

- **Docker**: For containerized development
- **VS Code**: With Go and React extensions
- **Make**: For build automation
- **curl/httpie**: For API testing

## Repository Setup

### Clone the Repository

```bash
git clone https://github.com/Levy-Tal/site-availability.git
cd site-availability
```

### Install Dependencies

```bash
# Install all dependencies (frontend, backend, docs)
make install
```

### Environment Setup

```bash
# Set up Go workspace
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

# Install Go tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

## Backend Development

### Setup

```bash
cd backend

# Download dependencies
go mod download

# Verify setup
go version
go mod verify
```

### Configuration

Create a development config file:

```bash
cp ../helpers/config/single-server.yaml config.yaml
```

Edit `config.yaml` for your environment:

```yaml
scrape_interval: 10s
log_level: debug
port: 8080

locations:
  - name: Development
    latitude: 40.712776
    longitude: -74.005974

apps:
  - name: test-app
    location: Development
    metric: up{instance="localhost:9090", job="prometheus"}
    prometheus: http://localhost:9090/
```

### Running the Backend+Frontend

```bash
# Build and run the server locally
make build
```

This command will:

- Build the frontend
- Copy the frontend build to the backend static directory
- Start the backend server with the built frontend

### Testing

```bash
# Run all tests
make test


```

## Development Workflow

### Making Changes

1. **Create a feature branch**:

   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**:

   - Backend changes in `backend/`
   - Frontend changes in `frontend/`
   - Documentation in `docs/`

3. **Test your changes**:

   ```bash
   # Run all tests
   make test
   ```

4. **Run pre-commit hooks**:

   ```bash
   make pre-commit
   ```

5. **Commit changes**:
   ```bash
   git add .
   git commit -m "feat: add new feature"
   ```

### Using Docker for Development

Start the complete stack:

```bash
# Run using Docker Compose
make run

# Stop containers
make down
```

This provides:

- Backend at http://localhost:8080
- Frontend at http://localhost:3000
- Prometheus at http://localhost:9090

### API Testing

Test the backend API:

```bash
# Health check
curl http://localhost:8080/health

# Get applications
curl http://localhost:8080/api/apps | jq

# Get locations
curl http://localhost:8080/api/locations | jq

# Metrics endpoint
curl http://localhost:8080/metrics
```

### Pre-commit Hooks

Install pre-commit hooks:

```bash
# Run pre-commit hooks
make pre-commit
```

## Debugging

### Backend Debugging

```bash
# Run with debug logging
LOG_LEVEL=debug make build

```

## Documentation

```bash
# Run docs website locally
make docs
```

## Common Issues

### Go Module Issues

```bash
# Clear module cache
go clean -modcache

# Update dependencies
go mod tidy
go mod download
```

### Node.js Issues

```bash
# Clear npm cache
npm cache clean --force

# Delete node_modules and reinstall
rm -rf node_modules package-lock.json
npm install
```

### Port Conflicts

```bash
# Check what's using port 8080
lsof -i :8080

# Kill process if needed
kill -9 <PID>
```

## Next Steps

- Read the [Architecture Guide](./architecture)
- Learn about [Contributing Guidelines](./contributing)
- Explore [Testing Strategies](./testing)
- Check out [API Documentation](../api/overview)
