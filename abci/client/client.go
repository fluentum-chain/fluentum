package abcicli

import (
	"fmt"
	"sync"

	"github.com/fluentum-chain/fluentum/libs/service"
	tmsync "github.com/fluentum-chain/fluentum/libs/sync"
	abci "github.com/fluentum-chain/fluentum/proto/tendermint/abci"
)

const (
	dialRetryIntervalSeconds = 3
	echoRetryIntervalSeconds = 1
)

// Client defines the interface for an ABCI client.
// All `Async` methods return a `ReqRes`, and all `Sync` methods return an error.
type Client interface {
	service.Service

	Error() error

	// Info/Query Connection
	InfoSync(req abci.RequestInfo) (*abci.ResponseInfo, error)
	InfoAsync(req abci.RequestInfo) *ReqRes

	QuerySync(req abci.RequestQuery) (*abci.ResponseQuery, error)
	QueryAsync(req abci.RequestQuery) *ReqRes

	// Mempool Connection
	CheckTxSync(req abci.RequestCheckTx) (*abci.ResponseCheckTx, error)
	CheckTxAsync(req abci.RequestCheckTx) *ReqRes

	// Consensus Connection
	InitChainSync(req abci.RequestInitChain) (*abci.ResponseInitChain, error)
	InitChainAsync(req abci.RequestInitChain) *ReqRes

	CommitSync() (*abci.ResponseCommit, error)
	CommitAsync() *ReqRes

	// State Sync Connection
	ListSnapshotsSync(req abci.RequestListSnapshots) (*abci.ResponseListSnapshots, error)
	ListSnapshotsAsync(req abci.RequestListSnapshots) *ReqRes

	OfferSnapshotSync(req abci.RequestOfferSnapshot) (*abci.ResponseOfferSnapshot, error)
	OfferSnapshotAsync(req abci.RequestOfferSnapshot) *ReqRes

	LoadSnapshotChunkSync(req abci.RequestLoadSnapshotChunk) (*abci.ResponseLoadSnapshotChunk, error)
	LoadSnapshotChunkAsync(req abci.RequestLoadSnapshotChunk) *ReqRes

	ApplySnapshotChunkSync(req abci.RequestApplySnapshotChunk) (*abci.ResponseApplySnapshotChunk, error)
	ApplySnapshotChunkAsync(req abci.RequestApplySnapshotChunk) *ReqRes

	// Utility
	FlushSync() error
	FlushAsync() *ReqRes

	EchoSync(msg string) (*abci.ResponseEcho, error)
	EchoAsync(msg string) *ReqRes
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

	mtx tmsync.Mutex
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
