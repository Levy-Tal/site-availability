---
sidebar_position: 2
---

# Terminology

This page defines key terms and concepts used throughout the Site Availability documentation.

## Core Concepts

### Source

A data source that provides monitoring information (e.g., Prometheus, HTTP endpoints). There are multiple sources (view sources in the docs). Every data source gets a 'config' section in the config.yaml, runs its scrape function against the config, and returns a list of apps with their statuses.

**Example:**

```yaml
- name: prometheus-main
  type: prometheus
  labels:
    service: "metric-monitoring"
    tier: "backend"
  config:
    url: http://prometheus:9090
    apps:
      - name: myApp
        location: London
        metric: up{instance="app:8080", job="app"}
        labels:
          app_type: "web-service"
          importance: "medium"
      - name: myPrometheus
        location: New York City
        metric: up{instance="localhost:9090", job="prometheus"}
        labels:
          app_type: "prometheus"
          importance: "low"
```

In this example, the Prometheus source will make an HTTP call to Prometheus to get the metrics value for each app. If the value is 1, then the app is up; if it's 0, then the app is down.

### App

An app is the smallest unit that has a status. An app can be a web app, database, server, network device, or any monitored service.

**Example:**

```yaml
apps:
  - name: myApp
    location: London
    metric: up{instance="app:8080", job="app"}
    labels:
      app_type: "web-service"
      importance: "medium"
```

### Location

A location is a geographical location that contains apps. Every app lives in a geographical location.

**Example:**

```yaml
locations:
  - name: New York City
    latitude: 40.712776
    longitude: -74.005974
```

### Status

#### Up

- **App**: App status is considered up if the source returned up
- **Location**: Location is considered up if all apps in it are up

#### Down

- **App**: App status is considered down if its source returned down
- **Location**: Location is considered down if at least one app is down in the location

#### Unavailable

- **App**: App is considered unavailable if the source did not return an answer regarding this app
- **Location**: Location is considered unavailable if one of the apps is unavailable and there is no app in down status

### Label

Every app has labels. Labels are used for filtering and authorization (for example, only Group A can view apps that have the label `group: A`).

There are 3 types of labels:

1. **Server Labels**: Applied to all apps in the server
2. **Source Labels**: Applied to all apps in the source
3. **App Labels**: Applied only to the specific app

Every app gets its labels from the sum of: **App Labels + Source Labels + Server Labels**.

## Configuration Terms

### Server Configuration

Settings that control the behavior of the Site Availability backend server.

### Source Configuration

Settings that define how to connect to and collect data from monitoring sources.
