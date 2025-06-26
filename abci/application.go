package abci

import (
	"context"
	"fmt"
	"sync"

	cmtabci "github.com/cometbft/cometbft/abci/types"
)

// Application implements the ABCI application interface for CometBFT v0.38.17
type Application struct {
	cmtabci.BaseApplication
	mtx sync.RWMutex
	
	// Application state
	state map[string][]byte
	height int64
	appHash []byte
}

// NewApplication creates a new ABCI application instance
func NewApplication() *Application {
	return &Application{
		state: make(map[string][]byte),
		height: 0,
		appHash: []byte("initial_hash"),
	}
}

// CheckTx validates a transaction for the mempool
func (app *Application) CheckTx(ctx context.Context, req *cmtabci.RequestCheckTx) (*cmtabci.ResponseCheckTx, error) {
	app.mtx.RLock()
	defer app.mtx.RUnlock()

	// Basic transaction validation
	if len(req.Tx) == 0 {
		return &cmtabci.ResponseCheckTx{
			Code: 1,
			Log:  "empty transaction",
		}, nil
	}

	// Check transaction format (simple key-value format for demo)
	if len(req.Tx) < 2 {
		return &cmtabci.ResponseCheckTx{
			Code: 1,
			Log:  "invalid transaction format",
		}, nil
	}

	// Estimate gas (simple heuristic)
	gasWanted := int64(len(req.Tx) * 10)

	return &cmtabci.ResponseCheckTx{
		Code:      0,
		GasWanted: gasWanted,
		GasUsed:   gasWanted,
		Log:       "transaction is valid",
	}, nil
}

// FinalizeBlock processes all transactions in a block
func (app *Application) FinalizeBlock(ctx context.Context, req *cmtabci.RequestFinalizeBlock) (*cmtabci.ResponseFinalizeBlock, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	// Update height
	app.height = req.Height

	// Process all transactions
	txResults := make([]*cmtabci.ExecTxResult, len(req.Txs))
	for i, tx := range req.Txs {
		// Validate transaction
		checkRes, err := app.CheckTx(ctx, &cmtabci.RequestCheckTx{Tx: tx})
		if err != nil {
			return nil, fmt.Errorf("failed to check transaction %d: %w", i, err)
		}

		if checkRes.Code != 0 {
			// Transaction is invalid
			txResults[i] = &cmtabci.ExecTxResult{
				Code: checkRes.Code,
				Log:  checkRes.Log,
			}
			continue
		}

		// Execute transaction
		execRes, err := app.executeTransaction(tx)
		if err != nil {
			txResults[i] = &cmtabci.ExecTxResult{
				Code: 1,
				Log:  fmt.Sprintf("execution error: %v", err),
			}
			continue
		}

		txResults[i] = &cmtabci.ExecTxResult{
			Code:      0,
			Data:      execRes.Data,
			Log:       execRes.Log,
			GasWanted: checkRes.GasWanted,
			GasUsed:   checkRes.GasUsed,
			Events:    execRes.Events,
		}
	}

	// Update app hash
	app.appHash = app.calculateAppHash()

	return &cmtabci.ResponseFinalizeBlock{
		TxResults: txResults,
		AppHash:   app.appHash,
		ConsensusParamUpdates: &cmtabci.ConsensusParams{
			Block: &cmtabci.BlockParams{
				MaxBytes: 22020096, // 21MB
				MaxGas:   15000000, // 15M gas
			},
			Evidence: &cmtabci.EvidenceParams{
				MaxAgeNumBlocks: 100000,
				MaxAgeDuration:  172800000000000, // 48 hours in nanoseconds
				MaxBytes:        1048576,         // 1MB
			},
			Validator: &cmtabci.ValidatorParams{
				PubKeyTypes: []string{"ed25519", "secp256k1"},
			},
		},
	}, nil
}

// Commit commits the current state
func (app *Application) Commit(ctx context.Context, req *cmtabci.RequestCommit) (*cmtabci.ResponseCommit, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	// Return the current app hash
	return &cmtabci.ResponseCommit{
		Data: app.appHash,
	}, nil
}

// Info returns application information
func (app *Application) Info(ctx context.Context, req *cmtabci.RequestInfo) (*cmtabci.ResponseInfo, error) {
	app.mtx.RLock()
	defer app.mtx.RUnlock()

	return &cmtabci.ResponseInfo{
		Data:             "tendermint-abci-app",
		Version:          "1.0.0",
		AppVersion:       1,
		LastBlockHeight:  app.height,
		LastBlockAppHash: app.appHash,
	}, nil
}

