package types

import (
	"context"
	cometbftabciv1 "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	cometbftcrypto "github.com/cometbft/cometbft/crypto"
	cometbfttypes "github.com/cometbft/cometbft/types"
)

// ABCI Types - Direct aliases to CometBFT v0.38.17 types for full compatibility

// Request types
type Request = cometbftabciv1.Request
type CheckTxRequest = cometbftabciv1.CheckTxRequest
type FinalizeBlockRequest = cometbftabciv1.FinalizeBlockRequest
type CommitRequest = cometbftabciv1.CommitRequest
type InfoRequest = cometbftabciv1.InfoRequest
type QueryRequest = cometbftabciv1.QueryRequest
type InitChainRequest = cometbftabciv1.InitChainRequest
type PrepareProposalRequest = cometbftabciv1.PrepareProposalRequest
type ProcessProposalRequest = cometbftabciv1.ProcessProposalRequest
type ExtendVoteRequest = cometbftabciv1.ExtendVoteRequest
type VerifyVoteExtensionRequest = cometbftabciv1.VerifyVoteExtensionRequest
type ListSnapshotsRequest = cometbftabciv1.ListSnapshotsRequest
type OfferSnapshotRequest = cometbftabciv1.OfferSnapshotRequest
type LoadSnapshotChunkRequest = cometbftabciv1.LoadSnapshotChunkRequest
type ApplySnapshotChunkRequest = cometbftabciv1.ApplySnapshotChunkRequest

// Response types
type Response = cometbftabciv1.Response
type CheckTxResponse = cometbftabciv1.CheckTxResponse
type FinalizeBlockResponse = cometbftabciv1.FinalizeBlockResponse
type CommitResponse = cometbftabciv1.CommitResponse
type InfoResponse = cometbftabciv1.InfoResponse
type QueryResponse = cometbftabciv1.QueryResponse
type InitChainResponse = cometbftabciv1.InitChainResponse
type PrepareProposalResponse = cometbftabciv1.PrepareProposalResponse
type ProcessProposalResponse = cometbftabciv1.ProcessProposalResponse
type ExtendVoteResponse = cometbftabciv1.ExtendVoteResponse
type VerifyVoteExtensionResponse = cometbftabciv1.VerifyVoteExtensionResponse
type ListSnapshotsResponse = cometbftabciv1.ListSnapshotsResponse
type OfferSnapshotResponse = cometbftabciv1.OfferSnapshotResponse
type LoadSnapshotChunkResponse = cometbftabciv1.LoadSnapshotChunkResponse
type ApplySnapshotChunkResponse = cometbftabciv1.ApplySnapshotChunkResponse

// Consensus and block types
type ConsensusParams = cometbftabciv1.ConsensusParams
type BlockParams = cometbftabciv1.BlockParams
type EvidenceParams = cometbftabciv1.EvidenceParams
type ValidatorParams = cometbftabciv1.ValidatorParams
type VersionParams = cometbftabciv1.VersionParams
type Header = cometbfttypes.Header
type BlockID = cometbfttypes.BlockID
type PartSetHeader = cometbfttypes.PartSetHeader

// Validator types
type Validator = cometbftabciv1.Validator
type ValidatorUpdate = cometbftabciv1.ValidatorUpdate
type PubKey = cometbftcrypto.PubKey

// Transaction and execution types
type ExecTxResult = cometbftabciv1.ExecTxResult
type Event = cometbftabciv1.Event
type EventAttribute = cometbftabciv1.EventAttribute
type TxResult = cometbftabciv1.TxResult

// Snapshot types
type Snapshot = cometbftabciv1.Snapshot

// Vote and commit types
type VoteInfo = cometbftabciv1.VoteInfo
type ExtendedVoteInfo = cometbftabciv1.ExtendedVoteInfo
type CommitInfo = cometbftabciv1.CommitInfo
type ExtendedCommitInfo = cometbftabciv1.ExtendedCommitInfo
type Misbehavior = cometbftabciv1.Misbehavior

// Proof types
type ProofOp = cometbftabciv1.ProofOp
type ProofOps = cometbftabciv1.ProofOps

// CheckTx types
type CheckTxType = cometbftabciv1.CheckTxType

// Response status types
type ResponseProcessProposal_Status = cometbftabciv1.ResponseProcessProposal_Status
type ResponseVerifyVoteExtension_Status = cometbftabciv1.ResponseVerifyVoteExtension_Status
type ResponseOfferSnapshot_Result = cometbftabciv1.ResponseOfferSnapshot_Result
type ResponseApplySnapshotChunk_Result = cometbftabciv1.ResponseApplySnapshotChunk_Result

