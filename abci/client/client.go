package client

import (
	"context"
	"fmt"

	cometbftabci "github.com/cometbft/cometbft/abci/types"
)

// Client matches CometBFT's ABCI 2.0 specification
type Client interface {
	// Echo method for testing
	Echo(context.Context, string) (*cometbftabci.ResponseEcho, error)
	
	// Mempool methods
	CheckTx(context.Context, *cometbftabci.RequestCheckTx) (*cometbftabci.ResponseCheckTx, error)
	CheckTxAsync(context.Context, *cometbftabci.RequestCheckTx) *ReqRes
	Flush(context.Context) error

	// Consensus methods
	FinalizeBlock(context.Context, *cometbftabci.RequestFinalizeBlock) (*cometbftabci.ResponseFinalizeBlock, error)
	FinalizeBlockAsync(context.Context, *cometbftabci.RequestFinalizeBlock) *ReqRes
	PrepareProposal(context.Context, *cometbftabci.RequestPrepareProposal) (*cometbftabci.ResponsePrepareProposal, error)
	ProcessProposal(context.Context, *cometbftabci.RequestProcessProposal) (*cometbftabci.ResponseProcessProposal, error)
	ExtendVote(context.Context, *cometbftabci.RequestExtendVote) (*cometbftabci.ResponseExtendVote, error)
	VerifyVoteExtension(context.Context, *cometbftabci.RequestVerifyVoteExtension) (*cometbftabci.ResponseVerifyVoteExtension, error)
	Commit(context.Context) (*cometbftabci.ResponseCommit, error)
	CommitAsync(context.Context) *ReqRes
	InitChain(context.Context, *cometbftabci.RequestInitChain) (*cometbftabci.ResponseInitChain, error)

	// Query methods
	Info(context.Context, *cometbftabci.RequestInfo) (*cometbftabci.ResponseInfo, error)
	Query(context.Context, *cometbftabci.RequestQuery) (*cometbftabci.ResponseQuery, error)

	// Snapshot methods
	ListSnapshots(context.Context, *cometbftabci.RequestListSnapshots) (*cometbftabci.ResponseListSnapshots, error)
	OfferSnapshot(context.Context, *cometbftabci.RequestOfferSnapshot) (*cometbftabci.ResponseOfferSnapshot, error)
	LoadSnapshotChunk(context.Context, *cometbftabci.RequestLoadSnapshotChunk) (*cometbftabci.ResponseLoadSnapshotChunk, error)
	ApplySnapshotChunk(context.Context, *cometbftabci.RequestApplySnapshotChunk) (*cometbftabci.ResponseApplySnapshotChunk, error)

	// Common
	Error() error
	SetResponseCallback(Callback)
	SetLogger(Logger)
	Close() error
}

//----------------------------------------

// NewClient returns a new ABCI client of the specified transport type.
// It returns an error if the transport is not "socket" or "grpc"
func NewClient(addr, transport string, mustConnect bool) (client Client, err error) {
	switch transport {
	case "socket":
		client = NewSocketClient(addr, mustConnect)
	case "grpc":
		client, err = NewGRPCClient(addr, nil)
	default:
		err = fmt.Errorf("unknown abci transport %s", transport)
	}
	return
}