// Query handles queries to the application state
func (app *Application) Query(ctx context.Context, req *cmtabci.RequestQuery) (*cmtabci.ResponseQuery, error) {
	app.mtx.RLock()
	defer app.mtx.RUnlock()

	// Handle different query paths
	switch req.Path {
	case "/store":
		// Query store
		if value, exists := app.state[string(req.Data)]; exists {
			return &cmtabci.ResponseQuery{
				Code:  0,
				Value: value,
				Log:   "found",
			}, nil
		}
		return &cmtabci.ResponseQuery{
			Code: 1,
			Log:  "not found",
		}, nil

	case "/height":
		// Query current height
		heightBytes := []byte(fmt.Sprintf("%d", app.height))
		return &cmtabci.ResponseQuery{
			Code:  0,
			Value: heightBytes,
			Log:   "current height",
		}, nil

	default:
		return &cmtabci.ResponseQuery{
			Code: 1,
			Log:  "unknown query path",
		}, nil
	}
}

// InitChain initializes the blockchain
func (app *Application) InitChain(ctx context.Context, req *cmtabci.RequestInitChain) (*cmtabci.ResponseInitChain, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	// Initialize with genesis state
	if req.AppStateBytes != nil {
		// Parse and apply genesis state
		// For demo purposes, we'll just store it
		app.state["genesis"] = req.AppStateBytes
	}

	// Set initial validators
	validators := make([]*cmtabci.ValidatorUpdate, len(req.Validators))
	for i, val := range req.Validators {
		validators[i] = &cmtabci.ValidatorUpdate{
			PubKey: val.PubKey,
			Power:  val.Power,
		}
	}

	return &cmtabci.ResponseInitChain{
		ConsensusParams: &cmtabci.ConsensusParams{
			Block: &cmtabci.BlockParams{
				MaxBytes: 22020096, // 21MB
				MaxGas:   15000000, // 15M gas
			},
			Evidence: &cmtabci.EvidenceParams{
				MaxAgeNumBlocks: 100000,
				MaxAgeDuration:  172800000000000, // 48 hours in nanoseconds
				MaxBytes:        1048576,         // 1MB
			},
			Validator: &cmtabci.ValidatorParams{
				PubKeyTypes: []string{"ed25519", "secp256k1"},
			},
		},
		Validators: validators,
		AppHash:    app.appHash,
	}, nil
}

// PrepareProposal prepares a block proposal
func (app *Application) PrepareProposal(ctx context.Context, req *cmtabci.RequestPrepareProposal) (*cmtabci.ResponsePrepareProposal, error) {
	app.mtx.RLock()
	defer app.mtx.RUnlock()

	// Select transactions for the block
	var selectedTxs [][]byte
	totalSize := 0

	for _, tx := range req.Txs {
		// Check if adding this transaction would exceed the limit
		if totalSize+len(tx) > int(req.MaxTxBytes) {
			break
		}

		// Validate transaction
		checkRes, err := app.CheckTx(ctx, &cmtabci.RequestCheckTx{Tx: tx})
		if err != nil {
			continue // Skip invalid transactions
		}

		if checkRes.Code == 0 {
			selectedTxs = append(selectedTxs, tx)
			totalSize += len(tx)
		}
	}

	return &cmtabci.ResponsePrepareProposal{
		Txs: selectedTxs,
	}, nil
}

// ProcessProposal validates a block proposal
func (app *Application) ProcessProposal(ctx context.Context, req *cmtabci.RequestProcessProposal) (*cmtabci.ResponseProcessProposal, error) {
	app.mtx.RLock()
	defer app.mtx.RUnlock()

	// Validate all transactions in the proposal
	for _, tx := range req.Txs {
		checkRes, err := app.CheckTx(ctx, &cmtabci.RequestCheckTx{Tx: tx})
		if err != nil {
			return &cmtabci.ResponseProcessProposal{
				Status: cmtabci.ResponseProcessProposal_REJECT,
			}, nil
		}

		if checkRes.Code != 0 {
			return &cmtabci.ResponseProcessProposal{
				Status: cmtabci.ResponseProcessProposal_REJECT,
			}, nil
		}
	}

	return &cmtabci.ResponseProcessProposal{
		Status: cmtabci.ResponseProcessProposal_ACCEPT,
	}, nil
}

// ExtendVote extends a vote with application-specific data
func (app *Application) ExtendVote(ctx context.Context, req *cmtabci.RequestExtendVote) (*cmtabci.ResponseExtendVote, error) {
	app.mtx.RLock()
	defer app.mtx.RUnlock()

	// Create vote extension data
	extensionData := []byte(fmt.Sprintf("height_%d", req.Height))

	return &cmtabci.ResponseExtendVote{
		VoteExtension: extensionData,
	}, nil
}

