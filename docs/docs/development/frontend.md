---
sidebar_position: 3
---

# Frontend Overview

This page covers the essential configuration and structure of the React frontend for Site Availability Monitoring.

## Main Files and Structure

- **src/App.js**: Main application component, sets up routing and layout.
- **src/index.js**: Entry point, renders the React app.
- **src/config.js**: Central configuration (API URL, feature flags, map settings, etc).
- **src/map.js**: Map rendering and logic for displaying locations and app statuses.
- **src/Sidebar.js**: Sidebar navigation and app list.
- **src/api/**: API helper modules for backend communication (e.g., appStatusAPI.js).
- **src/utils/**: Utility functions (e.g., storage.js for local storage helpers).
- **src/styles/**: Main CSS and custom styles.

## Running and Building

- Install dependencies:
  ```bash
  npm install
  ```
- Start development server:
  ```bash
  make build
  ```

---

For more details, see the code in the `frontend/` directory.
