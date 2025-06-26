# Fetch missing dependencies
Write-Host "Fetching missing dependencies..."
go get google.golang.org/grpc/status
go get google.golang.org/grpc/codes
go get github.com/golang/protobuf/proto

# (Optional) Add more go get lines here if needed

# Tidy up modules
Write-Host "Tidying up modules..."
go mod tidy

# Run all tests
Write-Host "Running all tests..."
go test ./... 