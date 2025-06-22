module github.com/fluentum-chain/fluentum/plugins/quantum_signer

go 1.24.4

require (
	github.com/fluentum-chain/fluentum/fluentum/core/crypto v0.0.0-00010101000000-000000000000
	github.com/fluentum-chain/dilithium v0.0.0-00010101000000-000000000000
)

replace (
	github.com/fluentum-chain/fluentum/fluentum/core/crypto => ../../core/crypto
	github.com/fluentum-chain/dilithium => ../../../stubs/dilithium
) 