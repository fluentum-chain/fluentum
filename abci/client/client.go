package abcicli

import (
	"context"
	"fmt"
	"sync"

	"github.com/cometbft/cometbft/abci/types"
	"github.com/fluentum-chain/fluentum/libs/service"
	tmsync "github.com/fluentum-chain/fluentum/libs/sync"
	abci "github.com/fluentum-chain/fluentum/proto/tendermint/abci"
)

const (
	dialRetryIntervalSeconds = 3
	echoRetryIntervalSeconds = 1
)

// Client matches CometBFT's v0.38.17 ABCI 2.0 specification
type Client interface {
	service.Service

	// Mempool methods
	CheckTx(context.Context, *types.RequestCheckTx) (*types.ResponseCheckTx, error)
	CheckTxAsync(context.Context, *types.RequestCheckTx) *ReqRes
	Flush(context.Context) error

	// Consensus methods
	FinalizeBlock(context.Context, *types.RequestFinalizeBlock) (*types.ResponseFinalizeBlock, error)
	PrepareProposal(context.Context, *types.RequestPrepareProposal) (*types.ResponsePrepareProposal, error)
	ProcessProposal(context.Context, *types.RequestProcessProposal) (*types.ResponseProcessProposal, error)
	ExtendVote(context.Context, *types.RequestExtendVote) (*types.ResponseExtendVote, error)
	VerifyVoteExtension(context.Context, *types.RequestVerifyVoteExtension) (*types.ResponseVerifyVoteExtension, error)
	Commit(context.Context, *types.RequestCommit) (*types.ResponseCommit, error)
	InitChain(context.Context, *types.RequestInitChain) (*types.ResponseInitChain, error)

	// Query methods
	Info(context.Context, *types.RequestInfo) (*types.ResponseInfo, error)
	Query(context.Context, *types.RequestQuery) (*types.ResponseQuery, error)

	// Snapshot methods
	ListSnapshots(context.Context, *types.RequestListSnapshots) (*types.ResponseListSnapshots, error)
	OfferSnapshot(context.Context, *types.RequestOfferSnapshot) (*types.ResponseOfferSnapshot, error)
	LoadSnapshotChunk(context.Context, *types.RequestLoadSnapshotChunk) (*types.ResponseLoadSnapshotChunk, error)
	ApplySnapshotChunk(context.Context, *types.RequestApplySnapshotChunk) (*types.ResponseApplySnapshotChunk, error)

	// Common
	Error() error
	SetResponseCallback(cb Callback)
	SetLogger(logger Logger)
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

type ReqRes struct {
	*abci.Request
	*sync.WaitGroup
	*abci.Response // Not set atomically, so be sure to use WaitGroup.

	mtx      tmsync.Mutex
	callback func(*abci.Response) // A single callback that may be set.
}

func NewReqRes(req *abci.Request) *ReqRes {
	return &ReqRes{
		Request:   req,
		WaitGroup: waitGroup1(),
		Response:  nil,
	}
}

func waitGroup1() (wg *sync.WaitGroup) {
	wg = &sync.WaitGroup{}
	wg.Add(1)
	return
}

func (reqRes *ReqRes) SetCallback(cb func(*abci.Response)) {
	reqRes.mtx.Lock()
	defer reqRes.mtx.Unlock()
	reqRes.callback = cb
}

func (reqRes *ReqRes) InvokeCallback() {
	reqRes.mtx.Lock()
	defer reqRes.mtx.Unlock()
	if reqRes.callback != nil {
		reqRes.callback(reqRes.Response)
	}
}
