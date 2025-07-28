---
sidebar_position: 4
---

# Testing Guide

This guide covers the essential testing commands and practices for Site Availability Monitoring.

## Backend (Go) Testing

- Run all tests:
  ```bash
  go test ./...
  ```
- Run with coverage:
  ```bash
  go test -cover ./...
  ```
- Run with race detection:
  ```bash
  go test -race ./...
  ```
- Generate coverage report:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Frontend (React) Testing

- Run all tests:
  ```bash
  npm test
  ```
- Run with coverage:

```bash
npm test -- --coverage
```

- Run specific test file:

```bash
  npm test -- AppStatusPanel.test.js
```

## Continuous Integration

- Backend and frontend tests are run automatically in CI (see `.github/workflows/`).
- Use `go test` and `npm test` locally before pushing changes.

## Best Practices

- Write unit tests for new features and bug fixes.
- Keep tests simple and focused.
- Use coverage reports to find gaps.
- Remove or update obsolete tests.

---

For more details, see the test files in the codebase.
