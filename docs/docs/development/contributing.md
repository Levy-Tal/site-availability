---
sidebar_position: 3
---

# Contributing Guide

Welcome! We're excited that you want to contribute to Site Availability Monitoring.

## Getting Started

### 1. Fork and Clone

```bash
# Fork the repository on GitHub, then clone your fork
git clone https://github.com/YOUR_USERNAME/site-availability.git
cd site-availability

# Add upstream remote
git remote add upstream https://github.com/Levy-Tal/site-availability.git
```

### 2. Set Up Development Environment

Follow the [Development Setup](./setup) guide to get your environment ready.

### 3. Create a Branch

```bash
# Sync with upstream
git fetch upstream
git checkout main
git merge upstream/main

# Create feature branch
git checkout -b feature/your-feature-name
```

## Contribution Types

### 🐛 Bug Reports

Before creating a bug report:

- Search existing issues
- Test with the latest version
- Provide minimal reproduction steps

Use the bug report template:

```markdown
**Bug Description**
A clear description of the bug.

**Steps to Reproduce**

1. Step one
2. Step two
3. See error

**Expected Behavior**
What should happen.

**Environment**

- OS: [e.g., Ubuntu 20.04]
- Go version: [e.g., 1.21]
- Node.js version: [e.g., 18.17]
```

### ✨ Feature Requests

For new features:

- Check if it aligns with project goals
- Discuss in GitHub Discussions first
- Consider implementation complexity

### 📝 Documentation

Documentation contributions are always welcome:

- Fix typos and grammar
- Add examples and clarifications
- Improve setup instructions
- Translate documentation

### 🔧 Code Contributions

We welcome:

- Bug fixes
- New features
- Performance improvements
- Test improvements
- Refactoring

## Development Guidelines

### Commit Messages

Use conventional commit format:

```
type(scope): description

[optional body]

[optional footer]
```

Types:

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Test additions/changes
- `chore`: Maintenance tasks

Examples:

```
feat(backend): add HMAC authentication
fix(frontend): resolve map rendering issue
docs(api): update endpoint documentation
test(scraping): add prometheus client tests
```

### Testing Requirements

#### Backend Tests

All Go code should have tests:

```go
func TestGetApplicationStatus(t *testing.T) {
    tests := []struct {
        name     string
        appName  string
        expected *Status
        wantErr  bool
    }{
        {
            name:     "valid application",
            appName:  "test-app",
            expected: &Status{Available: true},
            wantErr:  false,
        },
        {
            name:     "invalid application",
            appName:  "nonexistent",
            expected: nil,
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := GetApplicationStatus(tt.appName)

            if tt.wantErr && err == nil {
                t.Error("expected error but got none")
            }

            if !tt.wantErr && err != nil {
                t.Errorf("unexpected error: %v", err)
            }

            if !reflect.DeepEqual(result, tt.expected) {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

#### Frontend Tests

Use Jest and React Testing Library:

```javascript
import { render, screen, waitFor } from "@testing-library/react";
import { AppStatusPanel } from "./AppStatusPanel";

describe("AppStatusPanel", () => {
  test("renders application list", async () => {
    const mockApps = [
      { name: "app1", status: "up", location: "NYC" },
      { name: "app2", status: "down", location: "LA" },
    ];

    render(<AppStatusPanel applications={mockApps} />);

    await waitFor(() => {
      expect(screen.getByText("app1")).toBeInTheDocument();
      expect(screen.getByText("app2")).toBeInTheDocument();
    });
  });

  test("handles empty application list", () => {
    render(<AppStatusPanel applications={[]} />);
    expect(screen.getByText(/no applications/i)).toBeInTheDocument();
  });
});
```

## Review Process

### Pull Request Guidelines

1. **Title**: Use conventional commit format
2. **Description**: Explain what and why, not how
3. **Tests**: Include relevant tests
4. **Documentation**: Update docs if needed
5. **Breaking Changes**: Clearly mark and explain

### Review Checklist

Before submitting:

- [ ] Tests pass locally
- [ ] Code follows style guidelines
- [ ] Documentation is updated
- [ ] No sensitive information in commits
- [ ] Commit messages are clear
- [ ] PR description is complete

### Reviewer Guidelines

When reviewing:

- Be constructive and respectful
- Focus on code quality and maintainability
- Check for security issues
- Verify tests are adequate
- Ensure documentation is accurate

## Community Guidelines

### Code of Conduct

- Be respectful and inclusive
- Welcome newcomers
- Focus on constructive feedback
- Report inappropriate behavior

### Communication Channels

- **Issues**: Bug reports and feature requests
- **Discussions**: General questions and ideas
- **Pull Requests**: Code contributions
- **Security**: add email in the issue , and the team will make contact with you.

### Getting Help

1. Check existing documentation
2. Search existing issues
3. Ask in GitHub Discussions
4. Create an issue with detailed information

Thank you for contributing to Site Availability Monitoring! 🎉
