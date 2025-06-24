package main

import (
	"fmt"

	abci "github.com/cometbft/cometbft/api/client/cometbft/abci/v1"
)

type Application struct {
	abci.BaseApplication
	counter int64
}

func NewApplication() *Application {
	return &Application{}
}

func (app *Application) Info(req abci.RequestInfo) abci.ResponseInfo {
	return abci.ResponseInfo{Data: fmt.Sprintf("counter=%d", app.counter)}
}

func (app *Application) DeliverTx(req abci.RequestDeliverTx) abci.ResponseDeliverTx {
	app.counter++
	return abci.ResponseDeliverTx{Code: 0}
}

func (app *Application) CheckTx(req abci.RequestCheckTx) abci.ResponseCheckTx {
	return abci.ResponseCheckTx{Code: 0}
}

func (app *Application) Commit() abci.ResponseCommit {
	return abci.ResponseCommit{}
}

func (app *Application) Query(req abci.RequestQuery) abci.ResponseQuery {
	return abci.ResponseQuery{Value: []byte(fmt.Sprintf("counter=%d", app.counter))}
}
