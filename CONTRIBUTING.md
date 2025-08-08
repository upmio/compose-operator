# Contributing to Compose Operator

First off, thank you for considering contributing to Compose Operator! 🎉

The following is a set of guidelines for contributing to Compose Operator. These are mostly guidelines, not rules. Use your best judgment, and feel free to propose changes to this document in a pull request.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
- [Development Setup](#development-setup)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Documentation](#documentation)

## Code of Conduct

This project and everyone participating in it is governed by our [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the existing issues to avoid duplicates. When you are creating a bug report, please include as many details as possible:

**Use the bug report template:**

```markdown
**Describe the bug**
A clear and concise description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Deploy operator with config '...'
2. Create CRD '...'
3. See error

**Expected behavior**
A clear and concise description of what you expected to happen.

**Environment:**
- Kubernetes version: [e.g. v1.25.0]
- Operator version: [e.g. v1.0.0]
- Database versions: [e.g. MySQL 8.0.30]

**Additional context**
Add any other context about the problem here.
```

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, please include:

- **Use a clear and descriptive title**
- **Provide a step-by-step description** of the suggested enhancement
- **Explain why this enhancement would be useful**
- **Include examples** of how the feature would work

### Your First Code Contribution

Unsure where to begin contributing? You can start by looking through `good-first-issue` and `help-wanted` issues:

- **Good first issues** - issues which should only require a few lines of code
- **Help wanted issues** - issues which should be a bit more involved

### Pull Requests

1. Fork the repo and create your branch from `main`
2. If you've added code that should be tested, add tests
3. If you've changed APIs, update the documentation
4. Ensure the test suite passes
5. Make sure your code follows the coding standards
6. Submit your pull request!

## Development Setup

### Prerequisites

- Go 1.21 or later
- Docker
- Kubernetes cluster (for testing)
- kubectl configured
- Helm 3.x
- Make (for running Makefile commands)
- Git

### Local Development Environment

1. **Fork and clone the repository:**

```bash
git clone https://github.com/YOUR_USERNAME/compose-operator.git
cd compose-operator
```

2. **Install dependencies:**

```bash
# Install kubebuilder tools
make install-tools

# Install CRDs
make install
```

3. **Run tests:**

```bash
# Run unit tests
make test

# Run linting
make lint

# Run with coverage
make test-coverage
```

4. **Build and run locally:**

```bash
# Build the binary
make build

# Run the operator locally (requires kubeconfig)
make run
```

5. **Build container image:**

```bash
make docker-build IMG=compose-operator:dev
```

### Development Workflow

1. **Create a new branch:**

```bash
git checkout -b feature/your-feature-name
```

2. **Make your changes**

3. **Run tests and linting:**

```bash
make test lint
```

4. **Commit your changes:**

```bash
git add .
git commit -m "feat: add amazing new feature"
```

5. **Push to your fork:**

```bash
git push origin feature/your-feature-name
```

6. **Create a Pull Request**

## Pull Request Process

### Before Submitting

- [ ] Tests pass locally
- [ ] Code is properly formatted (`make fmt`)
- [ ] Linting passes (`make lint`)
- [ ] Documentation is updated if needed
- [ ] CHANGELOG.md is updated for significant changes

### PR Guidelines

1. **Title**: Use a clear and descriptive title
   - `feat: add Redis replication support`
   - `fix: resolve MySQL connection timeout issue`
   - `docs: update installation guide`

2. **Description**: Include:
   - What changes were made and why
   - Any breaking changes
   - Testing instructions
   - Screenshots (if UI changes)

3. **Size**: Keep PRs focused and reasonably sized
   - Large PRs are harder to review
   - Consider breaking large features into smaller PRs

### Review Process

1. **Automated checks** must pass (CI/CD, tests, linting)
2. **Code review** by at least one maintainer
3. **Documentation review** if docs are changed
4. **Final approval** by a maintainer

## Coding Standards

### Go Code Style

We follow standard Go conventions and use automated tools to enforce consistency:

```bash
# Format code
make fmt

# Run linter
make lint

# Vet code
make vet
```

### Code Organization

```
├── api/v1alpha1/          # API definitions
├── cmd/                   # Main applications
├── controller/            # Controllers
├── pkg/                   # Library code
│   ├── mysqlutil/         # MySQL utilities
│   ├── redisutil/         # Redis utilities
│   └── utils/             # Common utilities
├── config/                # Configuration files
└── examples/              # Example manifests
```

### Naming Conventions

- **Packages**: lowercase, single word when possible
- **Functions**: camelCase, exported functions start with uppercase
- **Variables**: camelCase
- **Constants**: UPPER_CASE with underscores
- **Files**: lowercase with underscores

### Error Handling

```go
// Good: Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to connect to database %s: %w", host, err)
}

// Good: Use specific error types
var ErrConnectionTimeout = errors.New("database connection timeout")

// Avoid: Generic error messages
if err != nil {
    return err
}
```

### Logging

Use structured logging with appropriate levels:

```go
log := ctrl.Log.WithName("mysql-controller")

// Info for normal operations
log.Info("Successfully configured replication", "source", source, "replica", replica)

// Error for error conditions
log.Error(err, "Failed to establish replication", "source", source)

// Debug for detailed troubleshooting
log.V(1).Info("Connecting to database", "host", host, "port", port)
```

## Testing Guidelines

### Test Structure

```
controller/
├── mysqlreplication/
│   ├── mysqlreplication_controller.go
│   ├── mysqlreplication_controller_test.go  # Controller tests
│   └── suite_test.go                        # Test suite setup
```

### Unit Tests

- Test individual functions and methods
- Use table-driven tests when appropriate
- Mock external dependencies

```go
func TestMysqlReplication(t *testing.T) {
    tests := []struct {
        name    string
        spec    v1alpha1.MysqlReplicationSpec
        want    error
    }{
        {
            name: "valid configuration",
            spec: v1alpha1.MysqlReplicationSpec{...},
            want: nil,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Integration Tests

- Test controller behavior with real Kubernetes API
- Use envtest for controller testing
- Test common scenarios and edge cases

### Test Coverage

- Aim for at least 70% code coverage
- Critical paths should have higher coverage
- Run coverage reports: `make test-coverage`

## Documentation

### Code Documentation

- Public functions and types must have godoc comments
- Comments should explain **why**, not just **what**
- Keep comments up to date with code changes

```go
// MysqlReplicationReconciler reconciles a MysqlReplication object.
// It manages the replication topology between MySQL instances by
// monitoring their status and configuring replication relationships.
type MysqlReplicationReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}
```

### README Updates

When adding new features or changing behavior:

1. Update relevant sections in README.md
2. Add or update examples
3. Update the feature comparison table if needed

### API Documentation

- Update API documentation in `doc/` directory
- Include example manifests
- Document breaking changes clearly

## Release Process

### Versioning

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR**: Incompatible API changes
- **MINOR**: New functionality in a backwards compatible manner
- **PATCH**: Backwards compatible bug fixes

### Changelog

Update `CHANGELOG.md` for significant changes:

```markdown
## [1.2.0] - 2024-01-15

### Added
- Redis replication support
- PostgreSQL streaming replication

### Fixed
- MySQL connection timeout handling
- Memory leak in Redis cluster monitoring

### Changed
- Improved error messages for configuration validation

## Community

### Communication

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: Questions and general discussion
- **Pull Requests**: Code contributions and reviews

### Getting Help

If you need help with development:

1. Check existing documentation
2. Search existing issues and discussions
3. Ask questions in GitHub Discussions
4. Join community meetings (if applicable)

## Recognition

Contributors are recognized in:

- Release notes
- CONTRIBUTORS.md file
- Annual contributor highlights

Thank you for contributing to Compose Operator! 🚀