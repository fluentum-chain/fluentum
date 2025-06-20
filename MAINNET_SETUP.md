# Fluentum Mainnet Setup Guide

This guide will help you set up and run the Fluentum mainnet on your Ubuntu server.

## 🚀 Quick Start

### 1. Initialize the Node

```bash
# Initialize Fluentum with default configuration
fluentum init my-node

# Or initialize with custom moniker
fluentum init my-node --moniker "My Fluentum Node"
```

### 2. Configure the Node

Edit the configuration file:
```bash
nano ~/.fluentum/config/config.toml
```

#### Essential Configuration Settings:

```toml
# Node identification
moniker = "My Fluentum Node"

# P2P Configuration
[p2p]
laddr = "tcp://0.0.0.0:26656"
external_address = "34.70.86.80:26656"
seeds = ""

# RPC Configuration
[rpc]
laddr = "tcp://0.0.0.0:26657"
cors_allowed_origins = ["*"]

# API Configuration
[api]
enable = true
swagger = true
laddr = "tcp://0.0.0.0:1317"

# Consensus Configuration
[consensus]
timeout_commit = "1s"
timeout_propose = "3s"
```

### 3. Configure Genesis

If you're running a validator node, you'll need the genesis file:

```bash
# Download genesis file (if available from mainnet)
wget https://raw.githubusercontent.com/fluentum-chain/mainnet/main/genesis.json -O ~/.fluentum/config/genesis.json

# Or create a new genesis for development
fluentum init --chain-id fluentum-mainnet-1
```

### 4. Start the Node

```bash
# Start the node in foreground
fluentum start

# Or start in background
nohup fluentum start > fluentum.log 2>&1 &
```

## 🔍 Verification Commands

### Check Node Status

```bash
# Check if node is running
ps aux | grep fluentum

# Check node logs
tail -f fluentum.log

# Check node info via RPC
curl -s http://localhost:26657/status | jq

# Check node info via CLI
fluentum status
```

### Check Network Connectivity

```bash
# Check if ports are open
netstat -tulpn | grep fluentum

# Check P2P port (26656)
curl -s http://localhost:26656

# Check RPC port (26657)
curl -s http://localhost:26657/status

# Check API port (1317)
curl -s http://localhost:1317/cosmos/base/tendermint/v1beta1/node_info
```

### Check Block Sync

```bash
# Check latest block
curl -s http://localhost:26657/block | jq '.result.block.header.height'

# Check sync status
curl -s http://localhost:26657/status | jq '.result.sync_info'

# Check validator set
curl -s http://localhost:26657/validators | jq '.result.validators | length'
```

## 🔧 Advanced Configuration

### Firewall Setup

```bash
# Allow Fluentum ports
sudo ufw allow 26656/tcp  # P2P
sudo ufw allow 26657/tcp  # RPC
sudo ufw allow 1317/tcp   # API
sudo ufw allow 9090/tcp   # Prometheus (optional)

# Check firewall status
sudo ufw status
```

### Systemd Service Setup

Create a systemd service for automatic startup:

```bash
sudo nano /etc/systemd/system/fluentum.service
```

Add the following content:

```ini
[Unit]
Description=Fluentum Node
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu
ExecStart=/usr/local/bin/fluentum start --home /home/ubuntu/.fluentum
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=fluentum

[Install]
WantedBy=multi-user.target
```

Enable and start the service:

```bash
sudo systemctl daemon-reload
sudo systemctl enable fluentum
sudo systemctl start fluentum
sudo systemctl status fluentum
```

### Monitoring Setup

Install monitoring tools:

```bash
# Install htop for system monitoring
sudo apt install htop

# Install jq for JSON parsing
sudo apt install jq

# Create monitoring script
cat > monitor.sh << 'EOF'
#!/bin/bash
echo "=== Fluentum Node Status ==="
echo "Time: $(date)"
echo ""

echo "=== System Resources ==="
echo "CPU Usage: $(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d'%' -f1)%"
echo "Memory Usage: $(free | grep Mem | awk '{printf("%.2f%%", $3/$2 * 100.0)}')"
echo "Disk Usage: $(df -h / | awk 'NR==2 {print $5}')"
echo ""

echo "=== Node Status ==="
if pgrep -x "fluentum" > /dev/null; then
    echo "✅ Fluentum is running"
    echo "Latest Block: $(curl -s http://localhost:26657/block | jq -r '.result.block.header.height // "N/A"')"
    echo "Sync Status: $(curl -s http://localhost:26657/status | jq -r '.result.sync_info.catching_up // "N/A"')"
else
    echo "❌ Fluentum is not running"
fi
echo ""

echo "=== Network Status ==="
echo "P2P Connections: $(curl -s http://localhost:26657/net_info | jq -r '.result.n_peers // "N/A"')"
echo "RPC Status: $(curl -s http://localhost:26657/status | jq -r '.result.node_info.rpc_address // "N/A"')"
EOF

chmod +x monitor.sh
```

