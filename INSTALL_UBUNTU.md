# Fluentum Core Ubuntu Installation Guide

This guide will help you install Fluentum Core on Ubuntu systems.

## Prerequisites

- Ubuntu 20.04 LTS or later
- Non-root user with sudo privileges
- Internet connection

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
```bash
# Download Go 1.24.4
wget https://go.dev/dl/go1.24.4.linux-amd64.tar.gz

# Extract to /usr/local
sudo tar -C /usr/local -xzf go1.24.4.linux-amd64.tar.gz

# Add Go to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify installation
go version
```

#### Step 4: Build and Install Fluentum
```bash
# Clone the repository (if not already done)
git clone https://github.com/fluentum-chain/fluentum.git
cd fluentum

# Build Fluentum Core with BadgerDB (no CGO dependencies)
CGO_ENABLED=0 BUILD_TAGS="tendermint,badgerdb" make build

# Install Fluentum Core
CGO_ENABLED=0 BUILD_TAGS="tendermint,badgerdb" make install
```

#### Step 5: Verify Installation
```bash
fluentum version
```

## Usage

After installation, you can use the following commands:

### Basic Commands
```bash
fluentum version          # Check version
fluentum --help          # Show all available commands
fluentum init            # Initialize a new node
fluentum node            # Start the node
fluentum testnet         # Generate testnet configuration
```

### Initialize a New Node
```bash
# Initialize with default configuration
fluentum init

# Initialize with custom home directory
fluentum init --home /path/to/custom/directory
```

### Start the Node
```bash
# Start with default configuration
fluentum node

# Start with custom configuration
fluentum node --home /path/to/config
```

### Generate Testnet
```bash
# Generate a 4-validator testnet
fluentum testnet

# Generate with custom parameters
fluentum testnet -v 8 -o ./my-testnet -chain-id my-chain
```

## Configuration

The default configuration files are created in `~/.fluentum/` when you run `fluentum init`.

### Key Configuration Files
- `~/.fluentum/config/config.toml` - Main configuration
- `~/.fluentum/config/genesis.json` - Genesis block configuration
- `~/.fluentum/config/node_key.json` - Node private key
- `~/.fluentum/config/priv_validator_key.json` - Validator private key

## Database Backends

Fluentum Core supports multiple database backends:

- **BadgerDB** (Default) - Pure Go implementation, no CGO required
- **LevelDB** - Requires CGO but widely supported
- **RocksDB** - High performance, requires CGO and additional dependencies
- **BoltDB** - Pure Go implementation

The installation script uses **BadgerDB** by default to avoid CGO dependencies and external library requirements.

## Troubleshooting

### Common Issues

1. **Permission Denied**
   ```bash
   sudo chown -R $USER:$USER ~/.fluentum/
   ```

2. **Go not found**
   ```bash
   export PATH=$PATH:/usr/local/go/bin
   source ~/.bashrc
   ```

3. **Build fails with RocksDB error**
   ```bash
   # Use BadgerDB instead (recommended)
   CGO_ENABLED=0 BUILD_TAGS="tendermint,badgerdb" make build
   
   # Or install RocksDB dependencies
   sudo apt install -y librocksdb-dev
   ```

4. **Port already in use**
   ```bash
   # Check what's using the port
   sudo netstat -tulpn | grep :26656
   
   # Kill the process or change port in config.toml
   ```

### Getting Help

- Check the logs: `tail -f ~/.fluentum/logs/fluentum.log`
- Run with verbose output: `fluentum node --log_level=debug`
- Check system resources: `htop` or `top`

## Development

For development purposes, you can also build without installation:

```bash
# Build only (creates binary in build/fluentum)
CGO_ENABLED=0 BUILD_TAGS="tendermint,badgerdb" make build

# Run the binary directly
./build/fluentum version
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