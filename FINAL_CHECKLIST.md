# Fluentum Final Deployment Checklist

This comprehensive checklist ensures a safe and successful deployment of Fluentum to mainnet.

## 📋 Pre-Deployment Checklist

### ✅ Code Quality & Testing
- [ ] All unit tests passing: `make test`
- [ ] Integration tests completed: `make test_integration`
- [ ] E2E tests passing: `make test_e2e` (Removed: E2E tests are no longer present in the codebase)
- [ ] Code coverage meets requirements: `make test_cover`
- [ ] Security audit completed
- [ ] Performance benchmarks validated
- [ ] Memory leak tests passed
- [ ] Network stress tests completed

### ✅ Build Verification
- [ ] Clean build successful: `make clean && make build`
- [ ] Binary integrity verified: `sha256sum build/fluentumd`
- [ ] Cross-platform builds tested
- [ ] Docker image builds successfully
- [ ] Binary size optimized
- [ ] Dependencies up to date: `go mod tidy && go mod verify`

### ✅ Configuration Validation
- [ ] Config files validated: `fluentum validate-genesis`
- [ ] All required config parameters set
- [ ] Network-specific settings configured
- [ ] Security settings reviewed
- [ ] Performance tuning applied
- [ ] Logging configuration optimized

## 📋 Version Compatibility

### ✅ Core Dependencies
- [ ] **CometBFT**: v0.38.6 (not v0.37.x)
- [ ] **Cosmos SDK**: v0.50.6 (not v0.47.x)
- [ ] **cometbft-db**: v0.9.1
- [ ] **IBC**: v7.3.1
- [ ] **Go**: 1.24.4+ (required)
- [ ] **gRPC**: v1.73.0

---

## 🧪 Testnet Validation

### Phase 1: Local Testnet
```bash
# 1. Create local testnet
make localnet

# 2. Verify node startup
./build/fluentumd start --home ~/.tendermint

# 3. Check node status
curl -s http://localhost:26657/status | jq

# 4. Test basic functionality
./build/fluentumd version
./build/fluentumd show_node_id
./build/fluentumd show_validator
```

### Phase 2: Multi-Node Testnet
```bash
# 1. Deploy testnet with multiple nodes
make testnet

# 2. Verify network connectivity
curl -s http://localhost:26657/net_info | jq '.result.n_peers'

# 3. Test consensus
curl -s http://localhost:26657/validators | jq '.result.validators | length'

# 4. Monitor block production
tail -f ~/.tendermint/logs/tendermint.log | grep "new block"
```

### Phase 3: Public Testnet
```bash
# 1. Deploy to public testnet
./scripts/deploy_testnet.sh

# 2. Verify external connectivity
curl -s http://YOUR_PUBLIC_IP:26657/status

# 3. Test cross-node communication
# Connect to other testnet nodes and verify sync

# 4. Load testing
./scripts/load_test.sh
```

### Testnet Validation Checklist
- [ ] Node starts successfully
- [ ] Network connectivity established
- [ ] Block synchronization working
- [ ] Consensus participation verified
- [ ] Transaction processing tested
- [ ] API endpoints responding
- [ ] Performance metrics acceptable
- [ ] Error handling validated
- [ ] Recovery procedures tested
- [ ] Security measures enforced

---

## 💾 Backup Procedures

### Critical Files to Backup

#### 1. Genesis File
```bash
# Backup genesis file
cp ~/.fluentum/config/genesis.json ~/.fluentum/config/genesis.json.backup.$(date +%Y%m%d_%H%M%S)

# Verify backup integrity
sha256sum ~/.fluentum/config/genesis.json*
```

#### 2. Validator Keys
```bash
# Backup private validator key
cp ~/.fluentum/config/priv_validator_key.json ~/.fluentum/config/priv_validator_key.json.backup.$(date +%Y%m%d_%H%M%S)

# Backup validator state
cp ~/.fluentum/data/priv_validator_state.json ~/.fluentum/data/priv_validator_state.json.backup.$(date +%Y%m%d_%H%M%S)

# Set proper permissions
chmod 600 ~/.fluentum/config/priv_validator_key.json.backup.*
```

