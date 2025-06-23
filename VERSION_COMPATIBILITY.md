# 🔧 Version Compatibility Update for CometBFT v0.37.4

This document outlines the version compatibility changes made to ensure Fluentum works correctly with CometBFT v0.37.4.

## 📋 Compatibility Requirements

### Target Versions
- **CometBFT**: v0.37.4
- **cometbft-db**: v0.8.0
- **Cosmos SDK**: v0.47.x
- **Go**: 1.20.x (recommended: 1.20.14)

## 🔄 Changes Made

### Updated Dependencies in go.mod

#### Before (Incompatible)
```go
github.com/cosmos/cosmos-sdk v0.50.6
github.com/cometbft/cometbft v0.38.6
github.com/cometbft/cometbft-db v0.9.0
```

#### After (Compatible)
```go
github.com/cosmos/cosmos-sdk v0.47.12
github.com/cometbft/cometbft v0.37.4
github.com/cometbft/cometbft-db v0.8.0
```

### Updated Replace Directives
```go
replace (
    // Ensure Cosmos SDK uses a compatible version for CometBFT v0.37.4
    github.com/cosmos/cosmos-sdk => github.com/cosmos/cosmos-sdk v0.47.12
    
    // Fix CometBFT version compatibility
    github.com/cometbft/cometbft => github.com/cometbft/cometbft v0.37.4
    github.com/cometbft/cometbft-db => github.com/cometbft/cometbft-db v0.8.0
    
    // Fix secp256k1 API compatibility for CometBFT v0.37.4
    github.com/btcsuite/btcd/btcec/v2 => github.com/btcsuite/btcd/btcec/v2 v2.2.1
)
```

## 🚀 Server Deployment Steps

### 1. Install Go 1.20.14 (Required)
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

### 2. Update Dependencies
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

### 3. Verify Compatibility
```bash
# Check that all dependencies are compatible
go list -m all | grep -E "(cometbft|cosmos-sdk)"

# Expected output should show:
# github.com/cometbft/cometbft v0.37.4
# github.com/cometbft/cometbft-db v0.8.0
# github.com/cosmos/cosmos-sdk v0.47.12
```

### 4. Build and Test
```bash
# Clean build
make clean
make build

# Run tests to ensure compatibility
make test

# Validate configuration
./build/fluentum validate-genesis
```

## ⚠️ Important Notes

### Why These Changes Were Necessary

1. **CometBFT v0.37.4 Compatibility**
   - Requires cometbft-db v0.8.0
   - Cosmos SDK v0.50.x introduces breaking changes
   - v0.47.x is the correct version for CometBFT v0.37.x

2. **API Compatibility**
   - Cosmos SDK v0.47.x has stable store interfaces
   - IBC v7.3.1 works well with v0.47.x
   - Protobuf and gRPC versions are compatible

### Potential Issues to Watch For

1. **Store Interface Changes**
   - Cosmos SDK v0.47.x uses stable store patterns
   - Should be compatible with existing code

2. **IBC Compatibility**
   - IBC v7.3.1 should work with Cosmos SDK v0.47.x
   - Monitor for any IBC-related errors

3. **Protobuf Compatibility**
   - Ensure all protobuf definitions are compatible
   - Check for any missing or changed message types

## 🔍 Verification Checklist

### Pre-Deployment
- [ ] Go 1.20.14 installed: `go version`
- [ ] `go mod tidy` runs without errors
- [ ] All tests pass: `make test`
- [ ] Build succeeds: `make build`
- [ ] Configuration validation passes: `./build/fluentum validate-genesis`

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
   - Consider CometBFT v0.36.x if v0.37.4 has issues
   - Use Cosmos SDK v0.46.x as fallback

3. **Contact Support**
   - Check CometBFT GitHub issues
   - Review Cosmos SDK compatibility matrix
   - Consult with development team

---

**✅ Version compatibility update completed!**

The project is now configured for CometBFT v0.37.4 with compatible dependencies. Run `go mod tidy` on the server to finalize the changes. 