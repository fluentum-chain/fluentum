package client

import (
	"context"
	"fmt"

	cometbftabciv1 "github.com/cometbft/cometbft/api/cometbft/abci/v1"
)

// Client matches CometBFT's ABCI 2.0 specification
type Client interface {
	// Echo method for testing
	Echo(context.Context, string) (*cometbftabciv1.EchoResponse, error)
	
	// Mempool methods
	CheckTx(context.Context, *cometbftabciv1.CheckTxRequest) (*cometbftabciv1.CheckTxResponse, error)
	CheckTxAsync(context.Context, *cometbftabciv1.CheckTxRequest) *ReqRes
	Flush(context.Context) error

	// Consensus methods
	FinalizeBlock(context.Context, *cometbftabciv1.FinalizeBlockRequest) (*cometbftabciv1.FinalizeBlockResponse, error)
	FinalizeBlockAsync(context.Context, *cometbftabciv1.FinalizeBlockRequest) *ReqRes
	PrepareProposal(context.Context, *cometbftabciv1.PrepareProposalRequest) (*cometbftabciv1.PrepareProposalResponse, error)
	ProcessProposal(context.Context, *cometbftabciv1.ProcessProposalRequest) (*cometbftabciv1.ProcessProposalResponse, error)
	ExtendVote(context.Context, *cometbftabciv1.ExtendVoteRequest) (*cometbftabciv1.ExtendVoteResponse, error)
	VerifyVoteExtension(context.Context, *cometbftabciv1.VerifyVoteExtensionRequest) (*cometbftabciv1.VerifyVoteExtensionResponse, error)
	Commit(context.Context) (*cometbftabciv1.CommitResponse, error)
	CommitAsync(context.Context) *ReqRes
	InitChain(context.Context, *cometbftabciv1.InitChainRequest) (*cometbftabciv1.InitChainResponse, error)

	// Query methods
	Info(context.Context, *cometbftabciv1.InfoRequest) (*cometbftabciv1.InfoResponse, error)
	Query(context.Context, *cometbftabciv1.QueryRequest) (*cometbftabciv1.QueryResponse, error)

	// Snapshot methods
	ListSnapshots(context.Context, *cometbftabciv1.ListSnapshotsRequest) (*cometbftabciv1.ListSnapshotsResponse, error)
	OfferSnapshot(context.Context, *cometbftabciv1.OfferSnapshotRequest) (*cometbftabciv1.OfferSnapshotResponse, error)
	LoadSnapshotChunk(context.Context, *cometbftabciv1.LoadSnapshotChunkRequest) (*cometbftabciv1.LoadSnapshotChunkResponse, error)
	ApplySnapshotChunk(context.Context, *cometbftabciv1.ApplySnapshotChunkRequest) (*cometbftabciv1.ApplySnapshotChunkResponse, error)

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
