package counter

import (
	"context"
	"encoding/binary"
	"fmt"

	abci "github.com/fluentum-chain/fluentum/abci/types"
)

// Return codes for the counter example
const (
	CodeTypeOK            uint32 = 0
	CodeTypeEncodingError uint32 = 1
	CodeTypeBadNonce      uint32 = 2
)

// ABCI request/response types (define minimal versions for this example)
type RequestInfo struct{}
type ResponseInfo struct{ Data string }
type RequestSetOption struct{ Key, Value string }
type ResponseSetOption struct{}
type RequestFinalizeBlock struct{ Tx []byte }
type ResponseDeliverTx struct {
	Code uint32
	Log  string
}
type RequestCheckTx struct{ Tx []byte }
type ResponseCheckTx struct {
	Code uint32
	Log  string
}
type ResponseCommit struct{ Data []byte }
type RequestQuery struct{ Path string }
type ResponseQuery struct {
	Value []byte
	Log   string
}

type Application struct {
	hashCount int
	txCount   int
	serial    bool
}

func NewApplication(serial bool) *Application {
	return &Application{serial: serial}
}

func (app *Application) Info(ctx context.Context, req *abci.InfoRequest) (*abci.InfoResponse, error) {
	return &abci.InfoResponse{Data: fmt.Sprintf("{\"hashes\":%v,\"txs\":%v}", app.hashCount, app.txCount)}, nil
}

func (app *Application) CheckTx(ctx context.Context, req *abci.CheckTxRequest) (*abci.CheckTxResponse, error) {
	if app.serial {
		if len(req.Tx) > 8 {
			return &abci.CheckTxResponse{
				Code: CodeTypeEncodingError,
				Log:  fmt.Sprintf("Max tx size is 8 bytes, got %d", len(req.Tx))}, nil
		}
		tx8 := make([]byte, 8)
		copy(tx8[len(tx8)-len(req.Tx):], req.Tx)
		txValue := binary.BigEndian.Uint64(tx8)
		if txValue < uint64(app.txCount) {
			return &abci.CheckTxResponse{
				Code: CodeTypeBadNonce,
				Log:  fmt.Sprintf("Invalid nonce. Expected >= %v, got %v", app.txCount, txValue)}, nil
		}
	}
	return &abci.CheckTxResponse{Code: CodeTypeOK}, nil
}

func (app *Application) FinalizeBlock(ctx context.Context, req *abci.FinalizeBlockRequest) (*abci.FinalizeBlockResponse, error) {
	results := make([]*abci.ExecTxResult, len(req.Txs))
	for i := range req.Txs {
		results[i] = &abci.ExecTxResult{Code: CodeTypeOK}
		app.txCount++
	}
	return &abci.FinalizeBlockResponse{
		TxResults: results,
	}, nil
}

func (app *Application) Commit(ctx context.Context, req *abci.CommitRequest) (*abci.CommitResponse, error) {
	app.hashCount++
	if app.txCount == 0 {
		return &abci.CommitResponse{}, nil
	}
	return &abci.CommitResponse{}, nil
}

func (app *Application) Query(ctx context.Context, reqQuery *abci.QueryRequest) (*abci.QueryResponse, error) {
	switch reqQuery.Path {
	case "hash":
		return &abci.QueryResponse{Value: []byte(fmt.Sprintf("%v", app.hashCount))}, nil
	case "tx":
		return &abci.QueryResponse{Value: []byte(fmt.Sprintf("%v", app.txCount))}, nil
	default:
		return &abci.QueryResponse{Log: fmt.Sprintf("Invalid query path. Expected hash or tx, got %v", reqQuery.Path)}, nil
	}
}

// Additional methods required by the Application interface
func (app *Application) PrepareProposal(ctx context.Context, req *abci.PrepareProposalRequest) (*abci.PrepareProposalResponse, error) {
	return &abci.PrepareProposalResponse{}, nil
}

func (app *Application) ProcessProposal(ctx context.Context, req *abci.ProcessProposalRequest) (*abci.ProcessProposalResponse, error) {
	return &abci.ProcessProposalResponse{}, nil
}

func (app *Application) ExtendVote(ctx context.Context, req *abci.ExtendVoteRequest) (*abci.ExtendVoteResponse, error) {
	return &abci.ExtendVoteResponse{}, nil
}

func (app *Application) VerifyVoteExtension(ctx context.Context, req *abci.VerifyVoteExtensionRequest) (*abci.VerifyVoteExtensionResponse, error) {
	return &abci.VerifyVoteExtensionResponse{}, nil
}

func (app *Application) InitChain(ctx context.Context, req *abci.InitChainRequest) (*abci.InitChainResponse, error) {
	return &abci.InitChainResponse{}, nil
}

func (app *Application) ListSnapshots(ctx context.Context, req *abci.ListSnapshotsRequest) (*abci.ListSnapshotsResponse, error) {
	return &abci.ListSnapshotsResponse{}, nil
}

func (app *Application) LoadSnapshotChunk(ctx context.Context, req *abci.LoadSnapshotChunkRequest) (*abci.LoadSnapshotChunkResponse, error) {
	return &abci.LoadSnapshotChunkResponse{}, nil
}

func (app *Application) OfferSnapshot(ctx context.Context, req *abci.OfferSnapshotRequest) (*abci.OfferSnapshotResponse, error) {
	return &abci.OfferSnapshotResponse{Result: abci.ResponseOfferSnapshot_REJECT}, nil
}

func (app *Application) ApplySnapshotChunk(ctx context.Context, req *abci.ApplySnapshotChunkRequest) (*abci.ApplySnapshotChunkResponse, error) {
	return &abci.ApplySnapshotChunkResponse{Result: abci.ResponseApplySnapshotChunk_ABORT}, nil
}

func (app *Application) Echo(ctx context.Context, req *abci.EchoRequest) (*abci.EchoResponse, error) {
	return &abci.EchoResponse{Message: req.Message}, nil
}
