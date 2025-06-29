package state

import (
	"fmt"

	mempl "github.com/fluentum-chain/fluentum/mempool"
	"github.com/fluentum-chain/fluentum/types"
)

// TxPreCheck returns a function to filter transactions before processing.
// The function limits the size of a transaction to the block's maximum data size.
func TxPreCheck(state State) mempl.PreCheckFunc {
	var validatorCount int
	if state.Validators != nil {
		validatorCount = state.Validators.Size()
	}
	fmt.Printf("[DEBUG] TxPreCheck: state.ConsensusParams.Block.MaxBytes=%d, validatorCount=%d\n", state.ConsensusParams.Block.MaxBytes, validatorCount)
	maxDataBytes := types.MaxDataBytesNoEvidence(
		state.ConsensusParams.Block.MaxBytes,
		validatorCount,
	)
	return mempl.PreCheckMaxBytes(maxDataBytes)
}

// TxPostCheck returns a function to filter transactions after processing.
// The function limits the gas wanted by a transaction to the block's maximum total gas.
func TxPostCheck(state State) mempl.PostCheckFunc {
	return mempl.PostCheckMaxGas(state.ConsensusParams.Block.MaxGas)
}
