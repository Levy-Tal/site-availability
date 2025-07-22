---
sidebar_position: 3
---

# Frontend Configuration

Configure the React frontend application for different environments and use cases.

## Configuration File

The main configuration is in `src/config.js`:

```javascript
const config = {
  // API Configuration
  apiUrl: process.env.REACT_APP_API_URL || "http://localhost:8080",

  // Update Settings
  updateInterval: parseInt(process.env.REACT_APP_UPDATE_INTERVAL) || 30000,

  // Map Configuration
  map: {
    defaultZoom: 2,
    maxZoom: 10,
    minZoom: 1,
    projection: "geoNaturalEarth1",
  },

  // Theme Settings
  theme: {
    darkMode: process.env.REACT_APP_DARK_MODE === "true",
    primaryColor: process.env.REACT_APP_PRIMARY_COLOR || "#1976d2",
  },

  // Feature Flags
  features: {
    realTimeUpdates: process.env.REACT_APP_REAL_TIME !== "false",
    historicalData: process.env.REACT_APP_HISTORICAL === "true",
    debugMode: process.env.NODE_ENV === "development",
  },
};

export default config;
```

## Environment Variables

### Development (.env.development)

```bash
REACT_APP_API_URL=http://localhost:8080
REACT_APP_UPDATE_INTERVAL=10000
REACT_APP_DARK_MODE=false
REACT_APP_DEBUG_MODE=true
```

### Production (.env.production)

```bash
REACT_APP_API_URL=https://api.example.com
REACT_APP_UPDATE_INTERVAL=30000
REACT_APP_DARK_MODE=false
REACT_APP_DEBUG_MODE=false
```

## Build Configuration

### package.json

```json
{
  "homepage": "/site-availability",
  "scripts": {
    "build:dev": "REACT_APP_ENV=development npm run build",
    "build:prod": "REACT_APP_ENV=production npm run build"
  }
}
```

### Custom Styling

Override default styles in `src/styles/main.css`:

```css
:root {
  --primary-color: #1976d2;
  --secondary-color: #dc004e;
  --success-color: #4caf50;
  --error-color: #f44336;
  --warning-color: #ff9800;
}

/* Custom map styling */
.world-map {
  background: #f5f5f5;
}

.marker-up {
  fill: var(--success-color);
}

.marker-down {
  fill: var(--error-color);
}
```

## Feature Configuration

### Real-time Updates

```javascript
// Enable/disable automatic updates
features: {
  realTimeUpdates: true,
  updateInterval: 30000
}
```

### Map Customization

```javascript
map: {
  defaultZoom: 2,
  maxZoom: 10,
  projection: 'geoNaturalEarth1',
  markerSize: 8,
  showLabels: true
}
```

### Theme Options

```javascript
theme: {
  darkMode: false,
  primaryColor: '#1976d2',
  animations: true,
  compactMode: false
}
```

## Deployment Configuration

### Docker

```dockerfile
# Build stage
FROM node:18-alpine as build
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
RUN npm run build

# Production stage
FROM nginx:alpine
COPY --from=build /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

### Nginx Configuration

```nginx
server {
    listen 80;
    root /usr/share/nginx/html;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api {
        proxy_pass http://backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

This covers the essential frontend configuration options for different deployment scenarios.
