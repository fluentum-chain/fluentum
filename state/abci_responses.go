package state

import (
	abci "github.com/fluentum-chain/fluentum/abci/types"
)

// ABCIResponses holds responses from the ABCI application for a block.
type ABCIResponses struct {
	FinalizeBlock *abci.FinalizeBlockResponse
}
