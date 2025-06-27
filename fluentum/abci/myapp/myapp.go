package main

import (
	"context"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
)

type Application struct {
	abci.BaseApplication
	counter int64
}

func NewApplication() *Application {
	return &Application{}
}

func (app *Application) Info(ctx context.Context, req *abci.RequestInfo) (*abci.ResponseInfo, error) {
	return &abci.ResponseInfo{Data: fmt.Sprintf("counter=%d", app.counter)}, nil
}

// FinalizeBlock processes all txs in the block and returns results for each
func (app *Application) FinalizeBlock(ctx context.Context, req *abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error) {
	txResults := make([]*abci.ExecTxResult, len(req.Txs))
	for i, tx := range req.Txs {
		result := app.processTx(tx)
		txResults[i] = result
	}
	return &abci.ResponseFinalizeBlock{TxResults: txResults}, nil
}

// processTx is a helper for custom transaction logic
func (app *Application) processTx(tx []byte) *abci.ExecTxResult {
	// Example: increment counter for each tx, always succeed
	app.counter++
	return &abci.ExecTxResult{Code: 0}
}

func (app *Application) Commit(ctx context.Context, req *abci.RequestCommit) (*abci.ResponseCommit, error) {
	return &abci.ResponseCommit{}, nil
}

func (app *Application) Query(ctx context.Context, req *abci.RequestQuery) (*abci.ResponseQuery, error) {
	return &abci.ResponseQuery{Value: []byte(fmt.Sprintf("counter=%d", app.counter))}, nil
}

func main() {
	// This is a simple ABCI application example
	// In a real implementation, this would be used with tendermint
	app := NewApplication()
	fmt.Printf("ABCI Application started with counter=%d\n", app.counter)
}
