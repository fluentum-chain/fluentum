# Fluentum Build and Install Script (PowerShell)
# This script builds and installs the Fluentum binary with the correct name

param(
    [switch]$SkipDeps,
    [switch]$SkipTests
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

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

# Check if Go is installed
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Error "Go is not installed. Please install Go first."
    exit 1
}

Write-Status "Go version: $(go version)"

# Set GOPATH if not set
if (-not $env:GOPATH) {
    $env:GOPATH = "$env:USERPROFILE\go"
    Write-Status "Setting GOPATH to $env:GOPATH"
}

# Create build directory
Write-Status "Creating build directory..."
New-Item -ItemType Directory -Force -Path "build" | Out-Null

# Clean previous builds
Write-Status "Cleaning previous builds..."
Remove-Item -Path "build\fluentumd" -ErrorAction SilentlyContinue
Remove-Item -Path "build\fluentum" -ErrorAction SilentlyContinue

# Build the binary with the correct name
Write-Status "Building Fluentum Core as fluentumd..."
$buildCmd = "CGO_ENABLED=0 go build -mod=readonly -ldflags `"-X github.com/fluentum-chain/fluentum/version.TMCoreSemVer=$(git describe --tags --always --dirty) -s -w`" -tags 'tendermint,badgerdb' -o build/fluentumd ./cmd/fluentum/"

try {
    Invoke-Expression $buildCmd
    if ($LASTEXITCODE -eq 0) {
        Write-Success "Build successful!"
    } else {
        Write-Error "Build failed!"
        exit 1
    }
} catch {
    Write-Error "Build failed with error: $($_.Exception.Message)"
    exit 1
}

# Create GOPATH/bin directory if it doesn't exist
Write-Status "Creating GOPATH/bin directory..."
$gopathBin = "$env:GOPATH\bin"
New-Item -ItemType Directory -Force -Path $gopathBin | Out-Null

# Copy binary to GOPATH/bin
Write-Status "Installing fluentumd to $gopathBin..."
Copy-Item "build\fluentumd" "$gopathBin\" -Force

# Verify installation
if (Test-Path "$gopathBin\fluentumd") {
    Write-Success "Installation successful!"
    Write-Status "Binary location: $gopathBin\fluentumd"
    $fileInfo = Get-Item "$gopathBin\fluentumd"
    Write-Status "Binary size: $([math]::Round($fileInfo.Length / 1MB, 2)) MB"
    
    # Test the binary
    Write-Status "Testing binary..."
    try {
        & "$gopathBin\fluentumd" version | Out-Null
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Binary test successful!"
            Write-Host ""
            Write-Host "🎉 Fluentum Core has been successfully installed as 'fluentumd'" -ForegroundColor Green
            Write-Host ""
            Write-Host "You can now use the following commands:" -ForegroundColor Cyan
            Write-Host "  fluentumd version     - Check version"
            Write-Host "  fluentumd init        - Initialize a new node"
            Write-Host "  fluentumd start       - Start the node"
            Write-Host ""
        } else {
            Write-Warning "Binary test failed, but installation completed"
            Write-Warning "You may need to check the binary manually"
        }
    } catch {
        Write-Warning "Binary test failed, but installation completed"
        Write-Warning "You may need to check the binary manually"
    }
} else {
    Write-Error "Installation failed! Binary not found at $gopathBin\fluentumd"
    exit 1
} 