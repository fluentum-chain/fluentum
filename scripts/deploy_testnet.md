# Fluentum Public Testnet Deployment Guide

This guide will help you deploy a Fluentum testnet across your 4 Ubuntu servers.

## Server Configuration

| Node Name | IP Address | P2P Port | RPC Port | API Port |
|-----------|------------|----------|----------|----------|
| fluentum-node1 | 34.44.129.207 | 26656 | 26657 | 1317 |
| fluentum-node2 | 34.44.82.114 | 26657 | 26658 | 1318 |
| fluentum-node3 | 34.68.180.153 | 26658 | 26659 | 1319 |
| fluentum-node4 | 34.72.252.153 | 26659 | 26660 | 1320 |

## Prerequisites

1. **Go 1.21+** installed on all servers
2. **Git** installed on all servers
3. **Make** installed on all servers
4. **Firewall ports** opened for P2P, RPC, and API communication
5. **SSH access** to all servers

## Step 1: Prepare All Servers

On each server, run these commands:

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install dependencies
sudo apt install -y build-essential git make curl wget

# Install Go (if not already installed)
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Clone the repository
git clone https://github.com/your-org/tendermint.git
cd tendermint

# Build the project
make build
```

## Step 2: Configure Firewall

On each server, open the necessary ports:

```bash
# Allow P2P, RPC, and API ports
sudo ufw allow 26656:26660/tcp
sudo ufw allow 1317:1320/tcp
sudo ufw allow 22/tcp  # SSH
sudo ufw enable
```

## Step 3: Deploy Nodes

### Option A: Using the Setup Script (Recommended)

On each server, run the setup script with the appropriate node name:

```bash
# On fluentum-node1 (34.44.129.207)
chmod +x scripts/setup_testnet.sh
./scripts/setup_testnet.sh fluentum-node1 1

# On fluentum-node2 (34.44.82.114)
chmod +x scripts/setup_testnet.sh
./scripts/setup_testnet.sh fluentum-node2 2

# On fluentum-node3 (34.68.180.153)
chmod +x scripts/setup_testnet.sh
./scripts/setup_testnet.sh fluentum-node3 3

# On fluentum-node4 (34.72.252.153)
chmod +x scripts/setup_testnet.sh
./scripts/setup_testnet.sh fluentum-node4 4
```

### Option B: Manual Setup

If you prefer manual setup, follow these steps on each server:

1. **Create directories:**
```bash
sudo mkdir -p /opt/fluentum
sudo chown $USER:$USER /opt/fluentum
mkdir -p /opt/fluentum/{config,data,logs}
```

2. **Initialize node:**
```bash
./build/fluentumd init <node-name> --chain-id fluentum-testnet-1 --home /opt/fluentum
```

3. **Generate keys:**
```bash
./build/fluentumd gen-node-key --home /opt/fluentum
./build/fluentumd gen-validator-key --home /opt/fluentum
```

4. **Configure the node** (see configuration templates below)

## Step 4: Start the Testnet

### Start Order (Important!)

Start the nodes in this order to ensure proper consensus:

1. **Start fluentum-node1 first:**
```bash
sudo systemctl start fluentum-testnet.service
sudo systemctl status fluentum-testnet.service
```

2. **Wait for node1 to be ready, then start fluentum-node2:**
```bash
sudo systemctl start fluentum-testnet.service
sudo systemctl status fluentum-testnet.service
```

3. **Continue with node3 and node4:**
```bash
sudo systemctl start fluentum-testnet.service
sudo systemctl status fluentum-testnet.service
```

## Step 5: Verify Testnet Health

### Check Node Status

On each node, check the status:

```bash
# Check service status
sudo systemctl status fluentum-testnet.service

# Check logs
sudo journalctl -u fluentum-testnet.service -f

# Check RPC endpoint
curl http://localhost:26657/status
```

### Check Network Connectivity

Test P2P connectivity between nodes:

```bash
# From any node, check if other nodes are reachable
curl http://34.44.129.207:26657/status  # node1
curl http://34.44.82.114:26658/status   # node2
curl http://34.68.180.153:26659/status  # node3
curl http://34.72.252.153:26660/status  # node4
```

### Check Consensus

Monitor consensus progress:

```bash
# Check block height
curl http://localhost:26657/status | jq '.result.sync_info.latest_block_height'

# Check validator set
curl http://localhost:26657/validators | jq '.result.validators | length'
```

## Step 6: Monitoring and Maintenance

### Log Monitoring

```bash
# Follow logs in real-time
sudo journalctl -u fluentum-testnet.service -f

# Check recent logs
sudo journalctl -u fluentum-testnet.service --since "1 hour ago"
```

### Performance Monitoring

```bash
# Check system resources
htop
df -h
free -h

# Check network connections
netstat -tulpn | grep fluentumd
```

### Backup and Recovery

```bash
# Backup important files
sudo cp -r /opt/fluentum/config /backup/fluentum-config-$(date +%Y%m%d)
sudo cp -r /opt/fluentum/data /backup/fluentum-data-$(date +%Y%m%d)
```

## Troubleshooting

### Common Issues

1. **Node won't start:**
   - Check logs: `sudo journalctl -u fluentum-testnet.service -f`
   - Verify configuration: `cat /opt/fluentum/config/config.toml`
   - Check permissions: `ls -la /opt/fluentum/`

2. **Nodes not connecting:**
   - Check firewall: `sudo ufw status`
   - Verify P2P ports are open
   - Check persistent_peers configuration

3. **Consensus issues:**
   - Ensure all nodes have the same genesis file
   - Check validator keys are unique
   - Verify chain ID is consistent

### Useful Commands

```bash
# Restart service
sudo systemctl restart fluentum-testnet.service

# Stop service
sudo systemctl stop fluentum-testnet.service

# Disable service
sudo systemctl disable fluentum-testnet.service

# Check service logs
sudo journalctl -u fluentum-testnet.service -n 100

# Check configuration
./build/fluentumd show-node-id --home /opt/fluentum
./build/fluentumd show-validator --home /opt/fluentum
```

## Security Considerations

1. **Firewall Configuration:**
   - Only open necessary ports
   - Consider using VPN for inter-node communication
   - Restrict RPC access to trusted IPs

2. **Key Management:**
   - Secure validator keys
   - Regular key rotation
   - Backup keys securely

3. **Monitoring:**
   - Set up alerts for node failures
   - Monitor disk space and memory usage
   - Track consensus performance

## Next Steps

Once your testnet is running:

1. **Deploy applications** that use the Fluentum blockchain
2. **Set up monitoring** and alerting systems
3. **Configure backup** and disaster recovery procedures
4. **Document** your deployment for team members
5. **Plan** for mainnet deployment

## Support

For issues and questions:

1. Check the logs first
2. Review this deployment guide
3. Check the Fluentum documentation
4. Contact the development team

---

**Note:** This is a testnet deployment. For production use, additional security measures and configurations will be required. 