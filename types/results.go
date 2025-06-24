package types

import (
	abci "github.com/cometbft/cometbft/api/client/cometbft/abci/v1"
	"github.com/fluentum-chain/fluentum/crypto/merkle"
)

// ABCIResults wraps the deliver tx results to return a proof.
type ABCIResults []*abci.ResponseDeliverTx

// NewResults creates a new ABCIResults from deliver tx responses
func NewResults(responses []*abci.ResponseDeliverTx) ABCIResults {
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
