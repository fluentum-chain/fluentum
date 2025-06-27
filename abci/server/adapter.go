package server

import (
	"context"

	cmtabci "github.com/cometbft/cometbft/abci/types"
	"github.com/fluentum-chain/fluentum/abci/types"
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
		ProofOps:  convertProofOps(resp.ProofOps),
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
		DecidedLastCommit:  convertCommitInfo(&req.DecidedLastCommit),
		Misbehavior:        convertMisbehavior(req.Misbehavior),
	})
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponseFinalizeBlock{
		Events:                convertEvents(resp.Events),
		TxResults:             convertExecTxResults(resp.TxResults),
		ValidatorUpdates:      resp.ValidatorUpdates,
		ConsensusParamUpdates: resp.ConsensusParamUpdates,
		AppHash:               resp.AppHash,
	}, nil
}

// Helper functions for type conversion
func convertEvents(events []*types.Event) []cmtabci.Event {
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

func convertEventAttributes(attrs []*types.EventAttribute) []cmtabci.EventAttribute {
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

func convertProofOps(ops *types.ProofOps) *cmtabci.ProofOps {
	if ops == nil {
		return nil
	}
	// For now, return nil to avoid compilation errors
	// TODO: Implement proper conversion when needed
	return nil
}

func convertConsensusParams(params *cmtabci.ConsensusParams) *types.ConsensusParams {
	if params == nil {
		return nil
	}
	// For now, return nil to avoid compilation errors
	// TODO: Implement proper conversion when needed
	return nil
}

func convertBlockParams(params *cmtabci.BlockParams) *types.BlockParams {
	if params == nil {
		return nil
	}
	// For now, return nil to avoid compilation errors
	// TODO: Implement proper conversion when needed
	return nil
}

func convertEvidenceParams(params *cmtabci.EvidenceParams) *types.EvidenceParams {
	if params == nil {
		return nil
	}
	// For now, return nil to avoid compilation errors
	// TODO: Implement proper conversion when needed
	return nil
}

func convertValidatorParams(params *cmtabci.ValidatorParams) *types.ValidatorParams {
	if params == nil {
		return nil
	}
	// For now, return nil to avoid compilation errors
	// TODO: Implement proper conversion when needed
	return nil
}

func convertVersionParams(params *cmtabci.VersionParams) *types.VersionParams {
	if params == nil {
		return nil
	}
	// For now, return nil to avoid compilation errors
	// TODO: Implement proper conversion when needed
	return nil
}

func convertConsensusParamsBack(params *types.ConsensusParams) *cmtabci.ConsensusParams {
	if params == nil {
		return nil
	}
	// For now, return nil to avoid compilation errors
	// TODO: Implement proper conversion when needed
	return nil
}

func convertBlockParamsBack(params *types.BlockParams) *cmtabci.BlockParams {
	if params == nil {
		return nil
	}
	// For now, return nil to avoid compilation errors
	// TODO: Implement proper conversion when needed
	return nil
}

func convertEvidenceParamsBack(params *types.EvidenceParams) *cmtabci.EvidenceParams {
	if params == nil {
		return nil
	}
	// For now, return nil to avoid compilation errors
	// TODO: Implement proper conversion when needed
	return nil
}

func convertValidatorParamsBack(params *types.ValidatorParams) *cmtabci.ValidatorParams {
	if params == nil {
		return nil
	}
	// For now, return nil to avoid compilation errors
	// TODO: Implement proper conversion when needed
	return nil
}

func convertVersionParamsBack(params *types.VersionParams) *cmtabci.VersionParams {
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

func convertHeader(header cmtabci.Header) *types.Header {
	return &types.Header{
		Version:            convertVersion(header.Version),
		ChainID:            header.ChainID,
		Height:             header.Height,
		Time:               header.Time,
		LastBlockID:        convertBlockID(header.LastBlockID),
		LastCommitHash:     header.LastCommitHash,
		DataHash:           header.DataHash,
		ValidatorsHash:     header.ValidatorsHash,
		NextValidatorsHash: header.NextValidatorsHash,
		ConsensusHash:      header.ConsensusHash,
		AppHash:            header.AppHash,
		LastResultsHash:    header.LastResultsHash,
		EvidenceHash:       header.EvidenceHash,
		ProposerAddress:    header.ProposerAddress,
	}
}

func convertVersion(version cmtabci.Version) *types.Version {
	return &types.Version{
		Block: version.Block,
		App:   version.App,
	}
}

func convertBlockID(id cmtabci.BlockID) *types.BlockID {
	return &types.BlockID{
		Hash:          id.Hash,
		PartSetHeader: convertPartSetHeader(id.PartSetHeader),
	}
}

func convertPartSetHeader(header cmtabci.PartSetHeader) *types.PartSetHeader {
	return &types.PartSetHeader{
		Total: header.Total,
		Hash:  header.Hash,
	}
}

func convertLastCommitInfo(info cmtabci.LastCommitInfo) *types.LastCommitInfo {
	return &types.LastCommitInfo{
		Round: info.Round,
		Votes: convertVoteInfos(info.Votes),
	}
}

func convertVoteInfos(votes []cmtabci.VoteInfo) []*types.VoteInfo {
	if votes == nil {
		return nil
	}
	result := make([]*types.VoteInfo, len(votes))
	for i, v := range votes {
		result[i] = &types.VoteInfo{
			Validator:       convertValidator(v.Validator),
			SignedLastBlock: v.SignedLastBlock,
		}
	}
	return result
}

func convertEvidence(evidence []cmtabci.Evidence) []*types.Evidence {
	if evidence == nil {
		return nil
	}
	result := make([]*types.Evidence, len(evidence))
	for i, e := range evidence {
		result[i] = &types.Evidence{
			Type:             e.Type,
			Validator:        convertValidator(e.Validator),
			Height:           e.Height,
			Time:             e.Time,
			TotalVotingPower: e.TotalVotingPower,
		}
	}
	return result
}

func convertValidator(v cmtabci.Validator) *types.Validator {
	return &types.Validator{
		Address: v.Address,
		Power:   v.Power,
	}
}

func convertCommitInfo(info *cmtabci.CommitInfo) *types.CommitInfo {
	if info == nil {
		return nil
	}
	return &types.CommitInfo{
		Round: info.Round,
		Votes: convertVoteInfos(info.Votes),
	}
}

func convertMisbehavior(misbehavior []cmtabci.Misbehavior) []*types.Misbehavior {
	if misbehavior == nil {
		return nil
	}
	result := make([]*types.Misbehavior, len(misbehavior))
	for i, m := range misbehavior {
		result[i] = &types.Misbehavior{
			Type:             m.Type,
			Validator:        convertValidator(m.Validator),
			Height:           m.Height,
			Time:             m.Time,
			TotalVotingPower: m.TotalVotingPower,
		}
	}
	return result
}

func convertExecTxResults(results []*types.ExecTxResult) []*cmtabci.ExecTxResult {
	if results == nil {
		return nil
	}
	cometResults := make([]*cmtabci.ExecTxResult, len(results))
	for i, r := range results {
		cometResults[i] = &cmtabci.ExecTxResult{
			Code:      r.Code,
			Data:      r.Data,
			Log:       r.Log,
			Info:      r.Info,
			GasWanted: r.GasWanted,
			GasUsed:   r.GasUsed,
			Events:    convertEvents(r.Events),
			Codespace: r.Codespace,
		}
	}
	return cometResults
}
