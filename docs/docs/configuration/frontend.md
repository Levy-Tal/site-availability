---
sidebar_position: 3
---

# Frontend Configuration

Configure the React frontend application settings and behavior.

## Configuration File

Edit `src/config.js` to configure the frontend:

```javascript
const config = {
  // Backend API URL
  apiUrl: process.env.REACT_APP_API_URL || "http://localhost:8080",

  // Map settings
  map: {
    updateInterval: 30000, // 30 seconds
    defaultZoom: 2,
    maxZoom: 10,
    minZoom: 1,
  },

  // Theme settings
  theme: {
    darkMode: true,
    primaryColor: "#1976d2",
    secondaryColor: "#dc004e",
  },

  // Features
  features: {
    realTimeUpdates: true,
    historicalData: true,
    alerts: true,
  },
};

export default config;
```

## Environment Variables

Configure via environment variables:

```bash
# API Configuration
REACT_APP_API_URL=http://localhost:8080
REACT_APP_MAP_UPDATE_INTERVAL=30000

# Feature Flags
REACT_APP_ENABLE_DARK_MODE=true
REACT_APP_ENABLE_ALERTS=true
REACT_APP_DEBUG_MODE=false

# Map Settings
REACT_APP_DEFAULT_ZOOM=2
REACT_APP_MAX_ZOOM=10
```

## Build Configuration

Configure build settings in `package.json`:

```json
{
  "homepage": "/site-availability",
  "scripts": {
    "build": "react-scripts build",
    "build:prod": "REACT_APP_API_URL=https://api.example.com npm run build"
  }
}
```

## Customization

### Custom Styling

Override default styles in `src/styles/main.css`:

```css
:root {
  --primary-color: #1976d2;
  --secondary-color: #dc004e;
  --background-color: #f5f5f5;
  --text-color: #333;
}

.custom-map-container {
  height: 100vh;
  width: 100%;
}
```

### Map Customization

Customize map appearance:

```javascript
const mapConfig = {
  style: "mapbox://styles/mapbox/dark-v10",
  center: [0, 20],
  zoom: 2,
  markers: {
    color: "#ff0000",
    size: "large",
  },
};
```
