package client

import (
	"context"
	"fmt"

	cmtabci "github.com/cometbft/cometbft/abci/types"
)

// Client matches CometBFT's ABCI 2.0 specification
type Client interface {
	// Mempool methods
	CheckTx(context.Context, *cmtabci.RequestCheckTx) (*cmtabci.ResponseCheckTx, error)
	CheckTxAsync(context.Context, *cmtabci.RequestCheckTx) *ReqRes
	Flush(context.Context) error

	// Consensus methods
	FinalizeBlock(context.Context, *cmtabci.RequestFinalizeBlock) (*cmtabci.ResponseFinalizeBlock, error)
	PrepareProposal(context.Context, *cmtabci.RequestPrepareProposal) (*cmtabci.ResponsePrepareProposal, error)
	ProcessProposal(context.Context, *cmtabci.RequestProcessProposal) (*cmtabci.ResponseProcessProposal, error)
	ExtendVote(context.Context, *cmtabci.RequestExtendVote) (*cmtabci.ResponseExtendVote, error)
	VerifyVoteExtension(context.Context, *cmtabci.RequestVerifyVoteExtension) (*cmtabci.ResponseVerifyVoteExtension, error)
	Commit(context.Context, *cmtabci.RequestCommit) (*cmtabci.ResponseCommit, error)
	InitChain(context.Context, *cmtabci.RequestInitChain) (*cmtabci.ResponseInitChain, error)

	// Query methods
	Info(context.Context, *cmtabci.RequestInfo) (*cmtabci.ResponseInfo, error)
	Query(context.Context, *cmtabci.RequestQuery) (*cmtabci.ResponseQuery, error)

	// Snapshot methods
	ListSnapshots(context.Context, *cmtabci.RequestListSnapshots) (*cmtabci.ResponseListSnapshots, error)
	OfferSnapshot(context.Context, *cmtabci.RequestOfferSnapshot) (*cmtabci.ResponseOfferSnapshot, error)
	LoadSnapshotChunk(context.Context, *cmtabci.RequestLoadSnapshotChunk) (*cmtabci.ResponseLoadSnapshotChunk, error)
	ApplySnapshotChunk(context.Context, *cmtabci.RequestApplySnapshotChunk) (*cmtabci.ResponseApplySnapshotChunk, error)

	// Common
	Error() error
	SetResponseCallback(Callback)
	SetLogger(Logger)
}

//----------------------------------------

// NewClient returns a new ABCI client of the specified transport type.
// It returns an error if the transport is not "socket" or "grpc"
func NewClient(addr, transport string, mustConnect bool) (client Client, err error) {
	switch transport {
	case "socket":
		client = NewSocketClient(addr, mustConnect)
	case "grpc":
		client = NewGRPCClient(addr, mustConnect)
	default:
		err = fmt.Errorf("unknown abci transport %s", transport)
	}
	return
}
