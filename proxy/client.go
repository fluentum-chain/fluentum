package proxy

import (
	"context"
	"fmt"

	cmtabci "github.com/cometbft/cometbft/abci/types"
	abcicli "github.com/fluentum-chain/fluentum/abci/client"
	"github.com/fluentum-chain/fluentum/abci/example/kvstore"
	abci "github.com/fluentum-chain/fluentum/abci/types"
	tmsync "github.com/fluentum-chain/fluentum/libs/sync"
)

//go:generate ../scripts/mockery_generate.sh ClientCreator

// ClientCreator creates new ABCI clients.
type ClientCreator interface {
	// NewABCIClient returns a new ABCI client.
	NewABCIClient() (abcicli.Client, error)
}

//----------------------------------------------------
// local proxy uses a mutex on an in-proc app

type localClientCreator struct {
	mtx *tmsync.Mutex
	app abci.Application
}

// NewLocalClientCreator returns a ClientCreator for the given app,
// which will be running locally.
func NewLocalClientCreator(app abci.Application) ClientCreator {
	return &localClientCreator{
		mtx: new(tmsync.Mutex),
		app: app,
	}
}

func (l *localClientCreator) NewABCIClient() (abcicli.Client, error) {
	// Create an adapter to convert local ABCI app to CometBFT ABCI app
	adapter := &ABCIAdapter{app: l.app}
	return abcicli.NewLocalClient(l.mtx, adapter), nil
}

//---------------------------------------------------------------
// remote proxy opens new connections to an external app process

type remoteClientCreator struct {
	addr        string
	transport   string
	mustConnect bool
}

// NewRemoteClientCreator returns a ClientCreator for the given address (e.g.
// "192.168.0.1") and transport (e.g. "tcp"). Set mustConnect to true if you
// want the client to connect before reporting success.
func NewRemoteClientCreator(addr, transport string, mustConnect bool) ClientCreator {
	return &remoteClientCreator{
		addr:        addr,
		transport:   transport,
		mustConnect: mustConnect,
	}
}

func (r *remoteClientCreator) NewABCIClient() (abcicli.Client, error) {
	remoteApp, err := abcicli.NewClient(r.addr, r.transport, r.mustConnect)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to proxy: %w", err)
	}

	return remoteApp, nil
}

// DefaultClientCreator returns a default ClientCreator, which will create a
// local client if addr is one of: 'kvstore', 'persistent_kvstore' or 'noop', otherwise - a remote client.
func DefaultClientCreator(addr, transport, dbDir string) ClientCreator {
	switch addr {
	case "kvstore":
		return NewLocalClientCreator(kvstore.NewApplication())
	case "persistent_kvstore":
		return NewLocalClientCreator(kvstore.NewPersistentKVStoreApplication(dbDir))
	case "e2e":
		// TODO: Fix e2e app to use local ABCI types
		panic("e2e app not yet compatible with local ABCI types")
	case "noop":
		return NewLocalClientCreator(abci.NewBaseApplication())
	default:
		mustConnect := false // loop retrying
		return NewRemoteClientCreator(addr, transport, mustConnect)
	}
}

// ABCIAdapter adapts local ABCI applications to CometBFT ABCI interface
type ABCIAdapter struct {
	app abci.Application
}

func NewABCIAdapter(app abci.Application) cmtabci.Application {
	return &ABCIAdapter{app: app}
}

