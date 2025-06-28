package server

import (
	"context"

	cmtabci "github.com/cometbft/cometbft/abci/types"
	cmtcrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	cmtcoretypes "github.com/cometbft/cometbft/types"
	cmttypes "github.com/cometbft/cometbft/types"

	"github.com/fluentum-chain/fluentum/abci/types"
	abcipb "github.com/fluentum-chain/fluentum/proto/tendermint/abci"
	statepb "github.com/fluentum-chain/fluentum/proto/tendermint/state"
)

// ABCIAdapter converts a local ABCI application to a CometBFT ABCI application
type ABCIAdapter struct {
	app types.Application
}

// NewABCIAdapter creates a new adapter that wraps a local ABCI application
func NewABCIAdapter(app types.Application) *ABCIAdapter {
	return &ABCIAdapter{app: app}
}

// Echo implements cmtabci.Application
func (a *ABCIAdapter) Echo(ctx context.Context, req *cmtabci.RequestEcho) (*cmtabci.ResponseEcho, error) {
	resp, err := a.app.Echo(ctx, &types.EchoRequest{Message: req.Message})
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponseEcho{Message: resp.Message}, nil
}

// Flush implements cmtabci.Application
func (a *ABCIAdapter) Flush(ctx context.Context, req *cmtabci.RequestFlush) (*cmtabci.ResponseFlush, error) {
	// Flush is not part of the local ABCI interface, so we just return success
	return &cmtabci.ResponseFlush{}, nil
}

// Info implements cmtabci.Application
func (a *ABCIAdapter) Info(ctx context.Context, req *cmtabci.RequestInfo) (*cmtabci.ResponseInfo, error) {
	resp, err := a.app.Info(ctx, &types.InfoRequest{
		Version:      req.Version,
		BlockVersion: req.BlockVersion,
		P2PVersion:   req.P2PVersion,
		AbciVersion:  req.AbciVersion,
	})
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

// CheckTx implements cmtabci.Application
func (a *ABCIAdapter) CheckTx(ctx context.Context, req *cmtabci.RequestCheckTx) (*cmtabci.ResponseCheckTx, error) {
	resp, err := a.app.CheckTx(ctx, &types.CheckTxRequest{
		Tx:   req.Tx,
		Type: types.CheckTxType(req.Type),
	})
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponseCheckTx{
		Code:      resp.Code,
		Data:      resp.Data,
		Log:       resp.Log,
		Info:      resp.Info,
		GasWanted: resp.GasWanted,
		GasUsed:   resp.GasUsed,
		Events:    convertEvents(resp.Events),
		Codespace: resp.Codespace,
	}, nil
}

// Query implements cmtabci.Application
func (a *ABCIAdapter) Query(ctx context.Context, req *cmtabci.RequestQuery) (*cmtabci.ResponseQuery, error) {
	resp, err := a.app.Query(ctx, &types.QueryRequest{
		Data:   req.Data,
		Path:   req.Path,
		Height: req.Height,
		Prove:  req.Prove,
	})
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
		ProofOps:  convertProofOpsFromCometBFT(resp.ProofOps),
		Height:    resp.Height,
		Codespace: resp.Codespace,
	}, nil
}

// Commit implements cmtabci.Application
func (a *ABCIAdapter) Commit(ctx context.Context, req *cmtabci.RequestCommit) (*cmtabci.ResponseCommit, error) {
	resp, err := a.app.Commit(ctx, &types.CommitRequest{})
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponseCommit{
		RetainHeight: resp.RetainHeight,
	}, nil
}

// InitChain implements cmtabci.Application
func (a *ABCIAdapter) InitChain(ctx context.Context, req *cmtabci.RequestInitChain) (*cmtabci.ResponseInitChain, error) {
	resp, err := a.app.InitChain(ctx, &types.InitChainRequest{
		Time:            req.Time,
		ChainId:         req.ChainId,
		ConsensusParams: req.ConsensusParams,
		Validators:      req.Validators,
		AppStateBytes:   req.AppStateBytes,
		InitialHeight:   req.InitialHeight,
	})
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponseInitChain{
		ConsensusParams: resp.ConsensusParams,
		Validators:      resp.Validators,
		AppHash:         resp.AppHash,
	}, nil
}

// ListSnapshots implements cmtabci.Application
func (a *ABCIAdapter) ListSnapshots(ctx context.Context, req *cmtabci.RequestListSnapshots) (*cmtabci.ResponseListSnapshots, error) {
	resp, err := a.app.ListSnapshots(ctx, &types.ListSnapshotsRequest{})
	if err != nil {
		return nil, err
	}
	snapshots := make([]*cmtabci.Snapshot, len(resp.Snapshots))
	for i, s := range resp.Snapshots {
		snapshots[i] = &cmtabci.Snapshot{
			Height:   s.Height,
			Format:   s.Format,
			Chunks:   s.Chunks,
			Hash:     s.Hash,
			Metadata: s.Metadata,
		}
	}
	return &cmtabci.ResponseListSnapshots{Snapshots: snapshots}, nil
}

