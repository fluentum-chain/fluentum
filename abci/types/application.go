package types

import (
	"context"

	abci "github.com/fluentum-chain/fluentum/proto/tendermint/abci"
)

// Type aliases for backward compatibility
type Request = abci.Request
type Response = abci.Response
type RequestInfo = abci.RequestInfo
type ResponseInfo = abci.ResponseInfo
type RequestSetOption = abci.RequestSetOption
type ResponseSetOption = abci.ResponseSetOption
type RequestQuery = abci.RequestQuery
type ResponseQuery = abci.ResponseQuery
type RequestCheckTx = abci.RequestCheckTx
type ResponseCheckTx = abci.ResponseCheckTx
type RequestInitChain = abci.RequestInitChain
type ResponseInitChain = abci.ResponseInitChain
type RequestCommit = abci.RequestCommit
type ResponseCommit = abci.ResponseCommit
type ResponseDeliverTx = abci.ResponseDeliverTx
type RequestListSnapshots = abci.RequestListSnapshots
type ResponseListSnapshots = abci.ResponseListSnapshots
type RequestOfferSnapshot = abci.RequestOfferSnapshot
type ResponseOfferSnapshot = abci.ResponseOfferSnapshot
type RequestLoadSnapshotChunk = abci.RequestLoadSnapshotChunk
type ResponseLoadSnapshotChunk = abci.ResponseLoadSnapshotChunk
type RequestApplySnapshotChunk = abci.RequestApplySnapshotChunk
type ResponseApplySnapshotChunk = abci.ResponseApplySnapshotChunk
type RequestEcho = abci.RequestEcho
type ResponseEcho = abci.ResponseEcho
type RequestFlush = abci.RequestFlush
type ResponseFlush = abci.ResponseFlush
type ValidatorUpdate = abci.ValidatorUpdate
type EventAttribute = abci.EventAttribute
type Event = abci.Event
type ExecTxResult = abci.ResponseDeliverTx

// Application is an interface that enables any finite, deterministic state machine
// to be driven by a blockchain-based replication engine via the ABCI.
// All methods take a RequestXxx argument and return a ResponseXxx argument,
// except CheckTx/DeliverTx, which take `tx []byte`, and `Commit`, which takes nothing.
type Application interface {
	// Info/Query Connection
	Info(context.Context, *RequestInfo) (*ResponseInfo, error)                // Return application info
	SetOption(context.Context, *RequestSetOption) (*ResponseSetOption, error) // Set application option
	Query(context.Context, *RequestQuery) (*ResponseQuery, error)             // Query for state

	// Mempool Connection
	CheckTx(context.Context, *RequestCheckTx) (*ResponseCheckTx, error) // Validate a tx for the mempool

	// Consensus Connection
	InitChain(context.Context, *RequestInitChain) (*ResponseInitChain, error) // Initialize blockchain w validators/other info from TendermintCore
	Commit(context.Context, *RequestCommit) (*ResponseCommit, error)          // Commit the state and return the application Merkle root hash

	// State Sync Connection
	ListSnapshots(context.Context, *RequestListSnapshots) (*ResponseListSnapshots, error)                // List available snapshots
	OfferSnapshot(context.Context, *RequestOfferSnapshot) (*ResponseOfferSnapshot, error)                // Offer a snapshot to the application
	LoadSnapshotChunk(context.Context, *RequestLoadSnapshotChunk) (*ResponseLoadSnapshotChunk, error)    // Load a snapshot chunk
	ApplySnapshotChunk(context.Context, *RequestApplySnapshotChunk) (*ResponseApplySnapshotChunk, error) // Apply a shapshot chunk
}

//-------------------------------------------------------
// BaseApplication is a base form of Application

var _ Application = (*BaseApplication)(nil)

type BaseApplication struct {
}

func NewBaseApplication() *BaseApplication {
	return &BaseApplication{}
}

func (BaseApplication) Info(ctx context.Context, req *RequestInfo) (*ResponseInfo, error) {
	return &ResponseInfo{}, nil
}

func (BaseApplication) SetOption(ctx context.Context, req *RequestSetOption) (*ResponseSetOption, error) {
	return &ResponseSetOption{}, nil
}

func (BaseApplication) CheckTx(ctx context.Context, req *RequestCheckTx) (*ResponseCheckTx, error) {
	return &ResponseCheckTx{Code: CodeTypeOK}, nil
}

