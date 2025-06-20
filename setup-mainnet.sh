#!/bin/bash

# Fluentum Mainnet Quick Setup Script
# Run this on your Ubuntu server at 34.70.86.80

set -e

echo "🚀 Setting up Fluentum Mainnet Node..."

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Check if fluentum is installed
if ! command -v fluentum &> /dev/null; then
    echo -e "${YELLOW}❌ Fluentum not found. Please install first.${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Fluentum found: $(fluentum version)${NC}"

# Initialize node
echo "📝 Initializing Fluentum node..."
fluentum init fluentum-mainnet --chain-id fluentum-mainnet-1 --moniker "Fluentum-Mainnet-Node"

# Configure external address
echo "🔧 Configuring external address..."
sed -i 's/external_address = ""/external_address = "34.70.86.80:26656"/' ~/.fluentum/config/config.toml

# Enable API
echo "🔧 Enabling API..."
sed -i 's/enable = false/enable = true/' ~/.fluentum/config/app.toml
sed -i 's/swagger = false/swagger = true/' ~/.fluentum/config/app.toml

# Configure RPC
echo "🔧 Configuring RPC..."
sed -i 's/laddr = "tcp:\/\/127.0.0.1:26657"/laddr = "tcp:\/\/0.0.0.0:26657"/' ~/.fluentum/config/config.toml

# Configure P2P
echo "🔧 Configuring P2P..."
sed -i 's/laddr = "tcp:\/\/0.0.0.0:26656"/laddr = "tcp:\/\/0.0.0.0:26656"/' ~/.fluentum/config/config.toml

# Setup firewall
echo "🔥 Configuring firewall..."
sudo ufw allow 26656/tcp
sudo ufw allow 26657/tcp
sudo ufw allow 1317/tcp

# Create monitoring script
echo "📊 Creating monitoring script..."
cat > monitor.sh << 'EOF'
#!/bin/bash
echo "=== Fluentum Node Status ==="
echo "Time: $(date)"
echo ""

if pgrep -x "fluentum" > /dev/null; then
    echo "✅ Fluentum is running"
    echo "Latest Block: $(curl -s http://localhost:26657/block | jq -r '.result.block.header.height // "N/A"')"
    echo "Sync Status: $(curl -s http://localhost:26657/status | jq -r '.result.sync_info.catching_up // "N/A"')"
    echo "Peers: $(curl -s http://localhost:26657/net_info | jq -r '.result.n_peers // "N/A"')"
else
    echo "❌ Fluentum is not running"
fi
EOF

chmod +x monitor.sh

# Create health check script
echo "🔍 Creating health check script..."
cat > health_check.sh << 'EOF'
#!/bin/bash
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "🔍 Fluentum Node Health Check"
echo "=============================="

if pgrep -x "fluentum" > /dev/null; then
    echo -e "${GREEN}✅ Process Status: Running${NC}"
else
    echo -e "${RED}❌ Process Status: Not Running${NC}"
    exit 1
fi

if curl -s http://localhost:26657/status > /dev/null; then
    echo -e "${GREEN}✅ RPC Endpoint: Accessible${NC}"
else
    echo -e "${RED}❌ RPC Endpoint: Not Accessible${NC}"
fi

SYNC_STATUS=$(curl -s http://localhost:26657/status | jq -r '.result.sync_info.catching_up // "unknown"')
if [ "$SYNC_STATUS" = "false" ]; then
    echo -e "${GREEN}✅ Sync Status: Caught Up${NC}"
elif [ "$SYNC_STATUS" = "true" ]; then
    echo -e "${YELLOW}⚠️  Sync Status: Catching Up${NC}"
else
    echo -e "${RED}❌ Sync Status: Unknown${NC}"
fi

LATEST_BLOCK=$(curl -s http://localhost:26657/block | jq -r '.result.block.header.height // "N/A"')
echo -e "${GREEN}📦 Latest Block: $LATEST_BLOCK${NC}"

PEER_COUNT=$(curl -s http://localhost:26657/net_info | jq -r '.result.n_peers // 0')
echo -e "${GREEN}🌐 Connected Peers: $PEER_COUNT${NC}"

echo "=============================="
echo "Health check completed!"
EOF

chmod +x health_check.sh

echo ""
echo -e "${GREEN}✅ Setup completed!${NC}"
echo ""
echo "Next steps:"
echo "1. Start the node: fluentum start"
echo "2. Check status: ./monitor.sh"
echo "3. Health check: ./health_check.sh"
echo ""
echo "Your node will be accessible at:"
echo "  RPC: http://34.70.86.80:26657"
echo "  API: http://34.70.86.80:1317"
echo "  P2P: 34.70.86.80:26656"
echo ""
echo "To start the node in background:"
echo "  nohup fluentum start > fluentum.log 2>&1 &"
echo ""
echo "To view logs:"
echo "  tail -f fluentum.log" 