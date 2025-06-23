# Fluentum Core Ubuntu Installation Guide

This guide will help you install Fluentum Core on Ubuntu systems.

## Prerequisites

- Ubuntu 20.04 LTS or later
- Non-root user with sudo privileges
- Internet connection

## Go Version Requirements

**⚠️ Important**: The project supports two Go version configurations:

### Option 1: Go 1.20.x (Recommended for Cosmos SDK v0.47.x)
- **Compatibility**: Cosmos SDK v0.47.12, CometBFT v0.37.2
- **Dependencies**: Pinned for Go 1.20 compatibility
- **Status**: ✅ Stable and tested

### Option 2: Go 1.22+ (For newer dependencies)
- **Compatibility**: Latest Cosmos SDK and CometBFT versions
- **Dependencies**: Auto-upgraded to latest compatible versions
- **Status**: ✅ Supported but may require dependency updates

## Quick Installation

### Option 1: Automated Installation Script

1. Clone the repository:
```bash
git clone https://github.com/fluentum-chain/fluentum.git
cd fluentum
```

2. Make the installation script executable:
```bash
chmod +x install-ubuntu.sh
```

3. Run the installation script:
```bash
./install-ubuntu.sh
```

### Option 2: Manual Installation

#### Step 1: Update System Packages
```bash
sudo apt update
sudo apt upgrade -y
```

#### Step 2: Install Dependencies
```bash
sudo apt install -y \
    build-essential \
    git \
    curl \
    wget \
    pkg-config \
    libssl-dev \
    libgmp-dev \
    libtool \
    autoconf \
    automake \
    cmake \
    clang \
    clang-format
```

#### Step 3: Install Go

**For Go 1.20.x (Recommended):**
```bash
# Download Go 1.20.14
wget https://go.dev/dl/go1.20.14.linux-amd64.tar.gz

# Extract to /usr/local
sudo tar -C /usr/local -xzf go1.20.14.linux-amd64.tar.gz

# Add Go to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify installation
go version
# Expected: go version go1.20.14 linux/amd64
```

**For Go 1.22+ (Newer dependencies):**
```bash
# Download Go 1.22.0
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz

# Extract to /usr/local
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz

# Add Go to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify installation
go version
# Expected: go version go1.22.0 linux/amd64
```

#### Step 4: Build and Install Fluentum
```bash
# Clone the repository (if not already done)
git clone https://github.com/fluentum-chain/fluentum.git
cd fluentum

# Install dependencies
go mod tidy

# Build Fluentum Core
make build

# Install Fluentum Core
make install
```

#### Step 5: Verify Installation
```bash
fluentumd version
```

## Usage

After installation, you can use the following commands:

### Basic Commands
```bash
fluentumd version          # Check version
fluentumd --help          # Show all available commands
fluentumd init            # Initialize a new node
fluentumd start           # Start the node
fluentumd testnet         # Generate testnet configuration
```

### Initialize a New Node
```bash
# Initialize with default configuration
fluentumd init --home ~/.cometbft

# Initialize with custom home directory
fluentumd init --home /path/to/custom/directory
```

### Start the Node
```bash
# Start with default configuration
fluentumd start --home ~/.cometbft

# Start with custom configuration
fluentumd start --home /path/to/config
```

### Generate Testnet
```bash
# Generate a 4-validator testnet
fluentumd testnet

# Generate with custom parameters
fluentumd testnet -v 8 -o ./my-testnet -chain-id my-chain
```

## Configuration

The default configuration files are created in `~/.cometbft/` when you run `fluentumd init`.

### Key Configuration Files
- `~/.cometbft/config/config.toml` - Main configuration
- `~/.cometbft/config/genesis.json` - Genesis block configuration
- `~/.cometbft/config/node_key.json` - Node private key
- `~/.cometbft/config/priv_validator_key.json` - Validator private key

## Database Backends

Fluentum Core supports multiple database backends:

- **PebbleDB** (Recommended) - High performance, Go implementation
- **LevelDB** - Requires CGO but widely supported
- **RocksDB** - High performance, requires CGO and additional dependencies
- **BoltDB** - Pure Go implementation

The installation uses **PebbleDB** by default for optimal performance.

## Dependency Management

### Current Pinned Versions (Go 1.20)

| Dependency | Version | Purpose |
|------------|---------|---------|
| **CometBFT** | v0.37.2 | Consensus engine |
| **Cosmos SDK** | v0.47.12 | Application framework |
| **gRPC** | v1.59.0 | RPC communication |
| **cosmossdk.io/store** | v1.0.2 | State management |

### Upgrading Dependencies (Go 1.22+)

If using Go 1.22+, dependencies will automatically upgrade:

```bash
# Update to latest compatible versions
go mod tidy

# Rebuild with new dependencies
make build
```

## Troubleshooting

### Common Issues

1. **Permission Denied**
   ```bash
   sudo chown -R $USER:$USER ~/.cometbft/
   ```

2. **Go not found**
   ```bash
   export PATH=$PATH:/usr/local/go/bin
   source ~/.bashrc
   ```

3. **Build fails with dependency errors**
   ```bash
   # For Go 1.20 compatibility issues
   go mod tidy
   make build
   
   # For newer Go versions
   go mod tidy
   make build
   ```

4. **Port already in use**
   ```bash
   # Check what's using the port
   sudo netstat -tulpn | grep :26656
   
   # Kill the process or change port in config.toml
   ```

5. **Go version compatibility errors**
   ```bash
   # Check current Go version
   go version
   
   # If using Go 1.22+ but need 1.20 compatibility
   # Downgrade Go or use go.mod replace directives
   ```

### Getting Help

- Check the logs: `tail -f ~/.cometbft/logs/fluentumd.log`
- Run with verbose output: `fluentumd start --log_level=debug --home ~/.cometbft`
- Check system resources: `htop` or `top`

## Development

For development purposes, you can also build without installation:

```bash
# Build only (creates binary in build/fluentumd)
make build

# Run the binary directly
./build/fluentumd version
```

## Security Notes

- Never run Fluentum as root
- Keep your private keys secure
- Regularly update your system and Fluentum
- Use firewall rules to restrict access to RPC endpoints

## Support

For issues and questions:
- GitHub Issues: https://github.com/fluentum-chain/fluentum/issues
- Documentation: https://docs.fluentum.tech
- Community: https://t.me/fluentum 