---
sidebar_position: 1
---

# Installation

This guide will walk you through installing Site Availability Monitoring on your system.

## Prerequisites

Before getting started, ensure you have the following installed on your system:

### Required Dependencies

- **Node.js** (v18.0 or above) - [Download here](https://nodejs.org/en/download/)
- **Go** (v1.19 or above) - [Download here](https://golang.org/dl/)
- **Git** - [Installation guide](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)

### Optional Dependencies

- **Docker** - [Installation guide](https://docs.docker.com/get-docker/)
- **Docker Compose** - [Installation guide](https://docs.docker.com/compose/install/)
- **Helm** (for Kubernetes) - [Installation guide](https://helm.sh/docs/intro/install/)
- **kubectl** (for Kubernetes) - [Installation guide](https://kubernetes.io/docs/tasks/tools/)

## Clone the Repository

First, clone the Site Availability Monitoring repository:

```bash
git clone https://github.com/Levy-Tal/site-availability.git
cd site-availability
```

## Backend Setup

### 1. Navigate to Backend Directory

```bash
cd backend
```

### 2. Download Dependencies

```bash
go mod download
```

### 3. Build the Application

```bash
go build -o site-availability main.go
```

### 4. Create Configuration File

Copy the example configuration and customize it for your environment:

```bash
cp ../helpers/config/single-server.yaml config.yaml
```

### 5. Run the Backend

```bash
./site-availability
```

The backend will start on `http://localhost:8080` by default.

## Frontend Setup

### 1. Navigate to Frontend Directory

```bash
cd ../frontend
```

### 2. Install Dependencies

```bash
npm install
```

### 3. Configure Frontend

Edit `src/config.js` to point to your backend:

```javascript
const config = {
  apiUrl: "http://localhost:8080",
  // ... other configuration
};
```

### 4. Start Development Server

```bash
npm start
```

The frontend will be available at `http://localhost:3000`.

## Verification

To verify your installation:

1. **Backend Health Check**: Visit `http://localhost:8080/health`
2. **Frontend**: Open `http://localhost:3000` in your browser
3. **Metrics Endpoint**: Check `http://localhost:8080/metrics` for Prometheus metrics

## Next Steps

- 📚 Continue with the [Quick Start Guide](./quick-start) for a complete setup
- 🐳 Try the [Docker Setup](./docker) for containerized deployment
- ⚙️ Learn about [Configuration](../configuration/overview) options

## Troubleshooting

### Common Issues

#### Port Already in Use

If you get a "port already in use" error:

```bash
# Check what's using the port
lsof -i :8080
# Kill the process if needed
kill -9 <PID>
```

#### Go Module Issues

If you encounter Go module issues:

```bash
go clean -modcache
go mod download
```

#### Node.js Issues

If npm install fails:

```bash
rm -rf node_modules package-lock.json
npm install
```

For more troubleshooting, see our [Troubleshooting Guide](../troubleshooting).
