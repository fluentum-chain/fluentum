# PowerShell build script for Fluentum AI-Validation Core with QMoE Consensus
# This script builds the QMoE validator as a shared library plugin

param(
    [string]$Action = "build"
)

# Configuration
$PluginName = "qmoe_validator"
$PluginDir = "fluentum/features/ai_validation"
$BuildDir = "build"
$OutputDir = "plugins"

# Colors for output
$Red = "Red"
$Green = "Green"
$Yellow = "Yellow"
$Blue = "Blue"

# Print colored output
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor $Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor $Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor $Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor $Red
}

# Check if Go is installed
function Test-Go {
    try {
        $goVersion = go version
        Write-Info "Found Go: $goVersion"
        return $true
    }
    catch {
        Write-Error "Go is not installed. Please install Go 1.19 or later."
        return $false
    }
}

# Check dependencies
function Test-Dependencies {
    Write-Info "Checking dependencies..."
    
    # Check for go.mod
    if (-not (Test-Path "go.mod")) {
        Write-Error "go.mod not found. Please run 'go mod init' first."
        return $false
    }
    
    # Check for CGO
    if ($env:CGO_ENABLED -ne "1") {
        Write-Warning "CGO is not enabled. Enabling CGO for plugin build..."
        $env:CGO_ENABLED = "1"
    }
    
    Write-Success "Dependencies check completed"
    return $true
}

# Create build directories
function New-BuildDirectories {
    Write-Info "Creating build directories..."
    
    if (-not (Test-Path $BuildDir)) {
        New-Item -ItemType Directory -Path $BuildDir | Out-Null
    }
    if (-not (Test-Path $OutputDir)) {
        New-Item -ItemType Directory -Path $OutputDir | Out-Null
    }
    if (-not (Test-Path "$BuildDir/$PluginDir")) {
        New-Item -ItemType Directory -Path "$BuildDir/$PluginDir" | Out-Null
    }
    
    Write-Success "Build directories created"
}

# Build the QMoE validator plugin
function Build-Plugin {
    Write-Info "Building QMoE validator plugin..."
    
    Push-Location $PluginDir
    
    try {
        # Build as shared library plugin
        Write-Info "Compiling QMoE validator with -buildmode=plugin..."
        
        $outputPath = "../../$OutputDir/${PluginName}.so"
        go build -buildmode=plugin -o $outputPath -ldflags="-s -w" .
        
        if ($LASTEXITCODE -eq 0) {
            Write-Success "QMoE validator plugin built successfully"
        }
        else {
            Write-Error "Failed to build QMoE validator plugin"
            return $false
        }
    }
    finally {
        Pop-Location
    }
    
    return $true
}

# Build for Windows
function Build-Windows {
    Write-Info "Building for Windows..."
    
    Push-Location $PluginDir
    
    try {
        $env:GOOS = "windows"
        $env:GOARCH = "amd64"
        $env:CGO_ENABLED = "1"
        
        $outputPath = "../../$OutputDir/${PluginName}_windows_amd64.dll"
        go build -buildmode=plugin -o $outputPath -ldflags="-s -w" .
        
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Built for Windows/amd64"
        }
        else {
            Write-Warning "Failed to build for Windows/amd64"
        }
    }
    finally {
        Pop-Location
        # Reset environment variables
        Remove-Item Env:GOOS -ErrorAction SilentlyContinue
        Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
    }
}

# Run tests
function Test-Plugin {
    Write-Info "Running tests..."
    
    Push-Location $PluginDir
    
    try {
        go test -v ./...
        
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Tests passed"
        }
        else {
            Write-Error "Tests failed"
            return $false
        }
    }
    finally {
        Pop-Location
    }
    
    return $true
}

# Run benchmarks
function Test-Benchmarks {
    Write-Info "Running benchmarks..."
    
    Push-Location $PluginDir
    
    try {
        go test -bench=. -benchmem ./...
    }
    finally {
        Pop-Location
    }
}

# Create plugin manifest
function New-PluginManifest {
    Write-Info "Creating plugin manifest..."
    
    $manifest = @{
        name = "QMoE Validator"
        version = "1.0.0"
        description = "Quantized Mixture-of-Experts consensus validator for Fluentum"
        type = "ai_validator"
        entry_point = "AIValidatorPlugin"
        platforms = @{
            "windows/amd64" = "${PluginName}_windows_amd64.dll"
        }
        dependencies = @{
            "go_version" = ">=1.19"
            "cgo" = $true
        }
        features = @(
            "quantized_moe_consensus",
            "predictive_batching",
            "dynamic_quantization",
            "gas_optimization",
            "adaptive_thresholds"
        )
        performance = @{
            "target_gas_savings" = "40%"
            "inference_time" = "<10ms"
            "memory_usage" = "<100MB"
        }
    }
    
    $manifest | ConvertTo-Json -Depth 10 | Out-File -FilePath "$OutputDir/plugin_manifest.json" -Encoding UTF8
    
    Write-Success "Plugin manifest created"
}

# Clean build artifacts
function Remove-BuildArtifacts {
    Write-Info "Cleaning build artifacts..."
    
    if (Test-Path $BuildDir) {
        Remove-Item -Recurse -Force $BuildDir
    }
    if (Test-Path $OutputDir) {
        Remove-Item -Recurse -Force $OutputDir
    }
    
    Write-Success "Build artifacts cleaned"
}

# Show help
function Show-Help {
    Write-Host "Fluentum AI-Validation Core Build Script" -ForegroundColor $Blue
    Write-Host ""
    Write-Host "Usage: .\build.ps1 [OPTIONS]" -ForegroundColor White
    Write-Host ""
    Write-Host "Options:" -ForegroundColor White
    Write-Host "  build           Build the QMoE validator plugin (default)" -ForegroundColor White
    Write-Host "  windows         Build for Windows platform" -ForegroundColor White
    Write-Host "  test            Run tests" -ForegroundColor White
    Write-Host "  benchmark       Run benchmarks" -ForegroundColor White
    Write-Host "  clean           Clean build artifacts" -ForegroundColor White
    Write-Host "  help            Show this help message" -ForegroundColor White
    Write-Host ""
    Write-Host "Environment variables:" -ForegroundColor White
    Write-Host "  CGO_ENABLED     Enable CGO (default: 1)" -ForegroundColor White
    Write-Host "  GOOS            Target operating system" -ForegroundColor White
    Write-Host "  GOARCH          Target architecture" -ForegroundColor White
    Write-Host ""
}

# Main function
function Main {
    param([string]$Action)
    
    Write-Info "Starting Fluentum AI-Validation Core build..."
    Write-Info "Action: $Action"
    
    switch ($Action.ToLower()) {
        "build" {
            if (-not (Test-Go)) { return }
            if (-not (Test-Dependencies)) { return }
            New-BuildDirectories
            if (Build-Plugin) {
                New-PluginManifest
            }
        }
        "windows" {
            if (-not (Test-Go)) { return }
            if (-not (Test-Dependencies)) { return }
            New-BuildDirectories
            Build-Windows
            New-PluginManifest
        }
        "test" {
            Test-Plugin
        }
        "benchmark" {
            Test-Benchmarks
        }
        "clean" {
            Remove-BuildArtifacts
        }
        "help" {
            Show-Help
        }
        default {
            Write-Error "Unknown action: $Action"
            Show-Help
            return
        }
    }
    
    Write-Success "Build process completed successfully!"
}

# Run main function
Main -Action $Action 