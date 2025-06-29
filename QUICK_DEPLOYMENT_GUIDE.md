# 🚀 Fluentum Quick Deployment Guide

## �� Prerequisites

### Go Version Requirements

**⚠️ Important:** Fluentum Core supports two Go version configurations:

#### Option 1: Go 1.20.x (Recommended for Cosmos SDK v0.47.x)
- **Compatibility:** Cosmos SDK v0.47.12, CometBFT v0.37.2
- **Dependencies:** Pinned for Go 1.20 compatibility
- **Status:** ✅ Stable and tested

#### Option 2: Go 1.22+ (For newer dependencies)
- **Compatibility:** Latest Cosmos SDK and CometBFT versions
- **Dependencies:** Auto-upgraded to latest compatible versions
- **Status:** ✅ Supported but may require dependency updates

### 1. Install Go

**For Go 1.20.x (Recommended):**
```bash
wget https://go.dev/dl/go1.20.14.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.20.14.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
go version
# Expected: go version go1.20.14 linux/amd64
```

**For Go 1.22+ (Newer dependencies):**
```bash
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
go version
# Expected: go version go1.22.0 linux/amd64
```

**Note:** If using Go 1.22+, run `go mod tidy` after installation to upgrade dependencies to compatible versions.

### 2. System Requirements

## 📋 Pre-Flight Checklist

### ✅ Before You Start
- [ ] All tests passing: `make test`
- [ ] Build successful: `make build`
- [ ] Testnet validation completed
- [ ] Backup procedures ready
- [ ] Documentation updated
- [ ] Team notified

---

## 🧪 Testnet Validation (Required)

### Step 1: Local Testnet
```bash
# Create and test local network
make localnet
./build/fluentumd start --home ~/.cometbft

# Verify functionality
curl -s http://localhost:26657/status | jq
./build/fluentumd show_node_id
./build/fluentumd show_validator
```

### Step 2: Multi-Node Testnet
```bash
# Deploy multi-node testnet
make testnet

# Verify network connectivity
curl -s http://localhost:26657/net_info | jq '.result.n_peers'
curl -s http://localhost:26657/validators | jq '.result.validators | length'
```

### Step 3: Public Testnet
```bash
# Deploy to public testnet
./scripts/deploy_testnet.sh

# Verify external connectivity
curl -s http://YOUR_PUBLIC_IP:26657/status
```

---

## 💾 Backup Critical Files

### Windows (PowerShell)
```powershell
# Run backup script
.\scripts\backup_fluentum.ps1

# Or manual backup
$Date = Get-Date -Format "yyyyMMdd_HHmmss"
Copy-Item "$env:USERPROFILE\.cometbft\config\genesis.json" "C:\backup\fluentum\genesis.json.$Date"
Copy-Item "$env:USERPROFILE\.cometbft\config\priv_validator_key.json" "C:\backup\fluentum\priv_validator_key.json.$Date"
Copy-Item "$env:USERPROFILE\.cometbft\config\node_key.json" "C:\backup\fluentum\node_key.json.$Date"
```

### Linux/macOS
```bash
# Run backup script
./scripts/backup_fluentum.sh

# Or manual backup
DATE=$(date +%Y%m%d_%H%M%S)
cp ~/.cometbft/config/genesis.json ~/.cometbft/config/genesis.json.backup.$DATE
cp ~/.cometbft/config/priv_validator_key.json ~/.cometbft/config/priv_validator_key.json.backup.$DATE
cp ~/.cometbft/config/node_key.json ~/.cometbft/config/node_key.json.backup.$DATE
```

---

## 🔧 Final Configuration

### Production config.toml
```toml
# Essential production settings
moniker = "Your Node Name"

[consensus]
timeout_commit = "1s"
timeout_propose = "3s"
create_empty_blocks = true

[p2p]
laddr = "tcp://0.0.0.0:26656"
external_address = "YOUR_PUBLIC_IP:26656"
seeds = "seed1@seed1.example.com:26656,seed2@seed2.example.com:26656"
persistent_peers = "validator1@validator1.example.com:26656"

[rpc]
laddr = "tcp://0.0.0.0:26657"
max_open_connections = 900
timeout_broadcast_tx_commit = "10s"

[instrumentation]
prometheus = true
prometheus_listen_addr = ":26660"
```

### Security Hardening
```bash
# Set proper permissions
chmod 600 ~/.cometbft/config/priv_validator_key.json
chmod 600 ~/.cometbft/config/node_key.json

# Firewall configuration
sudo ufw allow 26656/tcp  # P2P
sudo ufw allow 26657/tcp  # RPC (restrict to trusted IPs)
sudo ufw deny 26660/tcp   # Prometheus (internal only)
```

