package main

import (
	"encoding/binary"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/fluentum-chain/fluentum/abci/example/code"
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

func (app *Application) Info(req abci.RequestInfo) abci.ResponseInfo {
	return abci.ResponseInfo{Data: fmt.Sprintf("{\"hashes\":%v,\"txs\":%v}", app.hashCount, app.txCount)}
}

func (app *Application) SetOption(req abci.RequestSetOption) abci.ResponseSetOption {
	key, value := req.Key, req.Value
	if key == "serial" && value == "on" {
		app.serial = true
	} else {
		/*
			TODO Panic and have the ABCI server pass an exception.
			The client can call SetOptionSync() and get an `error`.
			return abci.ResponseSetOption{
				Error: fmt.Sprintf("Unknown key (%s) or value (%s)", key, value),
			}
		*/
		return abci.ResponseSetOption{}
	}

	return abci.ResponseSetOption{}
}

func (app *Application) DeliverTx(req abci.RequestDeliverTx) abci.ResponseDeliverTx {
	if app.serial {
		if len(req.Tx) > 8 {
			return abci.ResponseDeliverTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Max tx size is 8 bytes, got %d", len(req.Tx))}
		}
		tx8 := make([]byte, 8)
		copy(tx8[len(tx8)-len(req.Tx):], req.Tx)
		txValue := binary.BigEndian.Uint64(tx8)
		if txValue != uint64(app.txCount) {
			return abci.ResponseDeliverTx{
				Code: code.CodeTypeBadNonce,
				Log:  fmt.Sprintf("Invalid nonce. Expected %v, got %v", app.txCount, txValue)}
		}
	}
	app.txCount++
	return abci.ResponseDeliverTx{Code: code.CodeTypeOK}
}

func (app *Application) CheckTx(req abci.RequestCheckTx) abci.ResponseCheckTx {
	if app.serial {
		if len(req.Tx) > 8 {
			return abci.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Max tx size is 8 bytes, got %d", len(req.Tx))}
		}
		tx8 := make([]byte, 8)
		copy(tx8[len(tx8)-len(req.Tx):], req.Tx)
		txValue := binary.BigEndian.Uint64(tx8)
		if txValue < uint64(app.txCount) {
			return abci.ResponseCheckTx{
				Code: code.CodeTypeBadNonce,
				Log:  fmt.Sprintf("Invalid nonce. Expected >= %v, got %v", app.txCount, txValue)}
		}
	}
	return abci.ResponseCheckTx{Code: code.CodeTypeOK}
}

func (app *Application) Commit() (resp abci.ResponseCommit) {
	app.hashCount++
	if app.txCount == 0 {
		return abci.ResponseCommit{}
	}
	hash := make([]byte, 8)
	binary.BigEndian.PutUint64(hash, uint64(app.txCount))
	return abci.ResponseCommit{Data: hash}
}

func (app *Application) Query(reqQuery abci.RequestQuery) abci.ResponseQuery {
	switch reqQuery.Path {
	case "hash":
		return abci.ResponseQuery{Value: []byte(fmt.Sprintf("%v", app.hashCount))}
	case "tx":
		return abci.ResponseQuery{Value: []byte(fmt.Sprintf("%v", app.txCount))}
	default:
		return abci.ResponseQuery{Log: fmt.Sprintf("Invalid query path. Expected hash or tx, got %v", reqQuery.Path)}
	}
}
