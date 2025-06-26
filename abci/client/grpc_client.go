package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	cmtabci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var _ Client = (*grpcClient)(nil)

type grpcClient struct {
	client cmtproto.ABCIServiceClient
	conn   *grpc.ClientConn
	mtx    sync.Mutex
	logger Logger
}

// NewGRPCClient creates a new gRPC client
func NewGRPCClient(addr string, logger Logger) (Client, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	return &grpcClient{
		client: cmtproto.NewABCIServiceClient(conn),
		conn:   conn,
		logger: logger,
	}, nil
}

func (c *grpcClient) SetResponseCallback(cb Callback) {
	// gRPC client doesn't use callbacks in the same way
}

func (c *grpcClient) SetLogger(l Logger) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.logger = l
}

func (c *grpcClient) Error() error {
	if c.conn == nil {
		return ErrConnectionNotInitialized
	}
	return nil
}

// Mempool methods
func (c *grpcClient) CheckTx(ctx context.Context, req *cmtabci.RequestCheckTx) (*cmtabci.ResponseCheckTx, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	protoReq := &cmtproto.CheckTxRequest{
		Tx:   req.Tx,
		Type: cmtproto.CheckTxType(req.Type),
	}

	resp, err := c.client.CheckTx(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("CheckTx failed", "error", err)
		}
		return nil, fmt.Errorf("gRPC CheckTx failed: %w", err)
	}

	return &cmtabci.ResponseCheckTx{
		Code:      resp.Code,
		Data:      resp.Data,
		Log:       resp.Log,
		Info:      resp.Info,
		GasWanted: resp.GasWanted,
		GasUsed:   resp.GasUsed,
		Events:    fromProtoEvents(resp.Events),
		Codespace: resp.Codespace,
		Sender:    resp.Sender,
		Priority:  resp.Priority,
		MempoolError: resp.MempoolError,
	}, nil
}

func (c *grpcClient) CheckTxAsync(ctx context.Context, req *cmtabci.RequestCheckTx) *ReqRes {
	reqRes := NewReqRes(req)
	go func() {
		res, err := c.CheckTx(ctx, req)
		if err != nil {
			reqRes.Response = nil
			reqRes.Error = err
		} else {
			reqRes.Response = res
		}
		reqRes.Done()
	}()
	return reqRes
}

func (c *grpcClient) Flush(ctx context.Context) error {
	// gRPC doesn't need explicit flushing
	return nil
}

// Consensus methods
func (c *grpcClient) FinalizeBlock(ctx context.Context, req *cmtabci.RequestFinalizeBlock) (*cmtabci.ResponseFinalizeBlock, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	protoReq := &cmtproto.FinalizeBlockRequest{
		Txs:                req.Txs,
		DecidedLastCommit:  toProtoCommitInfo(req.DecidedLastCommit),
		Misbehavior:        toProtoMisbehavior(req.Misbehavior),
		Hash:               req.Hash,
		Height:             req.Height,
		Time:               req.Time,
		NextValidatorsHash: req.NextValidatorsHash,
		ProposerAddress:    req.ProposerAddress,
	}

	resp, err := c.client.FinalizeBlock(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("FinalizeBlock failed", "error", err)
		}
		return nil, fmt.Errorf("gRPC FinalizeBlock failed: %w", err)
	}

	return &cmtabci.ResponseFinalizeBlock{
		Events:                fromProtoEvents(resp.Events),
		TxResults:             fromProtoExecTxResults(resp.TxResults),
		ValidatorUpdates:      fromProtoValidatorUpdates(resp.ValidatorUpdates),
		ConsensusParamUpdates: fromProtoConsensusParams(resp.ConsensusParamUpdates),
		AppHash:               resp.AppHash,
		RetainHeight:          resp.RetainHeight,
	}, nil
}

func (c *grpcClient) FinalizeBlockAsync(ctx context.Context, req *cmtabci.RequestFinalizeBlock) *ReqRes {
	reqRes := NewReqRes(req)
	go func() {
		res, err := c.FinalizeBlock(ctx, req)
		if err != nil {
			reqRes.Response = nil
			reqRes.Error = err
		} else {
			reqRes.Response = res
		}
		reqRes.Done()
	}()
	return reqRes
}

