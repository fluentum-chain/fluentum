package myapp

import (
	"fmt"

	"github.com/fluentum-chain/fluentum/abci/types"
)

type Application struct {
	types.BaseApplication
	counter int64
}

func NewApplication() *Application {
	return &Application{}
}

func (app *Application) Info(req types.RequestInfo) types.ResponseInfo {
	return types.ResponseInfo{Data: fmt.Sprintf("counter=%d", app.counter)}
}

func (app *Application) DeliverTx(req types.RequestDeliverTx) types.ResponseDeliverTx {
	app.counter++
	return types.ResponseDeliverTx{Code: 0}
}

func (app *Application) CheckTx(req types.RequestCheckTx) types.ResponseCheckTx {
	return types.ResponseCheckTx{Code: 0}
}

func (app *Application) Commit() types.ResponseCommit {
	return types.ResponseCommit{}
}

func (app *Application) Query(req types.RequestQuery) types.ResponseQuery {
	return types.ResponseQuery{Data: []byte(fmt.Sprintf("counter=%d", app.counter))}
}
