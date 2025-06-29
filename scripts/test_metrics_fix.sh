#!/bin/bash

echo "=== Testing Metrics Fix ==="
echo "Current directory: $(pwd)"
echo "Binary location: $(which fluentumd)"
echo "Binary size: $(ls -lh $(which fluentumd) 2>/dev/null || echo 'Binary not found in PATH')"

echo ""
echo "=== Checking if binary was rebuilt ==="
if [ -f "build/fluentumd" ]; then
    echo "Build binary exists: $(ls -lh build/fluentumd)"
    echo "Build time: $(stat -c %y build/fluentumd 2>/dev/null || stat -f %Sm build/fluentumd 2>/dev/null || echo 'Cannot get build time')"
else
    echo "Build binary not found"
fi

echo ""
echo "=== Testing node creation with debug output ==="
# Test with minimal configuration to isolate the issue
export TENDERMINT_LOG_LEVEL=debug

echo "Creating test configuration..."
mkdir -p /tmp/test_fluentum/config
mkdir -p /tmp/test_fluentum/data

# Create minimal config
cat > /tmp/test_fluentum/config/config.toml << EOF
# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

proxy_app = "tcp://127.0.0.1:26658"
moniker = "test-node"
fast_sync = true

[consensus]
timeout_commit = "5s"
timeout_propose = "3s"

[p2p]
laddr = "tcp://0.0.0.0:26656"
external_address = ""
seeds = ""
max_num_inbound_peers = 40
max_num_outbound_peers = 10

[rpc]
laddr = "tcp://0.0.0.0:26657"
cors_allowed_origins = []
cors_allowed_methods = ["HEAD", "GET", "POST"]
cors_allowed_headers = ["Origin", "X-Requested-With", "Content-Type", "Accept"]

[mempool]
recheck = true
broadcast = true
wal_dir = "data/mempool.wal"

[instrumentation]
prometheus = false
prometheus_listen_addr = ":26660"
max_open_connections = 3
namespace = "tendermint"
EOF

echo "Testing node startup with minimal config..."
timeout 10s fluentumd start --home /tmp/test_fluentum --moniker test-node --chain-id test-chain-1 --log_level debug 2>&1 | head -20

echo ""
echo "=== Checking for any remaining Prometheus references ==="
if command -v grep >/dev/null 2>&1; then
    echo "Searching for PrometheusMetrics calls in binary..."
    strings $(which fluentumd) 2>/dev/null | grep -i prometheus | head -5
else
    echo "grep not available"
fi

echo ""
echo "=== Cleanup ==="
rm -rf /tmp/test_fluentum

echo "Test completed." 