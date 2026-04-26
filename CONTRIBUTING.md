# Contributing to Chronos

Thank you for your interest in contributing to Chronos! This document provides guidelines for contributing to the project.

## Code of Conduct

This project and everyone participating in it is governed by our [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## How Can I Contribute?

### Reporting Bugs

Before creating a bug report, please check if the issue already exists. When creating a bug report, include:

- A clear, descriptive title
- Steps to reproduce the issue
- Expected behavior vs actual behavior
- Your environment (OS, Go version, database type)
- Any relevant logs or error messages

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion:

- Use a clear, descriptive title
- Provide a detailed description of the proposed enhancement
- Explain why this enhancement would be useful
- Consider potential implementation approaches

### Pull Requests

1. Fork the repository
2. Create a branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`make check`)
5. Commit your changes (`git commit -m 'feat: add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## Development Setup

```bash
# Clone your fork
git clone https://github.com/your-username/chronos.git
cd chronos

# Install dependencies
go mod download

# Run tests
make test

# Build
make build
```

## Commit Message Guidelines

We follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `test:` Adding or updating tests
- `refactor:` Code refactoring
- `perf:` Performance improvements
- `chore:` Build process or auxiliary tool changes

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Keep functions focused and small
- Add tests for new functionality
- Update documentation for API changes
- Use meaningful variable and function names

## Testing

- All new functionality must include tests
- Run the full test suite before submitting: `make check`
- Aim for high test coverage on core logic
- Use table-driven tests where appropriate

## Documentation

- Update README.md if your changes affect usage
- Add GoDoc comments for exported functions and types
- Update AGENTS.md if architectural patterns change

## License

By contributing to Chronos, you agree that your contributions will be licensed under the MIT License.