// OfferSnapshot implements cmtabci.Application
func (a *ABCIAdapter) OfferSnapshot(ctx context.Context, req *cmtabci.RequestOfferSnapshot) (*cmtabci.ResponseOfferSnapshot, error) {
	resp, err := a.app.OfferSnapshot(ctx, &types.OfferSnapshotRequest{
		Snapshot: &types.Snapshot{
			Height:   req.Snapshot.Height,
			Format:   req.Snapshot.Format,
			Chunks:   req.Snapshot.Chunks,
			Hash:     req.Snapshot.Hash,
			Metadata: req.Snapshot.Metadata,
		},
		AppHash: req.AppHash,
	})
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponseOfferSnapshot{Result: cmtabci.ResponseOfferSnapshot_Result(resp.Result)}, nil
}

// LoadSnapshotChunk implements cmtabci.Application
func (a *ABCIAdapter) LoadSnapshotChunk(ctx context.Context, req *cmtabci.RequestLoadSnapshotChunk) (*cmtabci.ResponseLoadSnapshotChunk, error) {
	resp, err := a.app.LoadSnapshotChunk(ctx, &types.LoadSnapshotChunkRequest{
		Height: req.Height,
		Format: req.Format,
		Chunk:  req.Chunk,
	})
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponseLoadSnapshotChunk{Chunk: resp.Chunk}, nil
}

// ApplySnapshotChunk implements cmtabci.Application
func (a *ABCIAdapter) ApplySnapshotChunk(ctx context.Context, req *cmtabci.RequestApplySnapshotChunk) (*cmtabci.ResponseApplySnapshotChunk, error) {
	resp, err := a.app.ApplySnapshotChunk(ctx, &types.ApplySnapshotChunkRequest{
		Index:  req.Index,
		Chunk:  req.Chunk,
		Sender: req.Sender,
	})
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponseApplySnapshotChunk{
		Result:        cmtabci.ResponseApplySnapshotChunk_Result(resp.Result),
		RefetchChunks: resp.RefetchChunks,
		RejectSenders: resp.RejectSenders,
	}, nil
}

// FinalizeBlock implements cmtabci.Application
func (a *ABCIAdapter) FinalizeBlock(ctx context.Context, req *cmtabci.RequestFinalizeBlock) (*cmtabci.ResponseFinalizeBlock, error) {
	resp, err := a.app.FinalizeBlock(ctx, &types.FinalizeBlockRequest{
		Hash:               req.Hash,
		Height:             req.Height,
		Time:               req.Time,
		NextValidatorsHash: req.NextValidatorsHash,
		ProposerAddress:    req.ProposerAddress,
		Txs:                req.Txs,
		DecidedLastCommit:  req.DecidedLastCommit,
		Misbehavior:        req.Misbehavior,
	})
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponseFinalizeBlock{
		Events:                resp.Events,
		TxResults:             resp.TxResults,
		ValidatorUpdates:      resp.ValidatorUpdates,
		ConsensusParamUpdates: resp.ConsensusParamUpdates,
		AppHash:               resp.AppHash,
	}, nil
}

// ExtendVote implements cmtabci.Application
func (a *ABCIAdapter) ExtendVote(ctx context.Context, req *cmtabci.RequestExtendVote) (*cmtabci.ResponseExtendVote, error) {
	// TODO: Implement or forward to underlying app if needed
	return &cmtabci.ResponseExtendVote{}, nil
}

// PrepareProposal implements cmtabci.Application
func (a *ABCIAdapter) PrepareProposal(ctx context.Context, req *cmtabci.RequestPrepareProposal) (*cmtabci.ResponsePrepareProposal, error) {
	// TODO: Implement or forward to underlying app if needed
	return &cmtabci.ResponsePrepareProposal{}, nil
}

// ProcessProposal implements cmtabci.Application
func (a *ABCIAdapter) ProcessProposal(ctx context.Context, req *cmtabci.RequestProcessProposal) (*cmtabci.ResponseProcessProposal, error) {
	// TODO: Implement or forward to underlying app if needed
	return &cmtabci.ResponseProcessProposal{}, nil
}

// VerifyVoteExtension implements cmtabci.Application
func (a *ABCIAdapter) VerifyVoteExtension(ctx context.Context, req *cmtabci.RequestVerifyVoteExtension) (*cmtabci.ResponseVerifyVoteExtension, error) {
	// TODO: Implement or forward to underlying app if needed
	return &cmtabci.ResponseVerifyVoteExtension{}, nil
}

// Helper functions for type conversion
func convertEvents(events []types.Event) []cmtabci.Event {
	if events == nil {
		return nil
	}
	result := make([]cmtabci.Event, len(events))
	for i, e := range events {
		result[i] = cmtabci.Event{
			Type:       e.Type,
			Attributes: convertEventAttributes(e.Attributes),
		}
	}
	return result
}