---

## 🚀 Deployment Steps

### 1. Create Final Backup
```bash
# Windows
.\scripts\backup_fluentum.ps1

# Linux/macOS
./scripts/backup_fluentum.sh
```

### 2. Stop Existing Node (if running)
```bash
# Systemd service
sudo systemctl stop fluentum

# Or direct process
pkill -f fluentum
```

### 3. Update Configuration
```bash
# Copy production config
cp config.toml.production ~/.cometbft/config/config.toml

# Validate configuration
./build/fluentumd validate-genesis
```

### 4. Start Node
```bash
# Systemd service
sudo systemctl start fluentum

# Or direct start
./build/fluentumd start --home ~/.cometbft
```

### 5. Verify Deployment
```bash
# Run health check
./scripts/health_check.sh

# Or manual verification
curl -s http://localhost:26657/status | jq
curl -s http://localhost:26657/net_info | jq '.result.n_peers'
```

---

## 📊 Post-Deployment Monitoring

### Immediate Checks (First Hour)
- [ ] Node process running
- [ ] RPC endpoint responding
- [ ] Network connectivity established
- [ ] Block synchronization working
- [ ] No critical errors in logs

### Ongoing Monitoring
```bash
# Health check every 5 minutes
*/5 * * * * /path/to/scripts/health_check.sh

# Log monitoring
tail -f ~/.cometbft/logs/tendermint.log | grep -i error

# Performance monitoring
htop
df -h
free -h
```

### Alert Setup
- Node down alerts
- Sync issues alerts
- High resource usage alerts
- Error rate alerts

---

## 🚨 Emergency Procedures

### Node Won't Start
```bash
# Check logs
tail -f ~/.cometbft/logs/tendermint.log

# Validate configuration
./build/fluentumd validate-genesis

# Reset if needed (DANGEROUS - only if necessary)
./build/fluentumd unsafe-reset-all
```

### Not Syncing
```bash
# Check peers
curl -s http://localhost:26657/net_info | jq '.result.peers'

# Check seeds configuration
grep seeds ~/.cometbft/config/config.toml

# Verify network connectivity
ping seed1.example.com
```

### Data Corruption
```bash
# Stop node
sudo systemctl stop fluentum

# Restore from backup
cp ~/.cometbft/config/genesis.json.backup.* ~/.cometbft/config/genesis.json
cp ~/.cometbft/config/priv_validator_key.json.backup.* ~/.cometbft/config/priv_validator_key.json

# Restart node
sudo systemctl start fluentum
```

---

## 📞 Support Contacts

### Primary Contacts
- **Technical Lead**: [Contact Info]
- **DevOps Engineer**: [Contact Info]
- **Security Team**: [Contact Info]

### Escalation Levels
1. **Level 1**: Node operator attempts resolution
2. **Level 2**: Technical lead involvement
3. **Level 3**: Full team response
4. **Level 4**: External support engagement

---

## ✅ Success Criteria

### Technical Success
- [ ] Node running continuously for 24 hours
- [ ] All API endpoints responding
- [ ] Block synchronization working
- [ ] Network connectivity stable
- [ ] Performance metrics acceptable
- [ ] No critical errors in logs

### Operational Success
- [ ] Monitoring systems operational
- [ ] Backup procedures tested
- [ ] Documentation complete
- [ ] Team trained on procedures
- [ ] Support processes established

---

## 📚 Additional Resources

- **Full Checklist**: [FINAL_CHECKLIST.md](FINAL_CHECKLIST.md)
- **Mainnet Setup**: [MAINNET_SETUP.md](MAINNET_SETUP.md)
- **Server Setup**: [SERVER_SETUP.md](SERVER_SETUP.md)
- **Troubleshooting**: [TROUBLESHOOTING.md](TROUBLESHOOTING.md)

---

## 🔧 Dependency Management

### Current Dependency Versions (Go 1.20)
- CometBFT: v0.37.2
- Cosmos SDK: v0.47.12
- cometbft-db: v0.8.0

### Upgrading Dependencies (Go 1.22+)
If using Go 1.22+, run:
```bash
go mod tidy
make build
```
This will upgrade to the latest compatible dependencies (e.g., CometBFT v0.38+, Cosmos SDK v0.50+).

---

**🎉 Your Fluentum node is now deployed to mainnet!**

Monitor the node closely for the first 24 hours and ensure all systems are functioning correctly. 