func (c *grpcClient) PrepareProposal(ctx context.Context, req *cmtabci.RequestPrepareProposal) (*cmtabci.ResponsePrepareProposal, error) {
	ctx, cancel := contextWithTimeout(ctx, 0)
	defer cancel()

	if req.MaxTxBytes <= 0 {
		return nil, fmt.Errorf("invalid max tx bytes: %d", req.MaxTxBytes)
	}

	protoReq := &cmtproto.PrepareProposalRequest{
		MaxTxBytes:         req.MaxTxBytes,
		Txs:                req.Txs,
		LocalLastCommit:    toProtoExtendedCommitInfo(req.LocalLastCommit),
		Misbehavior:        toProtoMisbehavior(req.Misbehavior),
		Height:             req.Height,
		Time:               req.Time,
		NextValidatorsHash: req.NextValidatorsHash,
		ProposerAddress:    req.ProposerAddress,
	}

	resp, err := c.client.PrepareProposal(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("PrepareProposal failed", "err", err)
		}
		return nil, fmt.Errorf("gRPC PrepareProposal failed: %w", err)
	}

	return &cmtabci.ResponsePrepareProposal{
		TxRecords: resp.TxRecords,
		AppHash:   resp.AppHash,
	}, nil
}

func (c *grpcClient) ProcessProposal(ctx context.Context, req *cmtabci.RequestProcessProposal) (*cmtabci.ResponseProcessProposal, error) {
	ctx, cancel := contextWithTimeout(ctx, 0)
	defer cancel()

	if err := validateBlockHeight(req.Height); err != nil {
		return nil, fmt.Errorf("ProcessProposal validation failed: %w", err)
	}

	protoReq := &cmtproto.ProcessProposalRequest{
		Txs:                req.Txs,
		ProposedLastCommit: toProtoCommitInfo(req.ProposedLastCommit),
		Misbehavior:        toProtoMisbehavior(req.Misbehavior),
		Hash:               req.Hash,
		Height:             req.Height,
		Time:               req.Time,
		NextValidatorsHash: req.NextValidatorsHash,
		ProposerAddress:    req.ProposerAddress,
	}

	resp, err := c.client.ProcessProposal(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("ProcessProposal failed", "err", err)
		}
		return nil, fmt.Errorf("gRPC ProcessProposal failed: %w", err)
	}

	return &cmtabci.ResponseProcessProposal{
		Status: resp.Status,
	}, nil
}

func (c *grpcClient) ExtendVote(ctx context.Context, req *cmtabci.RequestExtendVote) (*cmtabci.ResponseExtendVote, error) {
	ctx, cancel := contextWithTimeout(ctx, 0)
	defer cancel()

	if err := validateBlockHeight(req.Height); err != nil {
		return nil, fmt.Errorf("ExtendVote validation failed: %w", err)
	}

	protoReq := &cmtproto.ExtendVoteRequest{
		Hash:   req.Hash,
		Height: req.Height,
	}

	resp, err := c.client.ExtendVote(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("ExtendVote failed", "err", err)
		}
		return nil, fmt.Errorf("gRPC ExtendVote failed: %w", err)
	}

	return &cmtabci.ResponseExtendVote{
		VoteExtension: resp.VoteExtension,
	}, nil
}

func (c *grpcClient) VerifyVoteExtension(ctx context.Context, req *cmtabci.RequestVerifyVoteExtension) (*cmtabci.ResponseVerifyVoteExtension, error) {
	ctx, cancel := contextWithTimeout(ctx, 0)
	defer cancel()

	if err := validateBlockHeight(req.Height); err != nil {
		return nil, fmt.Errorf("VerifyVoteExtension validation failed: %w", err)
	}

	protoReq := &cmtproto.VerifyVoteExtensionRequest{
		Hash:             req.Hash,
		ValidatorAddress: req.ValidatorAddress,
		Height:           req.Height,
		VoteExtension:    req.VoteExtension,
	}

	resp, err := c.client.VerifyVoteExtension(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("VerifyVoteExtension failed", "err", err)
		}
		return nil, fmt.Errorf("gRPC VerifyVoteExtension failed: %w", err)
	}

	return &cmtabci.ResponseVerifyVoteExtension{
		Status: resp.Status,
	}, nil
}

