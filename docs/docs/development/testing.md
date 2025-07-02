---
sidebar_position: 4
---

# Testing Guide

Comprehensive testing strategies for Site Availability Monitoring.

## Testing Philosophy

Our testing approach follows the testing pyramid:

```
       /\
      /  \    E2E Tests (Few)
     /____\
    /      \  Integration Tests (Some)
   /________\
  /          \ Unit Tests (Many)
 /____________\
```

## Backend Testing

### Unit Tests

Test individual functions and methods:

```go
// handlers/handlers_test.go
func TestHealthHandler(t *testing.T) {
    req := httptest.NewRequest("GET", "/health", nil)
    w := httptest.NewRecorder()

    HealthHandler(w, req)

    resp := w.Result()
    if resp.StatusCode != http.StatusOK {
        t.Errorf("expected status 200, got %d", resp.StatusCode)
    }

    body, _ := io.ReadAll(resp.Body)
    if !strings.Contains(string(body), "healthy") {
        t.Error("expected 'healthy' in response body")
    }
}
```

### Integration Tests

Test component interactions:

```go
// scraping/integration_test.go
func TestPrometheusIntegration(t *testing.T) {
    // Start test Prometheus server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[{"metric":{"__name__":"up","instance":"localhost:9090","job":"prometheus"},"value":[1609459200,"1"]}]}}`))
    }))
    defer server.Close()

    client := prometheus.NewClient(server.URL)
    result, err := client.Query("up")

    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

### Test Utilities

Create helper functions for common test scenarios:

```go
// testutil/helpers.go
func SetupTestConfig() *config.Config {
    return &config.Config{
        ScrapeInterval: time.Second * 10,
        LogLevel:       "debug",
        Port:           8080,
        Locations: []config.Location{
            {Name: "Test Location", Latitude: 40.7128, Longitude: -74.0060},
        },
        Apps: []config.App{
            {Name: "test-app", Location: "Test Location", Metric: "up", Prometheus: "http://localhost:9090"},
        },
    }
}

func SetupTestServer(config *config.Config) *httptest.Server {
    handler := setupRoutes(config)
    return httptest.NewServer(handler)
}
```

### Mocking

Use interfaces for easy mocking:

```go
// Define interface
type PrometheusClient interface {
    Query(query string) (*QueryResult, error)
}

// Mock implementation
type MockPrometheusClient struct {
    QueryFunc func(string) (*QueryResult, error)
}

func (m *MockPrometheusClient) Query(query string) (*QueryResult, error) {
    if m.QueryFunc != nil {
        return m.QueryFunc(query)
    }
    return nil, errors.New("not implemented")
}

// Test with mock
func TestScraper(t *testing.T) {
    mockClient := &MockPrometheusClient{
        QueryFunc: func(query string) (*QueryResult, error) {
            return &QueryResult{Value: "1"}, nil
        },
    }

    scraper := NewScraper(mockClient)
    result, err := scraper.Scrape("up")

    assert.NoError(t, err)
    assert.Equal(t, "1", result.Value)
}
```

### Running Backend Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...

# Run specific package tests
go test ./handlers/

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Frontend Testing

### Unit Tests

Test React components:

```javascript
// AppStatusPanel.test.js
import { render, screen } from "@testing-library/react";
import { AppStatusPanel } from "./AppStatusPanel";

describe("AppStatusPanel", () => {
  test("renders app status correctly", () => {
    const mockApps = [
      { name: "app1", status: "up", location: "NYC" },
      { name: "app2", status: "down", location: "LA" },
    ];

    render(<AppStatusPanel applications={mockApps} />);

    expect(screen.getByText("app1")).toBeInTheDocument();
    expect(screen.getByText("app2")).toBeInTheDocument();

    // Check status indicators
    expect(screen.getByTestId("status-up")).toBeInTheDocument();
    expect(screen.getByTestId("status-down")).toBeInTheDocument();
  });

  test("handles loading state", () => {
    render(<AppStatusPanel applications={null} loading={true} />);
    expect(screen.getByText(/loading/i)).toBeInTheDocument();
  });
});
```

### API Tests

Mock API calls:

```javascript
// api.test.js
import { getApplications } from "./appStatusAPI";

// Mock fetch
global.fetch = jest.fn();

describe("API tests", () => {
  beforeEach(() => {
    fetch.mockClear();
  });

  test("getApplications returns data", async () => {
    const mockData = [{ name: "app1", status: "up" }];

    fetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockData,
    });

    const result = await getApplications();
    expect(result).toEqual(mockData);
    expect(fetch).toHaveBeenCalledWith("http://localhost:8080/api/apps");
  });

  test("getApplications handles errors", async () => {
    fetch.mockRejectedValueOnce(new Error("Network error"));

    await expect(getApplications()).rejects.toThrow("Network error");
  });
});
```

### Component Integration Tests

Test component interactions:

```javascript
// Map.integration.test.js
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Map } from "./Map";

// Mock D3 and world map data
jest.mock("d3", () => ({
  select: jest.fn(() => ({
    append: jest.fn(() => ({ attr: jest.fn() })),
    selectAll: jest.fn(() => ({ data: jest.fn() })),
  })),
}));

describe("Map Integration", () => {
  test("renders map with application markers", async () => {
    const mockApps = [
      { name: "app1", status: "up", location: { lat: 40.7128, lon: -74.006 } },
    ];

    render(<Map applications={mockApps} />);

    await waitFor(() => {
      expect(screen.getByTestId("world-map")).toBeInTheDocument();
    });

    // Verify marker is rendered
    expect(screen.getByTestId("marker-app1")).toBeInTheDocument();
  });

  test("handles marker click events", async () => {
    const mockOnMarkerClick = jest.fn();
    const mockApps = [
      { name: "app1", status: "up", location: { lat: 40.7128, lon: -74.006 } },
    ];

    render(<Map applications={mockApps} onMarkerClick={mockOnMarkerClick} />);

    const marker = screen.getByTestId("marker-app1");
    await userEvent.click(marker);

    expect(mockOnMarkerClick).toHaveBeenCalledWith("app1");
  });
});
```

