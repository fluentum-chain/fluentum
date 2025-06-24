# Fluentum Testnet Quick Start Guide

This guide provides a quick overview of how to run the Fluentum node in testnet mode.

## 🚀 Quick Start (3 Steps)

### 1. Build Fluentum
```bash
make build
make install
```

### 2. Start Testnet (Choose One)

**Option A: Using Makefile (Recommended)**
```bash
# Initialize and start testnet
make init-testnet
make start-testnet

# Or start in background
make start-testnet-bg
```

**Option B: Using Startup Scripts**
```bash
# Linux/macOS
./scripts/start_testnet.sh

# Windows
.\scripts\start_testnet.ps1
```

**Option C: Manual Commands**
```bash
# Initialize
fluentumd init my-testnet --chain-id fluentum-testnet-1

# Start
fluentumd start --testnet --api --grpc --grpc-web
```

### 3. Verify It's Working
```bash
# Check status
fluentumd status

# Check endpoints
curl http://localhost:26657/status
curl http://localhost:1317/cosmos/base/tendermint/v1beta1/node_info
```

## 📋 What's Included

### ✅ Implemented Features

1. **Proper Start Command** - Full Cosmos SDK integration with testnet mode
2. **Configuration Management** - Automatic testnet configuration with faster block times
3. **Multiple Startup Options** - Makefile, scripts, and manual commands
4. **Cross-Platform Support** - Linux/macOS and Windows scripts
5. **Background Operation** - Run node in background with logging
6. **Health Monitoring** - Built-in health checks and status commands
7. **API Endpoints** - RPC, API, gRPC, and gRPC-Web servers
8. **Genesis Management** - Easy genesis account creation
9. **Network Configuration** - P2P, seeds, and persistent peers support

### 🔧 Configuration Options

| Setting | Testnet Default | Description |
|---------|----------------|-------------|
| Chain ID | `fluentum-testnet-1` | Testnet chain identifier |
| Block Time | 1 second | Faster blocks for testing |
| P2P Port | 26656 | Peer-to-peer communication |
| RPC Port | 26657 | Tendermint RPC |
| API Port | 1317 | Cosmos SDK API |
| gRPC Port | 9090 | gRPC server |
| gRPC-Web Port | 9091 | gRPC-Web server |

## 🛠️ Available Commands

### Makefile Commands
```bash
make init-testnet              # Initialize testnet node
make start-testnet             # Start testnet node
make start-testnet-bg          # Start in background
make stop-testnet              # Stop testnet node
make testnet-logs              # Show logs
make reset-testnet             # Reset node data
make testnet-genesis-account name=validator  # Create genesis account
make testnet-script            # Run startup script (Linux/macOS)
make testnet-script-win        # Run startup script (Windows)
```

### Direct Commands
```bash
fluentumd start --testnet                    # Start with testnet mode
fluentumd start --testnet --api --grpc       # Start with all APIs
fluentumd status                             # Check node status
fluentumd tendermint show-node-id            # Show node ID
fluentumd query bank balances <address>      # Check balances
```

## 🌐 Network Endpoints

Once running, your node will be accessible at:

- **RPC**: http://localhost:26657
- **API**: http://localhost:1317
- **gRPC**: localhost:9090
- **gRPC-Web**: localhost:9091
- **P2P**: localhost:26656

## 🧪 Testing Commands

```bash
# Create a test transaction
fluentumd tx fluentum create-fluentum 1 "Test" "Body" --chain-id fluentum-testnet-1 --keyring-backend test -y

# Send tokens
fluentumd tx bank send validator alice 1000ufluentum --chain-id fluentum-testnet-1 --keyring-backend test -y

# Query data
fluentumd query fluentum list-fluentum
fluentumd query bank balances validator
```

## 📁 File Structure

```
fluentum/
├── cmd/fluentum/main.go           # Main application with start command
├── scripts/
│   ├── start_testnet.sh          # Linux/macOS startup script
│   └── start_testnet.ps1         # Windows startup script
├── config/
│   └── testnet-config.toml       # Configuration template
├── docs/
│   └── testnet-setup.md          # Detailed setup guide
├── Makefile                      # Build and management commands
└── README_TESTNET.md             # This quick start guide
```

## 🔍 Troubleshooting

### Common Issues

1. **Port Already in Use**
   ```bash
   lsof -i :26657
   kill -9 <PID>
   ```

2. **Node Won't Start**
   ```bash
   make reset-testnet
   make init-testnet
   make start-testnet
   ```

3. **Can't Connect to Peers**
   ```bash
   # Check P2P configuration
   cat ~/.fluentum/config/config.toml | grep -A 10 "\[p2p\]"
   ```

### Logs and Monitoring
```bash
# View logs
make testnet-logs

# Health check
curl -s http://localhost:26657/status | jq '.result.sync_info'

# Monitor blocks
fluentumd query block
```

## 📚 Next Steps

1. **Read the Full Guide**: See `docs/testnet-setup.md` for detailed instructions
2. **Join the Network**: Connect to other testnet nodes
3. **Develop Applications**: Use the APIs to build on Fluentum
4. **Report Issues**: Create GitHub issues for bugs or improvements

## 🆘 Support

- **Documentation**: `docs/testnet-setup.md`
- **Issues**: GitHub Issues
- **Community**: Discord/Telegram (links in main README)

---

**Happy Testing! 🚀**

The Fluentum testnet is now ready for development and testing. Start building the future of decentralized finance! 