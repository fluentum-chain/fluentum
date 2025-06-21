# Fluentum Core Server Setup Guide

This guide helps you set up and build Fluentum Core on your server, including solutions for common issues.

## Quick Setup

### 1. Clone the Repository
```bash
git clone https://github.com/fluentum-chain/fluentum.git
cd fluentum
```

### 2. Fix Script Permissions
```bash
# Option 1: Use the fix script
chmod +x scripts/fix_permissions.sh
./scripts/fix_permissions.sh

# Option 2: Manual fix
chmod +x scripts/build.sh
```

### 3. Build the Project
```bash
# Option 1: Use the build script
./scripts/build.sh

# Option 2: Use make (recommended)
make build
```

## Common Issues and Solutions

### Issue 1: Permission Denied Error
**Error:** `-bash: ./scripts/build.sh: Permission denied`

**Solutions:**
```bash
# Fix permissions
chmod +x scripts/build.sh

# Or use bash directly
bash scripts/build.sh

# Or use make instead
make build
```

### Issue 2: Missing Dependencies
**Error:** `missing go.sum entry for module providing package...`

**Solutions:**
```bash
# Automatic fix (recommended)
make deps

# Manual fix
go mod tidy
go mod download
go mod verify
```

### Issue 3: Go Version Issues
**Error:** `go: go.mod requires go >= 1.24.4`

**Solutions:**
```bash
# Check Go version
go version

# Install/update Go if needed
# For Ubuntu/Debian:
sudo apt update
sudo apt install golang-go

# For CentOS/RHEL:
sudo yum install golang

# Or download from https://golang.org/dl/
```

### Issue 4: Missing Make
**Error:** `make: command not found`

**Solutions:**
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install make

# CentOS/RHEL
sudo yum install make

# macOS
xcode-select --install
```

## Build Options

### Development Build
```bash
# Full build with dependency management
make build

# Or use the build script
./scripts/build.sh
```

### CI/CD Build
```bash
# When dependencies are pre-managed
make build-only

# Or use the build script
./scripts/build.sh --skip-deps --skip-tests
```

### Quick Build
```bash
# Build without tests
./scripts/build.sh --skip-tests

# Build without dependency management
./scripts/build.sh --skip-deps
```

## Server-Specific Commands

### Ubuntu/Debian Server
```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install dependencies
sudo apt install -y git make golang-go

# Clone and build
git clone https://github.com/fluentum-chain/fluentum.git
cd fluentum
chmod +x scripts/build.sh
make build
```

### CentOS/RHEL Server
```bash
# Update system
sudo yum update -y

# Install dependencies
sudo yum install -y git make golang

# Clone and build
git clone https://github.com/fluentum-chain/fluentum.git
cd fluentum
chmod +x scripts/build.sh
make build
```

### Docker Server
```bash
# Build using Docker
docker build -t fluentum .

# Or run in container
docker run -it --rm -v $(pwd):/app -w /app golang:1.24.4-alpine sh -c "apk add --no-cache make && make build"
```

## Verification

### Check Build Success
```bash
# Verify binary exists
ls -la build/fluentum

# Check binary version
./build/fluentum version

# Test basic functionality
./build/fluentum help
```

### Check Dependencies
```bash
# Verify dependencies
make deps-check

# Or manually
go mod verify
```

## Troubleshooting

### Debug Build Issues
```bash
# Verbose build output
make build V=1

# Or use build script with verbose
./scripts/build.sh --verbose
```

### Check System Requirements
```bash
# Check Go version
go version

# Check available memory
free -h

# Check disk space
df -h

# Check CPU info
nproc
```

### Common Error Messages

#### "go: module lookup disabled by GOPROXY=off"
```bash
# Fix by setting GOPROXY
export GOPROXY=https://proxy.golang.org,direct
```

#### "fatal: not a git repository"
```bash
# Initialize git if needed
git init
git remote add origin https://github.com/fluentum-chain/fluentum.git
```

#### "make: *** No rule to make target 'build'"
```bash
# Check if Makefile exists
ls -la Makefile

# If missing, re-clone the repository
cd ..
rm -rf fluentum
git clone https://github.com/fluentum-chain/fluentum.git
cd fluentum
```

## Performance Optimization

### For Large Servers
```bash
# Set Go build flags for better performance
export GOFLAGS="-buildmode=exe -ldflags=-s -ldflags=-w"

# Use multiple cores for building
export GOMAXPROCS=$(nproc)
```

### For Small Servers
```bash
# Limit memory usage
export GOGC=50

# Use single core
export GOMAXPROCS=1
```

## Security Considerations

### Run as Non-Root User
```bash
# Create dedicated user
sudo useradd -m -s /bin/bash fluentum
sudo usermod -aG sudo fluentum

# Switch to user
sudo su - fluentum

# Clone and build as user
git clone https://github.com/fluentum-chain/fluentum.git
cd fluentum
make build
```

### Verify Binary Integrity
```bash
# Check binary checksum
sha256sum build/fluentum

# Verify it's not malicious
file build/fluentum
```

## Support

If you encounter issues:

1. Check this guide for common solutions
2. Run `make help` for available commands
3. Check the [BUILD_AUTOMATION.md](BUILD_AUTOMATION.md) for detailed build information
4. Review the [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines

For additional help, please check the project documentation or create an issue on GitHub. 