func (BaseApplication) Commit(ctx context.Context, req *RequestCommit) (*ResponseCommit, error) {
	return &ResponseCommit{}, nil
}

func (BaseApplication) Query(ctx context.Context, req *RequestQuery) (*ResponseQuery, error) {
	return &ResponseQuery{Code: CodeTypeOK}, nil
}

func (BaseApplication) InitChain(ctx context.Context, req *RequestInitChain) (*ResponseInitChain, error) {
	return &ResponseInitChain{}, nil
}

func (BaseApplication) ListSnapshots(ctx context.Context, req *RequestListSnapshots) (*ResponseListSnapshots, error) {
	return &ResponseListSnapshots{}, nil
}

func (BaseApplication) OfferSnapshot(ctx context.Context, req *RequestOfferSnapshot) (*ResponseOfferSnapshot, error) {
	return &ResponseOfferSnapshot{}, nil
}

func (BaseApplication) LoadSnapshotChunk(ctx context.Context, req *RequestLoadSnapshotChunk) (*ResponseLoadSnapshotChunk, error) {
	return &ResponseLoadSnapshotChunk{}, nil
}

func (BaseApplication) ApplySnapshotChunk(ctx context.Context, req *RequestApplySnapshotChunk) (*ResponseApplySnapshotChunk, error) {
	return &ResponseApplySnapshotChunk{}, nil
}

//-------------------------------------------------------

// GRPCApplication is a GRPC wrapper for Application
type GRPCApplication struct {
	app Application
}

func NewGRPCApplication(app Application) *GRPCApplication {
	return &GRPCApplication{app}
}

func (app *GRPCApplication) Echo(ctx context.Context, req *RequestEcho) (*ResponseEcho, error) {
	return &ResponseEcho{Message: req.Message}, nil
}

func (app *GRPCApplication) Flush(ctx context.Context, req *RequestFlush) (*ResponseFlush, error) {
	return &ResponseFlush{}, nil
}

func (app *GRPCApplication) Info(ctx context.Context, req *RequestInfo) (*ResponseInfo, error) {
	res, err := app.app.Info(ctx, req)
	return res, err
}

func (app *GRPCApplication) SetOption(ctx context.Context, req *RequestSetOption) (*ResponseSetOption, error) {
	res, err := app.app.SetOption(ctx, req)
	return res, err
}

// Legacy DeliverTx method (ABCI 1.0) is removed. Use FinalizeBlock for ABCI 2.0.

func (app *GRPCApplication) CheckTx(ctx context.Context, req *RequestCheckTx) (*ResponseCheckTx, error) {
	res, err := app.app.CheckTx(ctx, req)
	return res, err
}

func (app *GRPCApplication) Query(ctx context.Context, req *RequestQuery) (*ResponseQuery, error) {
	res, err := app.app.Query(ctx, req)
	return res, err
}

func (app *GRPCApplication) Commit(ctx context.Context, req *RequestCommit) (*ResponseCommit, error) {
	res, err := app.app.Commit(ctx, req)
	return res, err
}

func (app *GRPCApplication) InitChain(ctx context.Context, req *RequestInitChain) (*ResponseInitChain, error) {
	res, err := app.app.InitChain(ctx, req)
	return res, err
}

func (app *GRPCApplication) ListSnapshots(ctx context.Context, req *RequestListSnapshots) (*ResponseListSnapshots, error) {
	res, err := app.app.ListSnapshots(ctx, req)
	return res, err
}

func (app *GRPCApplication) OfferSnapshot(ctx context.Context, req *RequestOfferSnapshot) (*ResponseOfferSnapshot, error) {
	res, err := app.app.OfferSnapshot(ctx, req)
	return res, err
}

func (app *GRPCApplication) LoadSnapshotChunk(ctx context.Context, req *RequestLoadSnapshotChunk) (*ResponseLoadSnapshotChunk, error) {
	res, err := app.app.LoadSnapshotChunk(ctx, req)
	return res, err
}

func (app *GRPCApplication) ApplySnapshotChunk(ctx context.Context, req *RequestApplySnapshotChunk) (*ResponseApplySnapshotChunk, error) {
	res, err := app.app.ApplySnapshotChunk(ctx, req)
	return res, err
}
