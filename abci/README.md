# ABCI Application Interface for CometBFT v0.38.17

This package provides a complete implementation of the ABCI (Application Blockchain Interface) for CometBFT v0.38.17 compatibility. It includes the core interface, base implementations, optional extensions, and a complete example application.

## Overview

The ABCI interface enables any finite, deterministic state machine to be driven by a blockchain-based replication engine. This implementation provides:

- **Core Application Interface**: Complete ABCI 2.0 interface with all required methods
- **Base Implementation**: Default implementations for all ABCI methods
- **Optional Extensions**: Interfaces for advanced features like snapshots and vote extensions
- **Example Application**: A working key-value store application demonstrating the interface
- **Comprehensive Testing**: Full test coverage and verification tools

## Architecture

```
abci/
├── types/
│   ├── application.go      # Core Application interface
│   ├── baseapp.go          # BaseApplication with default implementations
│   ├── extensions.go       # Optional interfaces (Snapshotter, etc.)
│   ├── mempool_types.go    # Mempool-related types
│   ├── consensus_types.go  # Consensus-related types
│   ├── query_types.go      # Query/Info-related types
│   ├── snapshot_types.go   # Snapshot-related types
│   ├── common_types.go     # Common types and enums
│   ├── conversions.go      # Proto-to-ABCI type conversions
│   └── types_test.go       # Type tests
├── app.go                  # Example MyApp implementation
├── app_test.go            # Application tests
├── verify_compliance.go   # Interface compliance verification
├── main.go                # Demo and verification runner
└── README.md              # This file
```

## Core Components

### 1. Application Interface (`types/application.go`)

The main `Application` interface that all ABCI applications must implement:

```go
type Application interface {
    // Info/Query Connection
    Info(context.Context, *RequestInfo) (*ResponseInfo, error)
    Query(context.Context, *RequestQuery) (*ResponseQuery, error)

    // Mempool Connection
    CheckTx(context.Context, *RequestCheckTx) (*ResponseCheckTx, error)

    // Consensus Connection
    PrepareProposal(context.Context, *RequestPrepareProposal) (*ResponsePrepareProposal, error)
    ProcessProposal(context.Context, *RequestProcessProposal) (*ResponseProcessProposal, error)
    FinalizeBlock(context.Context, *RequestFinalizeBlock) (*ResponseFinalizeBlock, error)
    ExtendVote(context.Context, *RequestExtendVote) (*ResponseExtendVote, error)
    VerifyVoteExtension(context.Context, *RequestVerifyVoteExtension) (*ResponseVerifyVoteExtension, error)
    Commit(context.Context, *RequestCommit) (*ResponseCommit, error)
    InitChain(context.Context, *RequestInitChain) (*ResponseInitChain, error)

    // State Sync Connection
    ListSnapshots(context.Context, *RequestListSnapshots) (*ResponseListSnapshots, error)
    OfferSnapshot(context.Context, *RequestOfferSnapshot) (*ResponseOfferSnapshot, error)
    LoadSnapshotChunk(context.Context, *RequestLoadSnapshotChunk) (*ResponseLoadSnapshotChunk, error)
    ApplySnapshotChunk(context.Context, *RequestApplySnapshotChunk) (*ResponseApplySnapshotChunk, error)
}
```

### 2. Base Implementation (`types/baseapp.go`)

`BaseApplication` provides default implementations for all ABCI methods:

```go
type BaseApplication struct{}

// Applications can embed this and override only the methods they need
type MyApp struct {
    types.BaseApplication
    // Your custom fields
}
```

### 3. Optional Interfaces (`types/extensions.go`)

Optional interfaces for advanced features:

- `Snapshotter`: For state sync functionality
- `ValidatorSetUpdater`: For validator set management
- `ProposalProcessor`: For custom proposal logic
- `VoteExtensionProcessor`: For vote extension handling
- `StateManager`: For state management
- `TransactionProcessor`: For transaction processing
- `EventEmitter`: For event emission
- `GasMeter`: For gas metering
- `HeightManager`: For height management
- `ChainIDManager`: For chain ID management

## Example Application

The `MyApp` implementation demonstrates a complete key-value store application:

### Features

- **Key-Value Storage**: Simple SET/GET operations
- **Transaction Validation**: Format and size validation
- **Gas Metering**: Basic gas consumption tracking
- **Event Emission**: Transaction and block events
- **State Persistence**: App hash computation
- **Thread Safety**: Concurrent access protection

### Usage

```go
// Create a new application
app := abci.NewMyApp("my-chain")

// Initialize the chain
initReq := &types.RequestInitChain{
    ChainId:       "my-chain",
    InitialHeight: 1,
}
res, err := app.InitChain(context.Background(), initReq)

// Process transactions
txs := [][]byte{
    []byte("SET key1=value1"),
    []byte("SET key2=value2"),
    []byte("GET key1"),
}

finalizeRes, err := app.FinalizeBlock(context.Background(), &types.RequestFinalizeBlock{
    Height: 1,
    Txs:    txs,
})

// Commit the block
commitRes, err := app.Commit(context.Background(), &types.RequestCommit{})
```

