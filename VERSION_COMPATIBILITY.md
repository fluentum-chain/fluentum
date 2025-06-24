# 🔧 Version Compatibility Guide

This document outlines the version compatibility requirements and options for Fluentum Core.

## 📋 Compatibility Requirements

### Current Supported Configurations

#### Option 1: Go 1.20.x (Recommended for Cosmos SDK v0.47.x)
- **CometBFT**: v0.37.2
- **cometbft-db**: v0.8.0
- **Cosmos SDK**: v0.47.12
- **Go**: 1.20.x (recommended: 1.20.14)
- **Status**: ✅ Stable and tested

#### Option 2: Go 1.22+ (For newer dependencies)
- **CometBFT**: v0.38+
- **cometbft-db**: v0.9.0+
- **Cosmos SDK**: v0.50+
- **Go**: 1.22+ (recommended: 1.22.0)
- **Status**: ✅ Supported but may require dependency updates

## 🔄 Current Dependencies in go.mod

### Pinned Versions (Go 1.20 Compatible)
```go
github.com/cosmos/cosmos-sdk v0.47.12
github.com/cometbft/cometbft v0.37.2
github.com/cometbft/cometbft-db v0.8.0
google.golang.org/grpc v1.59.0
golang.org/x/exp v0.0.0-20231110203233-9a3e6036ecaa
golang.org/x/sys v0.15.0
cosmossdk.io/log v1.3.1
cosmossdk.io/store v1.0.2
```

### Replace Directives
```go
replace (
    // Ensure Cosmos SDK uses a compatible version for CometBFT v0.37.2
    github.com/cosmos/cosmos-sdk => github.com/cosmos/cosmos-sdk v0.47.12
    
    // Fix CometBFT version compatibility
    github.com/cometbft/cometbft => github.com/cometbft/cometbft v0.37.2
    github.com/cometbft/cometbft-db => github.com/cometbft/cometbft-db v0.8.0
    
    // Fix secp256k1 API compatibility for CometBFT v0.37.2
    github.com/btcsuite/btcd/btcec/v2 => github.com/btcsuite/btcd/btcec/v2 v2.2.1
    
    // Redirect cosmossdk.io packages to GitHub
    cosmossdk.io/core => github.com/cosmos/cosmos-sdk/core v0.11.0
    cosmossdk.io/db => github.com/cosmos/cosmos-sdk/db v0.11.0
)
```

## 🚀 Server Deployment Steps

### Option 1: Go 1.20.x Deployment (Recommended)

#### 1. Install Go 1.20.14
```bash
# For Ubuntu/Debian
wget https://go.dev/dl/go1.20.14.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.20.14.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# For CentOS/RHEL
wget https://go.dev/dl/go1.20.14.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.20.14.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify installation
go version
# Expected: go version go1.20.14 linux/amd64
```

#### 2. Update Dependencies
```bash
# Navigate to project directory
cd /path/to/fluentum

# Update go.mod with new versions
git pull origin main

# Clean and update dependencies
go mod tidy
go mod download
go mod verify
```

#### 3. Verify Compatibility
```bash
# Check that all dependencies are compatible
go list -m all | grep -E "(cometbft|cosmos-sdk)"

# Expected output should show:
# github.com/cometbft/cometbft v0.37.2
# github.com/cometbft/cometbft-db v0.8.0
# github.com/cosmos/cosmos-sdk v0.47.12
```

### Option 2: Go 1.22+ Deployment (Newer Dependencies)

#### 1. Install Go 1.22+
```bash
# Download Go 1.22.0
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify installation
go version
# Expected: go version go1.22.0 linux/amd64
```

#### 2. Upgrade Dependencies
```bash
# Navigate to project directory
cd /path/to/fluentum

# Update dependencies to latest compatible versions
go mod tidy

# This will automatically upgrade to:
# - CometBFT v0.38+
# - Cosmos SDK v0.50+
# - gRPC v1.71+
```

### 4. Build and Test
```bash
# Clean build
make clean
make build

# Run tests to ensure compatibility
make test

# Validate configuration
./build/fluentumd validate-genesis
```

## ⚠️ Important Notes

### Why These Changes Were Necessary

