# Fluentum Public Testnet Setup Script for Windows
# This script sets up a testnet node on one of the 4 servers

param(
    [string]$NodeName = "fluentum-node1",
    [int]$NodeIndex = 1
)

# Testnet configuration
$TESTNET_CHAIN_ID = "fluentum-testnet-1"
$TESTNET_HOME = "C:\fluentum"
$TESTNET_USER = $env:USERNAME

# Server configurations
$SERVERS = @{
    "fluentum-node1" = "34.44.129.207"
    "fluentum-node2" = "34.44.82.114"
    "fluentum-node3" = "34.68.180.153"
    "fluentum-node4" = "34.72.252.153"
}

# Validate node name
if (-not $SERVERS.ContainsKey($NodeName)) {
    Write-Error "Invalid node name: $NodeName"
    Write-Host "Valid options: $($SERVERS.Keys -join ', ')"
    exit 1
}

$SERVER_IP = $SERVERS[$NodeName]
$P2P_PORT = 26656 + $NodeIndex - 1
$RPC_PORT = 26657 + $NodeIndex - 1
$API_PORT = 1317 + $NodeIndex - 1

Write-Host "Setting up Fluentum Testnet Node" -ForegroundColor Blue
Write-Host "Node Name: $NodeName"
Write-Host "Server IP: $SERVER_IP"
Write-Host "P2P Port: $P2P_PORT"
Write-Host "RPC Port: $RPC_PORT"
Write-Host "API Port: $API_PORT"
Write-Host "Chain ID: $TESTNET_CHAIN_ID"
Write-Host ""

# Check if fluentumd binary exists
if (-not (Test-Path ".\build\fluentumd.exe")) {
    Write-Error "fluentumd binary not found. Please build the project first:"
    Write-Host "  make build"
    exit 1
}

# Create testnet directory structure
Write-Host "Creating testnet directory structure..." -ForegroundColor Blue
New-Item -ItemType Directory -Force -Path $TESTNET_HOME
New-Item -ItemType Directory -Force -Path "$TESTNET_HOME\config"
New-Item -ItemType Directory -Force -Path "$TESTNET_HOME\data"
New-Item -ItemType Directory -Force -Path "$TESTNET_HOME\logs"

# Initialize the node
Write-Host "Initializing node..." -ForegroundColor Blue
try {
    & ".\build\fluentumd.exe" init $NodeName --chain-id $TESTNET_CHAIN_ID --home $TESTNET_HOME
    Write-Host "Node initialized successfully" -ForegroundColor Green
} catch {
    Write-Host "Node initialization failed, but continuing..." -ForegroundColor Yellow
}

# Generate node key if not exists
if (-not (Test-Path "$TESTNET_HOME\config\node_key.json")) {
    Write-Host "Generating node key..." -ForegroundColor Blue
    & ".\build\fluentumd.exe" gen-node-key --home $TESTNET_HOME
    Write-Host "Node key generated" -ForegroundColor Green
}

# Generate validator key if not exists
if (-not (Test-Path "$TESTNET_HOME\config\priv_validator_key.json")) {
    Write-Host "Generating validator key..." -ForegroundColor Blue
    & ".\build\fluentumd.exe" gen-validator-key --home $TESTNET_HOME
    Write-Host "Validator key generated" -ForegroundColor Green
}

# Create testnet configuration
Write-Host "Creating testnet configuration..." -ForegroundColor Blue
$configContent = @"
# Fluentum Testnet Configuration
chain_id = "$TESTNET_CHAIN_ID"
moniker = "$NodeName"

# Database backend: goleveldb (compatible with Tendermint)
db_backend = "goleveldb"
db_dir = "data"

# P2P Configuration
[p2p]
laddr = "tcp://0.0.0.0:$P2P_PORT"
external_address = "$SERVER_IP`:$P2P_PORT"
seeds = ""
persistent_peers = ""

# RPC Configuration
[rpc]
laddr = "tcp://0.0.0.0:$RPC_PORT"
cors_allowed_origins = ["*"]
cors_allowed_methods = ["HEAD", "GET", "POST"]
cors_allowed_headers = ["*"]
max_open_connections = 900
unsafe = false

# API Configuration
[api]
enable = true
swagger = true
address = "tcp://0.0.0.0:$API_PORT"

# Consensus Configuration (optimized for testnet)
[consensus]
timeout_propose = "1s"
timeout_propose_delta = "500ms"
timeout_prevote = "1s"
timeout_prevote_delta = "500ms"
timeout_precommit = "1s"
timeout_precommit_delta = "500ms"
timeout_commit = "1s"
create_empty_blocks = true
create_empty_blocks_interval = "10s"

# Mempool Configuration
[mempool]
version = "v0"
recheck = true
broadcast = true
size = 5000
max_txs_bytes = 1073741824
cache_size = 10000

# State Sync Configuration
[statesync]
enable = true
temp_dir = "C:\temp\fluentum-statesync"

# Instrumentation
[instrumentation]
prometheus = true
prometheus_listen_addr = ":26660"
namespace = "tendermint"
"@

