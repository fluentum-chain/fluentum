package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	cmtabci "github.com/cometbft/cometbft/abci/types"
)

var (
	ErrConnectionNotInitialized = errors.New("connection not initialized")
	ErrTimeout                  = errors.New("request timeout")
	ErrInvalidRequest           = errors.New("invalid request")
)

const (
	dialRetryIntervalSeconds = 3
	echoRetryIntervalSeconds = 1
	defaultTimeout           = 10 * time.Second
)

// Helper to validate block height
func validateBlockHeight(height int64) error {
	if height <= 0 {
		return fmt.Errorf("invalid block height: %d", height)
	}
	return nil
}

// Helper to validate transaction data
func validateTxData(tx []byte) error {
	if len(tx) == 0 {
		return fmt.Errorf("empty transaction data")
	}
	return nil
}

// Helper to create context with timeout
func contextWithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout == 0 {
		timeout = defaultTimeout
	}
	return context.WithTimeout(ctx, timeout)
}

// Helper to check if application implements Snapshotter interface
func isSnapshotter(app cmtabci.Application) bool {
	_, ok := app.(cmtabci.Snapshotter)
	return ok
}

// Helper to process transaction results
func processTxResults(txs [][]byte, app cmtabci.Application) []*cmtabci.ExecTxResult {
	txResults := make([]*cmtabci.ExecTxResult, len(txs))
	for i, tx := range txs {
		// Execute each transaction
		res := app.DeliverTx(&cmtabci.RequestDeliverTx{Tx: tx})
		txResults[i] = &cmtabci.ExecTxResult{
			Code:      res.Code,
			Data:      res.Data,
			Log:       res.Log,
			Info:      res.Info,
			GasWanted: res.GasWanted,
			GasUsed:   res.GasUsed,
			Events:    res.Events,
		}
	}
	return txResults
} 
