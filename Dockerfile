# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    make \
    gcc \
    musl-dev \
    linux-headers \
    build-base

# Set working directory
WORKDIR /fluentum

# Copy source code
COPY . .

# Build the binary
RUN make build

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    jq

# Create necessary directories
RUN mkdir -p /fluentum/config /quantum-keys

# Copy binary from builder
COPY --from=builder /fluentum/build/tendermint /usr/bin/fluentumd

# Set environment variables
ENV FLUENTUM_HOME=/fluentum

# Create volumes for persistent data
VOLUME ["/fluentum/config", "/quantum-keys"]

# Expose ports
EXPOSE 26656 26657 26660

# Set entrypoint
ENTRYPOINT ["fluentumd"]

# Default command
CMD ["start", \
     "--quantum.key-file=/quantum-keys/key.json", \
     "--zk-prover-url=https://zk.fluentum.net", \
     "--home=/fluentum"] 