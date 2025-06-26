# Build the project
Write-Host "Building the project..."
go build ./...

# Run all tests
Write-Host "Running all tests..."
go test ./...

Write-Host "Build and test steps complete." 