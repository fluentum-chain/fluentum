# Install required protobuf plugins for Go
Write-Host "Installing protobuf plugins..." -ForegroundColor Green

# Install standard Go protobuf plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Also install gogofaster as backup (used in some parts of the codebase)
go install github.com/gogo/protobuf/protoc-gen-gogofaster@latest

Write-Host "Protobuf plugins installed successfully!" -ForegroundColor Green
Write-Host "You can now run: make proto-gen" -ForegroundColor Yellow 