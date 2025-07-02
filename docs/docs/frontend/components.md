---
sidebar_position: 2
---

# Frontend Components

Detailed documentation of React components used in the Site Availability Monitoring frontend.

## Component Hierarchy

```
App
├── Map
│   ├── WorldMap
│   ├── ApplicationMarker
│   └── Tooltip
├── Sidebar
│   ├── AppStatusPanel
│   ├── FilterControls
│   └── SortControls
└── ErrorBoundary
```

## Core Components

### App.js - Main Application Container

The root component that manages global state and coordinates data flow.

```javascript
import React, { useState, useEffect } from "react";
import Map from "./Map";
import Sidebar from "./Sidebar";
import { getApplications, getLocations } from "../api/appStatusAPI";
import config from "../config";

const App = () => {
  // State management
  const [appStatuses, setAppStatuses] = useState([]);
  const [locations, setLocations] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [lastUpdate, setLastUpdate] = useState(null);

  // Data fetching
  const fetchData = async () => {
    try {
      setLoading(true);
      setError(null);

      const [appsResponse, locationsResponse] = await Promise.all([
        getApplications(),
        getLocations(),
      ]);

      setAppStatuses(appsResponse.data || []);
      setLocations(locationsResponse.data || []);
      setLastUpdate(new Date());
    } catch (err) {
      console.error("Failed to fetch data:", err);
      setError(err);
    } finally {
      setLoading(false);
    }
  };

  // Auto-refresh data
  useEffect(() => {
    fetchData();

    const interval = setInterval(fetchData, config.updateInterval);
    return () => clearInterval(interval);
  }, []);

  // Manual refresh handler
  const handleRefresh = () => {
    fetchData();
  };

  return (
    <div className="app">
      <Map
        applications={appStatuses}
        locations={locations}
        onMarkerClick={handleMarkerClick}
      />
      <Sidebar
        applications={appStatuses}
        loading={loading}
        error={error}
        lastUpdate={lastUpdate}
        onRefresh={handleRefresh}
      />
    </div>
  );
};

export default App;
```

**Props:** None (root component)

**State:**

- `appStatuses`: Array of application status objects
- `locations`: Array of location objects
- `loading`: Boolean indicating data loading state
- `error`: Error object if fetch fails
- `lastUpdate`: Timestamp of last successful update

---

### Map.js - Interactive World Map

Renders the world map with application status markers using D3.js.

