package compat

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cometbft/cometbft/abci/types"
)

func AdaptResponse(res types.ResponseFinalizeBlock) storetypes.ResponseFinalizeBlock {
	return storetypes.ResponseFinalizeBlock{
		Events: res.Events,
		// TODO: Map other fields as needed for full compatibility
	}
}
