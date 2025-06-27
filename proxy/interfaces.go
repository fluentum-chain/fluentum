package proxy

import (
	"context"

	abcicli "github.com/fluentum-chain/fluentum/abci/client"
	abci "github.com/fluentum-chain/fluentum/abci/types"
)

// Core ABCI connections

type AppConnMempool interface {
	CheckTx(context.Context, *abci.CheckTxRequest) (*abci.CheckTxResponse, error)
	CheckTxAsync(*abci.CheckTxRequest) *abcicli.ReqRes
	Flush(context.Context) error
	SetResponseCallback(func(*abci.Request, *abci.Response))
	Error() error
	FlushAsync() *abcicli.ReqRes
}

type AppConnConsensus interface {
	FinalizeBlock(context.Context, *abci.FinalizeBlockRequest) (*abci.FinalizeBlockResponse, error)
	PrepareProposal(context.Context, *abci.PrepareProposalRequest) (*abci.PrepareProposalResponse, error)
	ProcessProposal(context.Context, *abci.ProcessProposalRequest) (*abci.ProcessProposalResponse, error)
	ExtendVote(context.Context, *abci.ExtendVoteRequest) (*abci.ExtendVoteResponse, error)
	VerifyVoteExtension(context.Context, *abci.VerifyVoteExtensionRequest) (*abci.VerifyVoteExtensionResponse, error)
	Commit(context.Context, *abci.CommitRequest) (*abci.CommitResponse, error)
	CommitSync(context.Context, *abci.CommitRequest) (*abci.CommitResponse, error)
}

type AppConnQuery interface {
	Info(context.Context, *abci.InfoRequest) (*abci.InfoResponse, error)
	Query(context.Context, *abci.QueryRequest) (*abci.QueryResponse, error)
	ABCIInfo(context.Context) (*abci.InfoResponse, error)
}

type AppConnSnapshot interface {
	ListSnapshots(context.Context, *abci.ListSnapshotsRequest) (*abci.ListSnapshotsResponse, error)
	OfferSnapshot(context.Context, *abci.OfferSnapshotRequest) (*abci.OfferSnapshotResponse, error)
	LoadSnapshotChunk(context.Context, *abci.LoadSnapshotChunkRequest) (*abci.LoadSnapshotChunkResponse, error)
	ApplySnapshotChunk(context.Context, *abci.ApplySnapshotChunkRequest) (*abci.ApplySnapshotChunkResponse, error)
}