func (a *ABCIAdapter) Echo(ctx context.Context, req *cmtabci.RequestEcho) (*cmtabci.ResponseEcho, error) {
	localReq := &abci.EchoRequest{Message: req.Message}
	resp, err := a.app.Echo(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponseEcho{Message: resp.Message}, nil
}

func (a *ABCIAdapter) Flush(ctx context.Context, req *cmtabci.RequestFlush) (*cmtabci.ResponseFlush, error) {
	return &cmtabci.ResponseFlush{}, nil
}

func (a *ABCIAdapter) Info(ctx context.Context, req *cmtabci.RequestInfo) (*cmtabci.ResponseInfo, error) {
	localReq := &abci.InfoRequest{
		Version:      req.Version,
		BlockVersion: req.BlockVersion,
		P2PVersion:   req.P2PVersion,
	}
	resp, err := a.app.Info(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponseInfo{
		Data:             resp.Data,
		Version:          resp.Version,
		AppVersion:       resp.AppVersion,
		LastBlockHeight:  resp.LastBlockHeight,
		LastBlockAppHash: resp.LastBlockAppHash,
	}, nil
}

func (a *ABCIAdapter) CheckTx(ctx context.Context, req *cmtabci.RequestCheckTx) (*cmtabci.ResponseCheckTx, error) {
	localReq := &abci.CheckTxRequest{
		Tx:   req.Tx,
		Type: abci.CheckTxType(req.Type),
	}
	resp, err := a.app.CheckTx(ctx, localReq)
	if err != nil {
		return nil, err
	}
	// Convert local events to CometBFT events if needed
	var events []cmtabci.Event
	if resp.Events != nil {
		events = make([]cmtabci.Event, len(resp.Events))
		for i, ev := range resp.Events {
			events[i] = cmtabci.Event{
				Type:       ev.Type,
				Attributes: nil, // Conversion for attributes if needed
			}
		}
	}
	return &cmtabci.ResponseCheckTx{
		Code:      resp.Code,
		Data:      resp.Data,
		Log:       resp.Log,
		Info:      resp.Info,
		GasWanted: resp.GasWanted,
		GasUsed:   resp.GasUsed,
		Events:    events,
		Codespace: resp.Codespace,
	}, nil
}

func (a *ABCIAdapter) Query(ctx context.Context, req *cmtabci.RequestQuery) (*cmtabci.ResponseQuery, error) {
	localReq := &abci.QueryRequest{
		Data:   req.Data,
		Path:   req.Path,
		Height: req.Height,
		Prove:  req.Prove,
	}
	resp, err := a.app.Query(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponseQuery{
		Code:      resp.Code,
		Log:       resp.Log,
		Info:      resp.Info,
		Index:     resp.Index,
		Key:       resp.Key,
		Value:     resp.Value,
		ProofOps:  nil, // Convert proof ops if needed
		Height:    resp.Height,
		Codespace: resp.Codespace,
	}, nil
}

func (a *ABCIAdapter) Commit(ctx context.Context, req *cmtabci.RequestCommit) (*cmtabci.ResponseCommit, error) {
	localReq := &abci.CommitRequest{}
	resp, err := a.app.Commit(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponseCommit{
		RetainHeight: resp.RetainHeight,
	}, nil
}

func (a *ABCIAdapter) InitChain(ctx context.Context, req *cmtabci.RequestInitChain) (*cmtabci.ResponseInitChain, error) {
	localReq := &abci.InitChainRequest{
		Time:            req.Time,
		ChainId:         req.ChainId,
		ConsensusParams: req.ConsensusParams,
		Validators:      req.Validators,
		AppStateBytes:   req.AppStateBytes,
		InitialHeight:   req.InitialHeight,
	}
	resp, err := a.app.InitChain(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponseInitChain{
		Validators:      resp.Validators,
		ConsensusParams: resp.ConsensusParams,
		AppHash:         resp.AppHash,
	}, nil
}

func (a *ABCIAdapter) FinalizeBlock(ctx context.Context, req *cmtabci.RequestFinalizeBlock) (*cmtabci.ResponseFinalizeBlock, error) {
	localReq := &abci.FinalizeBlockRequest{
		Txs:                req.Txs,
		DecidedLastCommit:  req.DecidedLastCommit,
		Misbehavior:        req.Misbehavior,
		Hash:               req.Hash,
		Height:             req.Height,
		Time:               req.Time,
		NextValidatorsHash: req.NextValidatorsHash,
		ProposerAddress:    req.ProposerAddress,
	}
	resp, err := a.app.FinalizeBlock(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponseFinalizeBlock{
		Events:                []cmtabci.Event{},         // Convert events
		TxResults:             []*cmtabci.ExecTxResult{}, // Convert tx results
		ValidatorUpdates:      resp.ValidatorUpdates,
		ConsensusParamUpdates: resp.ConsensusParamUpdates,
		AppHash:               resp.AppHash,
	}, nil
}

// Implement other required methods with default implementations
func (a *ABCIAdapter) ListSnapshots(ctx context.Context, req *cmtabci.RequestListSnapshots) (*cmtabci.ResponseListSnapshots, error) {
	localReq := &abci.ListSnapshotsRequest{}
	resp, err := a.app.ListSnapshots(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponseListSnapshots{
		Snapshots: resp.Snapshots,
	}, nil
}

func (a *ABCIAdapter) OfferSnapshot(ctx context.Context, req *cmtabci.RequestOfferSnapshot) (*cmtabci.ResponseOfferSnapshot, error) {
	localReq := &abci.OfferSnapshotRequest{
		Snapshot: req.Snapshot,
		AppHash:  req.AppHash,
	}
	resp, err := a.app.OfferSnapshot(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponseOfferSnapshot{
		Result: cmtabci.ResponseOfferSnapshot_Result(resp.Result),
	}, nil
}

func (a *ABCIAdapter) LoadSnapshotChunk(ctx context.Context, req *cmtabci.RequestLoadSnapshotChunk) (*cmtabci.ResponseLoadSnapshotChunk, error) {
	localReq := &abci.LoadSnapshotChunkRequest{
		Height: req.Height,
		Format: req.Format,
		Chunk:  req.Chunk,
	}
	resp, err := a.app.LoadSnapshotChunk(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponseLoadSnapshotChunk{
		Chunk: resp.Chunk,
	}, nil
}

func (a *ABCIAdapter) ApplySnapshotChunk(ctx context.Context, req *cmtabci.RequestApplySnapshotChunk) (*cmtabci.ResponseApplySnapshotChunk, error) {
	localReq := &abci.ApplySnapshotChunkRequest{
		Index:  req.Index,
		Chunk:  req.Chunk,
		Sender: req.Sender,
	}
	resp, err := a.app.ApplySnapshotChunk(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponseApplySnapshotChunk{
		Result:        cmtabci.ResponseApplySnapshotChunk_Result(resp.Result),
		RefetchChunks: resp.RefetchChunks,
		RejectSenders: resp.RejectSenders,
	}, nil
}

func (a *ABCIAdapter) PrepareProposal(ctx context.Context, req *cmtabci.RequestPrepareProposal) (*cmtabci.ResponsePrepareProposal, error) {
	localReq := &abci.PrepareProposalRequest{
		Txs: req.Txs,
	}
	resp, err := a.app.PrepareProposal(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponsePrepareProposal{
		Txs: resp.Txs,
	}, nil
}

func (a *ABCIAdapter) ProcessProposal(ctx context.Context, req *cmtabci.RequestProcessProposal) (*cmtabci.ResponseProcessProposal, error) {
	localReq := &abci.ProcessProposalRequest{
		Txs: req.Txs,
	}
	resp, err := a.app.ProcessProposal(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponseProcessProposal{
		Status: resp.Status,
	}, nil
}

func (a *ABCIAdapter) ExtendVote(ctx context.Context, req *cmtabci.RequestExtendVote) (*cmtabci.ResponseExtendVote, error) {
	localReq := &abci.ExtendVoteRequest{
		Hash:   req.Hash,
		Height: req.Height,
	}
	resp, err := a.app.ExtendVote(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponseExtendVote{
		VoteExtension: resp.VoteExtension,
	}, nil
}

func (a *ABCIAdapter) VerifyVoteExtension(ctx context.Context, req *cmtabci.RequestVerifyVoteExtension) (*cmtabci.ResponseVerifyVoteExtension, error) {
	localReq := &abci.VerifyVoteExtensionRequest{
		Hash:             req.Hash,
		ValidatorAddress: req.ValidatorAddress,
		Height:           req.Height,
		VoteExtension:    req.VoteExtension,
	}
	resp, err := a.app.VerifyVoteExtension(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponseVerifyVoteExtension{
		Status: resp.Status,
	}, nil
}
