package client

import (
	"context"
	"fmt"
	"net"

	cometbftabci "github.com/cometbft/cometbft/abci/types"
	tmlog "github.com/fluentum-chain/fluentum/libs/log"
)

// Client matches CometBFT's ABCI 2.0 specification
type Client interface {
	// Echo method for testing
	Echo(context.Context, string) (*cometbftabci.ResponseEcho, error)

	// Mempool methods
	CheckTx(context.Context, *cometbftabci.RequestCheckTx) (*cometbftabci.ResponseCheckTx, error)
	CheckTxAsync(context.Context, *cometbftabci.RequestCheckTx) *ReqRes
	Flush(context.Context) error
	FlushAsync(context.Context) *ReqRes

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
	SetLogger(tmlog.Logger)
	Start() error
	Stop() error
	Quit() <-chan struct{}
	Close() error
}

//----------------------------------------

// NewClient returns a new ABCI client of the specified transport type.
// It returns an error if the transport is not "socket" or "grpc"
func NewClient(addr, transport string, mustConnect bool) (client Client, err error) {
	switch transport {
	case "socket":
		// For socket transport, we need to establish a connection first
		// This is a simplified implementation - in practice you'd want proper connection handling
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to %s: %w", addr, err)
		}
		client = NewSocketClient(conn, nil) // Using nil logger for now
	case "grpc":
		// Note: gRPC client is not implemented in this version
		err = fmt.Errorf("gRPC transport not supported in this version")
	default:
		err = fmt.Errorf("unknown abci transport %s", transport)
	}
	return
}
