# Fluentum Testnet Setup Guide

This guide will help you set up and run a Fluentum node in testnet mode for development and testing purposes.

## 🚀 Quick Start

### Prerequisites

1. **Build Fluentum Core**
   ```bash
   # Build the binary
   make build
   
   # Install the binary
   make install
   ```

2. **Verify Installation**
   ```bash
   fluentumd version
   ```

### Option 1: Using the Startup Scripts

#### Linux/macOS
```bash
# Make the script executable
chmod +x scripts/start_testnet.sh

# Start with default settings
./scripts/start_testnet.sh

# Start with custom configuration
./scripts/start_testnet.sh -m my-node -c test-chain-1 -s node1:26656,node2:26656 -b
```

#### Windows
```powershell
# Start with default settings
.\scripts\start_testnet.ps1

# Start with custom configuration
.\scripts\start_testnet.ps1 -Moniker my-node -ChainId test-chain-1 -Seeds "node1:26656,node2:26656" -Background
```

### Option 2: Manual Setup

#### 1. Initialize the Node
```bash
# Initialize with default settings
fluentumd init my-testnet-node --chain-id fluentum-testnet-1

# Or with custom home directory
fluentumd init my-testnet-node --chain-id fluentum-testnet-1 --home ~/.fluentum-testnet
```

#### 2. Configure the Node

Edit the configuration files:

**config.toml** (Tendermint configuration):
```toml
# Node identification
moniker = "my-testnet-node"

# P2P Configuration
[p2p]
laddr = "tcp://0.0.0.0:26656"
external_address = ""
seeds = "node1:26656,node2:26656"  # Add your seed nodes
persistent_peers = "node3:26656,node4:26656"  # Add persistent peers

# RPC Configuration
[rpc]
laddr = "tcp://0.0.0.0:26657"
cors_allowed_origins = ["*"]

# Consensus Configuration (faster for testnet)
[consensus]
timeout_commit = "1s"
timeout_propose = "1s"
create_empty_blocks = true
create_empty_blocks_interval = "10s"
```

**app.toml** (Application configuration):
```toml
# API Configuration
[api]
enable = true
swagger = true
address = "tcp://0.0.0.0:1317"

# gRPC Configuration
[grpc]
enable = true
address = "0.0.0.0:9090"

# gRPC-Web Configuration
[grpc-web]
enable = true
address = "0.0.0.0:9091"
```

#### 3. Create Genesis Account (Optional)
```bash
# Add a key
fluentumd keys add validator --keyring-backend test

# Add genesis account
fluentumd add-genesis-account $(fluentumd keys show validator -a --keyring-backend test) 1000000000ufluentum,1000000000stake --keyring-backend test
```

#### 4. Start the Node
```bash
# Start in foreground
fluentumd start --testnet

# Start in background
nohup fluentumd start --testnet > fluentum.log 2>&1 &
```

## 🔧 Configuration Options

### Command Line Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--testnet` | Enable testnet mode with faster block times | false |
| `--api` | Enable the API server | false |
| `--grpc` | Enable the gRPC server | false |
| `--grpc-web` | Enable the gRPC-Web server | false |
| `--home` | Home directory | `~/.fluentum` |
| `--chain-id` | Chain ID | `fluentum-testnet-1` |
| `--moniker` | Node moniker | `fluentum-testnet-node` |

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `FLUENTUM_HOME` | Home directory | `~/.fluentum` |
| `FLUENTUM_CHAIN_ID` | Chain ID | `fluentum-testnet-1` |
| `FLUENTUM_MONIKER` | Node moniker | `fluentum-testnet-node` |

## 🌐 Network Configuration

### Default Ports

| Service | Port | Description |
|---------|------|-------------|
| P2P | 26656 | Peer-to-peer communication |
| RPC | 26657 | Tendermint RPC |
| API | 1317 | Cosmos SDK API |
| gRPC | 9090 | gRPC server |
| gRPC-Web | 9091 | gRPC-Web server |

### Firewall Configuration

```bash
# Linux/macOS
sudo ufw allow 26656/tcp  # P2P
sudo ufw allow 26657/tcp  # RPC
sudo ufw allow 1317/tcp   # API
sudo ufw allow 9090/tcp   # gRPC
sudo ufw allow 9091/tcp   # gRPC-Web

# Windows
netsh advfirewall firewall add rule name="Fluentum P2P" dir=in action=allow protocol=TCP localport=26656
netsh advfirewall firewall add rule name="Fluentum RPC" dir=in action=allow protocol=TCP localport=26657
netsh advfirewall firewall add rule name="Fluentum API" dir=in action=allow protocol=TCP localport=1317
netsh advfirewall firewall add rule name="Fluentum gRPC" dir=in action=allow protocol=TCP localport=9090
netsh advfirewall firewall add rule name="Fluentum gRPC-Web" dir=in action=allow protocol=TCP localport=9091
```

## 🔍 Verification Commands

