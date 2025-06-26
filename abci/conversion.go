package abci

import (
	cmtproto "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	cmtabci "github.com/cometbft/cometbft/abci/types"
)

// ToCometRequest converts ABCI Request to CometBFT proto Request
func ToCometRequest(req *cmtabci.Request) *cmtproto.Request {
	if req == nil {
		return nil
	}

	protoReq := &cmtproto.Request{
		RequestId: req.RequestID,
	}

	switch req := req.Value.(type) {
	case *cmtabci.Request_CheckTx:
		protoReq.Value = &cmtproto.Request_CheckTx{
			CheckTx: &cmtproto.CheckTxRequest{
				Tx:   req.CheckTx.Tx,
				Type: cmtproto.CheckTxType(req.CheckTx.Type),
			},
		}

	case *cmtabci.Request_FinalizeBlock:
		protoReq.Value = &cmtproto.Request_FinalizeBlock{
			FinalizeBlock: &cmtproto.FinalizeBlockRequest{
				Txs:                req.FinalizeBlock.Txs,
				DecidedLastCommit:  toProtoCommitInfo(req.FinalizeBlock.DecidedLastCommit),
				Misbehavior:        toProtoMisbehavior(req.FinalizeBlock.Misbehavior),
				Hash:               req.FinalizeBlock.Hash,
				Height:             req.FinalizeBlock.Height,
				Time:               req.FinalizeBlock.Time,
				NextValidatorsHash: req.FinalizeBlock.NextValidatorsHash,
				ProposerAddress:    req.FinalizeBlock.ProposerAddress,
			},
		}

	case *cmtabci.Request_Commit:
		protoReq.Value = &cmtproto.Request_Commit{
			Commit: &cmtproto.CommitRequest{},
		}

	case *cmtabci.Request_Info:
		protoReq.Value = &cmtproto.Request_Info{
			Info: &cmtproto.InfoRequest{
				Version:         req.Info.Version,
				BlockVersion:    req.Info.BlockVersion,
				P2PVersion:      req.Info.P2PVersion,
				AbciVersion:     req.Info.AbciVersion,
			},
		}

	case *cmtabci.Request_Query:
		protoReq.Value = &cmtproto.Request_Query{
			Query: &cmtproto.QueryRequest{
				Data:   req.Query.Data,
				Path:   req.Query.Path,
				Height: req.Query.Height,
				Prove:  req.Query.Prove,
			},
		}

	case *cmtabci.Request_InitChain:
		protoReq.Value = &cmtproto.Request_InitChain{
			InitChain: &cmtproto.InitChainRequest{
				Time:            req.InitChain.Time,
				ChainId:         req.InitChain.ChainId,
				ConsensusParams: toProtoConsensusParams(req.InitChain.ConsensusParams),
				Validators:      toProtoValidatorUpdates(req.InitChain.Validators),
				AppStateBytes:   req.InitChain.AppStateBytes,
				InitialHeight:   req.InitChain.InitialHeight,
			},
		}

	case *cmtabci.Request_PrepareProposal:
		protoReq.Value = &cmtproto.Request_PrepareProposal{
			PrepareProposal: &cmtproto.PrepareProposalRequest{
				MaxTxBytes: req.PrepareProposal.MaxTxBytes,
				Txs:        req.PrepareProposal.Txs,
				LocalLastCommit: toProtoExtendedCommitInfo(req.PrepareProposal.LocalLastCommit),
				Misbehavior:     toProtoMisbehavior(req.PrepareProposal.Misbehavior),
				Height:          req.PrepareProposal.Height,
				Time:            req.PrepareProposal.Time,
				NextValidatorsHash: req.PrepareProposal.NextValidatorsHash,
				ProposerAddress:    req.PrepareProposal.ProposerAddress,
			},
		}

	case *cmtabci.Request_ProcessProposal:
		protoReq.Value = &cmtproto.Request_ProcessProposal{
			ProcessProposal: &cmtproto.ProcessProposalRequest{
				Txs:                req.ProcessProposal.Txs,
				ProposedLastCommit: toProtoCommitInfo(req.ProcessProposal.ProposedLastCommit),
				Misbehavior:        toProtoMisbehavior(req.ProcessProposal.Misbehavior),
				Hash:               req.ProcessProposal.Hash,
				Height:             req.ProcessProposal.Height,
				Time:               req.ProcessProposal.Time,
				NextValidatorsHash: req.ProcessProposal.NextValidatorsHash,
				ProposerAddress:    req.ProcessProposal.ProposerAddress,
			},
		}

	case *cmtabci.Request_ExtendVote:
		protoReq.Value = &cmtproto.Request_ExtendVote{
			ExtendVote: &cmtproto.ExtendVoteRequest{
				Hash:   req.ExtendVote.Hash,
				Height: req.ExtendVote.Height,
			},
		}

	case *cmtabci.Request_VerifyVoteExtension:
		protoReq.Value = &cmtproto.Request_VerifyVoteExtension{
			VerifyVoteExtension: &cmtproto.VerifyVoteExtensionRequest{
				Hash:            req.VerifyVoteExtension.Hash,
				ValidatorProTxHash: req.VerifyVoteExtension.ValidatorProTxHash,
				Height:          req.VerifyVoteExtension.Height,
				VoteExtension:   req.VerifyVoteExtension.VoteExtension,
			},
		}

	case *cmtabci.Request_ListSnapshots:
		protoReq.Value = &cmtproto.Request_ListSnapshots{
			ListSnapshots: &cmtproto.ListSnapshotsRequest{},
		}

	case *cmtabci.Request_OfferSnapshot:
		protoReq.Value = &cmtproto.Request_OfferSnapshot{
			OfferSnapshot: &cmtproto.OfferSnapshotRequest{
				Snapshot: toProtoSnapshot(req.OfferSnapshot.Snapshot),
				AppHash:  req.OfferSnapshot.AppHash,
			},
		}

	case *cmtabci.Request_LoadSnapshotChunk:
		protoReq.Value = &cmtproto.Request_LoadSnapshotChunk{
			LoadSnapshotChunk: &cmtproto.LoadSnapshotChunkRequest{
				Height: req.LoadSnapshotChunk.Height,
				Format: req.LoadSnapshotChunk.Format,
				Chunk:  req.LoadSnapshotChunk.Chunk,
			},
		}

	case *cmtabci.Request_ApplySnapshotChunk:
		protoReq.Value = &cmtproto.Request_ApplySnapshotChunk{
			ApplySnapshotChunk: &cmtproto.ApplySnapshotChunkRequest{
				Index:  req.ApplySnapshotChunk.Index,
				Chunk:  req.ApplySnapshotChunk.Chunk,
				Sender: req.ApplySnapshotChunk.Sender,
			},
		}
	}

	return protoReq
}

