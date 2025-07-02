---
sidebar_position: 1
---

# Frontend Overview

The Site Availability Monitoring frontend is a modern React application that provides an interactive world map visualization for monitoring application status across geographic locations.

## Technology Stack

### Core Technologies

- **React 19**: Modern React with hooks and functional components
- **D3.js**: Data visualization and world map rendering
- **JavaScript (ES6+)**: Modern JavaScript features
- **CSS3**: Responsive styling and animations

### Build Tools

- **Create React App**: Development and build tooling
- **Webpack**: Module bundling (via CRA)
- **Babel**: JavaScript transpilation (via CRA)
- **ESLint**: Code linting
- **Prettier**: Code formatting

## Application Architecture

```
src/
├── components/         # React components
│   ├── Map.js         # World map visualization
│   ├── Sidebar.js     # Application status sidebar
│   └── AppStatusPanel.js # Individual app status
├── api/               # API integration
│   ├── appStatusAPI.js    # Application data
│   ├── docsAPI.js         # Documentation
│   └── scrapeIntervalAPI.js # Configuration
├── styles/            # Global styles
│   └── main.css       # Main stylesheet
├── config.js          # Configuration
├── App.js            # Main application
└── index.js          # Entry point
```

## Key Features

### Interactive World Map

- **SVG-based world map** using D3.js
- **Real-time status markers** showing application availability
- **Zoom and pan capabilities** for detailed exploration
- **Click interactions** for application details
- **Responsive design** that works on all devices

### Real-time Updates

- **Automatic data refresh** every 30 seconds (configurable)
- **WebSocket support** for instant updates (future)
- **Error handling** with graceful degradation
- **Loading states** for better user experience

### Status Visualization

- **Color-coded status indicators**: Green (up), Red (down), Yellow (unknown)
- **Location grouping** of applications
- **Historical data display** (when available)
- **Status change animations** for visual feedback

## Component Architecture

### App.js - Main Application

```javascript
const App = () => {
  const [appStatuses, setAppStatuses] = useState([]);
  const [locations, setLocations] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  // Data fetching and state management
  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, config.updateInterval);
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="app">
      <Map applications={appStatuses} locations={locations} />
      <Sidebar applications={appStatuses} loading={loading} error={error} />
    </div>
  );
};
```

### Map.js - World Map Component

```javascript
const Map = ({ applications, locations }) => {
  const svgRef = useRef();

  useEffect(() => {
    if (!applications.length) return;

    // D3.js map rendering logic
    const svg = d3.select(svgRef.current);

    // Render world map
    renderWorldMap(svg);

    // Add application markers
    renderApplicationMarkers(svg, applications);
  }, [applications, locations]);

  return (
    <div className="map-container">
      <svg ref={svgRef} className="world-map"></svg>
    </div>
  );
};
```

### Sidebar.js - Status Panel

```javascript
const Sidebar = ({ applications, loading, error }) => {
  const [filter, setFilter] = useState("all");
  const [sortBy, setSortBy] = useState("name");

  const filteredApps = useMemo(() => {
    return applications
      .filter((app) => filter === "all" || app.status === filter)
      .sort((a, b) => a[sortBy].localeCompare(b[sortBy]));
  }, [applications, filter, sortBy]);

  if (loading) return <div className="loading">Loading...</div>;
  if (error) return <div className="error">{error.message}</div>;

  return (
    <div className="sidebar">
      <div className="controls">
        <FilterControls filter={filter} onFilterChange={setFilter} />
        <SortControls sortBy={sortBy} onSortChange={setSortBy} />
      </div>

      <div className="app-list">
        {filteredApps.map((app) => (
          <AppStatusPanel key={app.name} application={app} />
        ))}
      </div>
    </div>
  );
};
```

## Data Flow

### Application State Management

```javascript
// State structure
const appState = {
  applications: [
    {
      name: "frontend",
      location: "New York",
      status: "up",
      lastCheck: "2023-12-01T10:00:00Z",
      responseTime: 0.142,
      coordinates: { latitude: 40.7128, longitude: -74.006 },
    },
  ],
  locations: [
    {
      name: "New York",
      latitude: 40.7128,
      longitude: -74.006,
      appsCount: 3,
    },
  ],
  loading: false,
  error: null,
};
```

### API Integration Flow

```javascript
// Data fetching pattern
const fetchApplications = async () => {
  try {
    setLoading(true);
    setError(null);

    const response = await fetch(`${config.apiUrl}/api/apps`);

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }

    const data = await response.json();
    setAppStatuses(data.data || []);
  } catch (err) {
    console.error("Failed to fetch applications:", err);
    setError(err);
  } finally {
    setLoading(false);
  }
};
```

## Styling and Theming

### CSS Architecture

```css
/* Global variables */
:root {
  --primary-color: #1976d2;
  --secondary-color: #dc004e;
  --success-color: #4caf50;
  --error-color: #f44336;
  --warning-color: #ff9800;
  --background-color: #f5f5f5;
  --text-color: #333;
  --border-radius: 4px;
  --box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

/* Component-based styling */
.app {
  display: flex;
  height: 100vh;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto",
    sans-serif;
}

.map-container {
  flex: 1;
  position: relative;
  background: var(--background-color);
}

.sidebar {
  width: 300px;
  background: white;
  border-left: 1px solid #ddd;
  overflow-y: auto;
}
```