$configContent | Out-File -FilePath "$TESTNET_HOME\config\config.toml" -Encoding UTF8
Write-Host "Testnet configuration created" -ForegroundColor Green

# Create genesis file if not exists
if (-not (Test-Path "$TESTNET_HOME\config\genesis.json")) {
    Write-Host "Creating genesis file..." -ForegroundColor Blue
    $genesisContent = @"
{
  "genesis_time": "2024-01-01T00:00:00Z",
  "chain_id": "$TESTNET_CHAIN_ID",
  "initial_height": "1",
  "consensus_params": {
    "block": {
      "max_bytes": 22020096,
      "max_gas": -1,
      "time_iota_ms": 1000
    },
    "evidence": {
      "max_age_num_blocks": 100000,
      "max_age_duration": "172800000000000",
      "max_bytes": 1048576
    },
    "validator": {
      "pub_key_types": ["ed25519"]
    },
    "version": {}
  },
  "validators": [],
  "app_hash": "",
  "app_state": {}
}
"@
    $genesisContent | Out-File -FilePath "$TESTNET_HOME\config\genesis.json" -Encoding UTF8
    Write-Host "Genesis file created" -ForegroundColor Green
}

# Create update peers script
Write-Host "Creating peer update script..." -ForegroundColor Blue
$updatePeersScript = @"
# Update persistent peers for testnet
`$TESTNET_HOME = "C:\fluentum"
`$CONFIG_FILE = "`$TESTNET_HOME\config\config.toml"

# Get all server IPs and P2P ports
`$SERVERS = @{
    "fluentum-node1" = "34.44.129.207:26656"
    "fluentum-node2" = "34.44.82.114:26657"
    "fluentum-node3" = "34.68.180.153:26658"
    "fluentum-node4" = "34.72.252.153:26659"
}

# Build persistent peers string (exclude current node)
`$CURRENT_NODE = (Get-Content `$CONFIG_FILE | Select-String "moniker" | ForEach-Object { `$_.Line.Split('"')[1] })
`$PERSISTENT_PEERS = ""

foreach (`$NODE in `$SERVERS.Keys) {
    if (`$NODE -ne `$CURRENT_NODE) {
        if (`$PERSISTENT_PEERS) {
            `$PERSISTENT_PEERS += ","
        }
        `$PERSISTENT_PEERS += `$SERVERS[`$NODE]
    }
}

# Update config file
`$configContent = Get-Content `$CONFIG_FILE
`$configContent = `$configContent -replace 'persistent_peers = ""', "persistent_peers = ``"`$PERSISTENT_PEERS``""
`$configContent | Out-File -FilePath `$CONFIG_FILE -Encoding UTF8

Write-Host "Updated persistent peers: `$PERSISTENT_PEERS"
"@

$updatePeersScript | Out-File -FilePath "$TESTNET_HOME\update_peers.ps1" -Encoding UTF8
Write-Host "Peer update script created" -ForegroundColor Green

# Create start script
Write-Host "Creating start script..." -ForegroundColor Blue
$startScript = @"
# Start Fluentum testnet node
Write-Host "Starting Fluentum testnet node: $NodeName"
Write-Host "Chain ID: $TESTNET_CHAIN_ID"
Write-Host "Home: $TESTNET_HOME"
Write-Host "P2P: $SERVER_IP`:$P2P_PORT"
Write-Host "RPC: http://$SERVER_IP`:$RPC_PORT"
Write-Host "API: http://$SERVER_IP`:$API_PORT"
Write-Host ""

# Update peers before starting
& "$TESTNET_HOME\update_peers.ps1"

# Start the node
& ".\build\fluentumd.exe" start `
    --home $TESTNET_HOME `
    --moniker $NodeName `
    --chain-id $TESTNET_CHAIN_ID `
    --testnet `
    --log_level info
"@

$startScript | Out-File -FilePath "$TESTNET_HOME\start_node.ps1" -Encoding UTF8
Write-Host "Start script created" -ForegroundColor Green

Write-Host "Testnet node setup complete!" -ForegroundColor Green
Write-Host ""
Write-Host "Node Information:"
Write-Host "  Name: $NodeName"
Write-Host "  IP: $SERVER_IP"
Write-Host "  P2P Port: $P2P_PORT"
Write-Host "  RPC Port: $RPC_PORT"
Write-Host "  API Port: $API_PORT"
Write-Host "  Chain ID: $TESTNET_CHAIN_ID"
Write-Host "  Home Directory: $TESTNET_HOME"
Write-Host ""
Write-Host "To start the node:"
Write-Host "  & `"$TESTNET_HOME\start_node.ps1`""
Write-Host ""
Write-Host "Configuration files:"
Write-Host "  Config: $TESTNET_HOME\config\config.toml"
Write-Host "  Genesis: $TESTNET_HOME\config\genesis.json"
Write-Host "  Node Key: $TESTNET_HOME\config\node_key.json"
Write-Host "  Validator Key: $TESTNET_HOME\config\priv_validator_key.json" 