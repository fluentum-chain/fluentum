# Install all missing Go dependencies for the project

# Core dependencies
Write-Host "Installing core dependencies..."
go get github.com/cometbft/cometbft/abci/types@v0.38.6
go get github.com/cometbft/cometbft/crypto/ed25519@v0.38.6
go get github.com/cometbft/cometbft/crypto/secp256k1@v0.38.6

# Tendermint (CometBFT) and DB
Write-Host "Installing Tendermint/CometBFT and DB dependencies..."
go get github.com/tendermint/tendermint
go get github.com/cometbft/cometbft-db

# Protobuf and related
Write-Host "Installing protobuf and related dependencies..."
go get github.com/gogo/protobuf/proto
go get github.com/gogo/protobuf/types
go get github.com/cosmos/gogoproto/gogoproto
go get github.com/cosmos/gogoproto/proto
go get github.com/cosmos/gogoproto/types
go get github.com/cosmos/gogoproto/jsonpb
go get github.com/cosmos/gogoproto/grpc

# Crypto and quantum
Write-Host "Installing crypto and quantum dependencies..."
go get github.com/cloudflare/circl/sign/dilithium
go get github.com/fluentum-chain/dilithium
go get github.com/decred/dcrd/dcrec/secp256k1/v4
go get github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa
go get github.com/btcsuite/btcd/btcec/v2
go get github.com/btcsuite/btcd/btcec/v2/ecdsa
go get golang.org/x/crypto/ed25519
go get golang.org/x/crypto/chacha20poly1305
go get golang.org/x/crypto/curve25519
go get golang.org/x/crypto/hkdf
go get golang.org/x/crypto/nacl/box
go get golang.org/x/crypto/ripemd160

# Google and cloud
Write-Host "Installing Google and cloud dependencies..."
go get cloud.google.com/go/kms/apiv1
go get google.golang.org/api/option
go mod download google.golang.org/genproto

# Networking and metrics
Write-Host "Installing networking and metrics dependencies..."
go get github.com/gorilla/websocket
go get github.com/rcrowley/go-metrics
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promhttp
go get github.com/go-kit/kit/metrics
go get github.com/go-kit/kit/metrics/discard
go get github.com/go-kit/kit/metrics/prometheus
go get github.com/minio/highwayhash
go get github.com/google/orderedcode
go get github.com/rs/cors
go get github.com/syndtr/goleveldb/leveldb
go get github.com/syndtr/goleveldb/leveldb/opt
go get github.com/syndtr/goleveldb/leveldb/util
go get github.com/libp2p/go-buffer-pool
go get github.com/creachadair/taskgroup
go get github.com/Workiva/go-datastructures/queue
go get github.com/gtank/merlin
go get github.com/go-kit/log
go get github.com/go-kit/log/level
go get github.com/go-kit/log/term
go get github.com/go-logfmt/logfmt
go get github.com/pkg/errors

# Viper and cobra
Write-Host "Installing Viper and Cobra dependencies..."
go get github.com/spf13/viper
go get github.com/spf13/cobra

# Other
Write-Host "Installing other dependencies..."
go get golang.org/x/net/context
go get golang.org/x/net/http2
go get golang.org/x/net/http2/hpack
go get golang.org/x/net/netutil
go get golang.org/x/net/trace

# Tidy up modules
Write-Host "Tidying up modules..."
go mod tidy

Write-Host "All missing dependencies installed and modules tidied." 