#### 3. Node Keys
```bash
# Backup node key
cp ~/.fluentum/config/node_key.json ~/.fluentum/config/node_key.json.backup.$(date +%Y%m%d_%H%M%S)

# Set proper permissions
chmod 600 ~/.fluentum/config/node_key.json.backup.*
```

#### 4. Configuration Files
```bash
# Backup entire config directory
tar -czf ~/.fluentum/config.backup.$(date +%Y%m%d_%H%M%S).tar.gz ~/.fluentum/config/

# Backup specific config files
cp ~/.fluentum/config/config.toml ~/.fluentum/config/config.toml.backup.$(date +%Y%m%d_%H%M%S)
```

#### 5. Data Directory
```bash
# Backup data directory (if needed)
tar -czf ~/.fluentum/data.backup.$(date +%Y%m%d_%H%M%S).tar.gz ~/.fluentum/data/
```

### Automated Backup Script
```bash
#!/bin/bash
# Create backup script: backup_fluentum.sh

BACKUP_DIR="/backup/fluentum"
DATE=$(date +%Y%m%d_%H%M%S)
FLUENTUM_HOME="$HOME/.fluentum"

# Create backup directory
mkdir -p "$BACKUP_DIR"

echo "🔄 Creating Fluentum backup at $DATE"

# Backup critical files
cp "$FLUENTUM_HOME/config/genesis.json" "$BACKUP_DIR/genesis.json.$DATE"
cp "$FLUENTUM_HOME/config/priv_validator_key.json" "$BACKUP_DIR/priv_validator_key.json.$DATE"
cp "$FLUENTUM_HOME/config/node_key.json" "$BACKUP_DIR/node_key.json.$DATE"
cp "$FLUENTUM_HOME/data/priv_validator_state.json" "$BACKUP_DIR/priv_validator_state.json.$DATE"
cp "$FLUENTUM_HOME/config/config.toml" "$BACKUP_DIR/config.toml.$DATE"

# Set proper permissions
chmod 600 "$BACKUP_DIR"/*.json.$DATE

# Create checksums
cd "$BACKUP_DIR"
sha256sum *.$DATE > "checksums.$DATE"

# Clean old backups (keep last 7 days)
find "$BACKUP_DIR" -name "*.$(date -d '7 days ago' +%Y%m%d)*" -delete

echo "✅ Backup completed: $BACKUP_DIR"
echo "📊 Backup size: $(du -sh "$BACKUP_DIR" | cut -f1)"
```

### Backup Verification
```bash
# Verify backup integrity
cd /backup/fluentum
sha256sum -c checksums.$(date +%Y%m%d_%H%M%S)

# Test restore procedure
mkdir -p /tmp/test_restore
cp genesis.json.* /tmp/test_restore/
cp priv_validator_key.json.* /tmp/test_restore/
```

### Backup Checklist
- [ ] Genesis file backed up
- [ ] Validator keys backed up
- [ ] Node keys backed up
- [ ] Configuration files backed up
- [ ] Backup integrity verified
- [ ] Backup permissions set correctly
- [ ] Backup location secure
- [ ] Restore procedure tested
- [ ] Automated backup script configured
- [ ] Backup retention policy set

---

## 📚 Documentation Updates

### 1. Node Operator Documentation

#### Update MAINNET_SETUP.md
```markdown
# Add new sections to MAINNET_SETUP.md

## New CLI Flags (v0.34+)

### Consensus Flags
- `--consensus.double_sign_check_height`: Height to check for double signing
- `--consensus.create_empty_blocks`: Enable/disable empty block creation
- `--consensus.create_empty_blocks_interval`: Interval between empty blocks

### P2P Flags
- `--p2p.external-address`: External address for peer discovery
- `--p2p.seeds`: Seed nodes for network discovery
- `--p2p.persistent_peers`: Persistent peer connections
- `--p2p.private_peer_ids`: Private peer IDs

### RPC Flags
- `--rpc.unsafe`: Enable unsafe RPC methods
- `--rpc.pprof_laddr`: Pprof debugging endpoint
- `--rpc.grpc_laddr`: gRPC server address

### Database Flags
- `--db_backend`: Database backend selection
- `--db_dir`: Database directory path
```