## Response Codes

The implementation includes standard ABCI response codes:

```go
const (
    CodeTypeOK                uint32 = 0
    CodeTypeInternalError     uint32 = 1
    CodeTypeEncodingError     uint32 = 2
    CodeTypeUnauthorized      uint32 = 3
    CodeTypeInsufficientFunds uint32 = 4
    CodeTypeUnknownRequest    uint32 = 5
    // ... more codes
)
```

## Testing and Verification

### Running Tests

```bash
# Run all tests
go test ./abci/...

# Run specific test files
go test ./abci/app_test.go
go test ./abci/types/types_test.go
```

### Interface Compliance Verification

```bash
# Run verification checks
go run ./abci/main.go
```

The verification checks:
- Interface compliance for all implementations
- Optional interface definitions
- Type compatibility with CometBFT
- Response code validation

### Manual Verification

```go
// Check interface compliance at compile time
var _ types.Application = (*MyApp)(nil)

// Run runtime verification
if err := abci.VerifyCompliance(); err != nil {
    log.Fatal(err)
}
```

## Migration Guide

### From Legacy ABCI

1. **Update Method Signatures**:
   ```go
   // Old
   func (app *MyApp) CheckTx(tx []byte) abci.ResponseCheckTx
   
   // New
   func (app *MyApp) CheckTx(ctx context.Context, req *types.RequestCheckTx) (*types.ResponseCheckTx, error)
   ```

2. **Replace BeginBlock/EndBlock**:
   ```go
   // Old
   func (app *MyApp) BeginBlock(req abci.RequestBeginBlock) abci.ResponseBeginBlock
   func (app *MyApp) EndBlock(req abci.RequestEndBlock) abci.ResponseEndBlock
   
   // New
   func (app *MyApp) FinalizeBlock(ctx context.Context, req *types.RequestFinalizeBlock) (*types.ResponseFinalizeBlock, error)
   ```

3. **Add Context and Error Handling**:
   ```go
   // All methods now require context and return errors
   func (app *MyApp) Commit(ctx context.Context, req *types.RequestCommit) (*types.ResponseCommit, error)
   ```

4. **Update Type Imports**:
   ```go
   // Use the new types package
   import "github.com/fluentum-chain/fluentum/abci/types"
   ```

### Implementation Steps

1. **Start with BaseApplication**:
   ```go
   type MyApp struct {
       types.BaseApplication
       // Your fields
   }
   ```

2. **Override Required Methods**:
   ```go
   func (app *MyApp) CheckTx(ctx context.Context, req *types.RequestCheckTx) (*types.ResponseCheckTx, error) {
       // Your implementation
   }
   
   func (app *MyApp) FinalizeBlock(ctx context.Context, req *types.RequestFinalizeBlock) (*types.ResponseFinalizeBlock, error) {
       // Your implementation
   }
   ```

3. **Add Optional Interfaces**:
   ```go
   func (app *MyApp) ListSnapshots(ctx context.Context, req *types.RequestListSnapshots) (*types.ResponseListSnapshots, error) {
       // Implement if you need snapshots
   }
   ```

4. **Test Interface Compliance**:
   ```go
   var _ types.Application = (*MyApp)(nil)
   ```

## Key Features

### Thread Safety

All implementations are thread-safe with proper mutex protection:

```go
type MyApp struct {
    types.BaseApplication
    mtx sync.RWMutex
    // ...
}

func (app *MyApp) FinalizeBlock(ctx context.Context, req *types.RequestFinalizeBlock) (*types.ResponseFinalizeBlock, error) {
    app.mtx.Lock()
    defer app.mtx.Unlock()
    // ...
}
```

### Context Support

All methods support context for cancellation and timeouts:

```go
func (app *MyApp) CheckTx(ctx context.Context, req *types.RequestCheckTx) (*types.ResponseCheckTx, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
        // Process transaction
    }
}
```

### Error Handling

Comprehensive error handling with proper response codes:

```go
func (app *MyApp) CheckTx(ctx context.Context, req *types.RequestCheckTx) (*types.ResponseCheckTx, error) {
    if len(req.Tx) == 0 {
        return &types.ResponseCheckTx{
            Code: types.CodeTypeEncodingError,
            Log:  "empty transaction",
        }, nil
    }
    // ...
}
```

### Gas Metering

Built-in gas metering support:

```go
type SimpleGasMeter struct {
    consumed int64
    limit    int64
}

func (gm *SimpleGasMeter) ConsumeGas(amount int64, descriptor string) error {
    if gm.consumed+amount > gm.limit {
        return fmt.Errorf("out of gas: %s", descriptor)
    }
    gm.consumed += amount
    return nil
}
```

## Dependencies

- Go 1.19+
- CometBFT v0.38.17
- testify (for testing)

## License

This implementation follows the same license as the main project.

## Contributing

1. Ensure all tests pass
2. Run interface compliance verification
3. Follow the existing code style
4. Add tests for new functionality
5. Update documentation as needed
