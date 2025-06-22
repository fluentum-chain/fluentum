# Migration script for Fluentum blockchain from Tendermint to CometBFT
# This script uses confix to migrate configuration files

param(
    [string]$CMTHOME = "$env:USERPROFILE\.cometbft",
    [string]$TMHOME = "$env:USERPROFILE\.tendermint"
)

Write-Host "Starting Fluentum blockchain migration from Tendermint to CometBFT..." -ForegroundColor Green

# Check if confix is installed
try {
    $null = Get-Command confix -ErrorAction Stop
    Write-Host "Confix is already installed." -ForegroundColor Green
} catch {
    Write-Host "Installing confix..." -ForegroundColor Yellow
    go install github.com/cometbft/confix@latest
}

Write-Host "Using CMTHOME: $CMTHOME" -ForegroundColor Cyan
Write-Host "Using TMHOME: $TMHOME" -ForegroundColor Cyan

# Create CometBFT directory if it doesn't exist
if (-not (Test-Path $CMTHOME)) {
    Write-Host "Creating CometBFT home directory..." -ForegroundColor Yellow
    New-Item -ItemType Directory -Path $CMTHOME -Force | Out-Null
}

# Migrate configuration if Tendermint config exists
if (Test-Path $TMHOME) {
    Write-Host "Migrating configuration from Tendermint to CometBFT..." -ForegroundColor Yellow
    
    # Copy existing config files
    try {
        Copy-Item -Path "$TMHOME\*" -Destination $CMTHOME -Recurse -Force
        Write-Host "Configuration files copied successfully." -ForegroundColor Green
    } catch {
        Write-Host "Warning: Some files could not be copied." -ForegroundColor Yellow
    }
    
    # Use confix to migrate the configuration
    Write-Host "Running confix migration..." -ForegroundColor Yellow
    try {
        confix migrate --home $CMTHOME --target-version v0.38.6
        Write-Host "Configuration migration completed!" -ForegroundColor Green
    } catch {
        Write-Host "Warning: Confix migration failed. You may need to manually update the configuration." -ForegroundColor Yellow
    }
} else {
    Write-Host "No existing Tendermint configuration found. Creating new CometBFT configuration..." -ForegroundColor Yellow
    
    # Initialize new CometBFT configuration
    try {
        fluentumd init --home $CMTHOME
        Write-Host "New CometBFT configuration created." -ForegroundColor Green
    } catch {
        Write-Host "Warning: Could not initialize new configuration. You may need to do this manually." -ForegroundColor Yellow
    }
}

# Update environment variables
Write-Host "Updating environment variables..." -ForegroundColor Yellow

# Set CMTHOME environment variable for current session
$env:CMTHOME = $CMTHOME

# Remove TMHOME from current session if it exists
if ($env:TMHOME) {
    Remove-Item Env:TMHOME
}

# Update user environment variables permanently
try {
    [Environment]::SetEnvironmentVariable("CMTHOME", $CMTHOME, "User")
    [Environment]::SetEnvironmentVariable("TMHOME", $null, "User")
    Write-Host "Environment variables updated permanently." -ForegroundColor Green
} catch {
    Write-Host "Warning: Could not update permanent environment variables. You may need to do this manually." -ForegroundColor Yellow
}

Write-Host "Migration completed successfully!" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "1. Run 'go mod tidy' to resolve dependencies" -ForegroundColor White
Write-Host "2. Build the application: 'make build'" -ForegroundColor White
Write-Host "3. Start the node: 'fluentumd start --home $CMTHOME'" -ForegroundColor White
Write-Host ""
Write-Host "Note: The old Tendermint configuration is still available at $TMHOME" -ForegroundColor Yellow
Write-Host "You can remove it after confirming everything works correctly." -ForegroundColor Yellow 