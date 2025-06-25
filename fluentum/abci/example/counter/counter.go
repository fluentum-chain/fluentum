package counter

import (
	"context"
	"encoding/binary"
	"fmt"

	"github.com/cometbft/cometbft/abci/types"
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

func (app *Application) Info(ctx context.Context, req types.RequestInfo) (types.ResponseInfo, error) {
	return types.ResponseInfo{Data: fmt.Sprintf("{\"hashes\":%v,\"txs\":%v}", app.hashCount, app.txCount)}, nil
}

func (app *Application) SetOption(ctx context.Context, req types.RequestSetOption) (types.ResponseSetOption, error) {
	key, value := req.Key, req.Value
	if key == "serial" && value == "on" {
		app.serial = true
	} else {
		return types.ResponseSetOption{}, nil
	}
	return types.ResponseSetOption{}, nil
}

func (app *Application) CheckTx(ctx context.Context, req types.RequestCheckTx) (types.ResponseCheckTx, error) {
	if app.serial {
		if len(req.Tx) > 8 {
			return types.ResponseCheckTx{
				Code: CodeTypeEncodingError,
				Log:  fmt.Sprintf("Max tx size is 8 bytes, got %d", len(req.Tx))}, nil
		}
		tx8 := make([]byte, 8)
		copy(tx8[len(tx8)-len(req.Tx):], req.Tx)
		txValue := binary.BigEndian.Uint64(tx8)
		if txValue < uint64(app.txCount) {
			return types.ResponseCheckTx{
				Code: CodeTypeBadNonce,
				Log:  fmt.Sprintf("Invalid nonce. Expected >= %v, got %v", app.txCount, txValue)}, nil
		}
	}
	return types.ResponseCheckTx{Code: CodeTypeOK}, nil
}

func (app *Application) FinalizeBlock(ctx context.Context, req types.RequestFinalizeBlock) (types.ResponseFinalizeBlock, error) {
	results := make([]*types.ExecTxResult, len(req.Txs))
	for i, tx := range req.Txs {
		results[i] = &types.ExecTxResult{Code: CodeTypeOK}
		app.txCount++
	}
	return types.ResponseFinalizeBlock{
		TxResults: results,
	}, nil
}

func (app *Application) Commit(ctx context.Context) (types.ResponseCommit, error) {
	app.hashCount++
	if app.txCount == 0 {
		return types.ResponseCommit{}, nil
	}
	hash := make([]byte, 8)
	binary.BigEndian.PutUint64(hash, uint64(app.txCount))
	return types.ResponseCommit{Data: hash}, nil
}

func (app *Application) Query(ctx context.Context, reqQuery types.RequestQuery) (types.ResponseQuery, error) {
	switch reqQuery.Path {
	case "hash":
		return types.ResponseQuery{Value: []byte(fmt.Sprintf("%v", app.hashCount))}, nil
	case "tx":
		return types.ResponseQuery{Value: []byte(fmt.Sprintf("%v", app.txCount))}, nil
	default:
		return types.ResponseQuery{Log: fmt.Sprintf("Invalid query path. Expected hash or tx, got %v", reqQuery.Path)}, nil
	}
}
