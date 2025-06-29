# Fluentum Node Setup Guide

This guide explains the fixes made to resolve the node key error and provides instructions for setting up and running the Fluentum node.

## Issues Fixed

### 1. Node Key Error
**Problem**: The node was failing to start with the error:
```
Error: failed to load or generate node key: open /tmp/fluentum-new-test/config/node_key.json: no such file or directory
```

**Solution**: Modified the `startNode` function in `cmd/fluentum/main.go` to:
- Ensure the home directory exists before starting
- Ensure the config directory exists before loading/generating node keys
- Ensure the data directory exists for validator state files
- Added proper error handling for directory creation

### 2. Feature Configuration
**Problem**: The feature configuration needed to be updated according to user requirements.

**Solution**: Updated `config/features.toml` to:
- Disable quantum_signing: `enabled = false`
- Enable state_sync: `enabled = true`
- Enable zk_rollup: `enabled = true`

## Current Feature Configuration

```toml
[features.quantum_signing]
enabled = false  # Disabled as requested

[features.state_sync]
enabled = true   # Enabled as requested

[features.zk_rollup]
enabled = true   # Enabled as requested
```

## Setup Instructions

### Option 1: Using the Setup Scripts (Recommended)

#### For Linux/macOS:
```bash
# Make the script executable
chmod +x scripts/init_and_start.sh

# Run the setup script
./scripts/init_and_start.sh /tmp/fluentum-new-test fluentum-node fluentum-mainnet-1 false
```

#### For Windows (PowerShell):
```powershell
# Run the PowerShell setup script
.\scripts\init_and_start.ps1 -HomeDir "/tmp/fluentum-new-test" -Moniker "fluentum-node" -ChainId "fluentum-mainnet-1" -Testnet $false
```

### Option 2: Manual Setup

1. **Initialize the node**:
   ```bash
   ./build/fluentumd init fluentum-node --chain-id fluentum-mainnet-1 --home /tmp/fluentum-new-test
   ```

2. **Generate node key** (if not already generated):
   ```bash
   ./build/fluentumd gen-node-key --home /tmp/fluentum-new-test
   ```

3. **Generate validator key** (if not already generated):
   ```bash
   ./build/fluentumd gen-validator-key --home /tmp/fluentum-new-test
   ```

4. **Start the node**:
   ```bash
   ./build/fluentumd start --home /tmp/fluentum-new-test --moniker fluentum-node --chain-id fluentum-mainnet-1
   ```

## Available Commands

The Fluentum node now supports the following commands:

- `fluentumd init [moniker]` - Initialize a new node
- `fluentumd start` - Start the node
- `fluentumd version` - Show version information
- `fluentumd gen-node-key` - Generate a node key
- `fluentumd gen-validator-key` - Generate a validator key

## Configuration Files

The node creates the following directory structure:
```
/tmp/fluentum-new-test/
├── config/
│   ├── config.toml          # Node configuration
│   ├── genesis.json         # Genesis file
│   ├── node_key.json        # Node key for P2P networking
│   └── priv_validator_key.json # Validator key for consensus
└── data/
    └── priv_validator_state.json # Validator state
```

## Feature Status

When the node starts, you should see output similar to:
```
[FeatureLoader] Features loaded and started: map[
  quantum_signing:map[enabled:false version:1.0.0] 
  state_sync:map[enabled:true version:1.0.0] 
  zk_rollup:map[enabled:true version:1.0.0]
]
```

## Troubleshooting

### Common Issues

1. **Permission denied errors**: Make sure you have write permissions to the home directory
2. **Port already in use**: The default ports are 26656 (P2P) and 26657 (RPC). Make sure these ports are available
3. **Binary not found**: Make sure to build the project first with `make build`

### Debug Mode

To run with debug logging:
```bash
./build/fluentumd start --home /tmp/fluentum-new-test --log_level debug
```

### Testnet Mode

To run in testnet mode with faster block times:
```bash
./build/fluentumd start --home /tmp/fluentum-new-test --testnet
```

## Next Steps

1. The node should now start successfully without the node key error
2. The feature configuration has been updated as requested
3. Use the provided scripts for easy setup and initialization
4. Monitor the node logs for any additional issues

For more information, refer to the main project documentation or contact the development team. 