#### Update Configuration Examples
```toml
# Add to config.toml examples

[consensus]
# New consensus parameters
double_sign_check_height = 0
create_empty_blocks = true
create_empty_blocks_interval = "0s"

[p2p]
# Enhanced P2P configuration
external_address = "YOUR_PUBLIC_IP:26656"
seeds = "node1@seed1.example.com:26656,node2@seed2.example.com:26656"
persistent_peers = "validator1@validator1.example.com:26656"
private_peer_ids = ""

[rpc]
# Enhanced RPC configuration
unsafe = false
pprof_laddr = ""
grpc_laddr = ""
```

### 2. Deployment Documentation

#### Create DEPLOYMENT_GUIDE.md
```markdown
# Fluentum Deployment Guide

## Pre-Deployment Checklist
1. Testnet validation completed
2. All backups created
3. Documentation updated
4. Monitoring configured
5. Security measures implemented

## Deployment Steps
1. Stop existing node (if any)
2. Create final backup
3. Update configuration
4. Start new node
5. Verify deployment
6. Monitor performance

## Rollback Procedures
1. Stop new node
2. Restore from backup
3. Restart previous version
4. Verify functionality

## Monitoring Setup
1. Configure Prometheus
2. Set up Grafana dashboards
3. Configure alerts
4. Test monitoring
```

### 3. CLI Reference Documentation

#### Update CLI_HELP.md
```markdown
# Fluentum CLI Reference

## Node Commands
- `fluentum start`: Start the node
- `fluentum init`: Initialize node configuration
- `fluentum reset`: Reset node data
- `fluentum unsafe-reset-all`: Reset all data (dangerous)

## Validator Commands
- `fluentum show_validator`: Show validator information
- `fluentum gen_validator`: Generate validator key
- `fluentum show_node_id`: Show node ID

## Network Commands
- `fluentum testnet`: Create testnet
- `fluentum light`: Run light client

## New Flags in v0.34+
- `--consensus.double_sign_check_height`
- `--p2p.external-address`
- `--rpc.unsafe`
- `--db_backend`
```

### 4. Troubleshooting Documentation

#### Update TROUBLESHOOTING.md
```markdown
# Fluentum Troubleshooting Guide

## Common Issues

### Node Won't Start
1. Check configuration: `fluentum validate-genesis`
2. Check logs: `tail -f ~/.fluentum/logs/tendermint.log`
3. Check permissions: `ls -la ~/.fluentum/config/`
4. Reset if needed: `fluentum unsafe-reset-all`

### Not Syncing
1. Check peers: `curl -s http://localhost:26657/net_info`
2. Check seeds configuration
3. Verify network connectivity
4. Check firewall settings

### Performance Issues
1. Monitor system resources
2. Check database performance
3. Optimize configuration
4. Consider hardware upgrades

## Recovery Procedures
1. Stop the node
2. Restore from backup
3. Verify configuration
4. Restart node
5. Monitor status
```

### Documentation Checklist
- [ ] MAINNET_SETUP.md updated with new flags
- [ ] Configuration examples updated
- [ ] CLI reference documentation created
- [ ] Troubleshooting guide updated
- [ ] Deployment procedures documented
- [ ] Rollback procedures documented
- [ ] Monitoring setup documented
- [ ] Security considerations documented
- [ ] Performance tuning guide created
- [ ] FAQ section added

---

## 🔧 Final Configuration

### 1. Production Configuration
```toml
# Production config.toml optimizations

[consensus]
timeout_commit = "1s"
timeout_propose = "3s"
create_empty_blocks = true
create_empty_blocks_interval = "0s"

[p2p]
max_num_inbound_peers = 40
max_num_outbound_peers = 10
persistent_peers = "validator1@validator1.example.com:26656,validator2@validator2.example.com:26656"
seeds = "seed1@seed1.example.com:26656,seed2@seed2.example.com:26656"

[rpc]
max_open_connections = 900
max_subscription_clients = 100
max_subscriptions_per_client = 5
timeout_broadcast_tx_commit = "10s"

[instrumentation]
prometheus = true
prometheus_listen_addr = ":26660"
```

### 2. Security Configuration
```bash
# Security hardening
chmod 600 ~/.fluentum/config/priv_validator_key.json
chmod 600 ~/.fluentum/config/node_key.json
chown -R fluentum:fluentum ~/.fluentum

