# AGENTS.md - Aceplay Development Guide

This document provides guidelines for AI agents working on the Aceplay project.

## Project Overview

Aceplay is a modern Go reimplementation of acestream-launcher. It automatically starts acestream-engine if not running and plays Ace Stream content using the user's preferred video player.

**Tech Stack:**
- Go 1.24.2 (minimum 1.22)
- Charm ecosystem (Bubble Tea, Huh, Lipgloss, Log)
- Cobra CLI framework
- Viper for configuration
- Resty for HTTP requests
- Testify for testing

---

## Build, Test, and Lint Commands

### Building

```bash
# Build binary for current platform
make build
# or
go build -o build/aceplay ./cmd

# Build with version info
make build
# Binary includes version, commit, and date from git

# Build for all platforms
make build-all

# Cross-compile manually
GOOS=linux GOARCH=amd64 go build -o build/aceplay ./cmd
```

### Testing

```bash
# Run all tests
make test
# or
go test -v -race ./...

# Run a single test
go test -v -run TestFunctionName ./internal/acestream

# Run tests in specific package
go test -v ./internal/config/...

# Run short tests
make test-short

# Run with coverage
make test-coverage
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Linting and Code Quality

```bash
# Format code
make fmt
# or
go fmt ./...

# Run go vet
make vet
# or
go vet ./...

# Run linter (requires golangci-lint)
make lint
# Install: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh

# Run all checks (fmt, vet, lint, test)
make check
```

### Development

```bash
# Install dependencies
make deps
# or
go mod download
go mod tidy

# Run the application
make run
# or
go run ./cmd

# Install binary to system
sudo make install

# Clean build artifacts
make clean
```

---

## Code Style Guidelines

### General Principles

- Write clean, readable Go code
- Follow standard Go conventions
- Keep functions focused and small
- Use meaningful variable and function names
- Add comments for exported functions and types

### Imports

Organize imports in three groups (standard library first, then external):

```go
import (
    "context"
    "fmt"
    "os"
    "time"

    "charm.land/lipgloss/v2"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
    "github.com/stretchr/testify/assert"
)
```

Use `go fmt` or organize imports automatically.

### Naming Conventions

- **Variables**: `camelCase` (e.g., `configPath`, `isFirstRun`)
- **Constants**: `PascalCase` or `camelCase` for unexported (e.g., `DefaultPlayer`, `maxRetries`)
- **Functions**: `PascalCase` for exported, `camelCase` for unexported (e.g., `LoadConfig`, `isValidURL`)
- **Types/Structs**: `PascalCase` (e.g., `Config`, `EngineConfig`)
- **Interfaces**: `PascalCase` with `er` suffix (e.g., `Reader`, `Writer`)

### Error Handling

- Always handle errors explicitly
- Return meaningful error messages with context
- Use `fmt.Errorf("...: %w", err)` for wrapping errors
- Avoid ignoring errors with `_`

```go
// Good
if err != nil {
    return fmt.Errorf("failed to load config: %w", err)
}

// Good - early return on error
cfg, err := Load(configPath)
if err != nil {
    return fmt.Errorf("error loading configuration: %w", err)
}
```

### Struct Tags

Use struct tags for configuration and serialization:

```go
type Config struct {
    Player string `mapstructure:"player"`
    Engine EngineConfig `mapstructure:"engine"`
    Timeout time.Duration `mapstructure:"timeout"`
}
```

### Testing

- Use `testify` assertions (`assert`, `require`)
- Test file naming: `module_test.go`
- Table-driven tests for multiple test cases

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name string
        input string
        want string
    }{
        {"case 1", "input1", "expected1"},
        {"case 2", "input2", "expected2"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := MyFunction(tt.input)
            assert.Equal(t, tt.want, result)
        })
    }
}
```

### Logging

Use the Charm log package:

```go
import "github.com/charmbracelet/log"

log.Info("starting application")
log.Errorf("failed to connect: %v", err)
```

### CLI Commands

- Use Cobra for CLI commands
- Group related functionality
- Provide helpful usage and descriptions

```go
var rootCmd = &cobra.Command{
    Use:   "aceplay",
    Short: "Play Ace Stream content",
    Long:  `Aceplay is a modern CLI for playing Ace Stream content.`,
}

var playCmd = &cobra.Command{
    Use:   "play [acestream-url]",
    Short: "Play an Ace Stream URL",
    Args:  cobra.ExactArgs(1),
    RunE:  runPlay,
}
```

### Configuration

- Use Viper for configuration management
- Support YAML config files
- Provide sensible defaults
- Allow environment variable overrides

### Dependencies

- Keep dependencies minimal
- Run `go mod tidy` after adding dependencies
- Avoid pulling unused dependencies

---

## Project Structure

```
aceplay/
├── cmd/                  # CLI entry point
│   └── main.go           # Main command
├── internal/
│   ├── acestream/         # Ace Stream client
│   ├── config/           # Configuration management
│   ├── player/           # Video player integration
│   ├── ui/               # UI components (Bubble Tea, Huh)
│   └── notify/           # Desktop notifications
├── pkg/                  # Reusable packages
├── scripts/              # Build/release scripts
├── Makefile              # Build commands
└── go.mod               # Dependencies
```

---

## Common Tasks

### Adding a New Command

1. Create command in appropriate package
2. Register with Cobra in `cmd/aceplay/main.go`
3. Add tests

### Adding Configuration

1. Add field to `Config` struct in `internal/config/config.go`
2. Add default value in `NewConfig()`
3. Use in code via `cfg.FieldName`

### Adding Tests

1. Create `*_test.go` file in same package
2. Use table-driven tests
3. Run with `go test -v -run TestName ./package/...`

---

## Resources

- [Go Documentation](https://go.dev/doc/)
- [Charm Ecosystem](https://charm.sh/)
- [Cobra Documentation](https://github.com/spf13/cobra)
- [Testify](https://github.com/stretchr/testify)