func (c *grpcClient) Commit(ctx context.Context, req *cmtabci.RequestCommit) (*cmtabci.ResponseCommit, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	protoReq := &cmtproto.CommitRequest{}

	resp, err := c.client.Commit(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("Commit failed", "error", err)
		}
		return nil, fmt.Errorf("gRPC Commit failed: %w", err)
	}

	return &cmtabci.ResponseCommit{
		Data:         resp.Data,
		RetainHeight: resp.RetainHeight,
	}, nil
}

func (c *grpcClient) CommitAsync(ctx context.Context, req *cmtabci.RequestCommit) *ReqRes {
	reqRes := NewReqRes(req)
	go func() {
		res, err := c.Commit(ctx, req)
		if err != nil {
			reqRes.Response = nil
			reqRes.Error = err
		} else {
			reqRes.Response = res
		}
		reqRes.Done()
	}()
	return reqRes
}

func (c *grpcClient) InitChain(ctx context.Context, req *cmtabci.RequestInitChain) (*cmtabci.ResponseInitChain, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	protoReq := &cmtproto.InitChainRequest{
		Time:            req.Time,
		ChainId:         req.ChainId,
		ConsensusParams: toProtoConsensusParams(req.ConsensusParams),
		Validators:      toProtoValidatorUpdates(req.Validators),
		AppStateBytes:   req.AppStateBytes,
		InitialHeight:   req.InitialHeight,
	}

	resp, err := c.client.InitChain(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("InitChain failed", "error", err)
		}
		return nil, fmt.Errorf("gRPC InitChain failed: %w", err)
	}

	return &cmtabci.ResponseInitChain{
		ConsensusParams: fromProtoConsensusParams(resp.ConsensusParams),
		Validators:      fromProtoValidatorUpdates(resp.Validators),
		AppHash:         resp.AppHash,
	}, nil
}

// Query methods
func (c *grpcClient) Info(ctx context.Context, req *cmtabci.RequestInfo) (*cmtabci.ResponseInfo, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	protoReq := &cmtproto.InfoRequest{
		Version:      req.Version,
		BlockVersion: req.BlockVersion,
		P2PVersion:   req.P2PVersion,
		AbciVersion:  req.AbciVersion,
	}

	resp, err := c.client.Info(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("Info failed", "error", err)
		}
		return nil, fmt.Errorf("gRPC Info failed: %w", err)
	}

	return &cmtabci.ResponseInfo{
		Data:             resp.Data,
		Version:          resp.Version,
		AppVersion:       resp.AppVersion,
		LastBlockHeight:  resp.LastBlockHeight,
		LastBlockAppHash: resp.LastBlockAppHash,
	}, nil
}

func (c *grpcClient) Query(ctx context.Context, req *cmtabci.RequestQuery) (*cmtabci.ResponseQuery, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	protoReq := &cmtproto.QueryRequest{
		Data:   req.Data,
		Path:   req.Path,
		Height: req.Height,
		Prove:  req.Prove,
	}

	resp, err := c.client.Query(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("Query failed", "error", err)
		}
		return nil, fmt.Errorf("gRPC Query failed: %w", err)
	}

	return &cmtabci.ResponseQuery{
		Code:      resp.Code,
		Log:       resp.Log,
		Info:      resp.Info,
		Index:     resp.Index,
		Key:       resp.Key,
		Value:     resp.Value,
		ProofOps:  fromProtoProofOps(resp.ProofOps),
		Height:    resp.Height,
		Codespace: resp.Codespace,
	}, nil
}

// Snapshot methods
func (c *grpcClient) ListSnapshots(ctx context.Context, req *cmtabci.RequestListSnapshots) (*cmtabci.ResponseListSnapshots, error) {
	ctx, cancel := contextWithTimeout(ctx, 0)
	defer cancel()

	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	protoReq := &cmtproto.ListSnapshotsRequest{}

	resp, err := c.client.ListSnapshots(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("ListSnapshots failed", "err", err)
		}
		return nil, fmt.Errorf("gRPC ListSnapshots failed: %w", err)
	}

	return &cmtabci.ResponseListSnapshots{
		Snapshots: fromProtoSnapshots(resp.Snapshots),
	}, nil
}

