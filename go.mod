module github.com/fluentum-chain/fluentum

go 1.23

toolchain go1.24.4

require (
    cloud.google.com/go/kms v1.15.7
	github.com/BurntSushi/toml v1.4.1-0.20240526193622-a339e1f7089c
	github.com/ChainSafe/go-schnorrkel v1.1.0
	github.com/Workiva/go-datastructures v1.1.5
	github.com/adlio/schema v1.3.6
	github.com/btcsuite/btcd/btcec/v2 v2.3.5
	github.com/btcsuite/btcutil v1.0.3-0.20201208143702-a53e38424cce
	github.com/bufbuild/buf v1.15.1
	github.com/cloudflare/circl v1.3.7
	github.com/cosmos/cosmos-proto v1.0.0-beta.5
	github.com/cosmos/cosmos-sdk v0.50.6
	github.com/cosmos/go-bip39 v1.0.0
	github.com/creachadair/taskgroup v0.13.0
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.0
	github.com/fluentum-chain/dilithium v0.0.0-00010101000000-000000000000
	github.com/fortytw2/leaktest v1.3.0
	github.com/go-kit/kit v0.13.0
	github.com/go-kit/log v0.2.1
	github.com/go-logfmt/logfmt v0.6.0
	github.com/gofrs/uuid v4.4.0+incompatible
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.5.4
	github.com/golangci/golangci-lint v1.52.0
	github.com/google/orderedcode v0.0.1
	github.com/google/uuid v1.6.0
	github.com/gorilla/websocket v1.5.3
	github.com/gtank/merlin v0.1.1
	github.com/iden3/go-iden3-crypto v0.0.17
	github.com/iden3/go-merkletree v0.1.0
	github.com/informalsystems/tm-load-test v1.3.0
	github.com/lib/pq v1.10.9
	github.com/libp2p/go-buffer-pool v0.1.0
	github.com/minio/highwayhash v1.0.3
	github.com/ory/dockertest v3.3.5+incompatible
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.20.5
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475
	github.com/rs/cors v1.11.1
	github.com/sasha-s/go-deadlock v0.3.5
	github.com/snikch/goodman v0.0.0-20171125024755-10e37e294daa
	github.com/spf13/cobra v1.8.1
	github.com/spf13/viper v1.19.0
	github.com/stretchr/testify v1.10.0
	github.com/syndtr/goleveldb v1.0.1-0.20220721030215-126854af5e6d
	github.com/vektra/mockery/v2 v2.23.1
	golang.org/x/crypto v0.32.0
	golang.org/x/exp v0.0.0-20250106191152-7588d65b2ba8
	golang.org/x/net v0.34.0
	golang.org/x/sys v0.29.0
	golang.org/x/text v0.21.0
	gonum.org/v1/gonum v0.15.1
	google.golang.org/api v0.171.0
	google.golang.org/genproto v0.0.0-20240227224415-6ceb2ff114de
	google.golang.org/grpc v1.70.0
	google.golang.org/protobuf v1.36.6
)

require (
    github.com/DataDog/zstd v1.5.6 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cockroachdb/errors v1.11.3 // indirect
	github.com/cockroachdb/fifo v0.0.0-20240816210425-c5d0cb0b6fc0 // indirect
	github.com/cockroachdb/logtags v0.0.0-20241215232642-bb51bb14a506 // indirect
	github.com/cockroachdb/pebble v1.1.4 // indirect
	github.com/cockroachdb/redact v1.1.5 // indirect
	github.com/cockroachdb/tokenbucket v0.0.0-20230807174530-cc333fc44b06 // indirect
	github.com/cosmos/gogoproto v1.7.0 // indirect
	github.com/dgraph-io/badger/v4 v4.5.1 // indirect
	github.com/dgraph-io/ristretto/v2 v2.1.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/getsentry/sentry-go v0.31.1 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/jmhodges/levigo v1.0.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/linxGnu/grocksdb v1.9.8 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/oasisprotocol/curve25519-voi v0.0.0-20230904125328-1f23a7beb09a // indirect
	github.com/petermattis/goid v0.0.0-20240813172612-4fcff4a6cae7 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.62.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/rogpeppe/go-internal v1.13.1 // indirect
	go.etcd.io/bbolt v1.4.0 // indirect
	go.opencensus.io v0.24.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241202173237-19429a94021a // indirect
)

replace github.com/gtank/merlin => github.com/gtank/merlin v0.1.1
replace github.com/decred/dcrd/dcrec/secp256k1/v4 => github.com/decred/dcrd/dcrec/secp256k1/v4 v4.1.0
replace github.com/fluentum-chain/dilithium => ./stubs/dilithium
replace github.com/btcsuite/btcd => github.com/btcsuite/btcd v0.22.1
replace github.com/golang/protobuf => github.com/golang/protobuf v1.5.4
replace github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2
replace github.com/fluentum-chain/fluentum => .
replace github.com/tendermint/tendermint => github.com/cometbft/cometbft v0.38.6
replace github.com/tendermint/tendermint-db => github.com/cometbft/cometbft-db v1.0.4