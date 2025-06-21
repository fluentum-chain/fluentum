# Fluentum Core Build Automation

This document explains the automated build system for Fluentum Core, which handles dependency management automatically to prevent build failures.

## Overview

The build automation system ensures that:
- Dependencies are properly managed before building
- `go mod tidy` is automatically run when needed
- Build failures due to missing `go.sum` entries are prevented
- CI/CD pipelines can be optimized for different scenarios

## Quick Start

### For Development
```bash
# Full build with automatic dependency management
make build

# Or use the build script
./scripts/build.sh          # Linux/Mac
.\scripts\build.ps1         # Windows
```

### For CI/CD
```bash
# When dependencies are pre-managed
make build-only

# Or use the build script with flags
./scripts/build.sh --skip-deps --skip-tests
```

## Makefile Targets

### Dependency Management
- `make deps` - Download and tidy dependencies
- `make deps-check` - Check dependencies without modifying files

### Build Targets
- `make build` - Build with automatic dependency management (recommended)
- `make build-only` - Build without dependency management (for CI/CD)
- `make install` - Install with automatic dependency management
- `make install-only` - Install without dependency management

### Other Targets
- `make clean` - Clean build artifacts
- `make test` - Run tests
- `make lint` - Run linter
- `make help` - Show all available targets

## Build Scripts

### Linux/Mac (Bash)
```bash
# Full build
./scripts/build.sh

# Build without dependency management
./scripts/build.sh --skip-deps

# Build without tests
./scripts/build.sh --skip-tests

# Quick build (no deps, no tests)
./scripts/build.sh --skip-deps --skip-tests

# Show help
./scripts/build.sh --help
```

### Windows (PowerShell)
```powershell
# Full build
.\scripts\build.ps1

# Build without dependency management
.\scripts\build.ps1 -SkipDeps

# Build without tests
.\scripts\build.ps1 -SkipTests

# Quick build (no deps, no tests)
.\scripts\build.ps1 -SkipDeps -SkipTests

# Show help
.\scripts\build.ps1 -Help
```

## CI/CD Integration

### GitHub Actions Example
```yaml
name: Build Fluentum Core

on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24.4'
          
      - name: Cache dependencies
        uses: actions/cache@v3
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: |
            ${{ runner.os }}-go-
            
      - name: Build
        run: make build-only  # Use build-only since deps are cached
```

### Docker Example
```dockerfile
FROM golang:1.24.4-alpine AS builder

WORKDIR /app
COPY . .

# Install make
RUN apk add --no-cache make

# Build with automatic dependency management
RUN make build

# Final stage
FROM alpine:latest
COPY --from=builder /app/build/fluentum /usr/local/bin/
CMD ["fluentum"]
```

## Troubleshooting

### Common Issues

#### 1. Missing go.sum entries
**Error:** `missing go.sum entry for module providing package...`

**Solution:** Run dependency management
```bash
make deps
# or
go mod tidy
```

#### 2. Build fails after updating go.mod
**Error:** Build fails with dependency errors

**Solution:** The build system should handle this automatically, but if not:
```bash
make deps
make build
```

#### 3. CI/CD build is slow
**Solution:** Use `build-only` target when dependencies are pre-managed
```bash
make build-only
```

### Dependency Management Commands

```bash
# Manual dependency management
go mod download    # Download dependencies
go mod tidy        # Tidy and update go.sum
go mod verify      # Verify dependencies

# Or use the make target
make deps          # All of the above
```

## Best Practices

### For Developers
1. Always use `make build` for local development
2. Run `make deps` after pulling changes that modify `go.mod`
3. Use `make test` to verify your changes
4. Run `make lint` before committing

### For CI/CD
1. Use `make build-only` when dependencies are cached
2. Cache the Go module cache to speed up builds
3. Use the build scripts for more control and better error reporting
4. Consider using `--skip-tests` for quick builds in development branches

### For Release Management
1. Use `make dist` for creating distribution builds
2. Always run full tests before releases
3. Verify dependencies with `make deps-check`

## Environment Variables

The build system respects these environment variables:

- `CGO_ENABLED` - Enable/disable CGO (default: 0)
- `BUILD_TAGS` - Build tags for conditional compilation
- `LDFLAGS` - Additional linker flags
- `TENDERMINT_BUILD_OPTIONS` - Build options (nostrip, race, cleveldb, etc.)

## Monitoring and Debugging

### Verbose Output
```bash
# Enable verbose make output
make build V=1

# Use verbose build script
./scripts/build.sh --verbose
```

### Dependency Analysis
```bash
# Check dependency graph
make draw_deps

# Analyze binary size
make get_deps_bin_size
```

## Migration Guide

### From Manual Build Process
**Before:**
```bash
go mod tidy
go build ./cmd/fluentum
```

**After:**
```bash
make build
```

### From Old Makefile
**Before:**
```bash
make build  # Might fail due to missing deps
```

**After:**
```bash
make build  # Automatically handles deps
make build-only  # For CI/CD when deps are ready
```

## Support

If you encounter issues with the build automation:

1. Check the troubleshooting section above
2. Run `make help` for available targets
3. Use the build scripts for detailed error reporting
4. Check the logs for specific error messages

For more information, see the main [CONTRIBUTING.md](CONTRIBUTING.md) file. 