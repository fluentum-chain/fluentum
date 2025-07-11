package counterlib

import (
	"context"
	"encoding/binary"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/fluentum-chain/fluentum/abci/example/code"
	"github.com/fluentum-chain/fluentum/abci/types"
)

type Application struct {
	abci.BaseApplication

	hashCount int
	txCount   int
	serial    bool
}

func NewApplication(serial bool) *Application {
	return &Application{serial: serial}
}

func (app *Application) Info(ctx context.Context, req *abci.RequestInfo) (*abci.ResponseInfo, error) {
	return &abci.ResponseInfo{Data: fmt.Sprintf("{\"hashes\":%v,\"txs\":%v}", app.hashCount, app.txCount)}, nil
}

func (app *Application) CheckTx(ctx context.Context, req *abci.RequestCheckTx) (*abci.ResponseCheckTx, error) {
	if app.serial {
		if len(req.Tx) > 8 {
			return &abci.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Max tx size is 8 bytes, got %d", len(req.Tx))}, nil
		}
		tx8 := make([]byte, 8)
		copy(tx8[len(tx8)-len(req.Tx):], req.Tx)
		txValue := binary.BigEndian.Uint64(tx8)
		if txValue < uint64(app.txCount) {
			return &abci.ResponseCheckTx{
				Code: code.CodeTypeBadNonce,
				Log:  fmt.Sprintf("Invalid nonce. Expected >= %v, got %v", app.txCount, txValue)}, nil
		}
	}
	return &abci.ResponseCheckTx{Code: code.CodeTypeOK}, nil
}

func (app *Application) FinalizeBlock(ctx context.Context, req *abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error) {
	txResults := make([]*abci.ExecTxResult, len(req.Txs))

	for i, tx := range req.Txs {
		if app.serial {
			if len(tx) > 8 {
				txResults[i] = &abci.ExecTxResult{
					Code: code.CodeTypeEncodingError,
					Log:  fmt.Sprintf("Max tx size is 8 bytes, got %d", len(tx))}
				continue
			}
			tx8 := make([]byte, 8)
			copy(tx8[len(tx8)-len(tx):], tx)
			txValue := binary.BigEndian.Uint64(tx8)
			if txValue != uint64(app.txCount) {
				txResults[i] = &abci.ExecTxResult{
					Code: code.CodeTypeBadNonce,
					Log:  fmt.Sprintf("Invalid nonce. Expected %v, got %v", app.txCount, txValue)}
				continue
			}
		}
		app.txCount++
		txResults[i] = &abci.ExecTxResult{Code: code.CodeTypeOK}
	}

	return &abci.ResponseFinalizeBlock{
		TxResults: txResults,
	}, nil
}

func (app *Application) Commit(ctx context.Context, req *abci.RequestCommit) (*types.CommitResponse, error) {
	app.hashCount++
	return &types.CommitResponse{}, nil
}

func (app *Application) Query(ctx context.Context, reqQuery *abci.RequestQuery) (*abci.ResponseQuery, error) {
	switch reqQuery.Path {
	case "hash":
		return &abci.ResponseQuery{Value: []byte(fmt.Sprintf("%v", app.hashCount))}, nil
	case "tx":
		return &abci.ResponseQuery{Value: []byte(fmt.Sprintf("%v", app.txCount))}, nil
	default:
		return &abci.ResponseQuery{Log: fmt.Sprintf("Invalid query path. Expected hash or tx, got %v", reqQuery.Path)}, nil
	}
}