func (c *grpcClient) OfferSnapshot(ctx context.Context, req *cmtabci.RequestOfferSnapshot) (*cmtabci.ResponseOfferSnapshot, error) {
	ctx, cancel := contextWithTimeout(ctx, 0)
	defer cancel()

	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	protoReq := &cmtproto.OfferSnapshotRequest{
		Snapshot: toProtoSnapshot(req.Snapshot),
		AppHash:  req.AppHash,
	}

	resp, err := c.client.OfferSnapshot(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("OfferSnapshot failed", "err", err)
		}
		return nil, fmt.Errorf("gRPC OfferSnapshot failed: %w", err)
	}

	return &cmtabci.ResponseOfferSnapshot{
		Result: resp.Result,
	}, nil
}

func (c *grpcClient) LoadSnapshotChunk(ctx context.Context, req *cmtabci.RequestLoadSnapshotChunk) (*cmtabci.ResponseLoadSnapshotChunk, error) {
	ctx, cancel := contextWithTimeout(ctx, 0)
	defer cancel()

	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	protoReq := &cmtproto.LoadSnapshotChunkRequest{
		Height: req.Height,
		Format: req.Format,
		Chunk:  req.Chunk,
	}

	resp, err := c.client.LoadSnapshotChunk(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("LoadSnapshotChunk failed", "err", err)
		}
		return nil, fmt.Errorf("gRPC LoadSnapshotChunk failed: %w", err)
	}

	return &cmtabci.ResponseLoadSnapshotChunk{
		Chunk: resp.Chunk,
	}, nil
}

func (c *grpcClient) ApplySnapshotChunk(ctx context.Context, req *cmtabci.RequestApplySnapshotChunk) (*cmtabci.ResponseApplySnapshotChunk, error) {
	ctx, cancel := contextWithTimeout(ctx, 0)
	defer cancel()

	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	protoReq := &cmtproto.ApplySnapshotChunkRequest{
		Index:  req.Index,
		Chunk:  req.Chunk,
		Sender: req.Sender,
	}

	resp, err := c.client.ApplySnapshotChunk(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("ApplySnapshotChunk failed", "err", err)
		}
		return nil, fmt.Errorf("gRPC ApplySnapshotChunk failed: %w", err)
	}

	return &cmtabci.ResponseApplySnapshotChunk{
		Result:        resp.Result,
		RefetchChunks: resp.RefetchChunks,
		RejectSenders: resp.RejectSenders,
	}, nil
}

func (c *grpcClient) Close() error {
	return c.conn.Close()
}

// Helper conversion functions
func convertEvents(pbEvents []*cmtproto.Event) []cmtabci.Event {
	if pbEvents == nil {
		return nil
	}
	events := make([]cmtabci.Event, len(pbEvents))
	for i, pbEvent := range pbEvents {
		events[i] = cmtabci.Event{
			Type:       pbEvent.Type,
			Attributes: convertAttributes(pbEvent.Attributes),
		}
	}
	return events
}

func convertAttributes(pbAttrs []*cmtproto.EventAttribute) []cmtabci.EventAttribute {
	if pbAttrs == nil {
		return nil
	}
	attrs := make([]cmtabci.EventAttribute, len(pbAttrs))
	for i, pbAttr := range pbAttrs {
		attrs[i] = cmtabci.EventAttribute{
			Key:   pbAttr.Key,
			Value: pbAttr.Value,
			Index: pbAttr.Index,
		}
	}
	return attrs
}

func convertTxResults(pbResults []*cmtproto.ExecTxResult) []*cmtabci.ExecTxResult {
	if pbResults == nil {
		return nil
	}
	results := make([]*cmtabci.ExecTxResult, len(pbResults))
	for i, pbResult := range pbResults {
		results[i] = &cmtabci.ExecTxResult{
			Code:      pbResult.Code,
			Data:      pbResult.Data,
			Log:       pbResult.Log,
			Info:      pbResult.Info,
			GasWanted: pbResult.GasWanted,
			GasUsed:   pbResult.GasUsed,
			Events:    convertEvents(pbResult.Events),
			Codespace: pbResult.Codespace,
		}
	}
	return results
}

func convertValidatorUpdates(pbUpdates []*cmtproto.ValidatorUpdate) []cmtabci.ValidatorUpdate {
	if pbUpdates == nil {
		return nil
	}
	updates := make([]cmtabci.ValidatorUpdate, len(pbUpdates))
	for i, pbUpdate := range pbUpdates {
		updates[i] = cmtabci.ValidatorUpdate{
			PubKey: convertPubKey(pbUpdate.PubKey),
			Power:  pbUpdate.Power,
		}
	}
	return updates
}

