package proxy

import (
	"context"

	abcicli "github.com/cometbft/cometbft/abci/client"
	"github.com/cometbft/cometbft/abci/types"
)

// Core ABCI connections

type AppConnMempool interface {
	CheckTx(context.Context, *types.RequestCheckTx) (*types.ResponseCheckTx, error)
	CheckTxAsync(*types.RequestCheckTx) *abcicli.ReqRes
	Flush(context.Context) error
}

type AppConnConsensus interface {
	FinalizeBlock(context.Context, *types.RequestFinalizeBlock) (*types.ResponseFinalizeBlock, error)
	PrepareProposal(context.Context, *types.RequestPrepareProposal) (*types.ResponsePrepareProposal, error)
	ProcessProposal(context.Context, *types.RequestProcessProposal) (*types.ResponseProcessProposal, error)
	ExtendVote(context.Context, *types.RequestExtendVote) (*types.ResponseExtendVote, error)
	VerifyVoteExtension(context.Context, *types.RequestVerifyVoteExtension) (*types.ResponseVerifyVoteExtension, error)
	Commit(context.Context) (*types.ResponseCommit, error)
}

type AppConnQuery interface {
	Info(context.Context, *types.RequestInfo) (*types.ResponseInfo, error)
	Query(context.Context, *types.RequestQuery) (*types.ResponseQuery, error)
	ABCIInfo(context.Context) (*types.ResponseInfo, error)
}

type AppConnSnapshot interface {
	ListSnapshots(context.Context, *types.RequestListSnapshots) (*types.ResponseListSnapshots, error)
	OfferSnapshot(context.Context, *types.RequestOfferSnapshot) (*types.ResponseOfferSnapshot, error)
	LoadSnapshotChunk(context.Context, *types.RequestLoadSnapshotChunk) (*types.ResponseLoadSnapshotChunk, error)
	ApplySnapshotChunk(context.Context, *types.RequestApplySnapshotChunk) (*types.ResponseApplySnapshotChunk, error)
}
