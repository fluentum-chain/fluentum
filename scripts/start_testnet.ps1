# Fluentum Testnet Startup Script for Windows
# This script helps you start a Fluentum node in testnet mode

param(
    [string]$Moniker = "fluentum-testnet-node",
    [string]$ChainId = "fluentum-testnet-1",
    [string]$HomeDir = "$env:USERPROFILE\.fluentum",
    [string]$Seeds = "",
    [string]$PersistentPeers = "",
    [string]$GenesisAccount = "",
    [string]$GenesisCoins = "1000000000ufluentum,1000000000stake",
    [switch]$Background,
    [switch]$Help
)

# Function to print colored output
function Write-Status {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

function Write-Header {
    Write-Host "================================" -ForegroundColor Blue
    Write-Host "  Fluentum Testnet Node Setup  " -ForegroundColor Blue
    Write-Host "================================" -ForegroundColor Blue
}

# Function to show usage
function Show-Usage {
    Write-Host "Usage: .\start_testnet.ps1 [OPTIONS]"
    Write-Host ""
    Write-Host "Options:"
    Write-Host "  -Moniker NAME        Node moniker (default: fluentum-testnet-node)"
    Write-Host "  -ChainId ID          Chain ID (default: fluentum-testnet-1)"
    Write-Host "  -HomeDir DIR         Home directory (default: `$env:USERPROFILE\.fluentum)"
    Write-Host "  -Seeds SEEDS         Comma-separated list of seed nodes"
    Write-Host "  -PersistentPeers PEERS Comma-separated list of persistent peers"
    Write-Host "  -GenesisAccount NAME Genesis account name"
    Write-Host "  -Background          Start node in background"
    Write-Host "  -Help                Show this help message"
    Write-Host ""
    Write-Host "Examples:"
    Write-Host "  .\start_testnet.ps1                                    # Start with default settings"
    Write-Host "  .\start_testnet.ps1 -Moniker my-node -ChainId test-chain-1"
    Write-Host "  .\start_testnet.ps1 -Seeds 'node1:26656,node2:26656'"
    Write-Host "  .\start_testnet.ps1 -GenesisAccount validator -Background"
    Write-Host ""
}

# Show help if requested
if ($Help) {
    Show-Usage
    exit 0
}

# Function to check if fluentumd is installed
function Test-Fluentumd {
    try {
        $version = fluentumd version 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Status "Found fluentumd: $version"
            return $true
        }
    }
    catch {
        Write-Error "fluentumd is not installed or not in PATH"
        Write-Host "Please build and install fluentumd first:"
        Write-Host "  make build"
        Write-Host "  make install"
        return $false
    }
    return $false
}

# Function to initialize node if needed
function Initialize-Node {
    param(
        [string]$HomeDir,
        [string]$Moniker,
        [string]$ChainId
    )

    $configDir = Join-Path $HomeDir "config"
    if (-not (Test-Path $configDir)) {
        Write-Status "Initializing new node..."
        fluentumd init $Moniker --chain-id $ChainId --home $HomeDir
        if ($LASTEXITCODE -eq 0) {
            Write-Status "Node initialized successfully"
        } else {
            Write-Error "Failed to initialize node"
            exit 1
        }
    } else {
        Write-Status "Node already initialized at $HomeDir"
    }
}

