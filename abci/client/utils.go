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

// Helper to check if application supports snapshots
func isSnapshotter(app cmtabci.Application) bool {
	// For now, we'll assume all applications support snapshots
	// In a real implementation, you'd check for specific interface methods
	return true
}

// Helper to process transaction results
func processTxResults(txs [][]byte, app cmtabci.Application) []*cmtabci.ExecTxResult {
	txResults := make([]*cmtabci.ExecTxResult, len(txs))

	// Create a FinalizeBlock request with all transactions
	req := &cmtabci.RequestFinalizeBlock{
		Txs: txs,
	}

	// Execute all transactions in a single FinalizeBlock call
	res, err := app.FinalizeBlock(context.Background(), req)
	if err != nil {
		// If FinalizeBlock fails, create error results for all transactions
		for i := range txs {
			txResults[i] = &cmtabci.ExecTxResult{
				Code: 1, // Use a simple error code
				Log:  fmt.Sprintf("FinalizeBlock failed: %v", err),
			}
		}
		return txResults
	}

	// Copy results from FinalizeBlock response
	for i, txRes := range res.TxResults {
		if i < len(txResults) {
			txResults[i] = txRes
		}
	}

	return txResults
}