```javascript
import React, { useRef, useEffect, useState } from "react";
import * as d3 from "d3";
import worldData from "../data/world-110m.json";

const Map = ({ applications, locations, onMarkerClick }) => {
  const svgRef = useRef();
  const [selectedApp, setSelectedApp] = useState(null);
  const [tooltip, setTooltip] = useState({
    visible: false,
    x: 0,
    y: 0,
    content: null,
  });

  useEffect(() => {
    if (!applications.length || !locations.length) return;

    const svg = d3.select(svgRef.current);
    const width = 960;
    const height = 500;

    // Clear previous content
    svg.selectAll("*").remove();

    // Set up projection
    const projection = d3
      .geoNaturalEarth1()
      .scale(width / 6.5)
      .translate([width / 2, height / 2]);

    const path = d3.geoPath().projection(projection);

    // Render world map
    svg
      .append("g")
      .selectAll("path")
      .data(worldData.features)
      .enter()
      .append("path")
      .attr("d", path)
      .attr("class", "country")
      .attr("fill", "#e0e0e0")
      .attr("stroke", "#ccc")
      .attr("stroke-width", 0.5);

    // Render application markers
    const markers = svg
      .append("g")
      .selectAll("circle")
      .data(applications)
      .enter()
      .append("circle")
      .attr(
        "cx",
        (d) => projection([d.coordinates.longitude, d.coordinates.latitude])[0],
      )
      .attr(
        "cy",
        (d) => projection([d.coordinates.longitude, d.coordinates.latitude])[1],
      )
      .attr("r", 8)
      .attr("class", (d) => `marker marker-${d.status}`)
      .attr("fill", (d) => getStatusColor(d.status))
      .attr("stroke", "#fff")
      .attr("stroke-width", 2)
      .style("cursor", "pointer")
      .on("click", handleMarkerClick)
      .on("mouseover", handleMarkerMouseOver)
      .on("mouseout", handleMarkerMouseOut);

    // Add application labels
    svg
      .append("g")
      .selectAll("text")
      .data(applications)
      .enter()
      .append("text")
      .attr(
        "x",
        (d) => projection([d.coordinates.longitude, d.coordinates.latitude])[0],
      )
      .attr(
        "y",
        (d) =>
          projection([d.coordinates.longitude, d.coordinates.latitude])[1] + 20,
      )
      .attr("text-anchor", "middle")
      .attr("class", "marker-label")
      .style("font-size", "12px")
      .style("fill", "#333")
      .text((d) => d.name);
  }, [applications, locations]);

  const getStatusColor = (status) => {
    switch (status) {
      case "up":
        return "#4caf50";
      case "down":
        return "#f44336";
      case "unknown":
        return "#ff9800";
      default:
        return "#9e9e9e";
    }
  };

  const handleMarkerClick = (event, d) => {
    setSelectedApp(d);
    if (onMarkerClick) {
      onMarkerClick(d);
    }
  };

  const handleMarkerMouseOver = (event, d) => {
    const [x, y] = d3.pointer(event);
    setTooltip({
      visible: true,
      x: x + 10,
      y: y - 10,
      content: (
        <div className="tooltip-content">
          <strong>{d.name}</strong>
          <div>
            Status: <span className={`status-${d.status}`}>{d.status}</span>
          </div>
          <div>Location: {d.location}</div>
          {d.responseTime && <div>Response: {d.responseTime}s</div>}
        </div>
      ),
    });
  };

  const handleMarkerMouseOut = () => {
    setTooltip({ visible: false, x: 0, y: 0, content: null });
  };

  return (
    <div className="map-container">
      <svg
        ref={svgRef}
        className="world-map"
        viewBox="0 0 960 500"
        preserveAspectRatio="xMidYMid meet"
      />

      {tooltip.visible && (
        <div
          className="tooltip"
          style={{
            position: "absolute",
            left: tooltip.x,
            top: tooltip.y,
            pointerEvents: "none",
          }}
        >
          {tooltip.content}
        </div>
      )}
    </div>
  );
};

export default Map;
```

**Props:**

- `applications`: Array of application objects with coordinates
- `locations`: Array of location objects
- `onMarkerClick`: Callback function when marker is clicked

**Features:**

- SVG-based world map rendering
- Interactive application markers
- Hover tooltips with status information
- Click handlers for application selection
- Responsive design with viewBox

---

### Sidebar.js - Application Status Panel

Displays application status list with filtering and sorting capabilities.

