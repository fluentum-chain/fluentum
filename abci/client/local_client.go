package abcicli

import (
	"context"

	extabci "github.com/fluentum-chain/fluentum/abci/types"
	"github.com/fluentum-chain/fluentum/libs/service"
	tmsync "github.com/fluentum-chain/fluentum/libs/sync"
	abci "github.com/fluentum-chain/fluentum/proto/tendermint/abci"
)

var _ Client = (*localClient)(nil)

// NOTE: use defer to unlock mutex because Application might panic (e.g., in
// case of malicious tx or query). It only makes sense for publicly exposed
// methods like CheckTx (/broadcast_tx_* RPC endpoint) or Query (/abci_query
// RPC endpoint), but defers are used everywhere for the sake of consistency.
type localClient struct {
	service.BaseService

	mtx *tmsync.Mutex
	extabci.Application
	Callback
}

var _ Client = (*localClient)(nil)

// NewLocalClient creates a local client, which will be directly calling the
// methods of the given app.
//
// Both Async and Sync methods ignore the given context.Context parameter.
func NewLocalClient(mtx *tmsync.Mutex, app extabci.Application) Client {
	if mtx == nil {
		mtx = new(tmsync.Mutex)
	}
	cli := &localClient{
		mtx:         mtx,
		Application: app,
	}
	cli.BaseService = *service.NewBaseService(nil, "localClient", cli)
	return cli
}

func (app *localClient) SetResponseCallback(cb Callback) {
	app.mtx.Lock()
	app.Callback = cb
	app.mtx.Unlock()
}

// TODO: change abci.Application to include Error()?
func (app *localClient) Error() error {
	return nil
}

func (app *localClient) FlushAsync() *ReqRes {
	// Do nothing
	return newLocalReqRes(&abci.Request{Value: &abci.Request_Flush{Flush: &abci.RequestFlush{}}}, nil)
}

func (app *localClient) EchoAsync(msg string) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	return app.callback(
		&abci.Request{Value: &abci.Request_Echo{Echo: &abci.RequestEcho{Message: msg}}},
		&abci.Response{Value: &abci.Response_Echo{Echo: &abci.ResponseEcho{Message: msg}}},
	)
}

func (app *localClient) InfoAsync(req abci.RequestInfo) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res, err := app.Application.Info(context.Background(), &req)
	if err != nil {
		// Handle error
		return app.callback(
			&abci.Request{Value: &abci.Request_Info{Info: &req}},
			&abci.Response{Value: &abci.Response_Info{Info: &abci.ResponseInfo{}}},
		)
	}
	return app.callback(
		&abci.Request{Value: &abci.Request_Info{Info: &req}},
		&abci.Response{Value: &abci.Response_Info{Info: res}},
	)
}

func (app *localClient) SetOptionAsync(req abci.RequestSetOption) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res, err := app.Application.SetOption(context.Background(), &req)
	if err != nil {
		// Handle error
		return app.callback(
			&abci.Request{Value: &abci.Request_SetOption{SetOption: &req}},
			&abci.Response{Value: &abci.Response_SetOption{SetOption: &abci.ResponseSetOption{}}},
		)
	}
	return app.callback(
		&abci.Request{Value: &abci.Request_SetOption{SetOption: &req}},
		&abci.Response{Value: &abci.Response_SetOption{SetOption: res}},
	)
}

func (app *localClient) CheckTxAsync(req abci.RequestCheckTx) *ReqRes {
	reqRes := NewReqRes(req)
	go func() {
		res, err := app.Application.CheckTx(context.Background(), &req)
		if err != nil {
			reqRes.ErrorCh <- err
		} else {
			reqRes.ResponseCh <- res
		}
	}()
	return reqRes
}

func (app *localClient) QueryAsync(req abci.RequestQuery) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res, err := app.Application.Query(context.Background(), &req)
	if err != nil {
		// Handle error
		return app.callback(
			&abci.Request{Value: &abci.Request_Query{Query: &req}},
			&abci.Response{Value: &abci.Response_Query{Query: &abci.ResponseQuery{}}},
		)
	}
	return app.callback(
		&abci.Request{Value: &abci.Request_Query{Query: &req}},
		&abci.Response{Value: &abci.Response_Query{Query: res}},
	)
}

func (app *localClient) CommitAsync() *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	// Return stub response
	req := &abci.Request{}
	res := &abci.Response{}

	return app.callback(req, res)
}

func (app *localClient) InitChainAsync(req abci.RequestInitChain) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res, err := app.Application.InitChain(context.Background(), &req)
	if err != nil {
		// Handle error
		return app.callback(
			&abci.Request{Value: &abci.Request_InitChain{InitChain: &req}},
			&abci.Response{Value: &abci.Response_InitChain{InitChain: &abci.ResponseInitChain{}}},
		)
	}
	return app.callback(
		&abci.Request{Value: &abci.Request_InitChain{InitChain: &req}},
		&abci.Response{Value: &abci.Response_InitChain{InitChain: res}},
	)
}

