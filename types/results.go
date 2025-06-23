package types

import (
	abci "github.com/fluentum-chain/fluentum/abci/types"
	"github.com/fluentum-chain/fluentum/crypto/merkle"
)

// ABCIResults wraps the finalize block results to return a proof.
type ABCIResults struct {
	FinalizeBlock *abci.ResponseFinalizeBlock
	// TODO: Add other fields as needed
}

// NewResults now takes a ResponseFinalizeBlock
func NewResults(response *abci.ResponseFinalizeBlock) ABCIResults {
	return ABCIResults{
		FinalizeBlock: response,
	}
}

// Hash returns a merkle hash of all results (update as needed for block-level)
func (a ABCIResults) Hash() []byte {
	if a.FinalizeBlock == nil || len(a.FinalizeBlock.TxResults) == 0 {
		return nil
	}
	bzs := make([][]byte, len(a.FinalizeBlock.TxResults))
	for i, res := range a.FinalizeBlock.TxResults {
		bz, err := res.Marshal()
		if err != nil {
			panic(err)
		}
		bzs[i] = bz
	}
	return merkle.HashFromByteSlices(bzs)
}

// ProveResult returns a merkle proof of one result from the set
func (a ABCIResults) ProveResult(i int) merkle.Proof {
	if a.FinalizeBlock == nil || len(a.FinalizeBlock.TxResults) == 0 {
		panic("no tx results")
	}
	bzs := make([][]byte, len(a.FinalizeBlock.TxResults))
	for i, res := range a.FinalizeBlock.TxResults {
		bz, err := res.Marshal()
		if err != nil {
			panic(err)
		}
		bzs[i] = bz
	}
	_, proofs := merkle.ProofsFromByteSlices(bzs)
	return *proofs[i]
}

func (a ABCIResults) toByteSlices() [][]byte {
	l := len(a)
	bzs := make([][]byte, l)
	for i := 0; i < l; i++ {
		bz, err := a[i].Marshal()
		if err != nil {
			panic(err)
		}
		bzs[i] = bz
	}
	return bzs
}

// deterministicResponseDeliverTx strips non-deterministic fields from
// ResponseDeliverTx and returns another ResponseDeliverTx.
func deterministicResponseDeliverTx(response *abci.ResponseDeliverTx) *abci.ResponseDeliverTx {
	return &abci.ResponseDeliverTx{
		Code:      response.Code,
		Data:      response.Data,
		GasWanted: response.GasWanted,
		GasUsed:   response.GasUsed,
	}
}
