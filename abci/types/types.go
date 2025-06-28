package types

import (
	"context"

	cometbftabci "github.com/cometbft/cometbft/abci/types"
	cometbftcrypto "github.com/cometbft/cometbft/crypto"
	cometbfttypes "github.com/cometbft/cometbft/types"
	tmproto "github.com/fluentum-chain/fluentum/proto/tendermint/types"
)

// ABCI Types - Direct aliases to CometBFT v0.38.17 types for full compatibility

// Request types
type Request = cometbftabci.Request
type EchoRequest = cometbftabci.RequestEcho
type CheckTxRequest = cometbftabci.RequestCheckTx
type FinalizeBlockRequest = cometbftabci.RequestFinalizeBlock
type CommitRequest = cometbftabci.RequestCommit
type InfoRequest = cometbftabci.RequestInfo
type QueryRequest = cometbftabci.RequestQuery
type InitChainRequest = cometbftabci.RequestInitChain
type PrepareProposalRequest = cometbftabci.RequestPrepareProposal
type ProcessProposalRequest = cometbftabci.RequestProcessProposal
type ExtendVoteRequest = cometbftabci.RequestExtendVote
type VerifyVoteExtensionRequest = cometbftabci.RequestVerifyVoteExtension
type ListSnapshotsRequest = cometbftabci.RequestListSnapshots
type OfferSnapshotRequest = cometbftabci.RequestOfferSnapshot
type LoadSnapshotChunkRequest = cometbftabci.RequestLoadSnapshotChunk
type ApplySnapshotChunkRequest = cometbftabci.RequestApplySnapshotChunk

// Request nested types
type Request_CheckTx = cometbftabci.Request_CheckTx

// Response types
type Response = cometbftabci.Response
type EchoResponse = cometbftabci.ResponseEcho
type CheckTxResponse = cometbftabci.ResponseCheckTx
type FinalizeBlockResponse = cometbftabci.ResponseFinalizeBlock
type CommitResponse = cometbftabci.ResponseCommit
type QueryResponse = cometbftabci.ResponseQuery
type InitChainResponse = cometbftabci.ResponseInitChain
type PrepareProposalResponse = cometbftabci.ResponsePrepareProposal
type ProcessProposalResponse = cometbftabci.ResponseProcessProposal
type ExtendVoteResponse = cometbftabci.ResponseExtendVote
type VerifyVoteExtensionResponse = cometbftabci.ResponseVerifyVoteExtension
type ListSnapshotsResponse = cometbftabci.ResponseListSnapshots
type LoadSnapshotChunkResponse = cometbftabci.ResponseLoadSnapshotChunk
type OfferSnapshotResponse = cometbftabci.ResponseOfferSnapshot
type ApplySnapshotChunkResponse = cometbftabci.ResponseApplySnapshotChunk

// Legacy response types (for backward compatibility)
// Note: ResponseDeliverTx, ResponseBeginBlock, and ResponseEndBlock were removed in newer CometBFT versions

// Response nested types
type Response_CheckTx = cometbftabci.Response_CheckTx

// Missing response types
type InfoResponse = cometbftabci.ResponseInfo
type ExecTxResult = cometbftabci.ExecTxResult

// Consensus and block types
type ConsensusParams = tmproto.ConsensusParams
type BlockParams = tmproto.BlockParams
type EvidenceParams = tmproto.EvidenceParams
type ValidatorParams = tmproto.ValidatorParams
type VersionParams = tmproto.VersionParams
type Header = cometbfttypes.Header
type BlockID = cometbfttypes.BlockID
type PartSetHeader = cometbfttypes.PartSetHeader

// Validator types
type Validator = cometbftabci.Validator
type ValidatorUpdate = cometbftabci.ValidatorUpdate
type PubKey = cometbftcrypto.PubKey

// Transaction and execution types
type Event = cometbftabci.Event
type EventAttribute = cometbftabci.EventAttribute
type TxResult = cometbftabci.TxResult

// Snapshot types
type Snapshot = cometbftabci.Snapshot

// Vote and commit types
type VoteInfo = cometbftabci.VoteInfo
type ExtendedVoteInfo = cometbftabci.ExtendedVoteInfo
type CommitInfo = cometbftabci.CommitInfo
type ExtendedCommitInfo = cometbftabci.ExtendedCommitInfo
type Misbehavior = cometbftabci.Misbehavior

// CheckTx types
type CheckTxType = cometbftabci.CheckTxType

// Response status types
// type ResponseProcessProposal_Status = cometbftabci.ResponseProcessProposal_Status
// type ResponseVerifyVoteExtension_Status = cometbftabci.ResponseVerifyVoteExtension_Status
type ResponseOfferSnapshot_Result = cometbftabci.ResponseOfferSnapshot_Result
type ResponseApplySnapshotChunk_Result = cometbftabci.ResponseApplySnapshotChunk_Result

// Constants for response statuses
const (
	ResponseProcessProposal_ACCEPT     = cometbftabci.ResponseProcessProposal_ACCEPT
	ResponseProcessProposal_REJECT     = cometbftabci.ResponseProcessProposal_REJECT
	ResponseVerifyVoteExtension_ACCEPT = cometbftabci.ResponseVerifyVoteExtension_ACCEPT
	ResponseVerifyVoteExtension_REJECT = cometbftabci.ResponseVerifyVoteExtension_REJECT
	ResponseOfferSnapshot_ACCEPT       = cometbftabci.ResponseOfferSnapshot_ACCEPT
	ResponseOfferSnapshot_REJECT       = cometbftabci.ResponseOfferSnapshot_REJECT
	ResponseOfferSnapshot_ABORT        = cometbftabci.ResponseOfferSnapshot_ABORT
	ResponseApplySnapshotChunk_ACCEPT  = cometbftabci.ResponseApplySnapshotChunk_ACCEPT
	ResponseApplySnapshotChunk_ABORT   = cometbftabci.ResponseApplySnapshotChunk_ABORT
)

// CheckTx type constants
const (
	CheckTxType_New     = cometbftabci.CheckTxType_New
	CheckTxType_Recheck = cometbftabci.CheckTxType_Recheck
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
	CodeTypeUnknownAddress    = uint32(7)
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
	Echo(context.Context, *EchoRequest) (*EchoResponse, error)

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
	Commit(context.Context, *CommitRequest) (*CommitResponse, error)
	InitChain(context.Context, *InitChainRequest) (*InitChainResponse, error)

	// State Sync Connection
	ListSnapshots(context.Context, *ListSnapshotsRequest) (*ListSnapshotsResponse, error)
	OfferSnapshot(context.Context, *OfferSnapshotRequest) (*OfferSnapshotResponse, error)
	LoadSnapshotChunk(context.Context, *LoadSnapshotChunkRequest) (*LoadSnapshotChunkResponse, error)
	ApplySnapshotChunk(context.Context, *ApplySnapshotChunkRequest) (*ApplySnapshotChunkResponse, error)
}

// SignedMsgType and vote type constants
// These map to the proto enum values for SignedMsgType
// and allow use of tmproto.PrevoteType, etc.
type SignedMsgType = tmproto.SignedMsgType

const (
	PrevoteType   SignedMsgType = 1  // SignedMsgType_SIGNED_MSG_TYPE_PREVOTE
	PrecommitType SignedMsgType = 2  // SignedMsgType_SIGNED_MSG_TYPE_PRECOMMIT
	ProposalType  SignedMsgType = 32 // SignedMsgType_SIGNED_MSG_TYPE_PROPOSAL
)
