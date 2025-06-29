# Fluentum Node Initialization and Startup Script (PowerShell)
# This script helps initialize and start a Fluentum node

param(
    [string]$HomeDir = "/tmp/fluentum-new-test",
    [string]$Moniker = "fluentum-node",
    [string]$ChainId = "fluentum-mainnet-1",
    [bool]$Testnet = $false
)

# Function to print colored output
function Write-Status {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

Write-Status "Fluentum Node Setup Script"
Write-Host "Home Directory: $HomeDir"
Write-Host "Moniker: $Moniker"
Write-Host "Chain ID: $ChainId"
Write-Host "Testnet Mode: $Testnet"
Write-Host ""

# Check if fluentumd binary exists
if (-not (Test-Path "./build/fluentumd.exe")) {
    Write-Error "fluentumd binary not found. Please build the project first:"
    Write-Host "  make build"
    exit 1
}

Write-Status "Step 1: Initializing node..."
try {
    & "./build/fluentumd.exe" init $Moniker --chain-id $ChainId --home $HomeDir
    Write-Success "Node initialized successfully"
} catch {
    Write-Warning "Node initialization failed, but continuing..."
}

Write-Status "Step 2: Checking configuration..."
$configPath = Join-Path $HomeDir "config\config.toml"
if (Test-Path $configPath) {
    Write-Success "Configuration file found"
} else {
    Write-Warning "Configuration file not found, creating minimal config..."
    
    # Create config directory
    $configDir = Join-Path $HomeDir "config"
    New-Item -ItemType Directory -Force -Path $configDir | Out-Null
    
    # Create minimal config.toml
    $configContent = @"
# Fluentum Node Configuration
chain_id = "$ChainId"
moniker = "$Moniker"

# Database backend: goleveldb (compatible with Tendermint)
db_backend = "goleveldb"
db_dir = "data"

[p2p]
laddr = "tcp://0.0.0.0:26656"

[rpc]
laddr = "tcp://0.0.0.0:26657"

[consensus]
timeout_commit = "5s"
timeout_propose = "3s"
"@
    
    $configContent | Out-File -FilePath $configPath -Encoding UTF8
    Write-Success "Minimal configuration created"
}

Write-Status "Step 3: Checking node key..."
$nodeKeyPath = Join-Path $HomeDir "config\node_key.json"
if (Test-Path $nodeKeyPath) {
    Write-Success "Node key found"
} else {
    Write-Warning "Node key not found, generating..."
    try {
        & "./build/fluentumd.exe" gen-node-key --home $HomeDir
        Write-Success "Node key generated"
    } catch {
        Write-Error "Failed to generate node key"
        exit 1
    }
}

Write-Status "Step 4: Checking validator key..."
$validatorKeyPath = Join-Path $HomeDir "config\priv_validator_key.json"
if (Test-Path $validatorKeyPath) {
    Write-Success "Validator key found"
} else {
    Write-Warning "Validator key not found, generating..."
    try {
        & "./build/fluentumd.exe" gen-validator-key --home $HomeDir
        Write-Success "Validator key generated"
    } catch {
        Write-Error "Failed to generate validator key"
        exit 1
    }
}

Write-Status "Step 5: Checking genesis file..."
$genesisPath = Join-Path $HomeDir "config\genesis.json"
if (Test-Path $genesisPath) {
    Write-Success "Genesis file found"
} else {
    Write-Warning "Genesis file not found, creating minimal genesis..."
    
    # Create minimal genesis.json
    $genesisContent = @"
{
  "genesis_time": "2024-01-01T00:00:00Z",
  "chain_id": "$ChainId",
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
    
    $genesisContent | Out-File -FilePath $genesisPath -Encoding UTF8
    Write-Success "Minimal genesis file created"
}

Write-Status "Step 6: Starting node..."
Write-Host ""
Write-Host "Starting Fluentum node with the following configuration:"
Write-Host "  Home Directory: $HomeDir"
Write-Host "  Moniker: $Moniker"
Write-Host "  Chain ID: $ChainId"
Write-Host "  RPC Endpoint: http://localhost:26657"
Write-Host "  P2P Endpoint: localhost:26656"
Write-Host ""

# Build the start command
$startCmd = "./build/fluentumd.exe start --home $HomeDir --moniker $Moniker --chain-id $ChainId"

if ($Testnet) {
    $startCmd += " --testnet"
    Write-Host "  Testnet Mode: Enabled"
}

Write-Host "Command: $startCmd"
Write-Host ""

# Ask user if they want to start the node
$response = Read-Host "Do you want to start the node now? (y/n)"
if ($response -eq "y" -or $response -eq "Y") {
    Write-Status "Starting Fluentum node..."
    Invoke-Expression $startCmd
} else {
    Write-Status "Node setup complete. To start the node manually, run:"
    Write-Host "  $startCmd"
}

Write-Success "Setup complete!" 