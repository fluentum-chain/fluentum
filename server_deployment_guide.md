# Fluentum Server Deployment Guide

## Overview
This guide covers deploying the Fluentum blockchain node on a server environment.

## Prerequisites
- Linux server (Ubuntu 20.04+ recommended)
- Go 1.18+ installed
- At least 4GB RAM
- 100GB+ storage space
- Open ports: 26656 (P2P), 26657 (RPC), 26660 (Prometheus)

## Deployment Options

### Option 1: Automated Setup Script

1. **Copy the setup script to your server:**
```bash
# Copy scripts/setup_validator.sh to your server
chmod +x setup_validator.sh
```

2. **Set environment variables:**
```bash
export MONIKER="your-node-name"
export VALIDATOR_NAME="your-validator"
export CHAIN_ID="fluentum-1"
export GENESIS_URL="https://your-genesis-url/genesis.json"  # Optional
export SEED_NODES="node1@ip1:26656,node2@ip2:26656"        # Optional
export PERSISTENT_PEERS="node1@ip1:26656,node2@ip2:26656"  # Optional
export ZK_PROVER_URL="https://zk.fluentum.net"             # Optional
```

3. **Run the setup script:**
```bash
./setup_validator.sh
```

This script will:
- Initialize the node
- Download genesis file (if URL provided)
- Configure the node for server deployment
- Create validator keys
- Generate quantum-resistant keys
- Create and start systemd service

### Option 2: Manual Setup

1. **Build and copy the binary:**
```bash
# On your development machine
go build -o fluentumd ./cmd/fluentum

# Copy to server
scp fluentumd user@server:/usr/local/bin/
```

2. **Initialize the node:**
```bash
# On server
fluentumd init your-node-name --chain-id fluentum-1
```

3. **Configure the node:**
```bash
# Edit config.toml
nano ~/.fluentumd/config/config.toml

# Key changes for server deployment:
# - Set laddr = "tcp://0.0.0.0:26657" for RPC
# - Set laddr = "tcp://0.0.0.0:26656" for P2P
# - Set prometheus = true
# - Configure seed_nodes and persistent_peers
```

4. **Create systemd service:**
```bash
sudo nano /etc/systemd/system/fluentumd.service
```

```ini
[Unit]
Description=Fluentum Validator Node
After=network.target

[Service]
Type=simple
User=fluentum
WorkingDirectory=/home/fluentum
ExecStart=/usr/local/bin/fluentumd start
Restart=on-failure
RestartSec=3
LimitNOFILE=4096

[Install]
WantedBy=multi-user.target
```

5. **Start the service:**
```bash
sudo systemctl daemon-reload
sudo systemctl enable fluentumd
sudo systemctl start fluentumd
```

### Option 3: Docker Deployment

1. **Build the Docker image:**
```bash
docker build -f DOCKER/Dockerfile -t fluentum/fluentumd .
```

2. **Run the container:**
```bash
docker run -d \
  --name fluentum-node \
  -p 26656:26656 \
  -p 26657:26657 \
  -p 26660:26660 \
  -v fluentum-data:/tendermint \
  -e MONIKER="your-node-name" \
  -e CHAIN_ID="fluentum-1" \
  fluentum/fluentumd
```

## Configuration

### Key Configuration Files

1. **config.toml** - Main node configuration
2. **genesis.json** - Chain genesis state
3. **priv_validator_key.json** - Validator private key
4. **node_key.json** - Node identity key

### Important Server Settings

```toml
# RPC Configuration
[rpc]
laddr = "tcp://0.0.0.0:26657"  # Allow external RPC access
cors_allowed_origins = ["*"]   # Configure CORS as needed

# P2P Configuration
[p2p]
laddr = "tcp://0.0.0.0:26656"  # Allow external P2P connections
external_address = "your-server-ip:26656"
seeds = "seed1@ip1:26656,seed2@ip2:26656"
persistent_peers = "peer1@ip1:26656,peer2@ip2:26656"

# Prometheus Metrics
instrumentation.prometheus = true
instrumentation.prometheus_listen_addr = ":26660"
```

## Monitoring and Health Checks

### Using the Health Check Script

```bash
# Copy the health check script
chmod +x scripts/health_check.sh

# Run health check
./scripts/health_check.sh

# Set up cron job for regular checks
crontab -e
# Add: */5 * * * * /path/to/health_check.sh
```

### Manual Health Checks

```bash
# Check node status
curl http://localhost:26657/status

# Check sync status
curl http://localhost:26657/status | jq '.result.sync_info.catching_up'

# Check connected peers
curl http://localhost:26657/net_info | jq '.result.n_peers'

# Check validator set
curl http://localhost:26657/validators | jq '.result.validators | length'
```

## Security Considerations

1. **Firewall Configuration:**
```bash
# Allow only necessary ports
sudo ufw allow 26656/tcp  # P2P
sudo ufw allow 26657/tcp  # RPC (restrict to trusted IPs)
sudo ufw allow 26660/tcp  # Prometheus (restrict to monitoring IPs)
```

2. **RPC Security:**
- Restrict RPC access to trusted IPs
- Use reverse proxy with authentication
- Consider using TLS certificates

3. **Key Management:**
- Store private keys securely
- Use hardware security modules (HSM) for production
- Regular key rotation

## Troubleshooting

### Common Issues

1. **Node not syncing:**
```bash
# Check peer connections
curl http://localhost:26657/net_info

# Check for errors in logs
tail -f ~/.fluentumd/logs/tendermint.log
```

2. **High memory usage:**
```bash
# Monitor memory usage
htop
# Consider increasing swap space
```

3. **Disk space issues:**
```bash
# Check disk usage
df -h
# Clean old data if needed
```

### Log Analysis

```bash
# View recent logs
tail -100 ~/.fluentumd/logs/tendermint.log

# Search for errors
grep -i error ~/.fluentumd/logs/tendermint.log

# Monitor in real-time
tail -f ~/.fluentumd/logs/tendermint.log
```

## Backup and Recovery

### Backup Strategy

```bash
# Backup configuration
tar -czf fluentum-config-backup.tar.gz ~/.fluentumd/config/

# Backup data (stop node first)
sudo systemctl stop fluentumd
tar -czf fluentum-data-backup.tar.gz ~/.fluentumd/data/
sudo systemctl start fluentumd
```

### Recovery Process

```bash
# Restore configuration
tar -xzf fluentum-config-backup.tar.gz

# Restore data (stop node first)
sudo systemctl stop fluentumd
tar -xzf fluentum-data-backup.tar.gz
sudo systemctl start fluentumd
```

## Performance Optimization

1. **Database Optimization:**
```toml
# Use pebble for better performance
db_backend = "pebble"
```

2. **Memory Optimization:**
```toml
# Adjust cache sizes
mempool.cache_size = 10000
```

3. **Network Optimization:**
```toml
# Optimize peer connections
p2p.max_num_inbound_peers = 40
p2p.max_num_outbound_peers = 10
```

## Support and Resources

- **Documentation:** Check the `docs/` directory
- **Health Monitoring:** Use the provided health check script
- **Logs:** Monitor `~/.fluentumd/logs/` directory
- **Metrics:** Access Prometheus metrics at `:26660`

## Next Steps

1. Deploy using your preferred method
2. Configure monitoring and alerts
3. Set up regular backups
4. Join the network and start validating
5. Monitor performance and adjust as needed 