func convertConsensusParams(pbParams *cmtproto.ConsensusParams) *cmtabci.ConsensusParams {
	if pbParams == nil {
		return nil
	}
	return &cmtabci.ConsensusParams{
		Block:     convertBlockParams(pbParams.Block),
		Evidence:  convertEvidenceParams(pbParams.Evidence),
		Validator: convertValidatorParams(pbParams.Validator),
		Version:   convertVersionParams(pbParams.Version),
	}
}

func convertHeader(header *cmtabci.Header) *cmtproto.Header {
	if header == nil {
		return nil
	}
	return &cmtproto.Header{
		Version:            header.Version,
		ChainId:            header.ChainId,
		Height:             header.Height,
		Time:               header.Time,
		LastBlockId:        convertBlockID(header.LastBlockId),
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

func convertBlockID(blockID cmtabci.BlockID) *cmtproto.BlockID {
	return &cmtproto.BlockID{
		Hash:          blockID.Hash,
		PartSetHeader: convertPartSetHeader(blockID.PartSetHeader),
	}
}

func convertPartSetHeader(psh cmtabci.PartSetHeader) *cmtproto.PartSetHeader {
	return &cmtproto.PartSetHeader{
		Total: psh.Total,
		Hash:  psh.Hash,
	}
}

func toProtoCommitInfo(commit *cmtabci.CommitInfo) *cmtproto.CommitInfo {
	if commit == nil {
		return nil
	}
	return &cmtproto.CommitInfo{
		Round: commit.Round,
		Votes: toProtoVoteInfos(commit.Votes),
	}
}

func fromProtoCommitInfo(commit *cmtproto.CommitInfo) *cmtabci.CommitInfo {
	if commit == nil {
		return nil
	}
	return &cmtabci.CommitInfo{
		Round: commit.Round,
		Votes: fromProtoVoteInfos(commit.Votes),
	}
}

func toProtoExtendedCommitInfo(commit *cmtabci.ExtendedCommitInfo) *cmtproto.ExtendedCommitInfo {
	if commit == nil {
		return nil
	}
	return &cmtproto.ExtendedCommitInfo{
		Round:             commit.Round,
		Votes:             toProtoExtendedVoteInfos(commit.Votes),
		QuorumHash:        commit.QuorumHash,
		ExtensionSignature: commit.ExtensionSignature,
	}
}

func fromProtoExtendedCommitInfo(commit *cmtproto.ExtendedCommitInfo) *cmtabci.ExtendedCommitInfo {
	if commit == nil {
		return nil
	}
	return &cmtabci.ExtendedCommitInfo{
		Round:             commit.Round,
		Votes:             fromProtoExtendedVoteInfos(commit.Votes),
		QuorumHash:        commit.QuorumHash,
		ExtensionSignature: commit.ExtensionSignature,
	}
}

func toProtoVoteInfos(votes []cmtabci.VoteInfo) []*cmtproto.VoteInfo {
	if votes == nil {
		return nil
	}
	protoVotes := make([]*cmtproto.VoteInfo, len(votes))
	for i, vote := range votes {
		protoVotes[i] = &cmtproto.VoteInfo{
			Validator:       toProtoValidator(vote.Validator),
			SignedLastBlock: vote.SignedLastBlock,
		}
	}
	return protoVotes
}

func fromProtoVoteInfos(votes []*cmtproto.VoteInfo) []cmtabci.VoteInfo {
	if votes == nil {
		return nil
	}
	abciVotes := make([]cmtabci.VoteInfo, len(votes))
	for i, vote := range votes {
		abciVotes[i] = cmtabci.VoteInfo{
			Validator:       fromProtoValidator(vote.Validator),
			SignedLastBlock: vote.SignedLastBlock,
		}
	}
	return abciVotes
}

func toProtoExtendedVoteInfos(votes []cmtabci.ExtendedVoteInfo) []*cmtproto.ExtendedVoteInfo {
	if votes == nil {
		return nil
	}
	protoVotes := make([]*cmtproto.ExtendedVoteInfo, len(votes))
	for i, vote := range votes {
		protoVotes[i] = &cmtproto.ExtendedVoteInfo{
			Validator:       toProtoValidator(vote.Validator),
			SignedLastBlock: vote.SignedLastBlock,
			VoteExtension:   vote.VoteExtension,
		}
	}
	return protoVotes
}

func fromProtoExtendedVoteInfos(votes []*cmtproto.ExtendedVoteInfo) []cmtabci.ExtendedVoteInfo {
	if votes == nil {
		return nil
	}
	abciVotes := make([]cmtabci.ExtendedVoteInfo, len(votes))
	for i, vote := range votes {
		abciVotes[i] = cmtabci.ExtendedVoteInfo{
			Validator:       fromProtoValidator(vote.Validator),
			SignedLastBlock: vote.SignedLastBlock,
			VoteExtension:   vote.VoteExtension,
		}
	}
	return abciVotes
}

func toProtoValidator(val cmtabci.Validator) *cmtproto.Validator {
	return &cmtproto.Validator{
		Address: val.Address,
		Power:   val.Power,
	}
}

func fromProtoValidator(val *cmtproto.Validator) cmtabci.Validator {
	if val == nil {
		return cmtabci.Validator{}
	}
	return cmtabci.Validator{
		Address: val.Address,
		Power:   val.Power,
	}
}

func toProtoValidatorUpdates(validators []cmtabci.ValidatorUpdate) []*cmtproto.ValidatorUpdate {
	if validators == nil {
		return nil
	}
	protoValidators := make([]*cmtproto.ValidatorUpdate, len(validators))
	for i, val := range validators {
		protoValidators[i] = &cmtproto.ValidatorUpdate{
			PubKey: toProtoPubKey(val.PubKey),
			Power:  val.Power,
		}
	}
	return protoValidators
}

func fromProtoValidatorUpdates(validators []*cmtproto.ValidatorUpdate) []cmtabci.ValidatorUpdate {
	if validators == nil {
		return nil
	}
	abciValidators := make([]cmtabci.ValidatorUpdate, len(validators))
	for i, val := range validators {
		abciValidators[i] = cmtabci.ValidatorUpdate{
			PubKey: fromProtoPubKey(val.PubKey),
			Power:  val.Power,
		}
	}
	return abciValidators
}

func toProtoPubKey(pubKey cmtabci.PubKey) *cmtproto.PubKey {
	return &cmtproto.PubKey{
		Sum: &cmtproto.PubKey_Ed25519{
			Ed25519: pubKey.Data,
		},
	}
}

func fromProtoPubKey(pubKey *cmtproto.PubKey) cmtabci.PubKey {
	if pubKey == nil {
		return cmtabci.PubKey{}
	}
	if ed25519 := pubKey.GetEd25519(); ed25519 != nil {
		return cmtabci.PubKey{
			Type: "ed25519",
			Data: ed25519,
		}
	}
	return cmtabci.PubKey{}
}

func toProtoConsensusParams(params *cmtabci.ConsensusParams) *cmtproto.ConsensusParams {
	if params == nil {
		return nil
	}
	return &cmtproto.ConsensusParams{
		Block:     toProtoBlockParams(params.Block),
		Evidence:  toProtoEvidenceParams(params.Evidence),
		Validator: toProtoValidatorParams(params.Validator),
		Version:   toProtoVersionParams(params.Version),
	}
}

func fromProtoConsensusParams(params *cmtproto.ConsensusParams) *cmtabci.ConsensusParams {
	if params == nil {
		return nil
	}
	return &cmtabci.ConsensusParams{
		Block:     fromProtoBlockParams(params.Block),
		Evidence:  fromProtoEvidenceParams(params.Evidence),
		Validator: fromProtoValidatorParams(params.Validator),
		Version:   fromProtoVersionParams(params.Version),
	}
}

func toProtoBlockParams(params *cmtabci.BlockParams) *cmtproto.BlockParams {
	if params == nil {
		return nil
	}
	return &cmtproto.BlockParams{
		MaxBytes: params.MaxBytes,
		MaxGas:   params.MaxGas,
	}
}

func fromProtoBlockParams(params *cmtproto.BlockParams) *cmtabci.BlockParams {
	if params == nil {
		return nil
	}
	return &cmtabci.BlockParams{
		MaxBytes: params.MaxBytes,
		MaxGas:   params.MaxGas,
	}
}

func toProtoEvidenceParams(params *cmtabci.EvidenceParams) *cmtproto.EvidenceParams {
	if params == nil {
		return nil
	}
	return &cmtproto.EvidenceParams{
		MaxAgeNumBlocks: params.MaxAgeNumBlocks,
		MaxAgeDuration:  params.MaxAgeDuration,
		MaxBytes:        params.MaxBytes,
	}
}

func fromProtoEvidenceParams(params *cmtproto.EvidenceParams) *cmtabci.EvidenceParams {
	if params == nil {
		return nil
	}
	return &cmtabci.EvidenceParams{
		MaxAgeNumBlocks: params.MaxAgeNumBlocks,
		MaxAgeDuration:  params.MaxAgeDuration,
		MaxBytes:        params.MaxBytes,
	}
}

func toProtoValidatorParams(params *cmtabci.ValidatorParams) *cmtproto.ValidatorParams {
	if params == nil {
		return nil
	}
	return &cmtproto.ValidatorParams{
		PubKeyTypes: params.PubKeyTypes,
	}
}

func fromProtoValidatorParams(params *cmtproto.ValidatorParams) *cmtabci.ValidatorParams {
	if params == nil {
		return nil
	}
	return &cmtabci.ValidatorParams{
		PubKeyTypes: params.PubKeyTypes,
	}
}

func toProtoVersionParams(params *cmtabci.VersionParams) *cmtproto.VersionParams {
	if params == nil {
		return nil
	}
	return &cmtproto.VersionParams{
		App: params.App,
	}
}

func fromProtoVersionParams(params *cmtproto.VersionParams) *cmtabci.VersionParams {
	if params == nil {
		return nil
	}
	return &cmtproto.VersionParams{
		App: params.App,
	}
}

func toProtoMisbehavior(misbehavior []cmtabci.Misbehavior) []*cmtproto.Misbehavior {
	if misbehavior == nil {
		return nil
	}
	protoMisbehavior := make([]*cmtproto.Misbehavior, len(misbehavior))
	for i, mis := range misbehavior {
		protoMisbehavior[i] = &cmtproto.Misbehavior{
			Type:             cmtproto.MisbehaviorType(mis.Type),
			Validator:        toProtoValidator(mis.Validator),
			Height:           mis.Height,
			Time:             mis.Time,
			TotalVotingPower: mis.TotalVotingPower,
		}
	}
	return protoMisbehavior
}

func fromProtoMisbehavior(misbehavior []*cmtproto.Misbehavior) []cmtabci.Misbehavior {
	if misbehavior == nil {
		return nil
	}
	abciMisbehavior := make([]cmtabci.Misbehavior, len(misbehavior))
	for i, mis := range misbehavior {
		abciMisbehavior[i] = cmtabci.Misbehavior{
			Type:             cmtabci.MisbehaviorType(mis.Type),
			Validator:        fromProtoValidator(mis.Validator),
			Height:           mis.Height,
			Time:             mis.Time,
			TotalVotingPower: mis.TotalVotingPower,
		}
	}
	return abciMisbehavior
}

func toProtoEvents(events []cmtabci.Event) []*cmtproto.Event {
	if events == nil {
		return nil
	}
	protoEvents := make([]*cmtproto.Event, len(events))
	for i, event := range events {
		protoEvents[i] = &cmtproto.Event{
			Type:       event.Type,
			Attributes: toProtoEventAttributes(event.Attributes),
		}
	}
	return protoEvents
}

func fromProtoEvents(events []*cmtproto.Event) []cmtabci.Event {
	if events == nil {
		return nil
	}
	abciEvents := make([]cmtabci.Event, len(events))
	for i, event := range events {
		abciEvents[i] = cmtabci.Event{
			Type:       event.Type,
			Attributes: fromProtoEventAttributes(event.Attributes),
		}
	}
	return abciEvents
}

func toProtoEventAttributes(attrs []cmtabci.EventAttribute) []*cmtproto.EventAttribute {
	if attrs == nil {
		return nil
	}
	protoAttrs := make([]*cmtproto.EventAttribute, len(attrs))
	for i, attr := range attrs {
		protoAttrs[i] = &cmtproto.EventAttribute{
			Key:   attr.Key,
			Value: attr.Value,
			Index: attr.Index,
		}
	}
	return protoAttrs
}

func fromProtoEventAttributes(attrs []*cmtproto.EventAttribute) []cmtabci.EventAttribute {
	if attrs == nil {
		return nil
	}
	abciAttrs := make([]cmtabci.EventAttribute, len(attrs))
	for i, attr := range attrs {
		abciAttrs[i] = cmtabci.EventAttribute{
			Key:   attr.Key,
			Value: attr.Value,
			Index: attr.Index,
		}
	}
	return abciAttrs
}

func toProtoExecTxResults(results []*cmtabci.ExecTxResult) []*cmtproto.ExecTxResult {
	if results == nil {
		return nil
	}
	protoResults := make([]*cmtproto.ExecTxResult, len(results))
	for i, result := range results {
		protoResults[i] = &cmtproto.ExecTxResult{
			Code:      result.Code,
			Data:      result.Data,
			Log:       result.Log,
			Info:      result.Info,
			GasWanted: result.GasWanted,
			GasUsed:   result.GasUsed,
			Events:    toProtoEvents(result.Events),
			Codespace: result.Codespace,
		}
	}
	return protoResults
}

func fromProtoExecTxResults(results []*cmtproto.ExecTxResult) []*cmtabci.ExecTxResult {
	if results == nil {
		return nil
	}
	abciResults := make([]*cmtabci.ExecTxResult, len(results))
	for i, result := range results {
		abciResults[i] = &cmtabci.ExecTxResult{
			Code:      result.Code,
			Data:      result.Data,
			Log:       result.Log,
			Info:      result.Info,
			GasWanted: result.GasWanted,
			GasUsed:   result.GasUsed,
			Events:    fromProtoEvents(result.Events),
			Codespace: result.Codespace,
		}
	}
	return abciResults
}

func toProtoProofOps(proofOps *cmtabci.ProofOps) *cmtproto.ProofOps {
	if proofOps == nil {
		return nil
	}
	return &cmtproto.ProofOps{
		Ops: toProtoProofOpList(proofOps.Ops),
	}
}

func fromProtoProofOps(proofOps *cmtproto.ProofOps) *cmtabci.ProofOps {
	if proofOps == nil {
		return nil
	}
	return &cmtabci.ProofOps{
		Ops: fromProtoProofOpList(proofOps.Ops),
	}
}

func toProtoProofOpList(ops []cmtabci.ProofOp) []*cmtproto.ProofOp {
	if ops == nil {
		return nil
	}
	protoOps := make([]*cmtproto.ProofOp, len(ops))
	for i, op := range ops {
		protoOps[i] = &cmtproto.ProofOp{
			Type: op.Type,
			Key:  op.Key,
			Data: op.Data,
		}
	}
	return protoOps
}

func fromProtoProofOpList(ops []*cmtproto.ProofOp) []cmtabci.ProofOp {
	if ops == nil {
		return nil
	}
	abciOps := make([]cmtabci.ProofOp, len(ops))
	for i, op := range ops {
		abciOps[i] = cmtabci.ProofOp{
			Type: op.Type,
			Key:  op.Key,
			Data: op.Data,
		}
	}
	return abciOps
}

func toProtoSnapshot(snapshot *cmtabci.Snapshot) *cmtproto.Snapshot {
	if snapshot == nil {
		return nil
	}
	return &cmtproto.Snapshot{
		Height:   snapshot.Height,
		Format:   snapshot.Format,
		Chunks:   snapshot.Chunks,
		Hash:     snapshot.Hash,
		Metadata: snapshot.Metadata,
	}
}

func fromProtoSnapshots(snapshots []*cmtproto.Snapshot) []*cmtabci.Snapshot {
	if snapshots == nil {
		return nil
	}
	abciSnapshots := make([]*cmtabci.Snapshot, len(snapshots))
	for i, snapshot := range snapshots {
		abciSnapshots[i] = &cmtabci.Snapshot{
			Height:   snapshot.Height,
			Format:   snapshot.Format,
			Chunks:   snapshot.Chunks,
			Hash:     snapshot.Hash,
			Metadata: snapshot.Metadata,
		}
	}
	return abciSnapshots
} 