# Firewall configuration
sudo ufw allow 26656/tcp  # P2P
sudo ufw allow 26657/tcp  # RPC (restrict to trusted IPs)
sudo ufw deny 26660/tcp   # Prometheus (internal only)
```

### 3. Monitoring Configuration
```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'fluentum'
    static_configs:
      - targets: ['localhost:26660']
    metrics_path: /metrics
```

---

## 🚀 Deployment Execution

### Final Deployment Steps
```bash
# 1. Create final backup
./backup_fluentum.sh

# 2. Stop existing node (if running)
sudo systemctl stop fluentum

# 3. Update configuration
cp config.toml.production ~/.fluentum/config/config.toml

# 4. Start node
sudo systemctl start fluentum

# 5. Verify deployment
./health_check.sh

# 6. Monitor logs
tail -f ~/.fluentum/logs/tendermint.log
```

### Deployment Verification
```bash
# Health check script
#!/bin/bash
echo "🔍 Verifying deployment..."

# Check process
if pgrep -x "fluentum" > /dev/null; then
    echo "✅ Process running"
else
    echo "❌ Process not running"
    exit 1
fi

# Check RPC
if curl -s http://localhost:26657/status > /dev/null; then
    echo "✅ RPC responding"
else
    echo "❌ RPC not responding"
    exit 1
fi

# Check sync
SYNC_STATUS=$(curl -s http://localhost:26657/status | jq -r '.result.sync_info.catching_up')
if [ "$SYNC_STATUS" = "false" ]; then
    echo "✅ Node synced"
else
    echo "⚠️  Node catching up"
fi

# Check peers
PEER_COUNT=$(curl -s http://localhost:26657/net_info | jq -r '.result.n_peers')
echo "🌐 Connected peers: $PEER_COUNT"

echo "✅ Deployment verification completed"
```

---

## 📊 Post-Deployment Monitoring

### 1. Immediate Monitoring (First 24 hours)
- [ ] Node status every 5 minutes
- [ ] Block production monitoring
- [ ] Network connectivity checks
- [ ] Error log monitoring
- [ ] Performance metrics tracking
- [ ] Resource usage monitoring

### 2. Ongoing Monitoring
- [ ] Daily health checks
- [ ] Weekly performance reviews
- [ ] Monthly security audits
- [ ] Quarterly backup verification
- [ ] Annual configuration reviews

### 3. Alert Configuration
- [ ] Node down alerts
- [ ] Sync issues alerts
- [ ] High resource usage alerts
- [ ] Error rate alerts
- [ ] Performance degradation alerts

---

## 🎯 Success Criteria

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
- [ ] Incident response plan ready

---

## 📞 Support & Escalation

### Primary Contacts
- **Technical Lead**: [Contact Info]
- **DevOps Engineer**: [Contact Info]
- **Security Team**: [Contact Info]

### Escalation Procedures
1. **Level 1**: Node operator attempts resolution
2. **Level 2**: Technical lead involvement
3. **Level 3**: Full team response
4. **Level 4**: External support engagement

### Emergency Procedures
- **Node Down**: Immediate restart procedures
- **Data Corruption**: Backup restoration
- **Security Breach**: Incident response plan
- **Performance Issues**: Scaling procedures

---

## ✅ Final Checklist Summary

### Pre-Deployment
- [ ] All tests passing
- [ ] Build verified
- [ ] Configuration validated
- [ ] Testnet validation completed

### Backup & Security
- [ ] Critical files backed up
- [ ] Backup integrity verified
- [ ] Security measures implemented
- [ ] Permissions set correctly

### Documentation
- [ ] Node operator docs updated
- [ ] CLI reference created
- [ ] Troubleshooting guide updated
- [ ] Deployment procedures documented

### Deployment
- [ ] Final backup created
- [ ] Configuration updated
- [ ] Node deployed successfully
- [ ] Health checks passing

### Post-Deployment
- [ ] Monitoring configured
- [ ] Alerts set up
- [ ] Team notified
- [ ] Support procedures established

---

**🎉 Congratulations! Your Fluentum deployment is ready for mainnet!**

For ongoing support and updates, please refer to the project documentation and community channels.