func convertEventAttributes(attrs []types.EventAttribute) []cmtabci.EventAttribute {
	if attrs == nil {
		return nil
	}
	result := make([]cmtabci.EventAttribute, len(attrs))
	for i, attr := range attrs {
		result[i] = cmtabci.EventAttribute{
			Key:   string(attr.Key),
			Value: string(attr.Value),
			Index: attr.Index,
		}
	}
	return result
}

func convertProofOps(ops *types.ProofOps) *cmtcrypto.ProofOps {
	if ops == nil {
		return nil
	}
	// Convert from local ProofOps to CometBFT ProofOps
	cometOps := &cmtcrypto.ProofOps{
		Ops: make([]cmtcrypto.ProofOp, len(ops.Ops)),
	}
	for i, op := range ops.Ops {
		cometOps.Ops[i] = cmtcrypto.ProofOp{
			Type: op.Type_,
			Key:  op.Key,
			Data: op.Data,
		}
	}
	return cometOps
}

func convertProofOpsFromCometBFT(ops *cmtcrypto.ProofOps) *cmtcrypto.ProofOps {
	// Since ops is already the correct type, just return it
	return ops
}

func convertConsensusParams(params *cmttypes.ConsensusParams) *types.ConsensusParams {
	if params == nil {
		return nil
	}
	// For now, return nil to avoid compilation errors
	// TODO: Implement proper conversion when needed
	return nil
}

func convertBlockParams(params *cmttypes.BlockParams) *types.BlockParams {
	if params == nil {
		return nil
	}
	// For now, return nil to avoid compilation errors
	// TODO: Implement proper conversion when needed
	return nil
}

func convertEvidenceParams(params *cmttypes.EvidenceParams) *types.EvidenceParams {
	if params == nil {
		return nil
	}
	// For now, return nil to avoid compilation errors
	// TODO: Implement proper conversion when needed
	return nil
}

func convertValidatorParams(params *cmttypes.ValidatorParams) *types.ValidatorParams {
	if params == nil {
		return nil
	}
	// For now, return nil to avoid compilation errors
	// TODO: Implement proper conversion when needed
	return nil
}

func convertVersionParams(params *cmttypes.VersionParams) *types.VersionParams {
	if params == nil {
		return nil
	}
	// For now, return nil to avoid compilation errors
	// TODO: Implement proper conversion when needed
	return nil
}

func convertConsensusParamsBack(params *types.ConsensusParams) *cmttypes.ConsensusParams {
	if params == nil {
		return nil
	}
	// For now, return nil to avoid compilation errors
	// TODO: Implement proper conversion when needed
	return nil
}

func convertBlockParamsBack(params *types.BlockParams) *cmttypes.BlockParams {
	if params == nil {
		return nil
	}
	// For now, return nil to avoid compilation errors
	// TODO: Implement proper conversion when needed
	return nil
}

func convertEvidenceParamsBack(params *types.EvidenceParams) *cmttypes.EvidenceParams {
	if params == nil {
		return nil
	}
	// For now, return nil to avoid compilation errors
	// TODO: Implement proper conversion when needed
	return nil
}

func convertValidatorParamsBack(params *types.ValidatorParams) *cmttypes.ValidatorParams {
	if params == nil {
		return nil
	}
	// For now, return nil to avoid compilation errors
	// TODO: Implement proper conversion when needed
	return nil
}

func convertVersionParamsBack(params *types.VersionParams) *cmttypes.VersionParams {
	if params == nil {
		return nil
	}
	// For now, return nil to avoid compilation errors
	// TODO: Implement proper conversion when needed
	return nil
}

func convertValidators(validators []cmtabci.ValidatorUpdate) []*types.ValidatorUpdate {
	if validators == nil {
		return nil
	}
	result := make([]*types.ValidatorUpdate, len(validators))
	for i, v := range validators {
		result[i] = &types.ValidatorUpdate{
			PubKey: v.PubKey,
			Power:  v.Power,
		}
	}
	return result
}

func convertValidatorsBack(validators []*types.ValidatorUpdate) []cmtabci.ValidatorUpdate {
	if validators == nil {
		return nil
	}
	result := make([]cmtabci.ValidatorUpdate, len(validators))
	for i, v := range validators {
		result[i] = cmtabci.ValidatorUpdate{
			PubKey: v.PubKey,
			Power:  v.Power,
		}
	}
	return result
}

func convertHeader(header cmtcoretypes.Header) *types.Header {
	// For now, return nil to avoid compilation errors
	// TODO: Implement proper conversion when needed
	return nil
}

func convertVersion(version *statepb.Version) *statepb.Version {
	return nil
}

func convertBlockID(id *types.BlockID) *types.BlockID {
	// For now, return nil to avoid compilation errors
	// TODO: Implement proper conversion when needed
	return nil
}

func convertPartSetHeader(header *types.PartSetHeader) *types.PartSetHeader {
	// For now, return nil to avoid compilation errors
	// TODO: Implement proper conversion when needed
	return nil
}

func convertLastCommitInfo(info *abcipb.LastCommitInfo) *abcipb.LastCommitInfo {
	return nil
}
