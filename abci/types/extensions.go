package types

import (
	cometbftabci "github.com/cometbft/cometbft/abci/types"
)

type (
	RequestListSnapshots       = cometbftabci.RequestListSnapshots
	ResponseListSnapshots      = cometbftabci.ResponseListSnapshots
	RequestOfferSnapshot       = cometbftabci.RequestOfferSnapshot
	ResponseOfferSnapshot      = cometbftabci.ResponseOfferSnapshot
	RequestLoadSnapshotChunk   = cometbftabci.RequestLoadSnapshotChunk
	ResponseLoadSnapshotChunk  = cometbftabci.ResponseLoadSnapshotChunk
	RequestApplySnapshotChunk  = cometbftabci.RequestApplySnapshotChunk
	ResponseApplySnapshotChunk = cometbftabci.ResponseApplySnapshotChunk
	RequestPrepareProposal     = cometbftabci.RequestPrepareProposal
	ResponsePrepareProposal    = cometbftabci.ResponsePrepareProposal
) 