### Running Frontend Tests

```bash
# Run all tests
npm test

# Run tests in watch mode
npm test -- --watch

# Run tests with coverage
npm test -- --coverage

# Run specific test file
npm test -- AppStatusPanel.test.js

# Update snapshots
npm test -- --updateSnapshot
```

## End-to-End Testing

### Cypress Setup

```bash
# Install Cypress
npm install --save-dev cypress

# Open Cypress
npx cypress open
```

### E2E Test Examples

```javascript
// cypress/e2e/app.cy.js
describe("Site Availability Monitoring", () => {
  beforeEach(() => {
    // Start backend server in test mode
    cy.exec("npm run start:test");
    cy.visit("http://localhost:3000");
  });

  it("displays the world map", () => {
    cy.get('[data-testid="world-map"]').should("be.visible");
  });

  it("shows application statuses", () => {
    cy.get('[data-testid="sidebar"]').should("contain", "Applications");
    cy.get('[data-testid="app-list"]').should("exist");
  });

  it("updates data in real-time", () => {
    // Wait for initial load
    cy.get('[data-testid="app-status"]').should("contain", "up");

    // Mock backend to return different status
    cy.intercept("GET", "/api/apps", { fixture: "apps-down.json" });

    // Wait for update
    cy.get('[data-testid="app-status"]', { timeout: 10000 }).should(
      "contain",
      "down",
    );
  });

  it("handles API errors gracefully", () => {
    cy.intercept("GET", "/api/apps", { statusCode: 500 });

    cy.get('[data-testid="error-message"]').should(
      "contain",
      "Failed to load applications",
    );
  });
});
```

### Visual Regression Testing

```javascript
// cypress/e2e/visual.cy.js
describe("Visual Regression Tests", () => {
  it("matches baseline screenshot", () => {
    cy.visit("http://localhost:3000");
    cy.get('[data-testid="world-map"]').should("be.visible");

    // Take screenshot and compare
    cy.matchImageSnapshot("world-map-baseline");
  });
});
```

## Performance Testing

### Load Testing with Artillery

```yaml
# artillery.yml
config:
  target: "http://localhost:8080"
  phases:
    - duration: 60
      arrivalRate: 10
  processor: "./test-functions.js"

scenarios:
  - name: "API Load Test"
    flow:
      - get:
          url: "/api/apps"
      - get:
          url: "/api/locations"
      - get:
          url: "/health"
```

### Benchmark Tests

```go
// benchmark_test.go
func BenchmarkHealthHandler(b *testing.B) {
    req := httptest.NewRequest("GET", "/health", nil)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        w := httptest.NewRecorder()
        HealthHandler(w, req)
    }
}

func BenchmarkPrometheusQuery(b *testing.B) {
    client := prometheus.NewClient("http://localhost:9090")

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := client.Query("up")
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## Test Data Management

### Fixtures

Create reusable test data:

```javascript
// cypress/fixtures/apps.json
[
  {
    name: "frontend",
    status: "up",
    location: "New York",
    last_check: "2023-12-01T10:00:00Z",
  },
  {
    name: "backend",
    status: "down",
    location: "London",
    last_check: "2023-12-01T10:00:00Z",
  },
];
```

### Test Database

For integration tests requiring data:

```go
func setupTestDB(t *testing.T) *sql.DB {
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        t.Fatal(err)
    }

    // Run migrations
    if err := runMigrations(db); err != nil {
        t.Fatal(err)
    }

    // Seed test data
    if err := seedTestData(db); err != nil {
        t.Fatal(err)
    }

    return db
}
```

## Continuous Integration

### GitHub Actions

```yaml
# .github/workflows/test.yml
name: Tests

on: [push, pull_request]

jobs:
  backend-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: 1.21

      - name: Run backend tests
        run: |
          cd backend
          go test -race -cover ./...

  frontend-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: 18

      - name: Install dependencies
        run: |
          cd frontend
          npm ci

      - name: Run frontend tests
        run: |
          cd frontend
          npm test -- --coverage --watchAll=false

  e2e-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run E2E tests
        run: |
          docker-compose -f docker-compose.test.yml up --abort-on-container-exit
```

## Best Practices

### Test Organization

- Group related tests in describe blocks
- Use clear, descriptive test names
- Follow AAA pattern (Arrange, Act, Assert)
- Keep tests independent and isolated

### Test Coverage

- Aim for 80%+ code coverage
- Focus on critical paths
- Don't sacrifice quality for coverage
- Use coverage reports to find gaps

### Test Maintenance

- Keep tests simple and focused
- Update tests when code changes
- Remove obsolete tests
- Refactor test code like production code

### Common Pitfalls

- **Flaky tests**: Use proper waits and timeouts
- **Slow tests**: Mock external dependencies
- **Brittle tests**: Avoid testing implementation details
- **Incomplete tests**: Test error scenarios too

Remember: Good tests are your safety net for confident refactoring and feature development!
