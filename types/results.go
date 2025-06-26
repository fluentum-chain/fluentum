package types

import (
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/fluentum-chain/fluentum/crypto/merkle"
)

// ABCIResults wraps the exec tx results to return a proof.
type ABCIResults []*abci.ExecTxResult

// NewResults creates a new ABCIResults from exec tx responses
func NewResults(responses []*abci.ExecTxResult) ABCIResults {
	return ABCIResults(responses)
}

// Hash returns a merkle hash of all results
func (a ABCIResults) Hash() []byte {
	if len(a) == 0 {
		return nil
	}
	bzs := make([][]byte, len(a))
	for i, res := range a {
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
	if len(a) == 0 {
		panic("no tx results")
	}
	bzs := make([][]byte, len(a))
	for i, res := range a {
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

// deterministicExecTxResult strips non-deterministic fields from
// ExecTxResult and returns another ExecTxResult.
func deterministicExecTxResult(response *abci.ExecTxResult) *abci.ExecTxResult {
	return &abci.ExecTxResult{
		Code:      response.Code,
		Data:      response.Data,
		GasWanted: response.GasWanted,
		GasUsed:   response.GasUsed,
	}
}
