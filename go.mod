module github.com/fluentum-chain/fluentum

go 1.22

toolchain go1.24.4

require (
	cloud.google.com/go/kms v1.15.5
	cosmossdk.io/log v1.3.1
	github.com/BurntSushi/toml v1.4.1-0.20240526193622-a339e1f7089c
	github.com/ChainSafe/go-schnorrkel v1.1.0
	github.com/Workiva/go-datastructures v1.1.5
	github.com/adlio/schema v1.3.3
	github.com/btcsuite/btcd/btcec/v2 v2.3.5
	github.com/btcsuite/btcutil v1.0.3-0.20201208143702-a53e38424cce
	github.com/bufbuild/buf v1.15.1
	github.com/cloudflare/circl v1.3.7
	github.com/cometbft/cometbft v0.38.6
	github.com/cometbft/cometbft-db v0.8.0
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
	github.com/gorilla/websocket v1.5.1
	github.com/gtank/merlin v0.1.1
	github.com/iden3/go-iden3-crypto v0.0.17
	github.com/iden3/go-merkletree v0.1.0
	github.com/informalsystems/tm-load-test v1.3.0
	github.com/lib/pq v1.10.9
	github.com/libp2p/go-buffer-pool v0.1.0
	github.com/minio/highwayhash v1.0.2
	github.com/ory/dockertest v3.3.5+incompatible
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.18.0
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475
	github.com/rs/cors v1.10.1
	github.com/sasha-s/go-deadlock v0.3.5
	github.com/snikch/goodman v0.0.0-20171125024755-10e37e294daa
	github.com/spf13/cobra v1.8.0
	github.com/spf13/viper v1.18.2
	github.com/stretchr/testify v1.10.0
	github.com/syndtr/goleveldb v1.0.1-0.20220721030215-126854af5e6d
	github.com/vektra/mockery/v2 v2.23.1
	golang.org/x/crypto v0.17.0
	golang.org/x/exp v0.0.0-20231110203233-9a3e6036ecaa
	golang.org/x/net v0.19.0
	golang.org/x/sys v0.15.0
	golang.org/x/text v0.14.0
	gonum.org/v1/gonum v0.12.0
	google.golang.org/api v0.154.0
	google.golang.org/genproto v0.0.0-20231120223509-83a465c0220f
	google.golang.org/grpc v1.59.0
	google.golang.org/protobuf v1.36.6
)

require cosmossdk.io/core v1.0.0 // indirect

replace github.com/gtank/merlin => github.com/gtank/merlin v0.1.1

replace github.com/decred/dcrd/dcrec/secp256k1/v4 => github.com/decred/dcrd/dcrec/secp256k1/v4 v4.1.0

replace github.com/fluentum-chain/dilithium => ./stubs/dilithium

replace github.com/btcsuite/btcd => github.com/btcsuite/btcd v0.22.1

replace github.com/golang/protobuf => github.com/golang/protobuf v1.5.4

replace github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2

replace (
	cosmossdk.io/core => github.com/cosmos/cosmos-sdk/core v0.11.0
	cosmossdk.io/db => github.com/cosmos/cosmos-sdk/db v0.11.0
	github.com/btcsuite/btcd/btcec/v2 => github.com/btcsuite/btcd/btcec/v2 v2.2.1
	github.com/cometbft/cometbft => github.com/cometbft/cometbft v0.37.0
	github.com/cometbft/cometbft-db => github.com/cometbft/cometbft-db v0.8.0
	github.com/cosmos/cosmos-sdk => github.com/cosmos/cosmos-sdk v0.47.5
	golang.org/x/text => golang.org/x/text v0.14.0
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20231120223509-83a465c0220f
	google.golang.org/genproto/googleapis/rpc => google.golang.org/genproto/googleapis/rpc v0.0.0-20231120223509-83a465c0220f
)

replace cosmossdk.io/store => github.com/cosmos/cosmos-sdk/store v0.47.12

replace cosmossdk.io/api => github.com/cosmos/cosmos-sdk/api v0.7.0