### Check Node Status
```bash
# Check if node is running
ps aux | grep fluentumd

# Check node info via RPC
curl -s http://localhost:26657/status | jq

# Check node info via CLI
fluentumd status

# Check node ID
fluentumd tendermint show-node-id
```

### Check Network Connectivity
```bash
# Check if ports are open
netstat -tulpn | grep fluentumd

# Check P2P connections
curl -s http://localhost:26657/net_info | jq

# Check consensus state
curl -s http://localhost:26657/consensus_state | jq
```

### Check Application State
```bash
# Check account balances
fluentumd query bank balances $(fluentumd keys show validator -a --keyring-backend test)

# Check Fluentum records
fluentumd query fluentum list-fluentum

# Check module parameters
fluentumd query fluentum params
```

## 🧪 Testing the Testnet

### Create a Test Transaction
```bash
# Create a Fluentum record
fluentumd tx fluentum create-fluentum 1 "Test Title" "Test Body" --chain-id fluentum-testnet-1 --keyring-backend test -y

# Send tokens
fluentumd tx bank send validator $(fluentumd keys show alice -a --keyring-backend test) 1000ufluentum --chain-id fluentum-testnet-1 --keyring-backend test -y
```

### Monitor Transactions
```bash
# Watch for new blocks
fluentumd query block

# Watch for new transactions
fluentumd query txs --events 'tx.height>0'

# Monitor logs
tail -f fluentum.log
```

## 🔧 Troubleshooting

### Common Issues

#### 1. Port Already in Use
```bash
# Check what's using the port
lsof -i :26657

# Kill the process
kill -9 <PID>
```

#### 2. Node Won't Start
```bash
# Check logs
tail -f fluentum.log

# Reset node (WARNING: This will delete all data)
fluentumd tendermint unsafe-reset-all

# Check configuration
fluentumd validate-genesis
```

#### 3. Can't Connect to Peers
```bash
# Check P2P configuration
cat ~/.fluentum/config/config.toml | grep -A 10 "\[p2p\]"

# Check if ports are open
telnet localhost 26656

# Check firewall settings
sudo ufw status
```

#### 4. API Not Accessible
```bash
# Check if API is enabled
cat ~/.fluentum/config/app.toml | grep -A 5 "\[api\]"

# Test API endpoint
curl -s http://localhost:1317/cosmos/base/tendermint/v1beta1/node_info
```

### Log Analysis

```bash
# View recent logs
tail -n 100 fluentum.log

# Search for errors
grep -i error fluentum.log

# Search for specific events
grep -i "new block" fluentum.log
grep -i "peer" fluentum.log
```

## 📊 Monitoring

### Health Check Script
```bash
#!/bin/bash
echo "=== Fluentum Testnet Health Check ==="
echo "Time: $(date)"
echo ""

# Check if process is running
if pgrep -x "fluentumd" > /dev/null; then
    echo "✅ Process: Running"
else
    echo "❌ Process: Not Running"
    exit 1
fi

# Check RPC endpoint
if curl -s http://localhost:26657/status > /dev/null; then
    echo "✅ RPC: Accessible"
else
    echo "❌ RPC: Not Accessible"
fi

# Check API endpoint
if curl -s http://localhost:1317/cosmos/base/tendermint/v1beta1/node_info > /dev/null; then
    echo "✅ API: Accessible"
else
    echo "❌ API: Not Accessible"
fi

# Get latest block
LATEST_BLOCK=$(curl -s http://localhost:26657/block | jq -r '.result.block.header.height // "N/A"')
echo "📦 Latest Block: $LATEST_BLOCK"

# Get peer count
PEER_COUNT=$(curl -s http://localhost:26657/net_info | jq -r '.result.n_peers // 0')
echo "🌐 Connected Peers: $PEER_COUNT"
```

### Prometheus Metrics

Enable Prometheus metrics in `config.toml`:
```toml
[instrumentation]
prometheus = true
prometheus_listen_addr = ":26660"
```

## 🔄 Updating the Testnet

### Stop the Node
```bash
# If running in foreground, use Ctrl+C
# If running in background
kill $(cat fluentum-testnet.pid)
```

### Update Configuration
```bash
# Edit configuration files
nano ~/.fluentum/config/config.toml
nano ~/.fluentum/config/app.toml
```

### Restart the Node
```bash
# Start again
./scripts/start_testnet.sh
```

## 📚 Additional Resources

- [Fluentum Documentation](../README.md)
- [Cosmos SDK Documentation](https://docs.cosmos.network/)
- [Tendermint Documentation](https://docs.tendermint.com/)
- [Community Discord](https://discord.gg/fluentum)

## 🆘 Support

If you encounter issues:

1. Check the troubleshooting section above
2. Search existing issues on GitHub
3. Create a new issue with:
   - Your operating system
   - Fluentum version (`fluentumd version`)
   - Error logs
   - Steps to reproduce 