# Function to configure the node
function Set-NodeConfiguration {
    param(
        [string]$HomeDir,
        [string]$Moniker,
        [string]$Seeds,
        [string]$PersistentPeers
    )

    $configFile = Join-Path $HomeDir "config\config.toml"
    $appConfigFile = Join-Path $HomeDir "config\app.toml"

    Write-Status "Configuring node..."

    # Update config.toml
    if (Test-Path $configFile) {
        $content = Get-Content $configFile -Raw
        
        # Set moniker
        $content = $content -replace 'moniker = ".*?"', "moniker = `"$Moniker`""
        
        # Configure P2P
        $content = $content -replace 'laddr = "tcp://127\.0\.0\.1:26656"', 'laddr = "tcp://0.0.0.0:26656"'
        
        # Set seeds if provided
        if ($Seeds) {
            $content = $content -replace 'seeds = ""', "seeds = `"$Seeds`""
        }
        
        # Set persistent peers if provided
        if ($PersistentPeers) {
            $content = $content -replace 'persistent_peers = ""', "persistent_peers = `"$PersistentPeers`""
        }
        
        # Configure RPC
        $content = $content -replace 'laddr = "tcp://127\.0\.0\.1:26657"', 'laddr = "tcp://0.0.0.0:26657"'
        
        # Configure consensus for testnet (faster block times)
        $content = $content -replace 'timeout_commit = "5s"', 'timeout_commit = "1s"'
        $content = $content -replace 'timeout_propose = "3s"', 'timeout_propose = "1s"'
        $content = $content -replace 'create_empty_blocks_interval = "0s"', 'create_empty_blocks_interval = "10s"'
        
        Set-Content $configFile $content -NoNewline
        Write-Status "config.toml updated"
    }

    # Update app.toml
    if (Test-Path $appConfigFile) {
        $content = Get-Content $appConfigFile -Raw
        
        # Enable API
        $content = $content -replace 'enable = false', 'enable = true'
        $content = $content -replace 'swagger = false', 'swagger = true'
        
        # Enable gRPC
        $content = $content -replace 'enable = false', 'enable = true'
        
        Set-Content $appConfigFile $content -NoNewline
        Write-Status "app.toml updated"
    }
}

# Function to create genesis account if needed
function New-GenesisAccount {
    param(
        [string]$HomeDir,
        [string]$AccountName,
        [string]$Coins
    )

    if ($AccountName -and $Coins) {
        Write-Status "Creating genesis account: $AccountName with $Coins"
        
        # Add key if it doesn't exist
        $keyExists = fluentumd keys show $AccountName --keyring-backend test --home $HomeDir 2>$null
        if ($LASTEXITCODE -ne 0) {
            fluentumd keys add $AccountName --keyring-backend test --home $HomeDir --output json --no-backup
        }

        # Get address
        $address = fluentumd keys show $AccountName -a --keyring-backend test --home $HomeDir
        
        # Add genesis account
        fluentumd add-genesis-account $address $Coins --home $HomeDir --keyring-backend test
        
        Write-Status "Genesis account created: $address"
    }
}

# Function to start the node
function Start-FluentumNode {
    param(
        [string]$HomeDir,
        [string]$ChainId,
        [bool]$Background
    )

    Write-Status "Starting Fluentum testnet node..."
    Write-Host "Chain ID: $ChainId"
    Write-Host "Home directory: $HomeDir"
    Write-Host "RPC endpoint: http://localhost:26657"
    Write-Host "API endpoint: http://localhost:1317"
    Write-Host "P2P endpoint: localhost:26656"
    Write-Host ""

    if ($Background) {
        Write-Status "Starting node in background..."
        $job = Start-Job -ScriptBlock {
            param($HomeDir, $ChainId)
            fluentumd start --home $HomeDir --chain-id $ChainId --testnet --api --grpc --grpc-web
        } -ArgumentList $HomeDir, $ChainId
        
        Write-Status "Node started in background (Job ID: $($job.Id))"
        Write-Status "To stop: Stop-Job $($job.Id); Remove-Job $($job.Id)"
        Write-Status "To view output: Receive-Job $($job.Id)"
    } else {
        Write-Status "Starting node in foreground..."
        fluentumd start --home $HomeDir --chain-id $ChainId --testnet --api --grpc --grpc-web
    }
}

# Main execution
Write-Header

# Check prerequisites
if (-not (Test-Fluentumd)) {
    exit 1
}

# Initialize node
Initialize-Node $HomeDir $Moniker $ChainId

# Configure node
Set-NodeConfiguration $HomeDir $Moniker $Seeds $PersistentPeers

# Create genesis account if specified
if ($GenesisAccount) {
    New-GenesisAccount $HomeDir $GenesisAccount $GenesisCoins
}

# Start the node
Start-FluentumNode $HomeDir $ChainId $Background

Write-Status "Setup complete!"
Write-Host ""
Write-Host "Node endpoints:"
Write-Host "  RPC:     http://localhost:26657"
Write-Host "  API:     http://localhost:1317"
Write-Host "  gRPC:    localhost:9090"
Write-Host "  gRPC-Web: localhost:9091"
Write-Host "  P2P:     localhost:26656"
Write-Host ""
Write-Host "Useful commands:"
Write-Host "  fluentumd status --home $HomeDir"
Write-Host "  fluentumd query bank balances --home $HomeDir"
Write-Host "  fluentumd tendermint show-node-id --home $HomeDir"
Write-Host "" 