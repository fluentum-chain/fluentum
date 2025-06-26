package abci

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"github.com/fluentum-chain/fluentum/abci/types"
	cmtabci "github.com/cometbft/cometbft/abci/types"
)

// MyApp is an example ABCI application implementation
type MyApp struct {
	types.BaseApplication
	
	// Application state
	state     map[string][]byte
	height    int64
	chainID   string
	events    []cmtabci.Event
	gasMeter  *SimpleGasMeter
	
	// Thread safety
	mtx sync.RWMutex
}

// NewMyApp creates a new instance of MyApp
func NewMyApp(chainID string) *MyApp {
	return &MyApp{
		state:    make(map[string][]byte),
		height:   0,
		chainID:  chainID,
		events:   []cmtabci.Event{},
		gasMeter: NewSimpleGasMeter(1000000), // 1M gas limit
	}
}

// Info returns application information
func (app *MyApp) Info(ctx context.Context, req *types.RequestInfo) (*types.ResponseInfo, error) {
	app.mtx.RLock()
	defer app.mtx.RUnlock()
	
	return &cmtabci.ResponseInfo{
		Data:             fmt.Sprintf("MyApp v1.0.0 (chain: %s)", app.chainID),
		Version:          "1.0.0",
		AppVersion:       1,
		LastBlockHeight:  app.height,
		LastBlockAppHash: app.getAppHash(),
	}, nil
}

// CheckTx validates a transaction for the mempool
func (app *MyApp) CheckTx(ctx context.Context, req *types.RequestCheckTx) (*types.ResponseCheckTx, error) {
	// Validate transaction format
	if len(req.Tx) == 0 {
		return &cmtabci.ResponseCheckTx{
			Code:      types.CodeTypeEncodingError,
			Log:       "empty transaction",
			GasWanted: 0,
			GasUsed:   0,
		}, nil
	}
	
	// Simple validation: check if transaction is at least 4 bytes
	if len(req.Tx) < 4 {
		return &cmtabci.ResponseCheckTx{
			Code:      types.CodeTypeEncodingError,
			Log:       "transaction too short",
			GasWanted: 0,
			GasUsed:   0,
		}, nil
	}
	
	// Estimate gas (simple: 1 gas per byte)
	gasWanted := int64(len(req.Tx))
	
	return &cmtabci.ResponseCheckTx{
		Code:      types.CodeTypeOK,
		Data:      []byte("valid"),
		Log:       "transaction valid",
		GasWanted: gasWanted,
		GasUsed:   0,
		Events:    []cmtabci.Event{},
		Codespace: "",
	}, nil
}

// FinalizeBlock processes all transactions in a block
func (app *MyApp) FinalizeBlock(ctx context.Context, req *types.RequestFinalizeBlock) (*types.ResponseFinalizeBlock, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()
	
	// Update height
	app.height = req.Height
	
	// Clear events for new block
	app.events = []cmtabci.Event{}
	
	// Process all transactions
	txResults := make([]*cmtabci.ExecTxResult, len(req.Txs))
	for i, tx := range req.Txs {
		result, err := app.processTransaction(ctx, tx)
		if err != nil {
			result = &cmtabci.ExecTxResult{
				Code:   types.CodeTypeInternalError,
				Log:    fmt.Sprintf("processing error: %v", err),
				GasUsed: 0,
			}
		}
		txResults[i] = result
	}
	
	// Create block event
	blockEvent := cmtabci.Event{
		Type: "block",
		Attributes: []cmtabci.EventAttribute{
			{Key: "height", Value: fmt.Sprintf("%d", app.height), Index: true},
			{Key: "num_txs", Value: fmt.Sprintf("%d", len(req.Txs)), Index: false},
		},
	}
	app.events = append(app.events, blockEvent)
	
	return &cmtabci.ResponseFinalizeBlock{
		TxResults:             txResults,
		ValidatorUpdates:      []cmtabci.ValidatorUpdate{},
		ConsensusParamUpdates: nil,
		AppHash:               app.getAppHash(),
		Events:                app.events,
	}, nil
}

// Commit commits the current state and returns the app hash
func (app *MyApp) Commit(ctx context.Context, req *types.RequestCommit) (*types.ResponseCommit, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()
	
	// In a real application, you would persist the state here
	// For this example, we just return the app hash
	
	return &cmtabci.ResponseCommit{
		Data:         app.getAppHash(),
		RetainHeight: app.height,
	}, nil
}

// InitChain initializes the blockchain
func (app *MyApp) InitChain(ctx context.Context, req *types.RequestInitChain) (*types.ResponseInitChain, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()
	
	// Set chain ID
	app.chainID = req.ChainId
	
	// Set initial height
	app.height = req.InitialHeight
	
	// Initialize state with genesis data if provided
	if req.AppStateBytes != nil && len(req.AppStateBytes) > 0 {
		// In a real application, you would deserialize and apply the genesis state
		app.state["genesis"] = req.AppStateBytes
	}
	
	return &cmtabci.ResponseInitChain{
		ConsensusParams: req.ConsensusParams,
		Validators:      req.Validators,
		AppHash:         app.getAppHash(),
	}, nil
}