```javascript
import React, { useState, useMemo } from "react";
import AppStatusPanel from "./AppStatusPanel";

const Sidebar = ({ applications, loading, error, lastUpdate, onRefresh }) => {
  const [filter, setFilter] = useState("all");
  const [sortBy, setSortBy] = useState("name");
  const [sortOrder, setSortOrder] = useState("asc");
  const [searchTerm, setSearchTerm] = useState("");

  // Filter and sort applications
  const filteredAndSortedApps = useMemo(() => {
    let filtered = applications;

    // Apply status filter
    if (filter !== "all") {
      filtered = filtered.filter((app) => app.status === filter);
    }

    // Apply search filter
    if (searchTerm) {
      filtered = filtered.filter(
        (app) =>
          app.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
          app.location.toLowerCase().includes(searchTerm.toLowerCase()),
      );
    }

    // Apply sorting
    filtered.sort((a, b) => {
      let aValue = a[sortBy];
      let bValue = b[sortBy];

      if (typeof aValue === "string") {
        aValue = aValue.toLowerCase();
        bValue = bValue.toLowerCase();
      }

      if (sortOrder === "asc") {
        return aValue < bValue ? -1 : aValue > bValue ? 1 : 0;
      } else {
        return aValue > bValue ? -1 : aValue < bValue ? 1 : 0;
      }
    });

    return filtered;
  }, [applications, filter, sortBy, sortOrder, searchTerm]);

  // Status summary
  const statusSummary = useMemo(() => {
    const summary = applications.reduce((acc, app) => {
      acc[app.status] = (acc[app.status] || 0) + 1;
      return acc;
    }, {});

    return {
      total: applications.length,
      up: summary.up || 0,
      down: summary.down || 0,
      unknown: summary.unknown || 0,
    };
  }, [applications]);

  if (loading) {
    return (
      <div className="sidebar">
        <div className="loading">
          <div className="spinner"></div>
          <p>Loading applications...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="sidebar">
        <div className="error">
          <h3>Error Loading Data</h3>
          <p>{error.message}</p>
          <button onClick={onRefresh} className="retry-button">
            Retry
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="sidebar">
      {/* Header */}
      <div className="sidebar-header">
        <h2>Applications</h2>
        <button onClick={onRefresh} className="refresh-button" title="Refresh">
          🔄
        </button>
      </div>

      {/* Status Summary */}
      <div className="status-summary">
        <div className="summary-item">
          <span className="count">{statusSummary.total}</span>
          <span className="label">Total</span>
        </div>
        <div className="summary-item status-up">
          <span className="count">{statusSummary.up}</span>
          <span className="label">Up</span>
        </div>
        <div className="summary-item status-down">
          <span className="count">{statusSummary.down}</span>
          <span className="label">Down</span>
        </div>
        <div className="summary-item status-unknown">
          <span className="count">{statusSummary.unknown}</span>
          <span className="label">Unknown</span>
        </div>
      </div>

      {/* Controls */}
      <div className="sidebar-controls">
        {/* Search */}
        <input
          type="text"
          placeholder="Search applications..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          className="search-input"
        />

        {/* Filter */}
        <select
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
          className="filter-select"
        >
          <option value="all">All Status</option>
          <option value="up">Up</option>
          <option value="down">Down</option>
          <option value="unknown">Unknown</option>
        </select>

        {/* Sort */}
        <select
          value={sortBy}
          onChange={(e) => setSortBy(e.target.value)}
          className="sort-select"
        >
          <option value="name">Name</option>
          <option value="location">Location</option>
          <option value="status">Status</option>
          <option value="lastCheck">Last Check</option>
        </select>

        <button
          onClick={() => setSortOrder(sortOrder === "asc" ? "desc" : "asc")}
          className="sort-order-button"
          title={`Sort ${sortOrder === "asc" ? "Descending" : "Ascending"}`}
        >
          {sortOrder === "asc" ? "↑" : "↓"}
        </button>
      </div>

      {/* Application List */}
      <div className="app-list">
        {filteredAndSortedApps.length === 0 ? (
          <div className="empty-state">
            <p>No applications match your criteria</p>
          </div>
        ) : (
          filteredAndSortedApps.map((app) => (
            <AppStatusPanel key={app.name} application={app} />
          ))
        )}
      </div>

      {/* Footer */}
      {lastUpdate && (
        <div className="sidebar-footer">
          <small>Last updated: {lastUpdate.toLocaleTimeString()}</small>
        </div>
      )}
    </div>
  );
};

export default Sidebar;
```

**Props:**

- `applications`: Array of application objects
- `loading`: Boolean indicating loading state
- `error`: Error object if present
- `lastUpdate`: Timestamp of last update
- `onRefresh`: Callback function for refresh action

**Features:**

- Real-time status summary
- Search functionality
- Status filtering
- Multiple sorting options
- Error handling and retry mechanism

---

### AppStatusPanel.js - Individual Application Status

Displays detailed status information for a single application.

