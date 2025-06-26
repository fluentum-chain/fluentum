package client

import (
	"context"
	"fmt"
	"sync"

	cmtabci "github.com/cometbft/cometbft/abci/types"
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
	Logger
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

func (app *localClient) SetLogger(logger Logger) {
	app.mtx.Lock()
	app.Logger = logger
	app.mtx.Unlock()
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

	return app.Application.CheckTx(req)
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
	return nil
}

// Consensus methods
func (app *localClient) FinalizeBlock(ctx context.Context, req *cmtabci.RequestFinalizeBlock) (*cmtabci.ResponseFinalizeBlock, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	// Validate request
	if err := validateBlockHeight(req.Height); err != nil {
		return nil, fmt.Errorf("FinalizeBlock validation failed: %w", err)
	}

	// Process transactions
	txResults := processTxResults(req.Txs, app.Application)

	// Call BeginBlock/EndBlock equivalents
	beginRes := app.Application.BeginBlock(&cmtabci.RequestBeginBlock{
		Hash:   req.Hash,
		Header: req.Header,
	})
	endRes := app.Application.EndBlock(&cmtabci.RequestEndBlock{
		Height: req.Height,
	})

	return &cmtabci.ResponseFinalizeBlock{
		TxResults:             txResults,
		ValidatorUpdates:      endRes.ValidatorUpdates,
		ConsensusParamUpdates: endRes.ConsensusParamUpdates,
		AppHash:               endRes.AppHash,
		Events:                append(beginRes.Events, endRes.Events...),
	}, nil
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

func (app *localClient) Commit(ctx context.Context, req *cmtabci.RequestCommit) (*cmtabci.ResponseCommit, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.Commit()
	return &cmtabci.ResponseCommit{
		Data:         res.Data,
		RetainHeight: res.RetainHeight,
	}, nil
}

func (app *localClient) InitChain(ctx context.Context, req *cmtabci.RequestInitChain) (*cmtabci.ResponseInitChain, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	return app.Application.InitChain(req)
}

// Query methods
func (app *localClient) Info(ctx context.Context, req *cmtabci.RequestInfo) (*cmtabci.ResponseInfo, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	return app.Application.Info(req)
}

func (app *localClient) Query(ctx context.Context, req *cmtabci.RequestQuery) (*cmtabci.ResponseQuery, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	return app.Application.Query(req)
}

// Snapshot methods
func (app *localClient) ListSnapshots(ctx context.Context, req *cmtabci.RequestListSnapshots) (*cmtabci.ResponseListSnapshots, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	if isSnapshotter(app.Application) {
		snapshotter := app.Application.(cmtabci.Snapshotter)
		return snapshotter.ListSnapshots(req)
	}
	return &cmtabci.ResponseListSnapshots{}, nil
}

func (app *localClient) OfferSnapshot(ctx context.Context, req *cmtabci.RequestOfferSnapshot) (*cmtabci.ResponseOfferSnapshot, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	if isSnapshotter(app.Application) {
		snapshotter := app.Application.(cmtabci.Snapshotter)
		return snapshotter.OfferSnapshot(req)
	}
	return &cmtabci.ResponseOfferSnapshot{
		Result: cmtabci.ResponseOfferSnapshot_REJECT,
	}, nil
}

func (app *localClient) LoadSnapshotChunk(ctx context.Context, req *cmtabci.RequestLoadSnapshotChunk) (*cmtabci.ResponseLoadSnapshotChunk, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	if isSnapshotter(app.Application) {
		snapshotter := app.Application.(cmtabci.Snapshotter)
		return snapshotter.LoadSnapshotChunk(req)
	}
	return nil, fmt.Errorf("application does not implement Snapshotter")
}

func (app *localClient) ApplySnapshotChunk(ctx context.Context, req *cmtabci.RequestApplySnapshotChunk) (*cmtabci.ResponseApplySnapshotChunk, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	if isSnapshotter(app.Application) {
		snapshotter := app.Application.(cmtabci.Snapshotter)
		return snapshotter.ApplySnapshotChunk(req)
	}
	return &cmtabci.ResponseApplySnapshotChunk{
		Result: cmtabci.ResponseApplySnapshotChunk_REJECT,
	}, nil
}
