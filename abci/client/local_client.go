package client

import (
	"context"
	"fmt"

	cmtabci "github.com/cometbft/cometbft/abci/types"
	tmlog "github.com/fluentum-chain/fluentum/libs/log"
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
	cmtabci.Application
	Callback
}

// NewLocalClient creates a local client, which will be directly calling the
// methods of the given app.
//
// Both Async and Sync methods ignore the given context.Context parameter.
func NewLocalClient(mtx *tmsync.Mutex, app cmtabci.Application) Client {
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

func (app *localClient) SetLogger(logger tmlog.Logger) {
	app.mtx.Lock()
	defer app.mtx.Unlock()
	app.BaseService.SetLogger(logger)
}

// TODO: change abci.Application to include Error()?
func (app *localClient) Error() error {
	return nil
}

// Mempool methods
func (app *localClient) CheckTx(ctx context.Context, req *cmtabci.RequestCheckTx) (*cmtabci.ResponseCheckTx, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	if err := validateTxData(req.Tx); err != nil {
		return nil, fmt.Errorf("CheckTx validation failed: %w", err)
	}

	return app.Application.CheckTx(ctx, req)
}

func (app *localClient) CheckTxAsync(ctx context.Context, req *cmtabci.RequestCheckTx) *ReqRes {
	reqRes := NewReqRes(&cmtabci.Request{Value: &cmtabci.Request_CheckTx{CheckTx: req}})
	go func() {
		res, err := app.CheckTx(ctx, req)
		if err != nil {
			reqRes.ErrorCh <- err
		} else {
			reqRes.ResponseCh <- res
		}
		reqRes.Done()
	}()
	return reqRes
}

func (app *localClient) Flush(ctx context.Context) error {
	// No-op for local client
	return nil
}

func (app *localClient) FlushAsync(ctx context.Context) *ReqRes {
	reqRes := NewReqRes(&cmtabci.Request{Value: &cmtabci.Request_Flush{Flush: &cmtabci.RequestFlush{}}})
	go func() {
		err := app.Flush(ctx)
		if err != nil {
			reqRes.ErrorCh <- err
		} else {
			reqRes.ResponseCh <- &cmtabci.ResponseFlush{}
		}
		reqRes.Done()
	}()
	return reqRes
}

// Consensus methods
func (app *localClient) FinalizeBlock(ctx context.Context, req *cmtabci.RequestFinalizeBlock) (*cmtabci.ResponseFinalizeBlock, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	if err := validateBlockHeight(req.Height); err != nil {
		return nil, fmt.Errorf("FinalizeBlock validation failed: %w", err)
	}

	return app.Application.FinalizeBlock(ctx, req)
}

func (app *localClient) FinalizeBlockAsync(ctx context.Context, req *cmtabci.RequestFinalizeBlock) *ReqRes {
	reqRes := NewReqRes(&cmtabci.Request{Value: &cmtabci.Request_FinalizeBlock{FinalizeBlock: req}})
	go func() {
		res, err := app.FinalizeBlock(ctx, req)
		if err != nil {
			reqRes.ErrorCh <- err
		} else {
			reqRes.ResponseCh <- res
		}
		reqRes.Done()
	}()
	return reqRes
}

func (app *localClient) PrepareProposal(ctx context.Context, req *cmtabci.RequestPrepareProposal) (*cmtabci.ResponsePrepareProposal, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	// Validate request
	if req.MaxTxBytes <= 0 {
		return nil, fmt.Errorf("invalid max tx bytes: %d", req.MaxTxBytes)
	}

	// For local client, we'll just return the transactions as-is
	// In a real implementation, this would involve transaction ordering and filtering
	return &cmtabci.ResponsePrepareProposal{
		Txs: req.Txs,
	}, nil
}

func (app *localClient) ProcessProposal(ctx context.Context, req *cmtabci.RequestProcessProposal) (*cmtabci.ResponseProcessProposal, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	// Validate request
	if err := validateBlockHeight(req.Height); err != nil {
		return nil, fmt.Errorf("ProcessProposal validation failed: %w", err)
	}

	// For local client, we'll accept all proposals
	// In a real implementation, this would involve validation logic
	return &cmtabci.ResponseProcessProposal{
		Status: cmtabci.ResponseProcessProposal_ACCEPT,
	}, nil
}

func (app *localClient) ExtendVote(ctx context.Context, req *cmtabci.RequestExtendVote) (*cmtabci.ResponseExtendVote, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	// Validate request
	if err := validateBlockHeight(req.Height); err != nil {
		return nil, fmt.Errorf("ExtendVote validation failed: %w", err)
	}

	// For local client, we'll return an empty extension
	// In a real implementation, this would involve creating vote extensions
	return &cmtabci.ResponseExtendVote{
		VoteExtension: []byte{},
	}, nil
}

func (app *localClient) VerifyVoteExtension(ctx context.Context, req *cmtabci.RequestVerifyVoteExtension) (*cmtabci.ResponseVerifyVoteExtension, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	// Validate request
	if err := validateBlockHeight(req.Height); err != nil {
		return nil, fmt.Errorf("VerifyVoteExtension validation failed: %w", err)
	}

	// For local client, we'll accept all vote extensions
	// In a real implementation, this would involve verification logic
	return &cmtabci.ResponseVerifyVoteExtension{
		Status: cmtabci.ResponseVerifyVoteExtension_ACCEPT,
	}, nil
}

func (app *localClient) Commit(ctx context.Context) (*cmtabci.ResponseCommit, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	// Construct a default RequestCommit
	return app.Application.Commit(ctx, &cmtabci.RequestCommit{})
}

func (app *localClient) CommitAsync(ctx context.Context) *ReqRes {
	req := &cmtabci.RequestCommit{}
	reqRes := NewReqRes(&cmtabci.Request{Value: &cmtabci.Request_Commit{Commit: req}})
	go func() {
		res, err := app.Commit(ctx)
		if err != nil {
			reqRes.ErrorCh <- err
		} else {
			reqRes.ResponseCh <- res
		}
		reqRes.Done()
	}()
	return reqRes
}

func (app *localClient) InitChain(ctx context.Context, req *cmtabci.RequestInitChain) (*cmtabci.ResponseInitChain, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	return app.Application.InitChain(ctx, req)
}

// Query methods
func (app *localClient) Info(ctx context.Context, req *cmtabci.RequestInfo) (*cmtabci.ResponseInfo, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	return app.Application.Info(ctx, req)
}

func (app *localClient) Query(ctx context.Context, req *cmtabci.RequestQuery) (*cmtabci.ResponseQuery, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	return app.Application.Query(ctx, req)
}

// Snapshot methods
func (app *localClient) ListSnapshots(ctx context.Context, req *cmtabci.RequestListSnapshots) (*cmtabci.ResponseListSnapshots, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()
	// Snapshotter not supported, return empty response
	return &cmtabci.ResponseListSnapshots{}, nil
}

func (app *localClient) OfferSnapshot(ctx context.Context, req *cmtabci.RequestOfferSnapshot) (*cmtabci.ResponseOfferSnapshot, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()
	// Snapshotter not supported, always reject
	return &cmtabci.ResponseOfferSnapshot{}, nil
}

func (app *localClient) LoadSnapshotChunk(ctx context.Context, req *cmtabci.RequestLoadSnapshotChunk) (*cmtabci.ResponseLoadSnapshotChunk, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()
	// Snapshotter not supported
	return nil, fmt.Errorf("application does not support snapshotting")
}

func (app *localClient) ApplySnapshotChunk(ctx context.Context, req *cmtabci.RequestApplySnapshotChunk) (*cmtabci.ResponseApplySnapshotChunk, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()
	// Snapshotter not supported, always reject
	return &cmtabci.ResponseApplySnapshotChunk{}, nil
}

func (app *localClient) Close() error {
	// No resources to close for local client
	return nil
}

func (app *localClient) Echo(ctx context.Context, msg string) (*cmtabci.ResponseEcho, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()
	if echoer, ok := app.Application.(interface {
		Echo(context.Context, string) (string, error)
	}); ok {
		resp, err := echoer.Echo(ctx, msg)
		return &cmtabci.ResponseEcho{Message: resp}, err
	}
	return &cmtabci.ResponseEcho{Message: msg}, nil
}

// Start starts the client
func (app *localClient) Start() error {
	// Local client is always ready
	return nil
}

// Stop stops the client
func (app *localClient) Stop() error {
	// Local client doesn't need to be stopped
	return nil
}

// Quit returns a channel that is closed when the client is stopped
func (app *localClient) Quit() <-chan struct{} {
	// Local client never quits
	ch := make(chan struct{})
	// TODO: Implement proper lifecycle management
	return ch
}
