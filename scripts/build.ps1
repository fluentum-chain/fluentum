# Fluentum Core Build Script (PowerShell)
# This script handles the build process with automatic dependency management

param(
    [switch]$SkipDeps,
    [switch]$SkipTests,
    [switch]$Verbose,
    [switch]$Help
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

# Function to check if command exists
function Test-Command {
    param([string]$Command)
    try {
        Get-Command $Command -ErrorAction Stop | Out-Null
        return $true
    }
    catch {
        return $false
    }
}

# Show help
if ($Help) {
    Write-Host "Usage: .\scripts\build.ps1 [OPTIONS]"
    Write-Host ""
    Write-Host "Options:"
    Write-Host "  -SkipDeps     Skip dependency management"
    Write-Host "  -SkipTests    Skip running tests"
    Write-Host "  -Verbose      Enable verbose output"
    Write-Host "  -Help         Show this help message"
    Write-Host ""
    Write-Host "Examples:"
    Write-Host "  .\scripts\build.ps1                    # Full build with deps and tests"
    Write-Host "  .\scripts\build.ps1 -SkipDeps          # Build without dependency management"
    Write-Host "  .\scripts\build.ps1 -SkipTests         # Build without running tests"
    Write-Host "  .\scripts\build.ps1 -SkipDeps -SkipTests  # Quick build only"
    exit 0
}

# Check prerequisites
function Test-Prerequisites {
    Write-Status "Checking prerequisites..."
    
    if (-not (Test-Command "go")) {
        Write-Error "Go is not installed. Please install Go 1.24.4 or later."
        exit 1
    }
    
    if (-not (Test-Command "make")) {
        Write-Error "Make is not installed. Please install make."
        exit 1
    }
    
    Write-Success "Prerequisites check passed"
}

# Function to handle dependencies
function Invoke-DependencyManagement {
    Write-Status "Managing dependencies..."
    
    # Check if go.mod exists
    if (-not (Test-Path "go.mod")) {
        Write-Error "go.mod file not found. Are you in the correct directory?"
        exit 1
    }
    
    # Run go mod tidy
    Write-Status "Running go mod tidy..."
    try {
        go mod tidy
        Write-Success "Dependencies updated successfully"
    }
    catch {
        Write-Error "Failed to update dependencies"
        exit 1
    }
    
    # Verify dependencies
    Write-Status "Verifying dependencies..."
    try {
        go mod verify
        Write-Success "Dependencies verified successfully"
    }
    catch {
        Write-Error "Dependency verification failed"
        exit 1
    }
}

# Function to build the project
function Invoke-Build {
    Write-Status "Building Fluentum Core..."
    
    # Clean previous builds
    Write-Status "Cleaning previous builds..."
    try {
        make clean
    }
    catch {
        Write-Warning "Clean failed, continuing with build..."
    }
    
    # Build the project
    Write-Status "Building with automatic dependency management..."
    try {
        make build
        Write-Success "Build completed successfully"
    }
    catch {
        Write-Error "Build failed"
        exit 1
    }
}

# Function to run tests
function Invoke-Tests {
    Write-Status "Running tests..."
    
    try {
        make test
        Write-Success "Tests passed"
    }
    catch {
        Write-Error "Tests failed"
        exit 1
    }
}

# Function to show build info
function Show-BuildInfo {
    Write-Status "Build Information:"
    Write-Host "  Go version: $(go version)"
    Write-Host "  Build target: $(Get-Location)\build\fluentum"
    Write-Host "  Build time: $(Get-Date)"
    
    if (Test-Path "build\fluentum") {
        $fileInfo = Get-Item "build\fluentum"
        Write-Host "  Binary size: $([math]::Round($fileInfo.Length / 1MB, 2)) MB"
        Write-Host "  Binary created: $($fileInfo.LastWriteTime)"
    }
}

# Main execution
try {
    Write-Host "=========================================="
    Write-Host "    Fluentum Core Build Script"
    Write-Host "=========================================="
    Write-Host ""
    
    # Check prerequisites
    Test-Prerequisites
    
    # Handle dependencies (unless skipped)
    if (-not $SkipDeps) {
        Invoke-DependencyManagement
    }
    else {
        Write-Warning "Skipping dependency management"
    }
    
    # Build the project
    Invoke-Build
    
    # Run tests (unless skipped)
    if (-not $SkipTests) {
        Invoke-Tests
    }
    else {
        Write-Warning "Skipping tests"
    }
    
    # Show build information
    Show-BuildInfo
    
    Write-Success "Build process completed successfully!"
    Write-Host ""
    Write-Host "Next steps:"
    Write-Host "  - Run 'make install' to install the binary"
    Write-Host "  - Run 'make init-node' to initialize a new node"
    Write-Host "  - Run 'make start' to start the node"
    Write-Host "  - Run 'make help' for more commands"
}
catch {
    Write-Error "Build process failed: $($_.Exception.Message)"
    exit 1
} 