```javascript
import React from "react";

const AppStatusPanel = ({ application }) => {
  const {
    name,
    location,
    status,
    lastCheck,
    responseTime,
    metric,
    prometheus_url,
  } = application;

  const formatTime = (timestamp) => {
    if (!timestamp) return "Never";

    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = now - date;
    const diffMins = Math.floor(diffMs / 60000);

    if (diffMins < 1) return "Just now";
    if (diffMins < 60) return `${diffMins}m ago`;

    const diffHours = Math.floor(diffMins / 60);
    if (diffHours < 24) return `${diffHours}h ago`;

    return date.toLocaleDateString();
  };

  const getStatusIcon = (status) => {
    switch (status) {
      case "up":
        return "✅";
      case "down":
        return "❌";
      case "unknown":
        return "❓";
      default:
        return "⚪";
    }
  };

  const getStatusClass = (status) => {
    return `status-indicator status-${status}`;
  };

  return (
    <div className={`app-status-panel app-status-${status}`}>
      {/* Status Indicator */}
      <div className={getStatusClass(status)}>
        <span
          className="status-icon"
          role="img"
          aria-label={`Status: ${status}`}
        >
          {getStatusIcon(status)}
        </span>
      </div>

      {/* Application Info */}
      <div className="app-info">
        <div className="app-header">
          <h3 className="app-name" title={name}>
            {name}
          </h3>
          <span className={`status-badge status-${status}`}>
            {status.toUpperCase()}
          </span>
        </div>

        <div className="app-details">
          <div className="detail-item">
            <span className="detail-label">Location:</span>
            <span className="detail-value">{location}</span>
          </div>

          <div className="detail-item">
            <span className="detail-label">Last Check:</span>
            <span className="detail-value" title={lastCheck}>
              {formatTime(lastCheck)}
            </span>
          </div>

          {responseTime && (
            <div className="detail-item">
              <span className="detail-label">Response Time:</span>
              <span className="detail-value">{responseTime.toFixed(3)}s</span>
            </div>
          )}

          {metric && (
            <div className="detail-item">
              <span className="detail-label">Metric:</span>
              <span className="detail-value metric-query" title={metric}>
                {metric.length > 40 ? `${metric.substring(0, 40)}...` : metric}
              </span>
            </div>
          )}
        </div>
      </div>

      {/* Actions */}
      <div className="app-actions">
        {prometheus_url && (
          <a
            href={prometheus_url}
            target="_blank"
            rel="noopener noreferrer"
            className="action-link"
            title="View in Prometheus"
          >
            📊
          </a>
        )}

        <button
          className="action-button"
          onClick={() => navigator.clipboard?.writeText(name)}
          title="Copy application name"
        >
          📋
        </button>
      </div>
    </div>
  );
};

export default React.memo(AppStatusPanel);
```

**Props:**

- `application`: Object containing application details
  - `name`: Application name
  - `location`: Geographic location
  - `status`: Current status ('up', 'down', 'unknown')
  - `lastCheck`: Timestamp of last check
  - `responseTime`: Response time in seconds
  - `metric`: Prometheus metric query
  - `prometheus_url`: URL to Prometheus instance

**Features:**

- Status visualization with icons and colors
- Relative time formatting
- Responsive layout
- Action buttons for external links
- Memoized for performance
- Accessibility support with ARIA labels

---

## Utility Components

### ErrorBoundary.js - Error Handling

```javascript
import React from "react";

class ErrorBoundary extends React.Component {
  constructor(props) {
    super(props);
    this.state = { hasError: false, error: null, errorInfo: null };
  }

  static getDerivedStateFromError(error) {
    return { hasError: true };
  }

  componentDidCatch(error, errorInfo) {
    this.setState({
      error: error,
      errorInfo: errorInfo,
    });

    // Log error to console or external service
    console.error("ErrorBoundary caught an error:", error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="error-boundary">
          <h2>Something went wrong</h2>
          <p>An unexpected error occurred. Please refresh the page.</p>

          {process.env.NODE_ENV === "development" && (
            <details style={{ whiteSpace: "pre-wrap", marginTop: "20px" }}>
              <summary>Error Details (Development Only)</summary>
              {this.state.error && this.state.error.toString()}
              <br />
              {this.state.errorInfo.componentStack}
            </details>
          )}

          <button
            onClick={() => window.location.reload()}
            className="error-retry-button"
          >
            Refresh Page
          </button>
        </div>
      );
    }

    return this.props.children;
  }
}

export default ErrorBoundary;
```

### LoadingSpinner.js - Loading Indicator

```javascript
import React from "react";

const LoadingSpinner = ({ size = "medium", message = "Loading..." }) => {
  const sizeClasses = {
    small: "spinner-small",
    medium: "spinner-medium",
    large: "spinner-large",
  };

  return (
    <div className="loading-container">
      <div className={`spinner ${sizeClasses[size]}`}>
        <div className="spinner-circle"></div>
      </div>
      {message && <p className="loading-message">{message}</p>}
    </div>
  );
};

export default LoadingSpinner;
```

These components provide a solid foundation for the Site Availability Monitoring frontend, with proper state management, error handling, and user experience considerations.