### Responsive Design

```css
/* Mobile-first responsive design */
@media (max-width: 768px) {
  .app {
    flex-direction: column;
  }

  .sidebar {
    width: 100%;
    height: 40vh;
    border-left: none;
    border-top: 1px solid #ddd;
  }

  .map-container {
    height: 60vh;
  }
}

@media (max-width: 480px) {
  .sidebar {
    height: 50vh;
  }

  .map-container {
    height: 50vh;
  }

  .app-status-panel {
    padding: 8px;
    font-size: 14px;
  }
}
```

## Configuration

### Frontend Configuration

```javascript
// src/config.js
const config = {
  // API settings
  apiUrl: process.env.REACT_APP_API_URL || "http://localhost:8080",

  // Update intervals
  updateInterval: parseInt(process.env.REACT_APP_UPDATE_INTERVAL) || 30000, // 30 seconds

  // Map settings
  map: {
    defaultZoom: 2,
    maxZoom: 10,
    minZoom: 1,
    projection: "geoNaturalEarth1",
  },

  // Theme settings
  theme: {
    darkMode: process.env.REACT_APP_DARK_MODE === "true",
    primaryColor: process.env.REACT_APP_PRIMARY_COLOR || "#1976d2",
  },

  // Feature flags
  features: {
    realTimeUpdates: process.env.REACT_APP_REAL_TIME !== "false",
    historicalData: process.env.REACT_APP_HISTORICAL === "true",
    debugMode: process.env.NODE_ENV === "development",
  },
};

export default config;
```

### Environment Variables

```bash
# .env file
REACT_APP_API_URL=http://localhost:8080
REACT_APP_UPDATE_INTERVAL=30000
REACT_APP_DARK_MODE=false
REACT_APP_PRIMARY_COLOR=#1976d2
REACT_APP_REAL_TIME=true
REACT_APP_HISTORICAL=false
```

## Performance Optimizations

### Component Optimization

```javascript
// Memoized components for performance
const AppStatusPanel = React.memo(({ application }) => {
  return (
    <div className="app-status-panel">
      <StatusIndicator status={application.status} />
      <div className="app-info">
        <h3>{application.name}</h3>
        <p>{application.location}</p>
      </div>
    </div>
  );
});

// Memoized expensive calculations
const Map = ({ applications }) => {
  const processedData = useMemo(() => {
    return applications.map((app) => ({
      ...app,
      coordinates: projectCoordinates(app.coordinates),
    }));
  }, [applications]);

  return <MapVisualization data={processedData} />;
};
```

### API Optimization

```javascript
// Debounced API calls
const debouncedFetch = useCallback(
  debounce(() => {
    fetchApplications();
  }, 1000),
  [],
);

// Efficient state updates
const updateApplicationStatus = useCallback((appName, newStatus) => {
  setAppStatuses((prev) =>
    prev.map((app) =>
      app.name === appName
        ? { ...app, status: newStatus, lastCheck: new Date().toISOString() }
        : app,
    ),
  );
}, []);
```

## Error Handling

### Error Boundaries

```javascript
class ErrorBoundary extends React.Component {
  constructor(props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error) {
    return { hasError: true, error };
  }

  componentDidCatch(error, errorInfo) {
    console.error("Frontend error:", error, errorInfo);
    // Send to error reporting service
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="error-boundary">
          <h2>Something went wrong</h2>
          <p>
            Please refresh the page or contact support if the problem persists.
          </p>
          <button onClick={() => window.location.reload()}>Refresh Page</button>
        </div>
      );
    }

    return this.props.children;
  }
}
```

### API Error Handling

```javascript
// Centralized error handling
const handleApiError = (error, operation) => {
  const errorMessage = {
    message: `Failed to ${operation}`,
    details: error.message,
    timestamp: new Date().toISOString(),
    operation,
  };

  // Log error
  console.error("API Error:", errorMessage);

  // Update UI state
  setError(errorMessage);

  // Show user notification
  showNotification(`Error: ${errorMessage.message}`, "error");
};
```

## Testing Strategy

### Component Testing

```javascript
// Example test
import { render, screen, waitFor } from "@testing-library/react";
import { App } from "./App";

// Mock API
jest.mock("./api/appStatusAPI", () => ({
  getApplications: jest.fn(() =>
    Promise.resolve([
      { name: "test-app", status: "up", location: "Test Location" },
    ]),
  ),
}));

test("renders application status", async () => {
  render(<App />);

  await waitFor(() => {
    expect(screen.getByText("test-app")).toBeInTheDocument();
  });
});
```

## Accessibility

### ARIA Support

```javascript
const StatusIndicator = ({ status }) => (
  <div
    className={`status-indicator status-${status}`}
    role="img"
    aria-label={`Application status: ${status}`}
    title={`Status: ${status}`}
  >
    <span className="sr-only">Status: {status}</span>
  </div>
);
```

### Keyboard Navigation

```css
/* Focus indicators */
.app-status-panel:focus,
.map-marker:focus {
  outline: 2px solid var(--primary-color);
  outline-offset: 2px;
}

/* Screen reader only content */
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}
```

The frontend provides a modern, responsive, and accessible interface for monitoring application availability with real-time updates and interactive visualizations.