// Query handles queries to the application state
func (app *MyApp) Query(ctx context.Context, req *types.RequestQuery) (*types.ResponseQuery, error) {
	app.mtx.RLock()
	defer app.mtx.RUnlock()
	
	// Simple key-value query
	if req.Path == "state" {
		key := string(req.Data)
		value, exists := app.state[key]
		if !exists {
			return &cmtabci.ResponseQuery{
				Code:   types.CodeTypeUnknownAddress,
				Log:    fmt.Sprintf("key not found: %s", key),
				Height: app.height,
			}, nil
		}
		
		return &cmtabci.ResponseQuery{
			Code:   types.CodeTypeOK,
			Value:  value,
			Height: app.height,
		}, nil
	}
	
	return &cmtabci.ResponseQuery{
		Code:   types.CodeTypeUnknownRequest,
		Log:    fmt.Sprintf("unknown query path: %s", req.Path),
		Height: app.height,
	}, nil
}

// Helper methods

// processTransaction processes a single transaction
func (app *MyApp) processTransaction(ctx context.Context, tx []byte) (*cmtabci.ExecTxResult, error) {
	// Reset gas meter for new transaction
	app.gasMeter.Reset()
	
	// Simple transaction format: first 4 bytes are command, rest is data
	if len(tx) < 4 {
		return &cmtabci.ExecTxResult{
			Code:     types.CodeTypeEncodingError,
			Log:      "transaction too short",
			GasUsed:  0,
		}, nil
	}
	
	command := string(tx[:4])
	data := tx[4:]
	
	// Consume gas for processing
	gasUsed := int64(len(tx))
	if err := app.gasMeter.ConsumeGas(gasUsed, "tx_processing"); err != nil {
		return &cmtabci.ExecTxResult{
			Code:     types.CodeTypeOutOfGas,
			Log:      "out of gas",
			GasUsed:  app.gasMeter.GasConsumed(),
		}, nil
	}
	
	var result *cmtabci.ExecTxResult
	
	switch command {
	case "SET ":
		result = app.handleSet(data)
	case "GET ":
		result = app.handleGet(data)
	default:
		result = &cmtabci.ExecTxResult{
			Code:     types.CodeTypeUnknownRequest,
			Log:      fmt.Sprintf("unknown command: %s", command),
			GasUsed:  gasUsed,
		}
	}
	
	// Add transaction event
	txEvent := cmtabci.Event{
		Type: "transaction",
		Attributes: []cmtabci.EventAttribute{
			{Key: "command", Value: command, Index: true},
			{Key: "gas_used", Value: fmt.Sprintf("%d", gasUsed), Index: false},
		},
	}
	app.events = append(app.events, txEvent)
	
	return result, nil
}

// handleSet handles SET commands
func (app *MyApp) handleSet(data []byte) *cmtabci.ExecTxResult {
	// Simple format: key=value
	parts := bytes.Split(data, []byte("="))
	if len(parts) != 2 {
		return &cmtabci.ExecTxResult{
			Code:     types.CodeTypeEncodingError,
			Log:      "invalid SET format, expected key=value",
			GasUsed:  int64(len(data)),
		}
	}
	
	key := string(parts[0])
	value := parts[1]
	
	app.state[key] = value
	
	return &cmtabci.ExecTxResult{
		Code:     types.CodeTypeOK,
		Data:     []byte("set"),
		Log:      fmt.Sprintf("set %s", key),
		GasUsed:  int64(len(data)),
	}
}

// handleGet handles GET commands
func (app *MyApp) handleGet(data []byte) *cmtabci.ExecTxResult {
	key := string(data)
	value, exists := app.state[key]
	
	if !exists {
		return &cmtabci.ExecTxResult{
			Code:     types.CodeTypeUnknownAddress,
			Log:      fmt.Sprintf("key not found: %s", key),
			GasUsed:  int64(len(data)),
		}
	}
	
	return &cmtabci.ExecTxResult{
		Code:     types.CodeTypeOK,
		Data:     value,
		Log:      fmt.Sprintf("get %s", key),
		GasUsed:  int64(len(data)),
	}
}

// getAppHash returns a simple app hash based on the current state
func (app *MyApp) getAppHash() []byte {
	// In a real application, you would compute a proper Merkle hash
	// For this example, we use a simple hash of the height and state size
	hash := fmt.Sprintf("%d-%d", app.height, len(app.state))
	return []byte(hash)
}

// SimpleGasMeter is a basic gas meter implementation
type SimpleGasMeter struct {
	consumed int64
	limit    int64
}

func NewSimpleGasMeter(limit int64) *SimpleGasMeter {
	return &SimpleGasMeter{
		consumed: 0,
		limit:    limit,
	}
}

func (gm *SimpleGasMeter) ConsumeGas(amount int64, descriptor string) error {
	if gm.consumed+amount > gm.limit {
		return fmt.Errorf("out of gas: %s", descriptor)
	}
	gm.consumed += amount
	return nil
}

func (gm *SimpleGasMeter) RefundGas(amount int64, descriptor string) {
	gm.consumed -= amount
	if gm.consumed < 0 {
		gm.consumed = 0
	}
}

func (gm *SimpleGasMeter) GasConsumed() int64 {
	return gm.consumed
}

func (gm *SimpleGasMeter) GasLimit() int64 {
	return gm.limit
}

func (gm *SimpleGasMeter) IsOutOfGas() bool {
	return gm.consumed >= gm.limit
}

func (gm *SimpleGasMeter) Reset() {
	gm.consumed = 0
} 