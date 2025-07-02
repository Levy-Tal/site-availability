---
sidebar_position: 1
---

# Development Setup

Set up your local development environment for contributing to Site Availability Monitoring.

## Prerequisites

### Required Software

- **Go** (v1.19+): [Download](https://golang.org/dl/)
- **Node.js** (v18.0+): [Download](https://nodejs.org/)
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

### Running the Backend

```bash
# Development mode with hot reload
go run main.go

# Or build and run
go build -o site-availability main.go
./site-availability
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./handlers/...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Frontend Development

### Setup

```bash
cd frontend

# Install dependencies
npm install

# Verify setup
npm --version
node --version
```

### Configuration

Edit `src/config.js` for development:

```javascript
const config = {
  apiUrl: "http://localhost:8080",
  map: {
    updateInterval: 10000, // 10 seconds for development
    defaultZoom: 3,
  },
  features: {
    debugMode: true,
  },
};
```

### Running the Frontend

```bash
# Start development server
npm start

# The app will open at http://localhost:3000
```

### Testing

```bash
# Run tests
npm test

# Run tests with coverage
npm test -- --coverage

# Run tests in watch mode
npm test -- --watch
```

### Building

```bash
# Create production build
npm run build

# Serve build locally
npx serve -s build
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
   # Backend tests
   cd backend && go test ./...

   # Frontend tests
   cd frontend && npm test
   ```

4. **Commit changes**:
   ```bash
   git add .
   git commit -m "feat: add new feature"
   ```

### Using Docker for Development

Start the complete stack:

```bash
cd helpers/docker-compose/single-server
docker-compose up -d
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

## Code Quality

### Linting

```bash
# Backend linting
cd backend
golangci-lint run

# Frontend linting
cd frontend
npm run lint
```

### Formatting

```bash
# Format Go code
gofmt -w .
goimports -w .

# Format JavaScript/React code
cd frontend
npm run format
```

### Pre-commit Hooks

Install pre-commit hooks:

```bash
# Install pre-commit
pip install pre-commit

# Install hooks
pre-commit install

# Run manually
pre-commit run --all-files
```

## IDE Setup

### VS Code

Recommended extensions:

- Go (by Google)
- ES7+ React/Redux/React-Native snippets
- Prettier
- ESLint
- REST Client

Create `.vscode/settings.json`:

```json
{
  "go.buildOnSave": "package",
  "go.lintOnSave": "package",
  "go.testOnSave": "package",
  "editor.formatOnSave": true,
  "eslint.autoFixOnSave": true
}
```

### GoLand/WebStorm

Configure Go modules and Node.js interpreter in IDE settings.

## Debugging

### Backend Debugging

```bash
# Run with debug logging
LOG_LEVEL=debug go run main.go

# Use delve debugger
go install github.com/go-delve/delve/cmd/dlv@latest
dlv debug main.go
```

### Frontend Debugging

```bash
# Enable debug mode
REACT_APP_DEBUG=true npm start

# Use React Developer Tools browser extension
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
