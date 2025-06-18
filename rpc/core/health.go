package core

import (
	ctypes "github.com/fluentum-chain/fluentum/rpc/core/types"
	rpctypes "github.com/fluentum-chain/fluentum/rpc/jsonrpc/types"
)

// Health gets node health. Returns empty result (200 OK) on success, no
// response - in case of an error.
// More: https://docs.tendermint.com/v0.34/rpc/#/Info/health
func Health(ctx *rpctypes.Context) (*ctypes.ResultHealth, error) {
	return &ctypes.ResultHealth{}, nil
}