// Constants for response statuses
const (
	ResponseProcessProposal_ACCEPT  = cometbftabciv1.ResponseProcessProposal_ACCEPT
	ResponseProcessProposal_REJECT  = cometbftabciv1.ResponseProcessProposal_REJECT
	ResponseVerifyVoteExtension_ACCEPT = cometbftabciv1.ResponseVerifyVoteExtension_ACCEPT
	ResponseVerifyVoteExtension_REJECT = cometbftabciv1.ResponseVerifyVoteExtension_REJECT
	ResponseOfferSnapshot_ACCEPT   = cometbftabciv1.ResponseOfferSnapshot_ACCEPT
	ResponseOfferSnapshot_REJECT   = cometbftabciv1.ResponseOfferSnapshot_REJECT
	ResponseApplySnapshotChunk_ACCEPT = cometbftabciv1.ResponseApplySnapshotChunk_ACCEPT
	ResponseApplySnapshotChunk_REJECT = cometbftabciv1.ResponseApplySnapshotChunk_REJECT
)

// CheckTx type constants
const (
	CheckTxType_New     = cometbftabciv1.CheckTxType_New
	CheckTxType_Recheck = cometbftabciv1.CheckTxType_Recheck
)

// Error codes (using CometBFT's constants)
const (
	CodeTypeOK                = uint32(0)
	CodeTypeInternalError     = uint32(1)
	CodeTypeEncodingError     = uint32(2)
	CodeTypeUnauthorized      = uint32(3)
	CodeTypeInsufficientFunds = uint32(4)
	CodeTypeUnknownRequest    = uint32(5)
	CodeTypeInvalidAddress    = uint32(6)
)

// Utility functions for common operations

// IsOK checks if a response code indicates success
func IsOK(code uint32) bool {
	return code == CodeTypeOK
}

// IsError checks if a response code indicates an error
func IsError(code uint32) bool {
	return code != CodeTypeOK
}

// NewEvent creates a new event with the given type
func NewEvent(eventType string) Event {
	return Event{
		Type: eventType,
	}
}

// AddAttribute adds an attribute to an event
func AddAttribute(event *Event, key, value string, index bool) {
	event.Attributes = append(event.Attributes, EventAttribute{
		Key:   key,
		Value: value,
		Index: index,
	})
}

// NewExecTxResult creates a new ExecTxResult
func NewExecTxResult(code uint32, data []byte, log string) *ExecTxResult {
	return &ExecTxResult{
		Code: code,
		Data: data,
		Log:  log,
	}
}

// NewResponseCheckTx creates a new ResponseCheckTx
func NewResponseCheckTx(code uint32, data []byte, log string, gasWanted, gasUsed int64) *CheckTxResponse {
	return &CheckTxResponse{
		Code:      code,
		Data:      data,
		Log:       log,
		GasWanted: gasWanted,
		GasUsed:   gasUsed,
	}
}

// NewResponseQuery creates a new ResponseQuery
func NewResponseQuery(code uint32, value []byte, log string) *QueryResponse {
	return &QueryResponse{
		Code:  code,
		Value: value,
		Log:   log,
	}
}

// Application interface defines the ABCI application methods
type Application interface {
	// Echo method for testing
	Echo(context.Context, string) (string, error)
	
	// Info/Query Connection
	Info(context.Context, *InfoRequest) (*InfoResponse, error)
	Query(context.Context, *QueryRequest) (*QueryResponse, error)

	// Mempool Connection
	CheckTx(context.Context, *CheckTxRequest) (*CheckTxResponse, error)

	// Consensus Connection
	PrepareProposal(context.Context, *PrepareProposalRequest) (*PrepareProposalResponse, error)
	ProcessProposal(context.Context, *ProcessProposalRequest) (*ProcessProposalResponse, error)
	FinalizeBlock(context.Context, *FinalizeBlockRequest) (*FinalizeBlockResponse, error)
	ExtendVote(context.Context, *ExtendVoteRequest) (*ExtendVoteResponse, error)
	VerifyVoteExtension(context.Context, *VerifyVoteExtensionRequest) (*VerifyVoteExtensionResponse, error)
	Commit(context.Context) (*CommitResponse, error)
	InitChain(context.Context, *InitChainRequest) (*InitChainResponse, error)

	// State Sync Connection
	ListSnapshots(context.Context, *ListSnapshotsRequest) (*ListSnapshotsResponse, error)
	OfferSnapshot(context.Context, *OfferSnapshotRequest) (*OfferSnapshotResponse, error)
	LoadSnapshotChunk(context.Context, *LoadSnapshotChunkRequest) (*LoadSnapshotChunkResponse, error)
	ApplySnapshotChunk(context.Context, *ApplySnapshotChunkRequest) (*ApplySnapshotChunkResponse, error)
} 