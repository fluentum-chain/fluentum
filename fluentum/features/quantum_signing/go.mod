module github.com/fluentum-chain/fluentum/features/quantum_signing

go 1.24.4

require (
	github.com/cloudflare/circl v1.3.3
	github.com/fluentum-chain/fluentum/core/plugin v0.0.0
	github.com/fluentum-chain/fluentum/core/crypto v0.0.0
	github.com/fluentum-chain/fluentum/version v0.0.0
)

require (
	golang.org/x/crypto v0.4.0 // indirect
	golang.org/x/sys v0.3.0 // indirect
)

// Replace with local path for development
replace github.com/cloudflare/circl => github.com/cloudflare/circl v1.3.3
replace github.com/fluentum-chain/fluentum/core/plugin => ../../core/plugin
replace github.com/fluentum-chain/fluentum/core/crypto => ../../core/crypto
replace github.com/fluentum-chain/fluentum/version => ../../version