func (app *localClient) ListSnapshotsAsync(req abci.RequestListSnapshots) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res, err := app.Application.ListSnapshots(context.Background(), &req)
	if err != nil {
		// Handle error
		return app.callback(
			&abci.Request{Value: &abci.Request_ListSnapshots{ListSnapshots: &req}},
			&abci.Response{Value: &abci.Response_ListSnapshots{ListSnapshots: &abci.ResponseListSnapshots{}}},
		)
	}
	return app.callback(
		&abci.Request{Value: &abci.Request_ListSnapshots{ListSnapshots: &req}},
		&abci.Response{Value: &abci.Response_ListSnapshots{ListSnapshots: res}},
	)
}

func (app *localClient) OfferSnapshotAsync(req abci.RequestOfferSnapshot) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res, err := app.Application.OfferSnapshot(context.Background(), &req)
	if err != nil {
		// Handle error
		return app.callback(
			&abci.Request{Value: &abci.Request_OfferSnapshot{OfferSnapshot: &req}},
			&abci.Response{Value: &abci.Response_OfferSnapshot{OfferSnapshot: &abci.ResponseOfferSnapshot{}}},
		)
	}
	return app.callback(
		&abci.Request{Value: &abci.Request_OfferSnapshot{OfferSnapshot: &req}},
		&abci.Response{Value: &abci.Response_OfferSnapshot{OfferSnapshot: res}},
	)
}

func (app *localClient) LoadSnapshotChunkAsync(req abci.RequestLoadSnapshotChunk) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res, err := app.Application.LoadSnapshotChunk(context.Background(), &req)
	if err != nil {
		// Handle error
		return app.callback(
			&abci.Request{Value: &abci.Request_LoadSnapshotChunk{LoadSnapshotChunk: &req}},
			&abci.Response{Value: &abci.Response_LoadSnapshotChunk{LoadSnapshotChunk: &abci.ResponseLoadSnapshotChunk{}}},
		)
	}
	return app.callback(
		&abci.Request{Value: &abci.Request_LoadSnapshotChunk{LoadSnapshotChunk: &req}},
		&abci.Response{Value: &abci.Response_LoadSnapshotChunk{LoadSnapshotChunk: res}},
	)
}

func (app *localClient) ApplySnapshotChunkAsync(req abci.RequestApplySnapshotChunk) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	// Return stub response
	localRes := &abci.ResponseApplySnapshotChunk{
		Result: abci.ResponseApplySnapshotChunk_ABORT,
	}

	return app.callback(
		&abci.Request{Value: &abci.Request_ApplySnapshotChunk{ApplySnapshotChunk: &req}},
		&abci.Response{Value: &abci.Response_ApplySnapshotChunk{ApplySnapshotChunk: localRes}},
	)
}

//-------------------------------------------------------

func (app *localClient) FlushSync() error {
	return nil
}

func (app *localClient) EchoSync(msg string) (*abci.ResponseEcho, error) {
	return &abci.ResponseEcho{Message: msg}, nil
}

func (app *localClient) InfoSync(req abci.RequestInfo) (*abci.ResponseInfo, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res, err := app.Application.Info(context.Background(), &req)
	return res, err
}

func (app *localClient) SetOptionSync(req abci.RequestSetOption) (*abci.ResponseSetOption, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res, err := app.Application.SetOption(context.Background(), &req)
	return res, err
}

func (app *localClient) CheckTxSync(req abci.RequestCheckTx) (*abci.ResponseCheckTx, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	// Return stub response
	return &abci.ResponseCheckTx{
		Code: 0, // OK
	}, nil
}

func (app *localClient) QuerySync(req abci.RequestQuery) (*abci.ResponseQuery, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res, err := app.Application.Query(context.Background(), &req)
	return res, err
}

func (app *localClient) CommitSync() (*abci.ResponseCommit, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	// Return stub response
	return &abci.ResponseCommit{}, nil
}

func (app *localClient) InitChainSync(req abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res, err := app.Application.InitChain(context.Background(), &req)
	return res, err
}

func (app *localClient) ListSnapshotsSync(req abci.RequestListSnapshots) (*abci.ResponseListSnapshots, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res, err := app.Application.ListSnapshots(context.Background(), &req)
	return res, err
}

func (app *localClient) OfferSnapshotSync(req abci.RequestOfferSnapshot) (*abci.ResponseOfferSnapshot, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res, err := app.Application.OfferSnapshot(context.Background(), &req)
	return res, err
}

func (app *localClient) LoadSnapshotChunkSync(
	req abci.RequestLoadSnapshotChunk) (*abci.ResponseLoadSnapshotChunk, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res, err := app.Application.LoadSnapshotChunk(context.Background(), &req)
	return res, err
}

func (app *localClient) ApplySnapshotChunkSync(
	req abci.RequestApplySnapshotChunk) (*abci.ResponseApplySnapshotChunk, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	// Return stub response
	return &abci.ResponseApplySnapshotChunk{
		Result: abci.ResponseApplySnapshotChunk_ABORT,
	}, nil
}

//-------------------------------------------------------

func (app *localClient) callback(req *abci.Request, res *abci.Response) *ReqRes {
	reqRes := NewReqRes(req)
	reqRes.Response = res
	reqRes.WaitGroup.Done()
	app.Callback(req, res)
	return reqRes
}

func newLocalReqRes(req *abci.Request, res *abci.Response) *ReqRes {
	reqRes := NewReqRes(req)
	reqRes.Response = res
	reqRes.WaitGroup.Done()
	return reqRes
}
