# Site Availability Monitoring

**Site Availability Monitoring** is an open-source application designed to monitor the availability of applications and services across multiple locations. It provides a visual representation of application statuses on a world map and integrates with Prometheus for metrics collection.

## Features

- Real-time site status monitoring
- Interactive world map visualization
- Historical data tracking
- Alert notifications
- New feature: Automated status reports generation

## Development

This project uses semantic versioning for releases. All commits should follow the conventional commit format.

---

## 🚀 Features

- **Frontend**: A React-based web interface that displays application statuses on a world map.  
- **Backend**: A Go-based server that fetches application statuses from Prometheus and serves them via REST APIs.  
- **Prometheus Integration**: Collects metrics such as uptime, memory usage, and application availability.  
- **Helm Chart**: Deploys the application to Kubernetes with customizable configurations.  
- **Docker Support**: Easily build and run the application using Docker and Docker Compose.  
- **Grafana Dashboards**: Pre-configured dashboards for visualizing metrics.  

---

## 📁 Project Structure

```text
.github/       - GitHub Actions workflows for CI/CD  
backend/       - Go-based backend server  
config/        - Configuration management  
handlers/      - HTTP handlers  
logging/       - Logging utilities  
scraping/      - Prometheus scraping logic  
chart/         - Helm chart for Kubernetes deployment  
frontend/      - React-based frontend application  
release/       - Packaged releases  
```

---

## ✅ Prerequisites

- [Node.js](https://nodejs.org/) (for frontend development)  
- [Go](https://golang.org/) (for backend development)  
- [Docker](https://www.docker.com/) (for containerization)  
- [Helm](https://helm.sh/) (for Kubernetes deployment)  

---

## 🛠️ Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/your-org/site-availability-monitoring.git
cd site-availability-monitoring
```

### 2. Build and Run Locally

#### Backend

```bash
cd backend
go run main.go
```

#### Frontend

```bash
cd frontend
npm install
npm start
```

### 3. Run with Docker Compose

```bash
docker-compose up --build
```

### 4. Deploy to Kubernetes with Helm

```bash
helm upgrade site-availability chart/ -f chart/values.yaml
```

---

## ⚙️ Configuration

### Backend Configuration

The backend reads configuration from `config.yaml`.  
Example:

```yaml
scrape_interval: 10s
locations:
- name: New York City
  latitude: 40.712776
  longitude: -74.005974
- name: Los Angeles
  latitude: 34.052235
  longitude: -118.243683
- name: Chicago
  latitude: 41.878113
  longitude: -87.629799
apps:
- name: app1
  location: New York City
  metric: up{instance="app:8080", job="app"}
  prometheus: http://prometheus:9090/
- name: app3
  location: Los Angeles
  metric: up{instance="localhost:9090", job="prometheus"}
  prometheus: http://prometheus:9090/
- name: app4
  location: Chicago
  metric: up{instance="app:8080", job="app"}
  prometheus: http://prometheus:9090/
```

### Helm Chart

Customize the deployment by editing `chart/values.yaml`.

### Environment Variables

- `CONFIG_FILE`: Path to the configuration file (default: `config.yaml`)  
- `CUSTOM_CA_PATH`: Path to custom CA certificates  

---

## 🧪 Testing

### Backend Tests

```bash
cd backend
go test ./...
```

### Frontend Tests

```bash
cd frontend
npm test
```

---

## 📊 Metrics

The backend exposes Prometheus metrics at `/metrics`. Example metrics:

- `server_uptime_seconds`: Total uptime of the server  
- `go_memstats_alloc_bytes`: Memory allocation by the Go runtime  

---

## 📈 Grafana Dashboards

Pre-configured Grafana dashboards are available in:  
`chart/grafana-dashboards/`

---

## ⚙️ CI/CD

GitHub Actions workflows are defined in `.github/workflows/`:

- **Test Workflow**: Runs backend and frontend tests  
- **Docker Workflow**: Builds and pushes Docker images  
- **Helm Workflow**: Packages and deploys Helm charts  
- **Security Workflow**: Scans for vulnerabilities using Trivy  

---

## 📄 License

This project is licensed under the **Apache License 2.0**.  
See the [LICENSE](LICENSE) file for details.

---

## 🤝 Contributing

Contributions are welcome! Please open an issue or submit a pull request.

---

## 📬 Contact

For questions or support, please contact:  
[**your-email@example.com**](mailto:your-email@example.com)
