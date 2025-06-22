# PowerShell script to fix compilation errors without updating dependencies

Write-Host "=== Fixing Installation Errors ===" -ForegroundColor Cyan
Write-Host ""

# Get Go module cache path
$goModCache = $env:GOMODCACHE
if (-not $goModCache) {
    $homeDir = $env:USERPROFILE
    $goModCache = Join-Path $homeDir "go\pkg\mod"
}

Write-Host "Go module cache: $goModCache" -ForegroundColor Yellow

# Fix 1: CometBFT secp256k1 assignment mismatch
Write-Host "Fixing CometBFT secp256k1 error..." -ForegroundColor Yellow

$cometbftPaths = @(
    "github.com\cometbft\cometbft@v0.38.0\crypto\secp256k1\secp256k1.go",
    "github.com\cometbft\cometbft@v0.37.0\crypto\secp256k1\secp256k1.go",
    "github.com\cometbft\cometbft@v0.36.0\crypto\secp256k1\secp256k1.go"
)

$cometbftFile = $null
foreach ($path in $cometbftPaths) {
    $fullPath = Join-Path $goModCache $path
    if (Test-Path $fullPath) {
        $cometbftFile = $fullPath
        Write-Host "Found CometBFT file: $cometbftFile" -ForegroundColor Green
        break
    }
}

if ($cometbftFile) {
    $content = Get-Content $cometbftFile -Raw
    $originalContent = $content
    
    # Fix the assignment mismatch error
    $content = $content -replace 'compactSig, err := ecdsa\.SignCompact', 'compactSig := ecdsa.SignCompact'
    
    if ($content -ne $originalContent) {
        $content | Set-Content $cometbftFile -NoNewline
        Write-Host "Fixed CometBFT secp256k1 assignment mismatch" -ForegroundColor Green
    } else {
        Write-Host "No changes needed for CometBFT secp256k1" -ForegroundColor Yellow
    }
} else {
    Write-Host "CometBFT secp256k1 file not found" -ForegroundColor Red
}

# Fix 2: Protobuf descriptor undefined constant
Write-Host "Fixing protobuf descriptor error..." -ForegroundColor Yellow

$protobufPaths = @(
    "github.com\golang\protobuf@v1.5.3\protoc-gen-go\descriptor\descriptor.pb.go",
    "github.com\golang\protobuf@v1.5.2\protoc-gen-go\descriptor\descriptor.pb.go",
    "github.com\golang\protobuf@v1.5.1\protoc-gen-go\descriptor\descriptor.pb.go"
)

$protobufFile = $null
foreach ($path in $protobufPaths) {
    $fullPath = Join-Path $goModCache $path
    if (Test-Path $fullPath) {
        $protobufFile = $fullPath
        Write-Host "Found protobuf file: $protobufFile" -ForegroundColor Green
        break
    }
}

if ($protobufFile) {
    $content = Get-Content $protobufFile -Raw
    $originalContent = $content
    
    # Fix the undefined constant error by adding a fallback
    $content = $content -replace 'descriptorpb\.Default_FileOptions_PhpGenericServices', 'false'
    
    if ($content -ne $originalContent) {
        $content | Set-Content $protobufFile -NoNewline
        Write-Host "Fixed protobuf descriptor undefined constant" -ForegroundColor Green
    } else {
        Write-Host "No changes needed for protobuf descriptor" -ForegroundColor Yellow
    }
} else {
    Write-Host "Protobuf descriptor file not found" -ForegroundColor Red
}

Write-Host ""
Write-Host "=== Summary ===" -ForegroundColor Cyan
Write-Host "Fixed compilation errors without updating dependencies" -ForegroundColor Green
Write-Host "You can now try running the installation again" -ForegroundColor White
Write-Host ""
Write-Host "If you still encounter errors, you may need to:" -ForegroundColor Yellow
Write-Host "1. Clear the Go module cache: go clean -modcache" -ForegroundColor White
Write-Host "2. Re-run this script after go mod download" -ForegroundColor White 