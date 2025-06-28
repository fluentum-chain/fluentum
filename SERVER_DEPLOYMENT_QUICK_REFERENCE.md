# Fluentum Server Deployment - Quick Reference

## 🚀 Quick Start (Linux Server)

### 1. Copy Files to Server
```bash
# Copy the built binary
scp fluentumd user@your-server:/tmp/

# Copy deployment script
scp scripts/deploy_server.sh user@your-server:/tmp/
chmod +x /tmp/deploy_server.sh
```

### 2. Run Deployment Script
```bash
# SSH to your server
ssh user@your-server

# Run the deployment script
/tmp/deploy_server.sh
```

## 📋 Manual Deployment Steps

### Step 1: Install Binary
```bash
sudo cp /tmp/fluentumd /usr/local/bin/
sudo chmod +x /usr/local/bin/fluentumd
```

### Step 2: Initialize Node
```bash
fluentumd init your-node-name --chain-id fluentum-1
```

### Step 3: Configure for Server
```bash
# Edit config file
nano ~/.fluentumd/config/config.toml

# Key changes:
# - laddr = "tcp://0.0.0.0:26657"  # RPC
# - laddr = "tcp://0.0.0.0:26656"  # P2P  
# - prometheus = true
# - cors_allowed_origins = ["*"]
```

### Step 4: Create Systemd Service
```bash
sudo nano /etc/systemd/system/fluentumd.service
```

```ini
[Unit]
Description=Fluentum Validator Node
After=network.target

[Service]
Type=simple
User=your-username
WorkingDirectory=/home/your-username
ExecStart=/usr/local/bin/fluentumd start
Restart=on-failure
RestartSec=3
LimitNOFILE=4096

[Install]
WantedBy=multi-user.target
```

### Step 5: Start Service
```bash
sudo systemctl daemon-reload
sudo systemctl enable fluentumd
sudo systemctl start fluentumd
```

## 🔧 Essential Commands

### Service Management
```bash
# Check status
sudo systemctl status fluentumd

# Start/Stop/Restart
sudo systemctl start fluentumd
sudo systemctl stop fluentumd
sudo systemctl restart fluentumd

# View logs
sudo journalctl -u fluentumd -f
```

### Node Health Checks
```bash
# Check node status
curl http://localhost:26657/status

# Check sync status
curl http://localhost:26657/status | jq '.result.sync_info.catching_up'

# Check connected peers
curl http://localhost:26657/net_info | jq '.result.n_peers'

# Check latest block
curl http://localhost:26657/status | jq '.result.sync_info.latest_block_height'
```

### Configuration
```bash
# View config
cat ~/.fluentumd/config/config.toml

# View genesis
cat ~/.fluentumd/config/genesis.json

# View validator key
cat ~/.fluentumd/config/priv_validator_key.json
```

## 🔒 Security Setup

### Firewall Configuration
```bash
# Allow necessary ports
sudo ufw allow 26656/tcp  # P2P
sudo ufw allow 26657/tcp  # RPC (restrict to trusted IPs)
sudo ufw allow 26660/tcp  # Prometheus (restrict to monitoring IPs)

# Enable firewall
sudo ufw enable
```

### RPC Security (Optional)
```bash
# Restrict RPC to specific IPs
sudo ufw allow from trusted-ip to any port 26657

# Or use reverse proxy with authentication
```

## 📊 Monitoring

### Health Check Script
```bash
# Copy health check script
scp scripts/health_check.sh user@your-server:/tmp/
chmod +x /tmp/health_check.sh

# Run health check
/tmp/health_check.sh

# Set up cron job
crontab -e
# Add: */5 * * * * /tmp/health_check.sh
```

### Prometheus Metrics
```bash
# Access metrics endpoint
curl http://localhost:26660/metrics

# Configure Prometheus to scrape :26660
```

## 🛠️ Troubleshooting

### Common Issues

1. **Node not syncing:**
```bash
# Check peer connections
curl http://localhost:26657/net_info

# Check for errors
sudo journalctl -u fluentumd -n 100 | grep -i error
```

2. **High memory usage:**
```bash
# Monitor memory
htop
# Consider increasing swap space
```

3. **Port already in use:**
```bash
# Check what's using the port
sudo netstat -tlnp | grep :26657
sudo netstat -tlnp | grep :26656
```

### Log Analysis
```bash
# View recent logs
sudo journalctl -u fluentumd -n 100

# Search for errors
sudo journalctl -u fluentumd | grep -i error

# Monitor in real-time
sudo journalctl -u fluentumd -f
```

## 📁 Important Files

- **Binary:** `/usr/local/bin/fluentumd`
- **Config:** `~/.fluentumd/config/config.toml`
- **Genesis:** `~/.fluentumd/config/genesis.json`
- **Keys:** `~/.fluentumd/config/priv_validator_key.json`
- **Data:** `~/.fluentumd/data/`
- **Logs:** `~/.fluentumd/logs/` (if configured)

## 🔄 Backup & Recovery

### Backup
```bash
# Backup config
tar -czf fluentum-config-backup.tar.gz ~/.fluentumd/config/

# Backup data (stop node first)
sudo systemctl stop fluentumd
tar -czf fluentum-data-backup.tar.gz ~/.fluentumd/data/
sudo systemctl start fluentumd
```

### Recovery
```bash
# Restore config
tar -xzf fluentum-config-backup.tar.gz

# Restore data (stop node first)
sudo systemctl stop fluentumd
tar -xzf fluentum-data-backup.tar.gz
sudo systemctl start fluentumd
```

## 📞 Support

- **Documentation:** Check `docs/` directory
- **Health Monitoring:** Use `scripts/health_check.sh`
- **Logs:** Monitor with `sudo journalctl -u fluentumd -f`
- **Metrics:** Access at `http://your-server:26660/metrics`

## 🎯 Next Steps

1. ✅ Deploy node using script or manual steps
2. 🔧 Configure firewall and security
3. 📊 Set up monitoring and alerts
4. 💾 Configure regular backups
5. 🌐 Join the network and start validating
6. 📈 Monitor performance and adjust as needed 