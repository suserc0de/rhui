# Contributing to RHUI

Thank you for your interest in contributing to **RHUI**! We welcome community contributions to make Go's component rendering experience even better.

## Development Workflow

### Prerequisites
- Go 1.22+ installed
- Git

### Running Tests
Make sure all unit tests, race detector tests, and benchmarks pass before submitting a Pull Request:

```bash
# Run all tests
go test -v ./tests/...

# Run race condition checks
go test -race ./tests/...

# Run benchmarks
go test -bench=. -benchmem ./tests/...
```

### Pull Request Guidelines
1. Fork the repository and create your feature branch from `master`.
2. Ensure new features are covered with comprehensive tests in `tests/`.
3. Adhere to Go standard code formatting (`gofmt` or `goimports`).
4. Write clear and concise commit messages.
