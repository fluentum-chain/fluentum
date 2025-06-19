module github.com/fluentum-chain/fluentum

go 1.23.0

toolchain go1.24.4

require (
	github.com/cloudflare/circl v1.6.1
	github.com/cosmos/cosmos-sdk v0.50.0
	github.com/iden3/go-iden3-crypto v0.0.17
	github.com/iden3/go-merkletree v0.1.0
	github.com/spf13/cobra v1.8.0
	github.com/stretchr/testify v1.10.0
	github.com/tendermint/tendermint v0.35.9
	golang.org/x/crypto v0.39.0
)

require (
	github.com/btcsuite/btcd/btcec/v2 v2.3.5
	github.com/btcsuite/btcd/btcutil v1.1.5
	github.com/btcsuite/btcd/chaincfg/chainhash v1.1.0
	github.com/cosmos/go-bip39 v1.0.0
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.5.3
	github.com/google/uuid v1.4.0
	github.com/gorilla/websocket v1.5.1
	github.com/gtank/merlin v0.1.1
	github.com/libp2p/go-buffer-pool v0.1.0
	github.com/minio/highwayhash v1.0.2
	github.com/oasisprotocol/curve25519-voi v0.0.0-20230904125328-1f23a7beb09a
	github.com/petermattis/goid v0.0.0-20230904192822-1876fd5063bc
	github.com/pkg/errors v0.9.1
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475
	github.com/rs/cors v1.10.1
	github.com/sasha-s/go-deadlock v0.3.1
	github.com/spf13/viper v1.17.0
	golang.org/x/net v0.21.0
	golang.org/x/sys v0.33.0
	google.golang.org/grpc v1.59.0
	google.golang.org/protobuf v1.36.5
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/Workiva/go-datastructures v1.1.5 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/creachadair/taskgroup v0.13.2 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.0 // indirect
	github.com/go-kit/kit v0.13.0 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/google/orderedcode v0.0.1 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_golang v1.18.0 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.45.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
)

replace github.com/gtank/merlin => github.com/gtank/merlin v0.1.1

replace github.com/decred/dcrd/dcrec/secp256k1/v4 => github.com/decred/dcrd/dcrec/secp256k1/v4 v4.1.0

replace github.com/fluentum-chain/dilithium => ./stubs/dilithium
