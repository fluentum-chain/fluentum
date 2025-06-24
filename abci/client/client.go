package abcicli

import (
	"fmt"
	"sync"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/fluentum-chain/fluentum/libs/service"
	tmsync "github.com/fluentum-chain/fluentum/libs/sync"
)

const (
	dialRetryIntervalSeconds = 3
	echoRetryIntervalSeconds = 1
)

// Client defines an interface for an ABCI client.
// All `Async` methods return a `ReqRes` object.
// All `Sync` methods return the appropriate protobuf ResponseXxx struct and an error.
// Note these are client errors, eg. ABCI socket connectivity issues.
// Application-related errors are reflected in response via ABCI error codes and logs.
type Client interface {
	service.Service

	SetResponseCallback(Callback)
	Error() error

	FlushAsync() *ReqRes
	EchoAsync(msg string) *ReqRes
	InfoAsync(abci.RequestInfo) *ReqRes
	SetOptionAsync(abci.RequestSetOption) *ReqRes
	DeliverTxAsync(abci.RequestDeliverTx) *ReqRes
	CheckTxAsync(abci.RequestCheckTx) *ReqRes
	QueryAsync(abci.RequestQuery) *ReqRes
	CommitAsync() *ReqRes
	InitChainAsync(abci.RequestInitChain) *ReqRes
	BeginBlockAsync(abci.RequestBeginBlock) *ReqRes
	EndBlockAsync(abci.RequestEndBlock) *ReqRes
	ListSnapshotsAsync(abci.RequestListSnapshots) *ReqRes
	OfferSnapshotAsync(abci.RequestOfferSnapshot) *ReqRes
	LoadSnapshotChunkAsync(abci.RequestLoadSnapshotChunk) *ReqRes
	ApplySnapshotChunkAsync(abci.RequestApplySnapshotChunk) *ReqRes

	FlushSync() error
	EchoSync(msg string) (*abci.ResponseEcho, error)
	InfoSync(abci.RequestInfo) (*abci.ResponseInfo, error)
	SetOptionSync(abci.RequestSetOption) (*abci.ResponseSetOption, error)
	DeliverTxSync(abci.RequestDeliverTx) (*abci.ResponseDeliverTx, error)
	CheckTxSync(abci.RequestCheckTx) (*abci.ResponseCheckTx, error)
	QuerySync(abci.RequestQuery) (*abci.ResponseQuery, error)
	CommitSync() (*abci.ResponseCommit, error)
	InitChainSync(abci.RequestInitChain) (*abci.ResponseInitChain, error)
	BeginBlockSync(abci.RequestBeginBlock) (*abci.ResponseBeginBlock, error)
	EndBlockSync(abci.RequestEndBlock) (*abci.ResponseEndBlock, error)
	ListSnapshotsSync(abci.RequestListSnapshots) (*abci.ResponseListSnapshots, error)
	OfferSnapshotSync(abci.RequestOfferSnapshot) (*abci.ResponseOfferSnapshot, error)
	LoadSnapshotChunkSync(abci.RequestLoadSnapshotChunk) (*abci.ResponseLoadSnapshotChunk, error)
	ApplySnapshotChunkSync(abci.RequestApplySnapshotChunk) (*abci.ResponseApplySnapshotChunk, error)
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

type Callback func(*abci.Request, *abci.Response)

type ReqRes struct {
	*abci.Request
	*sync.WaitGroup
	*abci.Response // Not set atomically, so be sure to use WaitGroup.

	mtx tmsync.Mutex

	// callbackInvoked as a variable to track if the callback was already
	// invoked during the regular execution of the request. This variable
	// allows clients to set the callback simultaneously without potentially
	// invoking the callback twice by accident, once when 'SetCallback' is
	// called and once during the normal request.
	callbackInvoked bool
	cb              func(*abci.Response) // A single callback that may be set.
}

func NewReqRes(req *abci.Request) *ReqRes {
	return &ReqRes{
		Request:   req,
		WaitGroup: waitGroup1(),
		Response:  nil,

		callbackInvoked: false,
		cb:              nil,
	}
}

// Sets sets the callback. If reqRes is already done, it will call the cb
// immediately. Note, reqRes.cb should not change if reqRes.done and only one
// callback is supported.
func (r *ReqRes) SetCallback(cb func(res *abci.Response)) {
	r.mtx.Lock()

	if r.callbackInvoked {
		r.mtx.Unlock()
		cb(r.Response)
		return
	}

	r.cb = cb
	r.mtx.Unlock()
}

// InvokeCallback invokes a thread-safe execution of the configured callback
// if non-nil.
func (r *ReqRes) InvokeCallback() {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if r.cb != nil {
		r.cb(r.Response)
	}
	r.callbackInvoked = true
}

// GetCallback returns the configured callback of the ReqRes object which may be
// nil. Note, it is not safe to concurrently call this in cases where it is
// marked done and SetCallback is called before calling GetCallback as that
// will invoke the callback twice and create a potential race condition.
//
// ref: https://github.com/fluentum-chain/fluentum/issues/5439
func (r *ReqRes) GetCallback() func(*abci.Response) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	return r.cb
}

func waitGroup1() (wg *sync.WaitGroup) {
	wg = &sync.WaitGroup{}
	wg.Add(1)
	return
}
