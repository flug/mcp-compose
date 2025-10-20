# Contributing to MCP Add Server

Thank you for your interest in contributing to MCP Add Server!

## Development Setup

1. Fork and clone the repository
2. Install Go 1.21 or higher
3. Install dependencies:
   ```bash
   go mod download
   ```

## Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out
```

## Code Style

We use standard Go formatting tools:

```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run
```

## Making Changes

1. Create a new branch for your feature/fix
2. Write tests for your changes
3. Ensure all tests pass
4. Ensure code is properly formatted
5. Submit a pull request

## Pull Request Guidelines

- Write clear, descriptive commit messages
- Include tests for new functionality
- Update documentation as needed
- Keep PRs focused on a single feature/fix
- Reference any related issues

## Project Structure

```
mcp-compose/
├── main.go              # CLI entry point
├── pkg/
│   ├── models/          # Data structures
│   ├── config/          # Configuration I/O
│   └── commands/        # Command implementations
```

## Adding New Commands

1. Create command file in `pkg/commands/`
2. Add command handler to `main.go`
3. Write tests in `pkg/commands/*_test.go`
4. Update README.md and CLAUDE.md
5. Add usage examples

## Reporting Issues

When reporting issues, please include:
- Go version
- Operating system
- Steps to reproduce
- Expected vs actual behavior
- Any relevant error messages

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and grow

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
