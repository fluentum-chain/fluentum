package abcicli

import (
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/fluentum-chain/fluentum/libs/service"
	tmsync "github.com/fluentum-chain/fluentum/libs/sync"
)

var _ Client = (*localClient)(nil)

// NOTE: use defer to unlock mutex because Application might panic (e.g., in
// case of malicious tx or query). It only makes sense for publicly exposed
// methods like CheckTx (/broadcast_tx_* RPC endpoint) or Query (/abci_query
// RPC endpoint), but defers are used everywhere for the sake of consistency.
type localClient struct {
	service.BaseService

	mtx *tmsync.Mutex
	abci.Application
	Callback
}

var _ Client = (*localClient)(nil)

// NewLocalClient creates a local client, which will be directly calling the
// methods of the given app.
//
// Both Async and Sync methods ignore the given context.Context parameter.
func NewLocalClient(mtx *tmsync.Mutex, app abci.Application) Client {
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

	res := app.Application.Info(req)
	return app.callback(
		&abci.Request{Value: &abci.Request_Info{Info: &req}},
		&abci.Response{Value: &abci.Response_Info{Info: &res}},
	)
}

func (app *localClient) SetOptionAsync(req abci.RequestSetOption) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.SetOption(req)
	return app.callback(
		&abci.Request{Value: &abci.Request_SetOption{SetOption: &req}},
		&abci.Response{Value: &abci.Response_SetOption{SetOption: &res}},
	)
}

func (app *localClient) DeliverTxAsync(params abci.RequestFinalizeBlock) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.DeliverTx(params)
	return app.callback(
		&abci.Request{Value: &abci.Request_DeliverTx{DeliverTx: &params}},
		&abci.Response{Value: &abci.Response_DeliverTx{DeliverTx: &res}},
	)
}

func (app *localClient) CheckTxAsync(req abci.RequestCheckTx) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.CheckTx(req)
	return app.callback(
		&abci.Request{Value: &abci.Request_CheckTx{CheckTx: &req}},
		&abci.Response{Value: &abci.Response_CheckTx{CheckTx: &res}},
	)
}

func (app *localClient) QueryAsync(req abci.RequestQuery) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.Query(req)
	return app.callback(
		&abci.Request{Value: &abci.Request_Query{Query: &req}},
		&abci.Response{Value: &abci.Response_Query{Query: &res}},
	)
}

func (app *localClient) CommitAsync() *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.Commit()
	return app.callback(
		&abci.Request{Value: &abci.Request_Commit{Commit: &abci.RequestCommit{}}},
		&abci.Response{Value: &abci.Response_Commit{Commit: &res}},
	)
}

func (app *localClient) InitChainAsync(req abci.RequestInitChain) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.InitChain(req)
	return app.callback(
		&abci.Request{Value: &abci.Request_InitChain{InitChain: &req}},
		&abci.Response{Value: &abci.Response_InitChain{InitChain: &res}},
	)
}

func (app *localClient) BeginBlockAsync(req abci.RequestFinalizeBlock) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.BeginBlock(req)
	return app.callback(
		&abci.Request{Value: &abci.Request_BeginBlock{BeginBlock: &req}},
		&abci.Response{Value: &abci.Response_BeginBlock{BeginBlock: &res}},
	)
}

func (app *localClient) EndBlockAsync(req abci.RequestFinalizeBlock) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.EndBlock(req)
	return app.callback(
		&abci.Request{Value: &abci.Request_EndBlock{EndBlock: &req}},
		&abci.Response{Value: &abci.Response_EndBlock{EndBlock: &res}},
	)
}

func (app *localClient) ListSnapshotsAsync(req abci.RequestListSnapshots) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.ListSnapshots(req)
	return app.callback(
		&abci.Request{Value: &abci.Request_ListSnapshots{ListSnapshots: &req}},
		&abci.Response{Value: &abci.Response_ListSnapshots{ListSnapshots: &res}},
	)
}

func (app *localClient) OfferSnapshotAsync(req abci.RequestOfferSnapshot) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.OfferSnapshot(req)
	return app.callback(
		&abci.Request{Value: &abci.Request_OfferSnapshot{OfferSnapshot: &req}},
		&abci.Response{Value: &abci.Response_OfferSnapshot{OfferSnapshot: &res}},
	)
}

func (app *localClient) LoadSnapshotChunkAsync(req abci.RequestLoadSnapshotChunk) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.LoadSnapshotChunk(req)
	return app.callback(
		&abci.Request{Value: &abci.Request_LoadSnapshotChunk{LoadSnapshotChunk: &req}},
		&abci.Response{Value: &abci.Response_LoadSnapshotChunk{LoadSnapshotChunk: &res}},
	)
}

func (app *localClient) ApplySnapshotChunkAsync(req abci.RequestApplySnapshotChunk) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.ApplySnapshotChunk(req)
	return app.callback(
		&abci.Request{Value: &abci.Request_ApplySnapshotChunk{ApplySnapshotChunk: &req}},
		&abci.Response{Value: &abci.Response_ApplySnapshotChunk{ApplySnapshotChunk: &res}},
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

	res := app.Application.Info(req)
	return &res, nil
}

func (app *localClient) SetOptionSync(req abci.RequestSetOption) (*abci.ResponseSetOption, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.SetOption(req)
	return &res, nil
}

func (app *localClient) DeliverTxSync(req abci.RequestFinalizeBlock) (*abci.ResponseDeliverTx, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.DeliverTx(req)
	return &res, nil
}

func (app *localClient) CheckTxSync(req abci.RequestCheckTx) (*abci.ResponseCheckTx, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.CheckTx(req)
	return &res, nil
}

func (app *localClient) QuerySync(req abci.RequestQuery) (*abci.ResponseQuery, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.Query(req)
	return &res, nil
}

func (app *localClient) CommitSync() (*abci.ResponseCommit, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.Commit()
	return &res, nil
}

func (app *localClient) InitChainSync(req abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.InitChain(req)
	return &res, nil
}

func (app *localClient) BeginBlockSync(req abci.RequestFinalizeBlock) (*abci.ResponseBeginBlock, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.BeginBlock(req)
	return &res, nil
}

func (app *localClient) EndBlockSync(req abci.RequestFinalizeBlock) (*abci.ResponseEndBlock, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.EndBlock(req)
	return &res, nil
}

func (app *localClient) ListSnapshotsSync(req abci.RequestListSnapshots) (*abci.ResponseListSnapshots, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.ListSnapshots(req)
	return &res, nil
}

func (app *localClient) OfferSnapshotSync(req abci.RequestOfferSnapshot) (*abci.ResponseOfferSnapshot, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.OfferSnapshot(req)
	return &res, nil
}

func (app *localClient) LoadSnapshotChunkSync(
	req abci.RequestLoadSnapshotChunk) (*abci.ResponseLoadSnapshotChunk, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.LoadSnapshotChunk(req)
	return &res, nil
}

func (app *localClient) ApplySnapshotChunkSync(
	req abci.RequestApplySnapshotChunk) (*abci.ResponseApplySnapshotChunk, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.ApplySnapshotChunk(req)
	return &res, nil
}

//-------------------------------------------------------

func (app *localClient) callback(req *abci.Request, res *abci.Response) *ReqRes {
	reqRes := NewReqRes(req)
	reqRes.Response = res
	reqRes.SetDone()
	app.Callback(req, res)
	return reqRes
}

func newLocalReqRes(req *abci.Request, res *abci.Response) *ReqRes {
	reqRes := NewReqRes(req)
	reqRes.Response = res
	reqRes.SetDone()
	return reqRes
}