// FromCometResponse converts CometBFT proto Response to ABCI Response
func FromCometResponse(res *cmtproto.Response) *cmtabci.Response {
	if res == nil {
		return nil
	}

	abciRes := &cmtabci.Response{
		RequestID: res.RequestId,
	}

	switch res := res.Value.(type) {
	case *cmtproto.Response_CheckTx:
		abciRes.Value = &cmtabci.Response_CheckTx{
			CheckTx: &cmtabci.ResponseCheckTx{
				Code:      res.CheckTx.Code,
				Data:      res.CheckTx.Data,
				Log:       res.CheckTx.Log,
				Info:      res.CheckTx.Info,
				GasWanted: res.CheckTx.GasWanted,
				GasUsed:   res.CheckTx.GasUsed,
				Events:    fromProtoEvents(res.CheckTx.Events),
				Codespace: res.CheckTx.Codespace,
				Sender:    res.CheckTx.Sender,
				Priority:  res.CheckTx.Priority,
				MempoolError: res.CheckTx.MempoolError,
			},
		}

	case *cmtproto.Response_FinalizeBlock:
		abciRes.Value = &cmtabci.Response_FinalizeBlock{
			FinalizeBlock: &cmtabci.ResponseFinalizeBlock{
				TxResults:             fromProtoExecTxResults(res.FinalizeBlock.TxResults),
				ConsensusParamUpdates: fromProtoConsensusParams(res.FinalizeBlock.ConsensusParamUpdates),
				AppHash:               res.FinalizeBlock.AppHash,
				RetainHeight:          res.FinalizeBlock.RetainHeight,
			},
		}

	case *cmtproto.Response_Commit:
		abciRes.Value = &cmtabci.Response_Commit{
			Commit: &cmtabci.ResponseCommit{
				Data: res.Commit.Data,
				RetainHeight: res.Commit.RetainHeight,
			},
		}

	case *cmtproto.Response_Info:
		abciRes.Value = &cmtabci.Response_Info{
			Info: &cmtabci.ResponseInfo{
				Data:             res.Info.Data,
				Version:          res.Info.Version,
				AppVersion:       res.Info.AppVersion,
				LastBlockHeight:  res.Info.LastBlockHeight,
				LastBlockAppHash: res.Info.LastBlockAppHash,
			},
		}

	case *cmtproto.Response_Query:
		abciRes.Value = &cmtabci.Response_Query{
			Query: &cmtabci.ResponseQuery{
				Code:      res.Query.Code,
				Log:       res.Query.Log,
				Info:      res.Query.Info,
				Index:     res.Query.Index,
				Key:       res.Query.Key,
				Value:     res.Query.Value,
				ProofOps:  fromProtoProofOps(res.Query.ProofOps),
				Height:    res.Query.Height,
				Codespace: res.Query.Codespace,
			},
		}

	case *cmtproto.Response_InitChain:
		abciRes.Value = &cmtabci.Response_InitChain{
			InitChain: &cmtabci.ResponseInitChain{
				ConsensusParams: fromProtoConsensusParams(res.InitChain.ConsensusParams),
				Validators:      fromProtoValidatorUpdates(res.InitChain.Validators),
				AppHash:         res.InitChain.AppHash,
			},
		}

	case *cmtproto.Response_PrepareProposal:
		abciRes.Value = &cmtabci.Response_PrepareProposal{
			PrepareProposal: &cmtabci.ResponsePrepareProposal{
				Txs: res.PrepareProposal.Txs,
			},
		}

	case *cmtproto.Response_ProcessProposal:
		abciRes.Value = &cmtabci.Response_ProcessProposal{
			ProcessProposal: &cmtabci.ResponseProcessProposal{
				Status: cmtabci.ResponseProcessProposal_Status(res.ProcessProposal.Status),
			},
		}

	case *cmtproto.Response_ExtendVote:
		abciRes.Value = &cmtabci.Response_ExtendVote{
			ExtendVote: &cmtabci.ResponseExtendVote{
				VoteExtension: res.ExtendVote.VoteExtension,
			},
		}

	case *cmtproto.Response_VerifyVoteExtension:
		abciRes.Value = &cmtabci.Response_VerifyVoteExtension{
			VerifyVoteExtension: &cmtabci.ResponseVerifyVoteExtension{
				Status: cmtabci.ResponseVerifyVoteExtension_Status(res.VerifyVoteExtension.Status),
			},
		}

	case *cmtproto.Response_ListSnapshots:
		abciRes.Value = &cmtabci.Response_ListSnapshots{
			ListSnapshots: &cmtabci.ResponseListSnapshots{
				Snapshots: fromProtoSnapshots(res.ListSnapshots.Snapshots),
			},
		}

	case *cmtproto.Response_OfferSnapshot:
		abciRes.Value = &cmtabci.Response_OfferSnapshot{
			OfferSnapshot: &cmtabci.ResponseOfferSnapshot{
				Result: cmtabci.ResponseOfferSnapshot_Result(res.OfferSnapshot.Result),
			},
		}

	case *cmtproto.Response_LoadSnapshotChunk:
		abciRes.Value = &cmtabci.Response_LoadSnapshotChunk{
			LoadSnapshotChunk: &cmtabci.ResponseLoadSnapshotChunk{
				Chunk: res.LoadSnapshotChunk.Chunk,
			},
		}

	case *cmtproto.Response_ApplySnapshotChunk:
		abciRes.Value = &cmtabci.Response_ApplySnapshotChunk{
			ApplySnapshotChunk: &cmtabci.ResponseApplySnapshotChunk{
				Result:         cmtabci.ResponseApplySnapshotChunk_Result(res.ApplySnapshotChunk.Result),
				RefetchChunks:  res.ApplySnapshotChunk.RefetchChunks,
				RejectSenders:  res.ApplySnapshotChunk.RejectSenders,
			},
		}
	}

	return abciRes
}