1. **Go Version Compatibility**
   - Go 1.20.x is required for Cosmos SDK v0.47.x
   - Go 1.22+ enables newer dependency features (`cmp`, `maps`, `slices`, `math/rand/v2`)
   - Dependencies are pinned for Go 1.20 compatibility

2. **CometBFT v0.37.2 Compatibility**
   - Requires cometbft-db v0.8.0
   - Cosmos SDK v0.50.x introduces breaking changes
   - v0.47.x is the correct version for CometBFT v0.37.x

3. **API Compatibility**
   - Cosmos SDK v0.47.x has stable store interfaces
   - IBC v7.3.1 works well with v0.47.x
   - Protobuf and gRPC versions are compatible

### Potential Issues to Watch For

1. **Go Version Conflicts**
   - Dependencies requiring Go 1.22+ features may cause build errors
   - Use `go mod tidy` to resolve version conflicts
   - Consider upgrading Go version if needed

2. **Store Interface Changes**
   - Cosmos SDK v0.47.x uses stable store patterns
   - Should be compatible with existing code

3. **IBC Compatibility**
   - IBC v7.3.1 should work with Cosmos SDK v0.47.x
   - Monitor for any IBC-related errors

4. **Protobuf Compatibility**
   - Ensure all protobuf definitions are compatible
   - Check for any missing or changed message types

## 🔍 Verification Checklist

### Pre-Deployment
- [ ] Go version installed: `go version`
- [ ] `go mod tidy` runs without errors
- [ ] All tests pass: `make test`
- [ ] Build succeeds: `make build`
- [ ] Configuration validation passes: `./build/fluentumd validate-genesis`

### Post-Deployment
- [ ] Node starts successfully
- [ ] RPC endpoints respond correctly
- [ ] Block synchronization works
- [ ] No compatibility errors in logs

## 📚 Additional Resources

### Official Compatibility Matrix
- [CometBFT v0.37.x Compatibility](https://docs.cometbft.com/v0.37/)
- [Cosmos SDK v0.47.x Documentation](https://docs.cosmos.network/v0.47/)
- [IBC v7.x Compatibility](https://ibc.cosmos.network/)

### Migration Guides
- [Cosmos SDK v0.47.x Migration Guide](https://docs.cosmos.network/v0.47/migrations/upgrading)
- [CometBFT v0.37.x Migration Guide](https://docs.cometbft.com/v0.37/migrations/)

## 🚨 Rollback Plan

If compatibility issues arise:

1. **Immediate Rollback**
   ```bash
   git checkout HEAD~1 go.mod go.sum
   go mod tidy
   make build
   ```

2. **Alternative Versions**
   - Consider CometBFT v0.36.x if v0.37.2 has issues
   - Use Cosmos SDK v0.46.x as fallback

3. **Go Version Downgrade**
   ```bash
   # If using Go 1.22+ and need 1.20 compatibility
   wget https://go.dev/dl/go1.20.14.linux-amd64.tar.gz
   sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.20.14.linux-amd64.tar.gz
   go mod tidy
   make build
   ```

4. **Contact Support**
   - Check CometBFT GitHub issues
   - Review Cosmos SDK compatibility matrix
   - Consult with development team

## 🔧 Dependency Management

### Upgrading Dependencies

To upgrade to newer versions with Go 1.22+:

```bash
# 1. Upgrade Go to 1.22+
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz

# 2. Update dependencies
go mod tidy

# 3. Rebuild
make build
```

### Pinning Dependencies

To pin dependencies for Go 1.20 compatibility:

```bash
# Pin specific versions
go get github.com/cometbft/cometbft@v0.37.2
go get github.com/cosmos/cosmos-sdk@v0.47.12
go get github.com/cometbft/cometbft-db@v0.8.0

# Update go.mod
go mod tidy
```

## Migration Notes

- ABCI++: `DeliverTx` replaced by `FinalizeBlock`. (Legacy: `BeginBlock`, `DeliverTx`, and `EndBlock` are no longer used in ABCI 2.0.)
- IAVL: Use `NewMutableTree(db, cacheSize, true)` for v1.0+.
- Run `go mod tidy` after dependency changes.

---

**✅ Version compatibility guide updated!**

The project supports both Go 1.20.x (recommended) and Go 1.22+ configurations. Choose the appropriate version based on your deployment requirements. 