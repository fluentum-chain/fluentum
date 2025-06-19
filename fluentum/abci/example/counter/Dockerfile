# Start from the official Golang image for building
FROM golang:1.21 as builder

WORKDIR /app

# Copy go.mod and go.sum first for dependency caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the counter demo app
RUN go build -o counter-demo ./fluentum/abci/example/counter/cmd/main.go

# Use a minimal base image for running
FROM debian:bullseye-slim
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/counter-demo /app/counter-demo

EXPOSE 26658

# Run the counter ABCI app
ENTRYPOINT ["/app/counter-demo"] 