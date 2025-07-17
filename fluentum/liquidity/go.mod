module github.com/fluentum-chain/fluentum/liquidity

go 1.24.4

require (
)

replace github.com/fluentum-chain/fluentum => ../../..
replace github.com/fluentum-chain/fluentum/core/plugin => ../core/plugin
replace github.com/fluentum-chain/fluentum/core/crypto => ../core/crypto
replace github.com/fluentum-chain/fluentum/x/fluentum => ../x/fluentum
replace github.com/fluentum-chain/fluentum/x/cex => ../x/cex
replace github.com/fluentum-chain/fluentum/x/dex => ../x/dex
replace github.com/fluentum-chain/fluentum/quantum => ../quantum
replace github.com/fluentum-chain/fluentum/zkprover => ../zkprover
replace github.com/fluentum-chain/fluentum/liquidity => .