## 🚨 Troubleshooting

### Common Issues

1. **Node won't start**
   ```bash
   # Check logs
   tail -f fluentum.log
   
   # Check configuration
   fluentum validate-genesis
   
   # Reset if needed
   fluentum unsafe-reset-all
   ```

2. **Not syncing**
   ```bash
   # Check peers
   curl -s http://localhost:26657/net_info | jq '.result.peers'
   
   # Add seeds manually
   # Edit config.toml and add seeds
   ```

3. **Port issues**
   ```bash
   # Check if ports are in use
   sudo netstat -tulpn | grep :26656
   sudo netstat -tulpn | grep :26657
   
   # Kill conflicting processes
   sudo pkill -f fluentum
   ```

4. **Permission issues**
   ```bash
   # Fix ownership
   sudo chown -R ubuntu:ubuntu ~/.fluentum
   
   # Fix permissions
   chmod 600 ~/.fluentum/config/node_key.json
   chmod 600 ~/.fluentum/config/priv_validator_key.json
   ```

### Log Analysis

```bash
# View real-time logs
tail -f fluentum.log

# Search for errors
grep -i error fluentum.log

# Search for warnings
grep -i warn fluentum.log

# Monitor specific events
grep -i "new block" fluentum.log
```

## 📊 Health Checks

### Automated Health Check Script

```bash
cat > health_check.sh << 'EOF'
#!/bin/bash

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "🔍 Fluentum Node Health Check"
echo "=============================="

# Check if process is running
if pgrep -x "fluentum" > /dev/null; then
    echo -e "${GREEN}✅ Process Status: Running${NC}"
else
    echo -e "${RED}❌ Process Status: Not Running${NC}"
    exit 1
fi

# Check RPC endpoint
if curl -s http://localhost:26657/status > /dev/null; then
    echo -e "${GREEN}✅ RPC Endpoint: Accessible${NC}"
else
    echo -e "${RED}❌ RPC Endpoint: Not Accessible${NC}"
fi

# Check sync status
SYNC_STATUS=$(curl -s http://localhost:26657/status | jq -r '.result.sync_info.catching_up // "unknown"')
if [ "$SYNC_STATUS" = "false" ]; then
    echo -e "${GREEN}✅ Sync Status: Caught Up${NC}"
elif [ "$SYNC_STATUS" = "true" ]; then
    echo -e "${YELLOW}⚠️  Sync Status: Catching Up${NC}"
else
    echo -e "${RED}❌ Sync Status: Unknown${NC}"
fi

# Check latest block
LATEST_BLOCK=$(curl -s http://localhost:26657/block | jq -r '.result.block.header.height // "N/A"')
echo -e "${GREEN}📦 Latest Block: $LATEST_BLOCK${NC}"

# Check peers
PEER_COUNT=$(curl -s http://localhost:26657/net_info | jq -r '.result.n_peers // 0')
echo -e "${GREEN}🌐 Connected Peers: $PEER_COUNT${NC}"

echo "=============================="
echo "Health check completed!"
EOF

chmod +x health_check.sh
```

## 🎯 Next Steps

1. **Run the health check**: `./health_check.sh`
2. **Monitor the node**: `./monitor.sh`
3. **Set up alerts**: Configure monitoring alerts for downtime
4. **Join the network**: Connect to other nodes in the network
5. **Become a validator**: Set up validator keys and stake tokens

## 📞 Support

If you encounter issues:
- Check the logs: `tail -f fluentum.log`
- Run health check: `./health_check.sh`
- Join our community: [Telegram](https://t.me/fluentum)
- Open an issue: [GitHub Issues](https://github.com/fluentum-chain/fluentum/issues)

---

**Your Fluentum node should now be running successfully on 34.70.86.80!** 🚀 