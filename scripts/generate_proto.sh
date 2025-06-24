#!/bin/bash

# Generate protobuf files with proper import paths
# This script handles the protobuf generation for the Fluentum project

set -e

echo "Generating protobuf files..."

# Create output directories if they don't exist
mkdir -p proto/tendermint/abci
mkdir -p proto/tendermint/crypto
mkdir -p proto/tendermint/types
mkdir -p proto/tendermint/p2p
mkdir -p proto/tendermint/state
mkdir -p proto/tendermint/version
mkdir -p proto/tendermint/rpc/grpc
mkdir -p proto/tendermint/blockchain
mkdir -p proto/tendermint/consensus
mkdir -p proto/tendermint/privval

# Generate protobuf files with proper import paths
# The -I flag adds include paths for imports

# Generate abci types
protoc \
    --proto_path=proto \
    --proto_path=third_party/proto \
    --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/tendermint/abci/types.proto

# Generate crypto files
protoc \
    --proto_path=proto \
    --proto_path=third_party/proto \
    --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/tendermint/crypto/proof.proto \
    proto/tendermint/crypto/keys.proto

# Generate types files
protoc \
    --proto_path=proto \
    --proto_path=third_party/proto \
    --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/tendermint/types/types.proto \
    proto/tendermint/types/params.proto \
    proto/tendermint/types/validator.proto \
    proto/tendermint/types/block.proto \
    proto/tendermint/types/canonical.proto \
    proto/tendermint/types/evidence.proto

# Generate other proto files
protoc \
    --proto_path=proto \
    --proto_path=third_party/proto \
    --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/tendermint/p2p/types.proto \
    proto/tendermint/p2p/pex.proto \
    proto/tendermint/p2p/conn.proto

protoc \
    --proto_path=proto \
    --proto_path=third_party/proto \
    --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/tendermint/state/types.proto \
    proto/tendermint/version/types.proto \
    proto/tendermint/rpc/grpc/types.proto

protoc \
    --proto_path=proto \
    --proto_path=third_party/proto \
    --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/tendermint/blockchain/types.proto \
    proto/tendermint/consensus/types.proto \
    proto/tendermint/consensus/wal.proto \
    proto/tendermint/privval/types.proto

echo "Protobuf generation completed successfully!" 