// Helper conversion functions

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
		Round: commit.Round,
		Votes: toProtoExtendedVoteInfos(commit.Votes),
	}
}

func fromProtoExtendedCommitInfo(commit *cmtproto.ExtendedCommitInfo) *cmtabci.ExtendedCommitInfo {
	if commit == nil {
		return nil
	}
	return &cmtabci.ExtendedCommitInfo{
		Round: commit.Round,
		Votes: fromProtoExtendedVoteInfos(commit.Votes),
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
			Validator:          toProtoValidator(vote.Validator),
			SignedLastBlock:    vote.SignedLastBlock,
			VoteExtension:      vote.VoteExtension,
			ExtensionSignature: vote.ExtensionSignature,
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
			Validator:          fromProtoValidator(vote.Validator),
			SignedLastBlock:    vote.SignedLastBlock,
			VoteExtension:      vote.VoteExtension,
			ExtensionSignature: vote.ExtensionSignature,
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
		return cmtabci.PubKey{Data: ed25519}
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
	return &cmtabci.VersionParams{
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
		Ops: toProtoProofOps(proofOps.Ops),
	}
}

func fromProtoProofOps(proofOps *cmtproto.ProofOps) *cmtabci.ProofOps {
	if proofOps == nil {
		return nil
	}
	return &cmtabci.ProofOps{
		Ops: fromProtoProofOps(proofOps.Ops),
	}
}

func toProtoProofOps(ops []cmtabci.ProofOp) []*cmtproto.ProofOp {
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

func fromProtoProofOps(ops []*cmtproto.ProofOp) []cmtabci.ProofOp {
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
``` 