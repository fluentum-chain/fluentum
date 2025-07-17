module github.com/fluentum-chain/fluentum/core

go 1.24.4

require (
	github.com/fluentum-chain/fluentum v0.0.0-00010101000000-000000000000
)

replace github.com/fluentum-chain/fluentum => ../..
replace github.com/fluentum-chain/fluentum/app => ../app
replace github.com/fluentum-chain/fluentum/core => .
replace github.com/fluentum-chain/fluentum/core/plugin => ./plugin
replace github.com/fluentum-chain/fluentum/core/crypto => ./crypto
replace github.com/fluentum-chain/fluentum/x/fluentum => ../x/fluentum
replace github.com/fluentum-chain/fluentum/x/cex => ../x/cex
replace github.com/fluentum-chain/fluentum/x/dex => ../x/dex
replace github.com/fluentum-chain/fluentum/quantum => ../quantum
replace github.com/fluentum-chain/fluentum/zkprover => ../zkprover
replace github.com/fluentum-chain/fluentum/liquidity => ../liquidity
