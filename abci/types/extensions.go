package types

import (
	"context"
)

// Snapshotter is an optional interface for applications that support state sync
// Applications can implement this interface to enable snapshot functionality
type Snapshotter interface {
	ListSnapshots(context.Context, *RequestListSnapshots) (*ResponseListSnapshots, error)
	OfferSnapshot(context.Context, *RequestOfferSnapshot) (*ResponseOfferSnapshot, error)
	LoadSnapshotChunk(context.Context, *RequestLoadSnapshotChunk) (*ResponseLoadSnapshotChunk, error)
	ApplySnapshotChunk(context.Context, *RequestApplySnapshotChunk) (*ResponseApplySnapshotChunk, error)
}

// ValidatorSetUpdater is an optional interface for applications that need to
// handle validator set updates outside of the normal ABCI flow
type ValidatorSetUpdater interface {
	ApplyValidatorSetUpdates(context.Context, []ValidatorUpdate) error
}

// ProposalProcessor is an optional interface for applications that need custom
// proposal preparation and processing logic
type ProposalProcessor interface {
	PrepareProposal(context.Context, *RequestPrepareProposal) (*ResponsePrepareProposal, error)
	ProcessProposal(context.Context, *RequestProcessProposal) (*ResponseProcessProposal, error)
}

// VoteExtensionProcessor is an optional interface for applications that need custom
// vote extension logic
type VoteExtensionProcessor interface {
	ExtendVote(context.Context, *RequestExtendVote) (*ResponseExtendVote, error)
	VerifyVoteExtension(context.Context, *RequestVerifyVoteExtension) (*ResponseVerifyVoteExtension, error)
}

// StateManager is an optional interface for applications that need custom
// state management logic
type StateManager interface {
	// GetState returns the current application state
	GetState() interface{}
	
	// SetState updates the application state
	SetState(state interface{}) error
	
	// SaveState persists the current state
	SaveState() error
	
	// LoadState loads the persisted state
	LoadState() error
}

// TransactionProcessor is an optional interface for applications that need custom
// transaction processing logic
type TransactionProcessor interface {
	// ProcessTransaction processes a single transaction and returns the result
	ProcessTransaction(ctx context.Context, tx []byte) (*ExecTxResult, error)
	
	// ValidateTransaction validates a transaction without processing it
	ValidateTransaction(ctx context.Context, tx []byte) error
}

// EventEmitter is an optional interface for applications that need custom
// event emission logic
type EventEmitter interface {
	// EmitEvent emits an event during transaction processing
	EmitEvent(event Event) error
	
	// GetEvents returns all events emitted during the current block
	GetEvents() []Event
	
	// ClearEvents clears all events for the current block
	ClearEvents()
}

// GasMeter is an optional interface for applications that need custom
// gas metering logic
type GasMeter interface {
	// ConsumeGas consumes the specified amount of gas
	ConsumeGas(amount int64, descriptor string) error
	
	// RefundGas refunds the specified amount of gas
	RefundGas(amount int64, descriptor string)
	
	// GasConsumed returns the total gas consumed
	GasConsumed() int64
	
	// GasLimit returns the gas limit
	GasLimit() int64
	
	// IsOutOfGas returns true if the gas meter is out of gas
	IsOutOfGas() bool
}

// HeightManager is an optional interface for applications that need custom
// height management logic
type HeightManager interface {
	// GetHeight returns the current block height
	GetHeight() int64
	
	// SetHeight sets the current block height
	SetHeight(height int64)
	
	// IncrementHeight increments the current block height
	IncrementHeight()
}

// ChainIDManager is an optional interface for applications that need custom
// chain ID management logic
type ChainIDManager interface {
	// GetChainID returns the current chain ID
	GetChainID() string
	
	// SetChainID sets the current chain ID
	SetChainID(chainID string)
	
	// ValidateChainID validates the provided chain ID
	ValidateChainID(chainID string) error
} 