// VerifyVoteExtension verifies a vote extension
func (app *Application) VerifyVoteExtension(ctx context.Context, req *cmtabci.RequestVerifyVoteExtension) (*cmtabci.ResponseVerifyVoteExtension, error) {
	app.mtx.RLock()
	defer app.mtx.RUnlock()

	// Verify vote extension format
	expectedData := []byte(fmt.Sprintf("height_%d", req.Height))
	if string(req.VoteExtension) == string(expectedData) {
		return &cmtabci.ResponseVerifyVoteExtension{
			Status: cmtabci.ResponseVerifyVoteExtension_ACCEPT,
		}, nil
	}

	return &cmtabci.ResponseVerifyVoteExtension{
		Status: cmtabci.ResponseVerifyVoteExtension_REJECT,
	}, nil
}

// ListSnapshots lists available snapshots
func (app *Application) ListSnapshots(ctx context.Context, req *cmtabci.RequestListSnapshots) (*cmtabci.ResponseListSnapshots, error) {
	app.mtx.RLock()
	defer app.mtx.RUnlock()

	// Return empty snapshot list for demo
	return &cmtabci.ResponseListSnapshots{
		Snapshots: []*cmtabci.Snapshot{},
	}, nil
}

// OfferSnapshot offers a snapshot to the application
func (app *Application) OfferSnapshot(ctx context.Context, req *cmtabci.RequestOfferSnapshot) (*cmtabci.ResponseOfferSnapshot, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	// For demo purposes, reject all snapshots
	return &cmtabci.ResponseOfferSnapshot{
		Result: cmtabci.ResponseOfferSnapshot_REJECT,
	}, nil
}

// LoadSnapshotChunk loads a snapshot chunk
func (app *Application) LoadSnapshotChunk(ctx context.Context, req *cmtabci.RequestLoadSnapshotChunk) (*cmtabci.ResponseLoadSnapshotChunk, error) {
	app.mtx.RLock()
	defer app.mtx.RUnlock()

	// Return empty chunk for demo
	return &cmtabci.ResponseLoadSnapshotChunk{
		Chunk: []byte{},
	}, nil
}

// ApplySnapshotChunk applies a snapshot chunk
func (app *Application) ApplySnapshotChunk(ctx context.Context, req *cmtabci.RequestApplySnapshotChunk) (*cmtabci.ResponseApplySnapshotChunk, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	// For demo purposes, reject all chunks
	return &cmtabci.ResponseApplySnapshotChunk{
		Result: cmtabci.ResponseApplySnapshotChunk_REJECT,
	}, nil
}

// Helper methods

func (app *Application) executeTransaction(tx []byte) (*cmtabci.ExecTxResult, error) {
	// Simple key-value transaction format: "key=value"
	txStr := string(tx)
	
	// Parse transaction
	var key, value string
	if len(txStr) > 0 {
		// Simple parsing for demo
		if len(txStr) >= 3 && txStr[1] == '=' {
			key = txStr[:1]
			value = txStr[2:]
		} else {
			key = txStr
			value = "default"
		}
	} else {
		return nil, fmt.Errorf("empty transaction")
	}

	// Store the key-value pair
	app.state[key] = []byte(value)

	return &cmtabci.ExecTxResult{
		Code: 0,
		Data: []byte(fmt.Sprintf("stored: %s=%s", key, value)),
		Log:  "transaction executed successfully",
		Events: []cmtabci.Event{
			{
				Type: "store",
				Attributes: []cmtabci.EventAttribute{
					{Key: "key", Value: key, Index: true},
					{Key: "value", Value: value, Index: false},
				},
			},
		},
	}, nil
}

func (app *Application) calculateAppHash() []byte {
	// Simple hash calculation for demo
	// In a real application, this would be a proper hash of the state
	hash := fmt.Sprintf("hash_%d_%d", app.height, len(app.state))
	return []byte(hash)
}

// GetState returns the current application state (for testing)
func (app *Application) GetState() map[string][]byte {
	app.mtx.RLock()
	defer app.mtx.RUnlock()
	
	stateCopy := make(map[string][]byte)
	for k, v := range app.state {
		stateCopy[k] = v
	}
	return stateCopy
}

// GetHeight returns the current height (for testing)
func (app *Application) GetHeight() int64 {
	app.mtx.RLock()
	defer app.mtx.RUnlock()
	return app.height
}

// GetAppHash returns the current app hash (for testing)
func (app *Application) GetAppHash() []byte {
	app.mtx.RLock()
	defer app.mtx.RUnlock()
	return app.appHash
} 